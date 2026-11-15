package power

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
)

// BuildSpec is what the resolver needs to know about a build (design 6.3). Attack is the skill the build plays at the
// level in question; Points are the skill points the build spends in total (for the summoner share).
type BuildSpec struct {
	Class  string
	Build  string
	Attack string
	Points map[string]int
	// AttackLevel is the level of Attack with the points the build has spent by the clvl asked for (items and quest
	// points are not counted); 0 when none were spent. The plain weapon swing is level 1.
	AttackLevel int
	Contact     bool   // the class's combat profile has a melee block: the hero stands in the pack
	Named       string // an explicit archetype in the build file wins
}

// Resolve places a build on an archetype by the rule of design 6.3, in data and with no class names: more than a third
// of the skill points in summon skills -> summoner; a melee skill -> bruiser; a weapon missile -> ranged_physical; a
// mana-gated skill -> caster, or hybrid_tank_caster when the class stands in the pack. An unknown class, an unknown
// skill or a build with no attack is an error naming it (AGENTS.md: never a silent fallback).
func (p *Profiles) Resolve(s BuildSpec) (*Archetype, error) {
	if s.Class == "" {
		return nil, fmt.Errorf("power: build %q has no class", s.Build)
	}
	if s.Named != "" {
		return p.Archetype(s.Named)
	}
	total, summon := 0, 0
	for sk, n := range s.Points {
		total += n
		if p.SkillKinds[sk] == "summon" {
			summon += n
		}
	}
	if total > 0 && summon*3 > total {
		return p.Archetype("summoner")
	}
	if s.Attack == "" {
		return nil, fmt.Errorf("power: build %q (%s) has no attack skill to place", s.Build, s.Class)
	}
	kind, ok := p.SkillKinds[s.Attack]
	if !ok {
		return nil, fmt.Errorf("power: build %q (%s): skill %q has no kind in skillKinds", s.Build, s.Class, s.Attack)
	}
	switch kind {
	case "melee":
		return p.Archetype("bruiser")
	case "missile":
		return p.Archetype("ranged_physical")
	case "summon":
		return p.Archetype("summoner")
	case "spell":
		if s.Contact {
			return p.Archetype("hybrid_tank_caster")
		}
		return p.Archetype("caster")
	}
	return nil, fmt.Errorf("power: build %q: skill %q has kind %q", s.Build, s.Attack, kind)
}

// buildFile is the part of a plandata build file the resolver reads.
type buildFile struct {
	Build      string `json:"build"`
	Class      string `json:"class"`
	Archetype  string `json:"archetype"`
	MainAttack []struct {
		From  int    `json:"from_clvl"`
		Until int    `json:"until_clvl"`
		Skill string `json:"skill"`
	} `json:"main_attack"`
	Rows []struct {
		Clvl   int `json:"clvl"`
		Skills []struct {
			Skill  string `json:"skill"`
			Points int    `json:"points"`
		} `json:"skills"`
	} `json:"rows"`
}

// ClassAttacks gives, for a class (lower case), the attack skills of its combat profile by minimum level, highest
// first (internal/combat/classdata bands); contact says whether the class has a melee block. Supplied by the caller so
// this package stays a leaf.
type ClassAttacks func(class string) (bands []Band, contact bool, err error)

// Band is one combat profile band: from minLevel the hero plays the first attack it owns; requires is the skill that
// opens the band ("" none).
type Band struct {
	MinLevel int
	Requires string
	Attack   []string
}

// SpecFromBuild reads a plandata build file (name without the build- prefix and .json) at clvl. The attack is the
// build's main_attack band for that level; builds without one use the class's combat profile bands, the first band
// whose requirement the build has paid a point for by clvl.
func SpecFromBuild(fsys fs.FS, file string, clvl int, ca ClassAttacks) (BuildSpec, error) {
	raw, err := fs.ReadFile(fsys, file)
	if err != nil {
		return BuildSpec{}, fmt.Errorf("power: build file %q: %w", file, err)
	}
	var b buildFile
	if err := json.Unmarshal(raw, &b); err != nil {
		return BuildSpec{}, fmt.Errorf("power: build file %q: %w", file, err)
	}
	s := BuildSpec{Class: strings.ToLower(b.Class), Build: b.Build, Named: b.Archetype, Points: map[string]int{}}
	for _, r := range b.Rows {
		for _, sk := range r.Skills {
			s.Points[sk.Skill] += max(sk.Points, 1)
		}
	}
	spent := map[string]int{}
	for _, r := range b.Rows {
		if r.Clvl > clvl {
			continue
		}
		for _, sk := range r.Skills {
			spent[sk.Skill] += max(sk.Points, 1)
		}
	}
	if s.Class == "" {
		return s, fmt.Errorf("power: build file %q has no class (a class-less fallback table is not a build)", file)
	}
	for _, m := range b.MainAttack {
		if clvl >= m.From && clvl <= m.Until {
			s.Attack = m.Skill
		}
	}
	bands, contact, err := ca(s.Class)
	if err != nil {
		return s, err
	}
	s.Contact = contact
	if s.Attack == "" {
		for _, bd := range bands {
			if clvl >= bd.MinLevel && (bd.Requires == "" || s.Points[bd.Requires] > 0) && len(bd.Attack) > 0 {
				s.Attack = firstKnown(bd.Attack, s.Points)
				if s.Attack != "" {
					break
				}
			}
		}
		if s.Attack == "" && len(bands) > 0 { // nothing bought yet: the profile's own opening attack
			if last := bands[len(bands)-1]; len(last.Attack) > 0 {
				s.Attack = last.Attack[0]
			}
		}
	}
	s.AttackLevel = spent[s.Attack]
	if s.Attack == PlainAttack {
		s.AttackLevel = 1
	}
	return s, nil
}

// AttackOf is the main attack of a resolved build in the estimator's terms. Synergy points and the gear speed are not
// known from a build file (flagged by the estimate); the caller adds what it reads.
func AttackOf(s BuildSpec) Attack {
	return Attack{Skill: s.Attack, Level: s.AttackLevel, SynergyPoints: -1}
}

// PlainAttack is the weapon swing every class has; its kind is melee.
const PlainAttack = "Attack"

// firstKnown is the first attack skill of a band the build invested in; "Attack" (the plain weapon swing) is melee by
// definition and stays last.
func firstKnown(attack []string, points map[string]int) string {
	for _, a := range attack {
		if points[a] > 0 || a == PlainAttack {
			return a
		}
	}
	return ""
}
