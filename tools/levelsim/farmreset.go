package levelsim

import (
	"fmt"
	"strings"
)

// The farm loop simulation (docs/farmloop.md): one zone farmed for an hour at a fixed character level, with a fresh
// game (Save and Exit, every area respawns) taken once a share of the zone is cleared, then the walk back to the same
// zone. It uses the supply model of docs/levelsim.md section 2 (rate = r0 (floor + (1-floor) f^decay), f = the share
// still alive) and equates the roam's coverage share with the killed share (U: coverage points are not monsters).

// FarmZone is a zone of the level plan's farm slices with the clvl the hero farms it at.
type FarmZone struct {
	Area int
	Clvl int
}

// FarmZones are the Paladin's farm slices of levelplan-normal.json (internal/plandata), at the middle of the band
// (entry + exit) / 2 (I: levelsim does not read the plan file, the table is a copy checked by TestFarmZonesMatchPlan).
var FarmZones = []FarmZone{{2, 2}, {3, 5}, {4, 7}, {9, 8}, {13, 8}, {6, 9}, {5, 10}, {7, 11}, {29, 13}, {35, 15}}

// FarmOpts selects the cycle rule. Pct 100 is today's clear-everything behaviour; Rule turns the value rule on:
// from Pct on, the cycle ends when the rate of the last RuleWindow minutes is below the cycle average a reset yields.
type FarmOpts struct {
	Pct        float64 // coverage percent from which a reset is allowed (100 = never before the zone is empty)
	Rule       bool
	RuleWindow float64 // minutes of the observed rate (default 3)
	WalkMin    float64 // minutes of the walk back from town; <0 = the graph's walk
}

// FarmRate is the XP per hour of zone a farmed at clvl cl for H minutes, with the cycle rule o. The level is held
// (the XP is put back after every cycle), so the numbers compare rules, not levelling. Cycles counts the fresh games.
func (e *Env) FarmRate(a *Area, cl int, o FarmOpts, H float64) (xpPerHour float64, cycles int) {
	if o.RuleWindow == 0 {
		o.RuleWindow = 3
	}
	base := (expTable[cl] + expTable[min(cl+1, MaxLevel)]) / 2
	s := NewState()
	s.XP = base
	s.WP = e.wp // every waypoint of the route is active (the plan's way_in)
	var gain float64
	minF := 1 - o.Pct/100
	for guard := 0; s.T < H && guard < 500; guard++ {
		walk := e.Travel(&s, a.ID) // town to the zone (waypoint hop or walk, graph minutes)
		if o.WalkMin >= 0 && walk > 0 {
			s.T += o.WalkMin - walk
			walk = o.WalkMin
		}
		t0, x0 := s.T-walk, s.XP
		type smp struct{ t, x float64 }
		var hist []smp
		if !o.Rule {
			e.Farm(&s, a, Exit{MinF: minF, TMax: H}, nil)
		} else {
			for s.T < H {
				e.Farm(&s, a, Exit{MaxMin: 0.1, TMax: H}, nil)
				hist = append(hist, smp{s.T, s.XP})
				if e.Remaining(&s, a) <= 0.001 {
					break // swept empty
				}
				if 1-e.Remaining(&s, a) < o.Pct/100 {
					continue
				}
				// the observed rate of the last window against the average of a cycle that resets now
				var old smp
				for _, h := range hist {
					if s.T-h.t >= o.RuleWindow {
						old = h
					}
				}
				if old.t == 0 {
					continue
				}
				now := (s.XP - old.x) / (s.T - old.t)
				over := e.P.RestartMin + walk
				if now < (s.XP-x0)/(s.T-t0+over) {
					break
				}
			}
		}
		gain += s.XP - base
		s.XP = base
		if s.T >= H {
			break
		}
		e.Restart(&s)
		cycles++
	}
	return gain / H * 60, cycles
}

// FarmRow is one zone of the comparison.
type FarmRow struct {
	Zone  string
	Clvl  int
	Base  float64         // XP/h clearing everything (pct 100)
	ByPct map[int]float64 // XP/h at a fixed reset share
	Rule  map[int]float64 // XP/h with the value rule from that share on
	Walk  float64         // minutes of the walk back
	Games map[int]int
}

// FarmReset runs every zone of FarmZones for an hour at the given reset shares.
func (e *Env) FarmReset(pcts []int, H float64) []FarmRow {
	var rows []FarmRow
	for _, z := range FarmZones {
		a := e.D.Areas[z.Area]
		if a == nil {
			continue
		}
		probe := NewState()
		probe.WP = e.wp
		w := e.Travel(&probe, a.ID)
		r := FarmRow{Zone: a.Name, Clvl: z.Clvl, ByPct: map[int]float64{}, Rule: map[int]float64{}, Walk: w, Games: map[int]int{}}
		r.Base, _ = e.FarmRate(a, z.Clvl, FarmOpts{Pct: 100, WalkMin: -1}, H)
		for _, p := range pcts {
			r.ByPct[p], r.Games[p] = e.FarmRate(a, z.Clvl, FarmOpts{Pct: float64(p), WalkMin: -1}, H)
			r.Rule[p], _ = e.FarmRate(a, z.Clvl, FarmOpts{Pct: float64(p), Rule: true, WalkMin: -1}, H)
		}
		rows = append(rows, r)
	}
	return rows
}

// FormatFarmReset prints the table: XP per hour clearing everything, then the gain in percent per reset share (fixed
// share, and with the value rule in brackets).
func FormatFarmReset(title string, rows []FarmRow, pcts []int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n%-20s %4s %6s %8s", title, "zone", "clvl", "walk", "xp/h@100")
	for _, p := range pcts {
		fmt.Fprintf(&b, " %13s", fmt.Sprintf("@%d (rule)", p))
	}
	b.WriteString("\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "%-20s %4d %5.1fm %8.0f", r.Zone, r.Clvl, r.Walk, r.Base)
		for _, p := range pcts {
			fmt.Fprintf(&b, " %+5.1f%% (%+4.1f)", 100*(r.ByPct[p]/r.Base-1), 100*(r.Rule[p]/r.Base-1))
		}
		b.WriteString("\n")
	}
	return b.String()
}
