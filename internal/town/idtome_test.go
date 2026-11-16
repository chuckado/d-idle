package town_test

import (
	"context"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/gear/geartest"
)

func idTomeRig(t *testing.T, cls d2.Class, gold int) *rig {
	t.Helper()
	geartest.Need(t)
	r := audit(t, gold, 12, true)
	r.W.Player.Class = cls
	r.W.Add("tsc", d2.ContInventory, 9, 3)
	return r
}

func idTomeQty(r *rig) int {
	for _, u := range r.W.Items {
		if u.Item != nil && r.W.DB.Items[u.Class].Code == "ibk" {
			return int(r.W.Stats[u.Addr][70])
		}
	}
	return -1
}

// Knob off (default): the chore never buys a tome, with or without gold; the manual errand is tested in playerid_sim_test.
func TestIDTomeChoreOffByDefault(t *testing.T) {
	for _, cls := range []d2.Class{d2.Sorceress, d2.Paladin} {
		r := idTomeRig(t, cls, 5000)
		r.Env.Tune.Town.IdentifyTome = 0
		rep := run(t, r)
		if rep.Bought["Tome of Identify top-up"] != 0 || idTomeQty(r) != -1 {
			t.Fatalf("class %v: bought %v tome %d", cls, rep.Bought, idTomeQty(r))
		}
	}
}

// Knob on: no tome -> buy and fill; tome full -> nothing; gold short -> nothing.
func TestIDTomeChoreBuysFillsAndSkips(t *testing.T) {
	for _, cls := range []d2.Class{d2.Sorceress, d2.Paladin} {
		r := idTomeRig(t, cls, 5000)
		r.Env.Tune.Town.IdentifyTome = 1
		run(t, r)
		if q := idTomeQty(r); q != 20 {
			t.Fatalf("class %v: tome holds %d, want 20", cls, q)
		}
		spent := r.W.Player.Gold
		run(t, r) // full: nothing more
		if r.W.Player.Gold != spent {
			t.Fatalf("class %v: a full tome still cost gold", cls)
		}
		poor := idTomeRig(t, cls, 150)
		poor.Env.Tune.Town.IdentifyTome = 1
		rep, err := poor.T.Chores(context.Background(), townPlan())
		if err != nil || idTomeQty(poor) != -1 {
			t.Fatalf("class %v: poor hero got a tome (%d) %v %v", cls, idTomeQty(poor), rep.Bought, err)
		}
	}
}

// No game tables needed: with town.identifyTome 0 (default) the chores never buy identify scrolls or a tome by themselves.
func TestNoAutoIdentifyBuyWithoutTables(t *testing.T) {
	r := newRig(t)
	r.W.Player.Gold = 5000
	r.W.Add("tsc", d2.ContInventory, 9, 3)
	if r.Env.Tune.Town.IdentifyTome != 0 {
		t.Fatal("town.identifyTome must default to 0")
	}
	rep, err := r.T.Chores(context.Background(), townPlan())
	if err != nil {
		t.Fatal(err)
	}
	for name, n := range rep.Bought {
		if n > 0 && (name == "Scroll of Identify" || name == "Tome of Identify" || name == "Tome of Identify top-up") {
			t.Fatalf("auto-bought %s x%d with the knob off", name, n)
		}
	}
	if idTomeQty(r) != -1 {
		t.Fatal("a tome appeared")
	}
}
