package bot

import (
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
