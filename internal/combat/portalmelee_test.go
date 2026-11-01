package combat

import "testing"

// Soak 2 death shape (card fixportalmelee): 34% life, no potions, a scroll, three or more attackers in reach.
func TestSafetyMeleeCutCastScenario(t *testing.T) {
	build := func(adj int) (*harness, SafetyAction) {
		h, _ := newPortalHarness(t)
		h.s.life = 34
		for i := 0; i < adj; i++ {
			h.s.addMon(uint32(10+i), clsFallen, 101+float64(i), 100, 500, 5, 6)
		}
		h.s.addMon(40, clsFallen, 106, 100, 500, 5, 6) // a fifth, further out
		snap, _ := h.s.snapshot()
		return h, h.f.Safety(snap)
	}
	if _, a := build(3); a != Chicken {
		t.Errorf("3 adjacent, no potion, scroll: got %v want Chicken", a)
	}
	if _, a := build(1); a != TownPortal {
		t.Errorf("1 adjacent: got %v want TownPortal (unchanged)", a)
	}
	if _, a := build(0); a != TownPortal {
		t.Errorf("none adjacent: got %v want TownPortal (unchanged)", a)
	}
}

func TestMeleeCastGuardOrder(t *testing.T) {
	heal := Belt{}
	heal.Cols[0] = BeltColumn{PotionHealing, 1, 2, false}
	rej := Belt{}
	rej.Cols[1] = BeltColumn{PotionRejuv, 1, 1, false}
	mk := func(adj int, b Belt, mut func(*SafetyInput)) SafetyInput {
		in := crowdIn(34, 1, adj, func(i *SafetyInput) { i.Belt = b; i.Tune.CrowdN = 0 })
		if mut != nil {
			mut(&in)
		}
		return in
	}
	for _, c := range []struct {
		name string
		in   SafetyInput
		want SafetyAction
	}{
		{"3 adjacent no potion: chicken", mk(3, Belt{}, nil), Chicken},
		{"3 adjacent, potion ready: drink", mk(3, heal, nil), DrinkLife},
		{"3 adjacent, rejuv ready: rejuv first", mk(3, rej, nil), DrinkRejuv},
		{"3 adjacent, potion on cooldown: base rules say nothing", mk(3, heal, func(i *SafetyInput) { i.HealReady = false }), SafetyNone},
		{"2 adjacent no potion: portal as before", mk(2, Belt{}, nil), TownPortal},
		{"no scroll: no portal action to replace", mk(3, Belt{}, func(i *SafetyInput) { i.HasPortal = false }), SafetyNone},
		{"quiet (not in a fight): portal as before", mk(3, Belt{}, func(i *SafetyInput) { i.InFight = false }), TownPortal},
	} {
		if got := Decide(c.in); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}
