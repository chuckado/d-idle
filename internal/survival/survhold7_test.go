package survival

import (
	"math"
	"testing"
)

// The 211-like sequence (card survhold7): Burial Grounds modelled at 152 xp/min realizes 7 over 25 minutes while the
// unvisited Mausoleum is modelled at 160. Both builds run the same rule; the knobs are set in the test so it does not
// depend on which archetype row params.json enables.

func knobs(e env, arch string, on bool) {
	row := ArchTune{YieldMinMinutes: e.p.yieldMin(arch), YieldShrinkMinutes: e.p.yieldShrink(arch)}
	if on {
		row.YieldStarveRatio, row.YieldStarveMinutes, row.YieldStarveMul, row.UnvisitedFloor, row.UnvisitedBelow = 0.25, 3, 0.5, 0.3, 0.8
	}
	e.p.YieldStarveRatio, e.p.UnvisitedFloor = 0, 0
	e.p.Archetypes = map[string]ArchTune{arch: row}
}

// scene sets candidate 0 to prior 152, candidate 1 to prior 160 and feeds candidate 0 `minutes` of stay at `realized` xp/min.
func scene(e env, arch string, hero int, minutes, realized float64) (*Ledger, Input) {
	h := sorceress(hero)
	if arch == "bruiser" {
		h = paladin(hero)
	}
	in := activeIn(e, h, arch, t0)
	in.Cands[0].Eco.XPPerMin, in.Cands[1].Eco.XPPerMin = 152, 160
	l := NewLedger(e.p)
	if minutes > 0 {
		l.Observe(in.Cands[0].Key, t0, Event{Kind: EvXP, XP: realized * minutes, Seconds: minutes * 60}, in.Power)
	}
	return l, in
}

func TestYieldStarvedZoneWorthWhatItPaid(t *testing.T) {
	for _, b := range twoBuilds {
		t.Run(b.name, func(t *testing.T) {
			e := newEnv(t, b.arch)
			knobs(e, b.arch, false)
			l, in := scene(e, b.arch, 14, 25, 7)
			off := l.judge(in, in.Cands[0], 0, 1)
			if off.xp < 30 {
				t.Fatalf("premise: the blend alone keeps the zone high, got %.1f", off.xp)
			}
			knobs(e, b.arch, true)
			on := l.judge(in, in.Cands[0], 0, 1)
			if !on.yieldStarved || math.Abs(on.xp-7) > 0.01 {
				t.Fatalf("xp must be the realized 7, got %.1f starved %v", on.xp, on.yieldStarved)
			}
			if on.value > off.value*0.5*(7/off.xp)*1.01 {
				t.Fatalf("value %.2f not cut to realized*mul (off %.2f xp %.1f)", on.value, off.value, off.xp)
			}
			// too few minutes: the trust floor holds, the blend stands
			l2, in2 := scene(e, b.arch, 14, 2, 7)
			if v := l2.judge(in2, in2.Cands[0], 0, 1); v.yieldStarved {
				t.Fatal("two minutes must not starve a zone")
			}
		})
	}
}

func TestUnvisitedPriorDiscountBounded(t *testing.T) {
	for _, b := range twoBuilds {
		t.Run(b.name, func(t *testing.T) {
			e := newEnv(t, b.arch)
			knobs(e, b.arch, true)
			l, in := scene(e, b.arch, 14, 25, 7)
			d := l.unvisitedDiscount(in)
			if d != 0.3 { // 7/152 is far below the floor
				t.Fatalf("discount %v, want the 0.3 floor", d)
			}
			if v := l.judge(in, in.Cands[1], 0, d); math.Abs(v.xp-48) > 0.01 {
				t.Fatalf("Mausoleum prior 160 must read 48, got %.1f", v.xp)
			}
			// a half-way hero is discounted by the ratio, a hero who beats the prior is never boosted
			l, in = scene(e, b.arch, 14, 25, 76)
			if d := l.unvisitedDiscount(in); math.Abs(d-0.5) > 0.01 {
				t.Fatalf("discount %v, want 0.5", d)
			}
			l, in = scene(e, b.arch, 14, 25, 600)
			if d := l.unvisitedDiscount(in); d != 1 {
				t.Fatalf("discount %v, want 1", d)
			}
			// no data, or the shadow: no discount
			l, in = scene(e, b.arch, 14, 0, 0)
			if d := l.unvisitedDiscount(in); d != 1 {
				t.Fatalf("no data: discount %v", d)
			}
			l, in = scene(e, b.arch, 14, 25, 7)
			in.Active = false
			if d := l.unvisitedDiscount(in); d != 1 {
				t.Fatalf("shadow: discount %v", d)
			}
		})
	}
}

// The Paladin-like case: realized matches the prior. Nothing may change, with the knobs on or off, and with the knobs
// off a starved zone's decision line equals the old behaviour.
func TestMatchingYieldIsUnchanged(t *testing.T) {
	for _, b := range twoBuilds {
		t.Run(b.name, func(t *testing.T) {
			var lines [2]string
			for i, on := range []bool{false, true} {
				e := newEnv(t, b.arch)
				knobs(e, b.arch, on)
				l, in := scene(e, b.arch, 14, 25, 150)
				d, err := l.Choose(in)
				if err != nil {
					t.Fatal(err)
				}
				lines[i] = d.Line()
			}
			if lines[0] != lines[1] {
				t.Fatalf("matching yield changed the decision:\n off %s\n on  %s", lines[0], lines[1])
			}
		})
	}
}

func TestKnobsOffIsToday(t *testing.T) {
	e := newEnv(t, "caster")
	knobs(e, "caster", false)
	l, in := scene(e, "caster", 14, 25, 7)
	v := l.judge(in, in.Cands[0], 0, l.unvisitedDiscount(in))
	b, _ := l.BlendYield(in.Cands[0].Key, 152, "caster")
	if v.yieldStarved || v.xp != b {
		t.Fatalf("knobs off: xp %.2f blend %.2f starved %v", v.xp, b, v.yieldStarved)
	}
}
