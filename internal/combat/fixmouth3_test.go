package combat

import (
	"context"
	"testing"

	"d2lod-bot/internal/input"
)

// guardR15 mirrors nav.NearMouth after FIXMOUTH3 (radius 15, the guard held for the whole stay) around the Den exit (35,135).
type guardR15 struct {
	Mover
	walks []input.Point
}

func (m *guardR15) NearMouth(p input.Point) bool {
	dx, dy := p.X-35, p.Y-135
	return dx*dx+dy*dy <= 225
}
func (m *guardR15) GoTo(ctx context.Context, p input.Point) error {
	m.walks = append(m.walks, p)
	return m.Mover.GoTo(ctx, p)
}

// FIXMOUTH3, V (DRILLS17): the engage approach to the Zombie (43,141) is 10.0 from the exit and the click (42,146) that fired
// the trigger is 13.0 from it: neither is walked to, a point 27 away still is.
func TestMoveSkipsLoggedBounceTargets(t *testing.T) {
	h := newHarness(t)
	mv := &guardR15{Mover: h.s}
	h.f.mv = mv
	for _, p := range []input.Point{{X: 43, Y: 141}, {X: 42, Y: 146}} {
		if err := h.f.move(context.Background(), p); err != nil {
			t.Fatal(err)
		}
	}
	if len(mv.walks) != 0 {
		t.Fatalf("walked toward the mouth: %v", mv.walks)
	}
	if err := h.f.move(context.Background(), input.Point{X: 61, Y: 148}); err != nil || len(mv.walks) != 1 {
		t.Fatalf("a far point must be walked to: %v %v", mv.walks, err)
	}
}
