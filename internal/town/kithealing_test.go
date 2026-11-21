package town_test

import (
	"context"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/town"
)

// Proving Ground act1more (act1.tools.malus, 2026-11-20): the route guard wants KitMinHP healing potions (belt and pack)
// while the belt plan of a 4 slot belt with a mana column holds fewer. The shop saw nothing to buy, the chores step
// ended "still needed" and the hero never restocked, with gold in the pack or in the stash.
func TestKitHealingBoughtWhenPlanIsFull(t *testing.T) {
	for _, tc := range []struct {
		name  string
		class d2.Class
	}{{"paladin", d2.Paladin}, {"sorceress", d2.Sorceress}} {
		r, _ := withdrawRig(t, 1500, 0, 22, false)
		r.W.Player.Class = tc.class
		// three healing and one mana: every column the plan reserves is full, the guard counts 3 < KitMinHP
		for i := 0; i < 3; i++ {
			r.W.Add("hp1", d2.ContBelt, i, 0)
		}
		r.W.Add("mp1", d2.ContBelt, 3, 0)
		r.W.Add("tsc", d2.ContInventory, 9, 3)
		rep, err := r.T.Chores(context.Background(), town.ChorePlan{Potions: true, Scrolls: true, Gold: true})
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if rep.Bought["Minor Healing Potion"] < 1 {
			t.Errorf("%s: bought %v, want a healing potion toward the kit minimum of %d", tc.name, rep.Bought, r.Env.Tune.Resource.KitMinHP)
		}
	}
}

// The stash pays: a purse of 6 gold, 3 healing potions carried, gold in the stash.
func TestKitHealingWithdrawsFromStash(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		r, c := withdrawRig(t, 6, 5000, 22, false)
		r.W.Player.Class = class
		for i := 0; i < 3; i++ {
			r.W.Add("hp1", d2.ContBelt, i, 0)
		}
		r.W.Add("mp1", d2.ContBelt, 3, 0)
		r.W.Add("tsc", d2.ContInventory, 9, 3)
		if _, err := r.T.Chores(context.Background(), town.ChorePlan{Potions: true, Scrolls: true, Gold: true}); err != nil {
			t.Fatal(err)
		}
		if len(c.moves) == 0 || c.moves[0] >= 0 {
			t.Errorf("class %d: moves %v, want a withdrawal for the kit potion", class, c.moves)
		}
	}
}

// Below kitGateLevel the minimum does not apply (the Den hero keeps its small belt shopping).
func TestKitHealingNotBelowGateLevel(t *testing.T) {
	r, _ := withdrawRig(t, 1500, 0, 3, false)
	for i := 0; i < 2; i++ {
		r.W.Add("hp1", d2.ContBelt, i, 0)
	}
	r.W.Add("mp1", d2.ContBelt, 2, 0)
	r.W.Add("mp1", d2.ContBelt, 3, 0)
	r.W.Add("tsc", d2.ContInventory, 9, 3)
	rep, err := r.T.Chores(context.Background(), town.ChorePlan{Potions: true, Scrolls: true, Gold: true})
	if err != nil {
		t.Fatal(err)
	}
	if n := rep.Bought["Minor Healing Potion"]; n != 0 {
		t.Errorf("bought %d healing potions below the gate level", n)
	}
}
