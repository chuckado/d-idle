package town_test

import (
	"context"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/town"
)

// tpfix diagnosis 2026-11-18: 8 gold in the pack, 128000 in the stash, a healing potion still carried (so the old
// empty-belt rule stayed quiet), no portal scroll and no tome: the stash must pay the shopping list before the shop.
func TestWithdrawForScrollsAndTome(t *testing.T) {
	for _, tc := range []struct {
		name  string
		class d2.Class
	}{{"sorceress", d2.Sorceress}, {"barbarian", d2.Barbarian}} {
		r, c := withdrawRig(t, 8, 128000, 11, true) // full belt: potions are not the reason
		r.W.Player.Class = tc.class
		rep, err := r.T.Chores(context.Background(), town.ChorePlan{Potions: true, Scrolls: true, Gold: true})
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if len(c.moves) != 1 || c.moves[0] >= 0 {
			t.Fatalf("%s: moves %v, want one withdrawal", tc.name, c.moves)
		}
		if got := -c.moves[0]; got < 500 || got > 5000 {
			t.Errorf("%s: withdrew %d, want the shopping list (not the whole stash)", tc.name, got)
		}
		if len(rep.Bought) == 0 {
			t.Errorf("%s: nothing bought after the withdrawal: %+v", tc.name, rep.Steps)
		}
	}
}

// An empty or unreadable stash (0) never withdraws.
func TestNoShopWithdrawWhenStashReadsZero(t *testing.T) {
	r, c := withdrawRig(t, 8, 0, 11, true)
	if _, err := r.T.Chores(context.Background(), town.ChorePlan{Potions: true, Scrolls: true, Gold: true}); err != nil {
		t.Fatal(err)
	}
	if len(c.moves) != 0 {
		t.Errorf("moves %v with an unreadable stash", c.moves)
	}
}
