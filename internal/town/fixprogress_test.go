package town_test

import (
	"context"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/gear/geartest"
	"d2lod-bot/internal/town"
)

// FIXPROGRESS (Gate 1 rerun): the game said "no portal source (empty)" while the pack read a tome with charges, so the
// shop saw nothing to buy, the next field check asked for a scroll trip again, and the bot walked town and back for
// 43 virtual minutes with 600 gold.
func emptySourceRig(t *testing.T) *rig {
	t.Helper()
	r := newRig(t)
	r.W.Player.Gold = 700
	r.W.Player.Level = 14
	tome := r.W.Add("tbk", d2.ContInventory, 9, 3)
	r.W.Stats[tome.Addr] = d2.Stats{70: 2}
	for i := 0; i < 4; i++ {
		r.W.Add("hp1", d2.ContBelt, i, 0)
		r.W.Add("mp1", d2.ContBelt, i, 1)
	}
	return r
}

func TestEmptySourceRestocksAgainstTheGame(t *testing.T) {
	geartest.Need(t)
	r := emptySourceRig(t)
	r.T.NotePortalEmpty()
	rep, err := r.T.Chores(context.Background(), town.FullPlan())
	if err != nil {
		t.Fatal(err)
	}
	if rep.Bought["Scroll of Town Portal"] == 0 {
		t.Fatalf("nothing bought although the game found the tome empty: %+v %+v", rep.Bought, rep.Steps)
	}
}

func TestTwoEmptyVisitsQuietTheScrollTrip(t *testing.T) {
	geartest.Need(t)
	r := emptySourceRig(t)
	for i := 0; i < 2; i++ {
		r.T.NotePortalEmpty()
		if _, err := r.T.Chores(context.Background(), town.FullPlan()); err != nil {
			t.Fatal(err)
		}
	}
	// the next field check finds the source empty again (the game's word; the bought scrolls are gone from the pack)
	items := r.W.Items[:0]
	for _, u := range r.W.Items {
		if r.W.DB.Items[u.Class].Code != "tsc" {
			items = append(items, u)
		}
	}
	r.W.Items = items
	r.W.Player.Gold = 700
	r.T.NotePortalEmpty()
	s, _ := r.W.Snapshot()
	if ok, why := r.T.NeedsTrip(s); ok {
		t.Fatalf("scroll trip raised inside the quiet time: %q", why)
	}
	r.Env.Clock.Sleep(11 * time.Minute)
	if ok, why := r.T.NeedsTrip(s); !ok || why != "no town portal scroll" {
		t.Fatalf("after the quiet time: %v %q", ok, why)
	}
}

func TestOneEmptyVisitDoesNotQuietIt(t *testing.T) {
	geartest.Need(t)
	r := emptySourceRig(t)
	r.T.NotePortalEmpty()
	if _, err := r.T.Chores(context.Background(), town.FullPlan()); err != nil {
		t.Fatal(err)
	}
	r.T.NotePortalEmpty()
	s, _ := r.W.Snapshot()
	if ok, _ := r.T.NeedsTrip(s); !ok {
		t.Fatal("one empty visit must not silence the trip")
	}
}
