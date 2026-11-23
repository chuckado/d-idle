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
		// the shop spends after the withdrawal, so the purse itself ends lower: judge the amount moved
		if got := -c.moves[0]; got < 300+50*11 || got > 300+50*11+2000 {
			t.Errorf("%v: withdrew %d, want the carry line (850) plus the plan's spend", class, got)
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
		// the pack keeps the carry line plus the plan's spend before the shop, the rest is banked
		if kept := 8000 - c.moves[0]; kept < 300+50*11 || kept > 300+50*11+2000 {
			t.Errorf("%v: kept %d before the shop (moves %v), want the carry line (850) plus the plan's spend", class, kept, c.moves)
		}
	}
}

// The default path (town.plan 0, withdrawShop and bankFirst on): a thin purse over a rich stash is topped up to the
// carry line (300 + 50 per level), not to the shopping list. Must fail without the carry line in withdraw. Two classes.
func TestShopWithdrawTopsUpToTheCarryLine(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		r, c := withdrawRig(t, 150, 500000, 19, true)
		r.W.Player.Class = class
		if _, err := r.T.Chores(context.Background(), town.ChorePlan{Potions: true, Scrolls: true, Gold: true}); err != nil {
			t.Fatalf("%v: %v", class, err)
		}
		if len(c.moves) == 0 || c.moves[0] >= 0 {
			t.Fatalf("%v: moves %v, want a withdrawal first", class, c.moves)
		}
		if got := -c.moves[0]; got < 300+50*19-150 {
			t.Errorf("%v: withdrew %d, want at least the carry line (1250) minus the purse", class, got)
		}
	}
}
