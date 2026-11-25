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

// V live 2026-11-24: a purse of 859 paid the shopping list, so "withdraw: not needed", and the purse fell to 188 over five
// trips with 1.1 million in the stash. A boss trip (BossPurse) withdraws up to the boss purse; a plan without it does not.
// Two classes. Must fail without the bosspurse chore.
func TestBossPurseIsWithdrawn(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		r, c := withdrawRig(t, 859, 1100000, 20, true)
		r.W.Player.Class = class
		if _, err := r.T.Chores(context.Background(), town.ChorePlan{Potions: true, Scrolls: true, Gold: true, BossPurse: 3000}); err != nil {
			t.Fatalf("%v: %v", class, err)
		}
		if len(c.moves) == 0 || c.moves[0] >= 0 || -c.moves[0] != 3000-859 {
			t.Fatalf("%v: moves %v, want a withdrawal of %d first", class, c.moves, 3000-859)
		}
	}
}

func TestNoBossPurseWithoutBossTrip(t *testing.T) {
	r, c := withdrawRig(t, 859, 1100000, 20, true)
	if _, err := r.T.Chores(context.Background(), town.ChorePlan{Potions: true, Scrolls: true, Gold: true}); err != nil {
		t.Fatal(err)
	}
	for _, m := range c.moves {
		if m < 0 && -m > 1000 {
			t.Fatalf("moves %v: a plain trip took the boss purse", c.moves)
		}
	}
}

// V live 2026-11-24 00:16Z: the boss purse withdrawn at the start of the visit went back to the stash in the same visit (the
// stash chore kept only the carry line: "stash -1556", left 1300). With BossPurse the purse stays. Two classes. Needs game tables.
func TestBossPurseStaysInThePack(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		r, c := withdrawRig(t, 859, 1100000, 20, true)
		r.W.Player.Class = class
		if _, err := r.T.Chores(context.Background(), town.ChorePlan{Potions: true, Scrolls: true, Gold: true, Stash: true, BossPurse: 3000}); err != nil {
			t.Fatalf("%v: %v", class, err)
		}
		for _, m := range c.moves[1:] {
			if m > 0 {
				t.Fatalf("%v: moves %v: gold went back to the stash during the boss trip", class, c.moves)
			}
		}
		if r.W.Player.Gold < 2000 {
			t.Errorf("%v: purse %d at the end of the visit", class, r.W.Player.Gold)
		}
	}
}
