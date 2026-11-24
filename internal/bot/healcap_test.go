package bot

import (
	"d2lod-bot/internal/task"
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
)

// V lab 2026-11-23 (necro.act1-tools): in town the healing floor doubles to 4, a Necromancer's belt holds two healing
// columns beside the mana ones, the shop buys nothing for a full belt, and every visit asked again ("[2 healing potions
// heal]", 11 trips). The floor is capped by what the belt plan can hold. Two classes: the Necromancer (cap 2) and the
// Paladin (cap 3). Must fail while the town floor is 4 for both.
func TestTownHealingFloorIsCappedByTheBelt(t *testing.T) {
	for _, tc := range []struct {
		class  d2.Class
		cap    int
		belt   int
		wantHP bool
	}{
		{d2.Necromancer, 2, 2, false},
		{d2.Necromancer, 2, 1, true},
		{d2.Paladin, 3, 3, false},
		{d2.Paladin, 3, 2, true},
	} {
		g := newRig(t)
		g.w.set(func(w *world) { w.level, w.area, w.gold0, w.class = 20, 1, 800, tc.class })
		ids := make([]uint32, tc.belt)
		for i := range ids {
			ids[i] = uint32(100 + i)
		}
		g.w.belt(ids...)
		s, _ := g.w.snapshot()
		plan, _, _ := g.r.choresDue(s)
		got := strings.Contains(plan.Reason, "healing potions")
		if got != tc.wantHP {
			t.Errorf("%v belt %d (cap %d): healing-potion chore %v, want %v (reason %q)", tc.class, tc.belt, tc.cap, got, tc.wantHP, plan.Reason)
		}
	}
}

// The owner's try anyway skips readiness words only: the potion trip (a survival reaction) is the same with it on and off.
// Necromancer and Paladin. Must fail if the override ever reached choresDue.
func TestTryAnywayLeavesTheSurvivalTripsAlone(t *testing.T) {
	for _, class := range []d2.Class{d2.Necromancer, d2.Paladin} {
		g := newRig(t)
		g.w.set(func(w *world) { w.level, w.area, w.gold0, w.class = 12, 4, 800, class })
		g.w.belt() // no potion at all
		s, _ := g.w.snapshot()
		plan0, emerg0, ok0 := g.r.choresDue(s)
		task.SetTry(true)
		plan1, emerg1, ok1 := g.r.choresDue(s)
		task.SetTry(false)
		if !ok0 || !emerg0 || ok0 != ok1 || emerg0 != emerg1 || plan0.Reason != plan1.Reason {
			t.Errorf("%v: trip off %v/%v/%q, on %v/%v/%q", class, ok0, emerg0, plan0.Reason, ok1, emerg1, plan1.Reason)
		}
	}
}
