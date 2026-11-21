package readiness

import "testing"

// Cases from docs/research/decomp37-combat.md (B1, Q5): penalty first, then cap (75 + max bonus), floor -100, then
// the percent absorb. Two classes, three difficulties.
func TestWorstResistGameOrder(t *testing.T) {
	p := params{"cap": 75, "poison": 0}
	cases := []struct {
		name  string
		raw   int
		max   int
		abs   int
		diff  int
		basis ResistBasis
		want  int
	}{
		{"normal raw 90 caps", 90, 0, 0, 0, ResistRaw, 75},
		{"nightmare raw 90 is 50 (old order: 35)", 90, 0, 0, 1, ResistRaw, 50},
		{"hell raw 120 is 20 (old order: -25)", 120, 0, 0, 2, ResistRaw, 20},
		{"hell raw 0 floors at -100", 0, 0, 0, 2, ResistRaw, -100},
		{"hell raw 190 with +20 max is 90", 190, 20, 0, 2, ResistRaw, 90},
		{"cap limit 95", 200, 40, 0, 0, ResistRaw, 95},
		{"absorb 20 percent on res 50: 100-(50*80/100)", 50, 0, 20, 0, ResistRaw, 60},
		{"absorb capped at 40 percent", 50, 0, 60, 0, ResistRaw, 70},
		{"shown numbers keep no second penalty", 75, 0, 0, 2, ResistWithPenalty, 75},
		{"shown -30 in Hell stays -30", -30, 0, 0, 2, ResistWithPenalty, -30},
	}
	for _, class := range []string{"sorceress", "paladin"} {
		for _, c := range cases {
			h := hero(class, "hell", 80)
			h.Resist = Resists{Fire: c.raw, Cold: 1000, Lightning: 1000, Poison: 1000}
			h.MaxBonus = Resists{Fire: c.max, Cold: 100, Lightning: 100, Poison: 100} // the others sit at the 95 limit
			h.AbsorbPct = Resists{Fire: c.abs}
			h.Basis = c.basis
			if got, _ := worstResist(p, h, c.diff); got != min(c.want, 95) {
				t.Errorf("%s %s: got %d want %d", class, c.name, got, c.want)
			}
		}
	}
}
