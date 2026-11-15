package levelsim

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"d2lod-bot/internal/danger"
	"d2lod-bot/internal/power"
)

// Shadow replay (survival slice S-B, docs/survival/power.md): feeds the `zone entry`, `safety decision`, `character
// died`, `merc` and `combat potion` lines of a recorded bot log through the power estimate and prints, per stay in a
// zone, the margin the model gives for that zone and what the stay really cost. The pass rule of design section 9:
// stays that ended in a chicken, a death or a merc loss should have a bad-pack margin (low estimate) under the dial's
// enter line in at least 80 percent of cases; stays of 10 minutes or more without a potion should have one above it in
// at least 80 percent. The log has no defence or damage, so those terms stay default and are flagged in the rows.

// Stay is one stay in a zone, from one `zone entry` line to the next.
type Stay struct {
	Area     int
	Clvl     int
	Start    time.Time
	Minutes  float64
	Potions  int
	Chickens int
	Deaths   int
	MercLost bool
	Margin   power.ZoneMargin
	Skipped  string // why no margin was computed ("area not in the profile", ...)
}

// Class of a stay for the pass rule.
func (s Stay) Class() string {
	switch {
	case s.Chickens > 0 || s.Deaths > 0 || s.MercLost:
		return "bad"
	case s.Minutes >= 10 && s.Potions == 0:
		return "good"
	}
	return "other"
}

// ShadowReport is the result of one replay.
type ShadowReport struct {
	Stays                      []Stay
	Bad, BadHit, Good, GoodHit int
	Dial                       string
}

// MinCases is the number of cases a class needs before the pass rule says PASS or FAIL.
const MinCases = 3

// Verdict applies the 80 percent rule.
func (r ShadowReport) Verdict() string {
	if r.Bad < MinCases || r.Good < MinCases {
		return fmt.Sprintf("n/a (needs %d bad and %d good stays, have %d and %d)", MinCases, MinCases, r.Bad, r.Good)
	}
	if r.BadHit*5 >= r.Bad*4 && r.GoodHit*5 >= r.Good*4 {
		return "PASS"
	}
	return "FAIL"
}

// ShadowOpts selects the model inputs of a replay.
type ShadowOpts struct {
	Tab  *danger.Table
	M    *power.Model
	Arch *power.Archetype
	Diff string
	Dial string
}

// ShadowReplay reads a bot log (text key=value or JSON lines) and evaluates every stay.
func ShadowReplay(r io.Reader, o ShadowOpts) (*ShadowReport, error) {
	if o.Tab == nil || o.M == nil || o.Arch == nil {
		return nil, fmt.Errorf("levelsim: shadow replay needs a danger table, a model and an archetype")
	}
	dial := o.Dial
	if dial == "" {
		dial = o.M.P.DefaultDial
	}
	rep := &ShadowReport{Dial: dial}
	var cur *Stay
	var hero power.Hero
	var last time.Time
	closeStay := func(end time.Time) {
		if cur == nil {
			return
		}
		cur.Minutes = end.Sub(cur.Start).Minutes()
		rep.Stays = append(rep.Stays, *cur)
		cur = nil
	}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		kv := textKV(strings.TrimSpace(sc.Text()))
		ts, err := time.Parse(time.RFC3339, kv["time"])
		if err == nil {
			last = ts
		}
		msg := kv["msg"]
		switch {
		case msg == "zone entry":
			closeStay(last)
			area, _ := strconv.Atoi(kv["area"])
			hero = heroOfEntry(kv)
			cur = &Stay{Area: area, Clvl: hero.Level, Start: last}
			z, err := o.Tab.Find(area, o.Diff)
			switch {
			case err != nil:
				cur.Skipped = "area not in the profile"
			case hero.MaxLife <= 0:
				cur.Skipped = "no life in the entry line"
			default:
				if cur.Margin, err = o.M.Zone(hero, o.Arch, z, dial); err != nil {
					cur.Skipped = err.Error()
				}
			}
		case cur == nil:
		case msg == "safety decision" && kv["action"] == "chicken":
			cur.Chickens++
		case msg == "character died":
			cur.Deaths++
		case msg == "combat potion":
			cur.Potions++
		case strings.Contains(strings.ToLower(msg), "merc") && strings.Contains(strings.ToLower(msg), "died"):
			cur.MercLost = true
		}
	}
	closeStay(last)
	for i := range rep.Stays {
		s := &rep.Stays[i]
		if s.Skipped != "" {
			continue
		}
		low := s.Margin.Bad.Low
		switch s.Class() {
		case "bad":
			rep.Bad++
			if low < s.Margin.Enter {
				rep.BadHit++
			}
		case "good":
			rep.Good++
			if low >= s.Margin.Enter {
				rep.GoodHit++
			}
		}
	}
	return rep, sc.Err()
}

// heroOfEntry builds the hero of a `zone entry` line: life, mana, healing potions and merc state are read, everything
// else stays default (flagged).
func heroOfEntry(kv map[string]string) power.Hero {
	h := power.Hero{Class: "log"}
	h.Level, _ = strconv.Atoi(kv["clvl"])
	if _, mx, ok := strings.Cut(kv["life"], "/"); ok {
		h.MaxLife, _ = strconv.Atoi(mx)
		h.SetQuality(power.TermLife, power.Measured)
	}
	if _, mx, ok := strings.Cut(kv["mana"], "/"); ok {
		h.MaxMana, _ = strconv.Atoi(mx)
	}
	// terms of the zone entry line since card survcal (docs/survival/replay-2026-11-14.md); older logs lack them and stay default
	if v, err := strconv.Atoi(kv["def"]); err == nil {
		h.Defence = v
		h.SetQuality(power.TermDefence, power.Measured)
	}
	if v, err := strconv.Atoi(kv["block"]); err == nil {
		h.BlockPct = v
		h.SetQuality(power.TermBlock, power.Derived)
	}
	if lo, hi, ok := strings.Cut(kv["dmg"], "-"); ok {
		h.DmgMin, _ = strconv.Atoi(lo)
		h.DmgMax, _ = strconv.Atoi(hi)
		h.DmgPct, _ = strconv.Atoi(kv["dmgpct"])
		h.AR, _ = strconv.Atoi(kv["ar"])
		if h.DmgMax > 0 {
			h.SetQuality(power.TermOffence, power.Derived)
		}
	}
	if n, err := strconv.Atoi(kv["heal"]); err == nil {
		h.Potions = n
		h.SetQuality(power.TermPotions, power.Derived)
	}
	switch kv["merc"] {
	case "alive":
		h.Merc = power.Merc{Hired: true, Alive: true}
		h.SetQuality(power.TermMerc, power.Derived)
	case "dead":
		h.Merc = power.Merc{Hired: true}
		h.SetQuality(power.TermMerc, power.Derived)
	}
	return h
}

// Text prints the replay: one row per stay and the verdict.
func (r *ShadowReport) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "== shadow replay (dial %s) ==\n", r.Dial)
	fmt.Fprintf(&b, "%-8s %5s %4s %7s %5s %5s %5s %7s %6s %6s %-9s %s\n", "start", "area", "clvl", "minutes", "pots", "chick", "death", "class", "bad", "low", "word", "note")
	for _, s := range r.Stays {
		note := s.Skipped
		if note == "" && s.MercLost {
			note = "merc lost"
		}
		if s.Skipped != "" {
			fmt.Fprintf(&b, "%-8s %5d %4d %7.1f %5d %5d %5d %7s %6s %6s %-9s %s\n", s.Start.Format("15:04:05"), s.Area, s.Clvl, s.Minutes, s.Potions, s.Chickens, s.Deaths, s.Class(), "-", "-", "-", note)
			continue
		}
		fmt.Fprintf(&b, "%-8s %5d %4d %7.1f %5d %5d %5d %7s %6.2f %6.2f %-9s %s\n", s.Start.Format("15:04:05"), s.Area, s.Clvl, s.Minutes, s.Potions, s.Chickens, s.Deaths, s.Class(), s.Margin.Bad.Best, s.Margin.Bad.Low, s.Margin.Word(), note)
	}
	fmt.Fprintf(&b, "pass rule (design 9): bad stays under the enter line %d of %d, good stays above it %d of %d -> %s\n", r.BadHit, r.Bad, r.GoodHit, r.Good, r.Verdict())
	return b.String()
}

// textKV reads `key=value key="quoted value"` log lines.
func textKV(line string) map[string]string {
	kv := map[string]string{}
	for len(line) > 0 {
		line = strings.TrimLeft(line, " ")
		k, rest, ok := strings.Cut(line, "=")
		if !ok {
			break
		}
		var v string
		if strings.HasPrefix(rest, `"`) {
			end := 1
			for end < len(rest) && (rest[end] != '"' || rest[end-1] == '\\') {
				end++
			}
			if end >= len(rest) {
				break
			}
			v, _ = strconv.Unquote(rest[:end+1])
			line = rest[end+1:]
		} else {
			v, line, _ = strings.Cut(rest, " ")
		}
		kv[k] = v
	}
	return kv
}
