package combat

import (
	"context"
	"errors"
	"testing"

	"d2lod-bot/internal/input"
)

// jumpMover models a navigator after a position jump: its planner knows nothing until Sync reads the new position.
type jumpMover struct {
	Mover
	synced bool
	syncs  int
}

func (m *jumpMover) Sync() error { m.syncs++; m.synced = true; return nil }
func (m *jumpMover) PathTo(p input.Point) ([]input.Point, error) {
	if !m.synced {
		return nil, errors.New("nav: no path: unknown map")
	}
	return m.Mover.PathTo(p)
}

// DRILLS5 F4: the first Clear after a position jump must sync the map, not call everything unreachable.
func TestClearSyncsMapAfterJump(t *testing.T) {
	h := newHarness(t)
	h.f.sight = func(a, b input.Point) bool { return false }
	jm := &jumpMover{Mover: h.s}
	h.f.mv = jm
	h.s.addMon(10, clsFallen, 135, 100, 24, 0, 0)
	res, err := h.f.Clear(context.Background(), ClearOpts{})
	if err != nil || jm.syncs == 0 {
		t.Fatalf("syncs %d err %v", jm.syncs, err)
	}
	for _, ig := range res.Ignored {
		if ig.Why == IgnoreUnreachable {
			t.Fatalf("unreachable after a jump: %+v", res.Ignored)
		}
	}
}

func TestKillSyncsMap(t *testing.T) {
	h := newHarness(t)
	jm := &jumpMover{Mover: h.s}
	h.f.mv = jm
	h.s.addMon(10, clsFallen, 110, 100, 24, 0, 0)
	_ = h.f.Kill(context.Background(), 10)
	if jm.syncs == 0 {
		t.Fatal("Kill did not sync")
	}
}
