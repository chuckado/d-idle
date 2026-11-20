package town_test

import (
	"context"
	"errors"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/town"
)

// A level only a portal leads to (Tristram) is refused at once while the standing town portal leads elsewhere, for both a
// Sorceress and a Barbarian kit (Proving Ground cain.chain: the stale portal of a Dark Wood trip took the hero back there).
func TestBackToFieldPortalOnlyRefusesStalePortal(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Barbarian} {
		r := newRig(t)
		r.W.Player.Class = class
		r.W.Player.Area, r.W.PortalLevel = 5, 1 // Dark Wood, the portal leads to camp
		r.W.Add("tsc", d2.ContInventory, 0, 0)
		if err := r.T.ToTown(context.Background()); err != nil {
			t.Fatalf("class %v: ToTown: %v", class, err)
		}
		r.W.PortalLevel = 5
		err := r.T.BackToField(context.Background(), 38)
		if !errors.Is(err, town.ErrPortalStale) {
			t.Fatalf("class %v: err %v, want ErrPortalStale", class, err)
		}
		if r.W.Player.Area == 5 || len(r.Mv.Calls) != 0 {
			t.Errorf("class %v: moved (area %d, calls %v)", class, r.W.Player.Area, r.Mv.Calls)
		}
	}
}

// A portal cast in Tristram leads back to Tristram, so the same call is allowed then.
func TestBackToFieldPortalOnlyAllowedWhenCastThere(t *testing.T) {
	r := newRig(t)
	r.W.Player.Area, r.W.PortalLevel = 38, 1
	r.W.Add("tsc", d2.ContInventory, 0, 0)
	if err := r.T.ToTown(context.Background()); err != nil {
		t.Fatal(err)
	}
	r.W.PortalLevel = 38
	if err := r.T.BackToField(context.Background(), 38); err != nil || r.W.Player.Area != 38 {
		t.Fatalf("err %v area %d", err, r.W.Player.Area)
	}
}
