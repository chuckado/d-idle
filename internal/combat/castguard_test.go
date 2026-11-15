package combat

import (
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/env"
)

func guardMon(hit int) *data.Monster {
	m := &data.Monster{}
	m.A1Max[d2.Normal] = hit
	return m
}

func guardTargets(n int, m *data.Monster, elite bool) []Target {
	hs := make([]Target, n)
	for i := range hs {
		hs[i] = Target{Mon: m, Dist: 3, Elite: elite}
	}
	return hs
}

// portalDecision runs the whole path: build the guard from the hostiles, then decide with the pure survival rule at the
// given life, with a healing potion ready (the fallback when the cast is refused) and no rejuvenation.
func portalDecision(cls d2.Class, mode, life int, hs []Target) (SafetyAction, SafetyInput) {
	bt, ct := env.DefaultBotTuning(), env.DefaultCombatTuning()
	bt.PortalGuard = mode
	p := &d2.Player{Class: cls, Difficulty: d2.Normal}
	var belt Belt
	belt.Cols[0] = BeltColumn{PotionHealing, 3, 2, false}
	in := SafetyInput{Life: life, Max: 200, Belt: belt, InFight: true, HasPortal: true, HealReady: true, Tune: ct}
	in.Crowd, in.Adjacent = Crowd(p, hs, float64(ct.CrowdAdjacent))
	in.CastGuard = castGuard(bt, ct, p, hs)
	// no healing potion in the belt for the base decision: it opens a portal (decideBase), the guard may refuse it
	in.Belt = Belt{}
	a := Decide(in)
	return a, in
}

func TestPortalGuardTwoBuilds(t *testing.T) {
	for _, cls := range []d2.Class{d2.Sorceress, d2.Paladin} {
		weak, boss := guardMon(4), guardMon(120)
		cases := []struct {
			name string
			life int
			hs   []Target
			mode int
			want SafetyAction
		}{
			{"3 weak, old rule blocks", 100, guardTargets(3, weak, false), 0, Chicken},
			{"3 weak, new rule allows", 100, guardTargets(3, weak, false), 1, TownPortal},
			{"1 strong boss, old rule allows", 100, guardTargets(1, boss, true), 0, TownPortal},
			{"1 strong boss, new rule blocks", 100, guardTargets(1, boss, true), 1, Chicken},
			{"3 weak at very low life, new blocks", 20, guardTargets(3, weak, false), 1, Chicken},
			{"unknown damage, 3 adjacent: old rule", 100, append(guardTargets(2, weak, false), Target{Dist: 3}), 1, Chicken},
			{"unknown damage, 1 adjacent: old rule allows", 100, []Target{{Dist: 3}}, 1, TownPortal},
		}
		for _, c := range cases {
			a, in := portalDecision(cls, c.mode, c.life, c.hs)
			if a != c.want {
				t.Errorf("%v %s: got %v want %v", cls, c.name, a, c.want)
			}
			if b, why := in.castBlocked(); b && why == "" {
				t.Errorf("%v %s: blocked without a reason", cls, c.name)
			}
		}
	}
}

// Mode 1 never allows a cast the old rule blocked unless every adjacent monster has a damage term and the sum is safe.
func TestPortalGuardNeverLooserWithoutTerms(t *testing.T) {
	hs := append(guardTargets(3, guardMon(6), false), Target{Dist: 3})
	_, in := portalDecision(d2.Paladin, 1, 200, hs)
	b, why := in.castBlocked()
	if !b || !strings.Contains(why, "count rule") {
		t.Fatalf("blocked=%v why=%q", b, why)
	}
}

func TestPortalGuardReason(t *testing.T) {
	_, in := portalDecision(d2.Sorceress, 1, 100, guardTargets(1, guardMon(120), true))
	if b, why := in.castBlocked(); !b || !strings.Contains(why, "expected damage") {
		t.Fatalf("blocked=%v why=%q", b, why)
	}
}
