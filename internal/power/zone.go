package power

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"d2lod-bot/internal/danger"
)

// ZoneMargin is the margin of a zone's three reference encounters (design 5.1); Bad gates.
type ZoneMargin struct {
	Area      int     `json:"area"`
	Name      string  `json:"name"`
	Diff      string  `json:"diff"`
	Archetype string  `json:"archetype"`
	Dial      string  `json:"dial"`
	Enter     float64 `json:"enter"`
	Stay      float64 `json:"stay"`
	Typical   Result  `json:"typical"`
	Bad       Result  `json:"bad"`
	Elite     Result  `json:"elite"`
}

// meanMon is the share-weighted monster of one reach among the zone's spawns; false when the zone has none.
func meanMon(z *danger.Zone, ranged bool) (Mon, bool) {
	var out Mon
	var w float64
	seen := map[string]bool{}
	for _, mo := range z.Monsters {
		if (mo.Reach == "ranged") != ranged {
			continue
		}
		w += mo.Share
		out.Life += mo.Share * float64(mo.Life)
		out.Def += mo.Share * float64(mo.Def)
		out.ToHit += mo.Share * float64(mo.ToHit)
		out.Dmg += mo.Share * mo.Dmg
		if mo.Unblockable {
			out.Unblockable += mo.Share
		}
		for _, h := range mo.Hazards {
			if !seen[h] {
				seen[h] = true
				out.Hazards = append(out.Hazards, h)
			}
		}
	}
	if w == 0 {
		return Mon{}, false
	}
	out.Life, out.Def, out.ToHit, out.Dmg = out.Life/w, out.Def/w, out.ToHit/w, out.Dmg/w
	out.Unblockable /= w
	out.Ranged = ranged
	sort.Strings(out.Hazards)
	return out, true
}

// PackOf builds the reference pack of an encounter from the zone's spawns. A zone with no monsters of one reach fills
// that reach's members with the other reach's mean monster (as melee, so the contact rule still applies). leaderLife
// and leaderDmg scale the first member (the elite's unique); 1 and 1 leave the pack alone.
func PackOf(z *danger.Zone, e danger.Encounter, leaderLife, leaderDmg float64) (Pack, error) {
	mm, mok := meanMon(z, false)
	rm, rok := meanMon(z, true)
	if !mok && !rok {
		return Pack{}, fmt.Errorf("power: zone %d %q (%s) has no monsters", z.Area, z.Name, z.Diff)
	}
	if !mok {
		mm = rm
		mm.Ranged = false
	}
	if !rok {
		rm = mm
		rm.Ranged = true
	}
	pk := Pack{MLvl: z.MLvl}
	for i := 0; i < e.Melee; i++ {
		pk.Mons = append(pk.Mons, mm)
	}
	for i := 0; i < e.Ranged; i++ {
		pk.Mons = append(pk.Mons, rm)
	}
	if len(pk.Mons) > 0 && (leaderLife != 1 || leaderDmg != 1) {
		l := pk.Mons[0]
		l.Life *= leaderLife
		l.Dmg *= leaderDmg
		pk.Mons[0] = l
	}
	return pk, nil
}

// Zone evaluates the three reference encounters of z for the hero under the dial.
func (m *Model) Zone(h Hero, a *Archetype, z *danger.Zone, dial string) (ZoneMargin, error) {
	dl, err := m.P.Dial(dial)
	if err != nil {
		return ZoneMargin{}, err
	}
	zm := ZoneMargin{Area: z.Area, Name: z.Name, Diff: z.Diff, Archetype: a.ID, Dial: dl.ID, Enter: dl.Enter, Stay: dl.Stay}
	for _, c := range []struct {
		dst  *Result
		e    danger.Encounter
		l, d float64
	}{
		{&zm.Typical, z.Encounters.Typical, 1, 1},
		{&zm.Bad, z.Encounters.Bad, 1, 1},
		{&zm.Elite, z.Encounters.Elite, m.P.Params.EliteLifeMul, m.P.Params.EliteDmgMul},
	} {
		pk, err := PackOf(z, c.e, c.l, c.d)
		if err != nil {
			return zm, err
		}
		if *c.dst, err = m.Margin(h, a, pk); err != nil {
			return zm, err
		}
	}
	return zm, nil
}

// Word is the label the watch UI shows instead of a number (user decision 2026-11-14, decisions-2026-11-14.md item 2).
// The estimate that decides entry is the low one.
func (z ZoneMargin) Word() string {
	switch {
	case z.Bad.Low >= z.Enter:
		return "safe"
	case z.Bad.Best >= z.Stay:
		return "risky"
	}
	return "too hard"
}

// Weakest names the least certain term of the bad pack ("defence_default"), "" when everything is measured.
func (z ZoneMargin) Weakest() string {
	rank := map[string]int{TermDefence: 0, TermOffence: 1, TermPotions: 2, TermMerc: 3, TermResist: 4, TermBlock: 5}
	best, bi := "", 99
	for _, f := range z.Bad.Flags {
		term, _, _ := strings.Cut(f, "_")
		if strings.HasSuffix(f, "_default") && rank[term] < bi {
			best, bi = f, rank[term]
		}
	}
	if best == "" && len(z.Bad.Flags) > 0 {
		return z.Bad.Flags[0]
	}
	return best
}

// ShadowLine is the log line `zone shadow:` written next to every `planner` line. The numbers are for the log and for
// developers; the UI shows Word only.
func (z ZoneMargin) ShadowLine(clvl int) string {
	f := func(v float64) string { return fmt.Sprintf("%.2f", math.Round(v*100)/100) }
	return fmt.Sprintf("zone shadow: area=%d name=%q diff=%s clvl=%d arch=%s dial=%s word=%s bad=%s low=%s typical=%s elite=%s need=%s stay=%s perKill=%.1fs potions=%.1f weakest=%s dps=%.1f src=%s",
		z.Area, z.Name, z.Diff, clvl, z.Archetype, z.Dial, z.Word(), f(z.Bad.Best), f(z.Bad.Low), f(z.Typical.Best), f(z.Elite.Best),
		f(z.Enter), f(z.Stay), z.Bad.Terms.PerKill, z.Bad.Terms.Potions, orNone(z.Weakest()), z.Bad.Terms.HeroDPS, z.Bad.Terms.Offence)
}

func orNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}
