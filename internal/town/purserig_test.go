package town_test

import (
	"context"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/town"
)

// Stage 3 of town.plan on two classes (needs the game tables; skips without them). A thin purse over a big stash is
// topped up to the purse target, and a fat purse banks down to it: one settle, whatever bankFirst and withdrawShop say.
func TestStage3PurseSettles(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Barbarian} {
		r, c := withdrawRig(t, 8, 128000, 11, true)
		r.W.Player.Class = class
		r.Env.Tune.Town.Plan = 3
		r.Env.Tune.Town.WithdrawShop, r.Env.Tune.Town.BankFirst = 0, 0
		if _, err := r.T.Chores(context.Background(), town.ChorePlan{Potions: true, Scrolls: true, Gold: true}); err != nil {
			t.Fatalf("%v: %v", class, err)
		}
		if len(c.moves) == 0 || c.moves[0] >= 0 {
			t.Fatalf("%v: moves %v, want a withdrawal first", class, c.moves)
		}
		if p := r.W.Player; p.Gold+p.StashGold < 128000-2000 || p.Gold < 300+50*11-100 {
			t.Errorf("%v: purse %d stash %d after the settle", class, p.Gold, p.StashGold)
		}
	}
}

func TestStage3BanksDownToTheTarget(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Barbarian} {
		r, c := withdrawRig(t, 8000, 0, 11, true)
		r.W.Player.Class = class
		r.Env.Tune.Town.Plan = 3
		r.Env.Tune.Town.BankFirst = 0
		if _, err := r.T.Chores(context.Background(), town.ChorePlan{Potions: true, Scrolls: true, Gold: true}); err != nil {
			t.Fatalf("%v: %v", class, err)
		}
		if len(c.moves) != 1 || c.moves[0] <= 0 {
			t.Fatalf("%v: moves %v, want one deposit", class, c.moves)
		}
		if p := r.W.Player; p.Gold < 300+50*11 || p.Gold > 300+50*11+1000 || p.StashGold < 6000 {
			t.Errorf("%v: purse %d stash %d, want the carry line plus the plan in the purse", class, p.Gold, p.StashGold)
		}
	}
}
