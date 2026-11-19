package buildfile

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// The one-line chat notation:
//
//	didle1:pal:Zeal20,Holy Shield20@24,Fanaticism20|str=60,dex=70/3,vit=rest
//
// "Skill20" is a goal ("Zeal" to 20), "@24" its when.clvl; after "|" come stat goals ("dex=70/3" is to 70, per 3;
// "vit=rest"). The class is the three letter code. Name, notes, author, phases, respec, quest points and when.after
// do not fit and make Encode fail rather than drop them silently.
const linePrefix = "didle1:"

var (
	goalRe = regexp.MustCompile(`^(.+?)\s*(\d+)(?:@(\d+))?$`)
	statRe = regexp.MustCompile(`^(str|dex|vit|ene)=(\d+|rest)(?:/(\d+))?(?:@(\d+))?$`)
)

// Encode writes the one-line form.
func Encode(b *Build) (string, error) {
	if len(b.Phases) > 0 || b.RespecAt != nil || len(b.Quest) > 0 {
		return "", fmt.Errorf("buildfile: phases, respec_at and quest_points have no one-line form")
	}
	code := ""
	for c, n := range classCode {
		if n == b.Class {
			code = c
		}
	}
	if code == "" {
		return "", fmt.Errorf("buildfile: unknown class %q", b.Class)
	}
	var sk []string
	for _, g := range b.Skills {
		if g.Hold || (g.When != nil && g.When.After != nil) {
			return "", fmt.Errorf("buildfile: %q uses hold or when.after", g.Skill)
		}
		if strings.ContainsAny(g.Skill, ",|@:") {
			return "", fmt.Errorf("buildfile: skill name %q has a reserved character", g.Skill)
		}
		s := g.Skill + strconv.Itoa(g.To)
		if c := whenClvl(g); c > 0 {
			s += "@" + strconv.Itoa(c)
		}
		sk = append(sk, s)
	}
	out := linePrefix + code + ":" + strings.Join(sk, ",")
	var st []string
	for _, g := range b.Stats {
		if g.When != nil && g.When.After != nil {
			return "", fmt.Errorf("buildfile: stat goal uses when.after")
		}
		s := g.Stat + "="
		if g.To.Rest {
			s += "rest"
		} else {
			s += strconv.Itoa(g.To.N)
		}
		if g.Per > 0 {
			s += "/" + strconv.Itoa(g.Per)
		}
		if c := whenClvl(Goal{When: g.When}); c > 0 {
			s += "@" + strconv.Itoa(c)
		}
		st = append(st, s)
	}
	if len(st) > 0 {
		out += "|" + strings.Join(st, ",")
	}
	return out, nil
}

// Decode reads the one-line form. The result has Name "shared build" and needs Validate like any file.
func Decode(line string) (*Build, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, linePrefix) {
		return nil, fmt.Errorf("buildfile: line must start with %q", linePrefix)
	}
	rest := line[len(linePrefix):]
	skills, stats, _ := strings.Cut(rest, "|")
	code, goals, ok := strings.Cut(skills, ":")
	class, known := classCode[code]
	if !ok || !known {
		return nil, fmt.Errorf("buildfile: class code %q (want ama sor nec pal bar dru ass)", code)
	}
	b := &Build{Schema: Schema, Name: "shared build", Class: class}
	for _, tok := range strings.Split(goals, ",") {
		m := goalRe.FindStringSubmatch(strings.TrimSpace(tok))
		if m == nil {
			return nil, fmt.Errorf("buildfile: skill token %q (want Name<count> or Name<count>@<clvl>)", tok)
		}
		to, _ := strconv.Atoi(m[2])
		g := Goal{Skill: m[1], To: to}
		if m[3] != "" {
			c, _ := strconv.Atoi(m[3])
			g.When = &When{Clvl: c}
		}
		b.Skills = append(b.Skills, g)
	}
	if stats != "" {
		for _, tok := range strings.Split(stats, ",") {
			m := statRe.FindStringSubmatch(strings.TrimSpace(tok))
			if m == nil {
				return nil, fmt.Errorf("buildfile: stat token %q (want str=60, dex=70/3, vit=rest)", tok)
			}
			g := StatGoal{Stat: m[1]}
			if m[2] == "rest" {
				g.To.Rest = true
			} else {
				g.To.N, _ = strconv.Atoi(m[2])
			}
			if m[3] != "" {
				g.Per, _ = strconv.Atoi(m[3])
			}
			if m[4] != "" {
				c, _ := strconv.Atoi(m[4])
				g.When = &When{Clvl: c}
			}
			b.Stats = append(b.Stats, g)
		}
	}
	return b, nil
}

// ImportKolbot maps a kolbot Config.AutoSkill.Build given as JSON with skill NAMES instead of ids
// ([["Blessed Hammer",20,true],["Concentration",20]]) to goals, and a Config.AutoStat.Build
// ([["strength",60],["dexterity","block"],["vitality","all"]]) to stat goals. satisfy true or missing (kolbot's
// default) becomes hold: later entries wait. "all" becomes "rest". "block" has no equivalent here and is an error
// (it needs the block chance formula, an open question in the README). Kolbot counts above 20 are clamped by kolbot
// itself; here they stay and Validate reports them.
func ImportKolbot(skillsJSON, statsJSON []byte) ([]Goal, []StatGoal, error) {
	var goals []Goal
	if len(skillsJSON) > 0 {
		var raw [][]json.RawMessage
		if err := json.Unmarshal(skillsJSON, &raw); err != nil {
			return nil, nil, fmt.Errorf("kolbot skills: %w", err)
		}
		for i, e := range raw {
			if len(e) < 2 || len(e) > 3 {
				return nil, nil, fmt.Errorf("kolbot skills[%d]: want [name, count] or [name, count, satisfy]", i)
			}
			var name string
			var n int
			if json.Unmarshal(e[0], &name) != nil || json.Unmarshal(e[1], &n) != nil {
				return nil, nil, fmt.Errorf("kolbot skills[%d]: want a skill name (ids are not accepted) and a count", i)
			}
			hold := true
			if len(e) == 3 && json.Unmarshal(e[2], &hold) != nil {
				return nil, nil, fmt.Errorf("kolbot skills[%d]: satisfy must be true or false", i)
			}
			goals = append(goals, Goal{Skill: name, To: n, Hold: hold})
		}
	}
	var stats []StatGoal
	if len(statsJSON) > 0 {
		var raw [][]json.RawMessage
		if err := json.Unmarshal(statsJSON, &raw); err != nil {
			return nil, nil, fmt.Errorf("kolbot stats: %w", err)
		}
		short := map[string]string{"strength": "str", "dexterity": "dex", "vitality": "vit", "energy": "ene"}
		for i, e := range raw {
			var name string
			if len(e) != 2 || json.Unmarshal(e[0], &name) != nil || short[name] == "" {
				return nil, nil, fmt.Errorf("kolbot stats[%d]: want [\"strength\"|\"dexterity\"|\"vitality\"|\"energy\", target]", i)
			}
			g := StatGoal{Stat: short[name]}
			var s string
			if json.Unmarshal(e[1], &s) == nil {
				if s != "all" {
					return nil, nil, fmt.Errorf("kolbot stats[%d]: target %q has no equivalent (only a number or \"all\")", i, s)
				}
				g.To.Rest = true
			} else if err := json.Unmarshal(e[1], &g.To.N); err != nil {
				return nil, nil, fmt.Errorf("kolbot stats[%d]: target must be a number or \"all\"", i)
			}
			stats = append(stats, g)
		}
	}
	return goals, stats, nil
}
