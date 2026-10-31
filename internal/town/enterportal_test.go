package town_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/input"
	"d2lod-bot/internal/town"
)

// deafPortal drops the next n interactions with a town portal (class 59) as a lockstep d2d did (it read the click as
// a walk, drill D2.t) and counts them all.
type deafPortal struct {
	input.Controller
	n, seen int
	portal  func(uint32) bool
}

func (d *deafPortal) Interact(t input.Target) error {
	if d.portal(t.ID) {
		d.seen++
		if d.n > 0 {
			d.n--
			return nil
		}
	}
	return d.Controller.Interact(t)
}

// Scenario (drill D2.t, docs/d2d-portal-entry.md): a town portal stands open next to the player in the field and the
// first try to enter it does nothing (a lockstep d2d read the click as a walk). ToTown must try that portal again,
// not read a scroll: a new cast only replaces it with the same kind of portal. Live it spent all 3 scrolls.
func TestScenario_OpenPortalBesideThePlayer(t *testing.T) {
	scrolls := func(r *rig) int { return count(r.W, d2.ContInventory) } // the pack holds only the scrolls
	field := func(t *testing.T, n, deaf int) (*rig, *deafPortal) {
		r := newRig(t)
		r.W.Player.Area, r.W.PortalLevel = 2, 1
		for i := 0; i < n; i++ {
			r.W.Add("tsc", d2.ContInventory, i, 0)
		}
		d := &deafPortal{Controller: r.Env.In, n: deaf, portal: func(id uint32) bool {
			for _, o := range r.W.Objs {
				if o.ID == id {
					return o.Class == town.ObjPortal
				}
			}
			return false
		}}
		r.Env.In = d
		return r, d
	}
	ctx := context.Background()

	t.Run("open portal, first entry ignored: entered on the retry, no scroll read", func(t *testing.T) {
		r, d := field(t, 3, 1)
		r.W.AddObject(59, r.W.Player.X+10, r.W.Player.Y)
		if err := r.T.ToTown(ctx); err != nil || r.W.Player.Area != 1 {
			t.Fatalf("err %v area %d", err, r.W.Player.Area)
		}
		if got := scrolls(r); got != 3 || d.seen != 2 {
			t.Fatalf("scrolls %d of 3 left, %d portal interactions (want 3, 2)", got, d.seen)
		}
	})

	t.Run("no portal in reach: one scroll is read and its portal entered", func(t *testing.T) {
		r, _ := field(t, 2, 0)
		if err := r.T.ToTown(ctx); err != nil || r.W.Player.Area != 1 {
			t.Fatalf("err %v area %d", err, r.W.Player.Area)
		}
		if got := scrolls(r); got != 1 {
			t.Fatalf("scrolls %d of 2 left (want 1)", got)
		}
	})

	t.Run("open portal never leads: no scroll read, walk back instead", func(t *testing.T) {
		r, _ := field(t, 2, 1<<30)
		r.W.AddObject(59, r.W.Player.X+10, r.W.Player.Y)
		err := r.T.ToTown(ctx)
		if got := scrolls(r); got != 2 {
			t.Fatalf("scrolls %d of 2 left: a cast cannot fix a portal that does not take us", got)
		}
		if err != nil && !errors.Is(err, town.ErrNoPortal) {
			t.Fatalf("err %v", err)
		}
		if r.W.Player.Area != 1 {
			t.Fatalf("Blood Moor borders the camp: the walk fallback should arrive (area %d, err %v, calls %v)", r.W.Player.Area, err, r.Mv.Calls)
		}
	})
}

// walkingIn is an input whose Interact walks to the portal itself (d2d townInput).
type walkingIn struct{ input.Controller }

func (walkingIn) InteractWalks(input.Target) bool { return true }

// DRILLS6 F8: the pinned character cannot GoTo the portal; with a walking Interact the town layer must not try.
func TestEnterPortalDoesNotWalkWhenInteractWalks(t *testing.T) {
	r := newRig(t)
	r.W.Player.Area, r.W.PortalLevel = 2, 1
	r.Env.In = walkingIn{r.Env.In}
	r.W.AddObject(59, r.W.Player.X+10, r.W.Player.Y)
	if err := r.T.ToTown(context.Background()); err != nil || r.W.Player.Area != 1 {
		t.Fatalf("err %v area %d", err, r.W.Player.Area)
	}
	for _, c := range r.Mv.Calls {
		if strings.HasPrefix(c, "goto") {
			t.Fatalf("walked to the portal: %v", r.Mv.Calls)
		}
	}
}
