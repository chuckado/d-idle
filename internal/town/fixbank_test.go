package town_test

import (
	"context"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/input"
	"d2lod-bot/internal/town"
)

// d2dGold models the d2d town controller: the amount prompt starts empty and TypeChar is hidden from the town
// package (embedding the interface hides it), so the only way to move gold is the MoveGold capability.
type d2dGold struct {
	input.Controller
	r     *rig
	moves []int
}

func (c *d2dGold) MoveGold(n int, deposit bool) error {
	c.moves = append(c.moves, n)
	p := &c.r.W.Player
	if deposit && n <= p.Gold {
		p.Gold -= n
		p.StashGold += n
	}
	return nil
}

func goldRig(t *testing.T, gold, clvl int) (*rig, *d2dGold) {
	r := audit(t, gold, clvl, true)
	c := &d2dGold{Controller: r.In, r: r}
	r.Env.In = c
	return r, c
}

// FIXBANK drill D3.s2: a partial deposit on a controller that cannot type reaches the stash through MoveGold.
func TestFixBankPartialDepositViaMoveGold(t *testing.T) {
	r, c := goldRig(t, 2000, 5) // keep line 550 at clvl 5
	rep, err := r.T.Chores(context.Background(), town.ChorePlan{Gold: true})
	if err != nil || rep.Errors != 0 {
		t.Fatalf("%v %+v", err, rep.Steps)
	}
	if r.W.Player.StashGold != 1450 || len(c.moves) != 1 {
		t.Fatalf("stash %d purse %d moves %v", r.W.Player.StashGold, r.W.Player.Gold, c.moves)
	}
}

// FIXBANK drill D3.s: depositing the whole purse works too.
func TestFixBankFullDepositViaMoveGold(t *testing.T) {
	r, c := goldRig(t, 1500, 5)
	if err := r.T.OpenStashForTest(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := r.T.DepositForTest(context.Background(), 1500); err != nil {
		t.Fatal(err)
	}
	if r.W.Player.StashGold != 1500 || r.W.Player.Gold != 0 || len(c.moves) != 1 {
		t.Fatalf("stash %d purse %d moves %v", r.W.Player.StashGold, r.W.Player.Gold, c.moves)
	}
}

// The stash cap still clamps the amount handed to MoveGold.
func TestFixBankDepositRespectsStashCap(t *testing.T) {
	r, c := goldRig(t, 2000, 5)
	r.W.Player.StashGold = town.StashGoldCap - 100
	if _, err := r.T.Chores(context.Background(), town.ChorePlan{Gold: true}); err != nil {
		t.Fatal(err)
	}
	if len(c.moves) != 1 || c.moves[0] != 100 {
		t.Fatalf("moves %v, want [100]", c.moves)
	}
}

// FIXBANK drill D3.r0: a character that cannot pay does not open the smith at all.
func TestFixBankPoorCharacterDoesNotRepair(t *testing.T) {
	r := audit(t, 0, 1, true)
	r.W.RepairBtn = town.RepairAll
	for _, u := range r.W.Items {
		if u.Item.Where == d2.ContEquipment {
			r.W.Stats[u.Addr] = d2.Stats{72: 1, 73: 20}
		}
	}
	rep, err := r.T.Chores(context.Background(), town.ChorePlan{Repair: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range rep.Steps {
		if st.Name == "repair" && (st.Ran || st.Err != "") {
			t.Fatalf("repair step %+v at gold 0", st)
		}
	}
	r.W.Player.Gold = 500
	rep, _ = r.T.Chores(context.Background(), town.ChorePlan{Repair: true})
	if !rep.Repaired {
		t.Fatalf("a character with 500 gold was not repaired: %+v", rep.Steps)
	}
}
