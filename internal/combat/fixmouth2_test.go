package combat

import (
	"context"
	"testing"

	"d2lod-bot/internal/input"
)

// denMouth mirrors nav.NearMouth (radius 12) around the Den of Evil exit (35,135), geometry V from the DRILLS16 logs.
type denMouth struct {
	Mover
	walks []input.Point
}

func (m *denMouth) NearMouth(p input.Point) bool {
	dx, dy := p.X-35, p.Y-135
	return dx*dx+dy*dy <= 144
}
func (m *denMouth) GoTo(ctx context.Context, p input.Point) error {
	m.walks = append(m.walks, p)
	return m.Mover.GoTo(ctx, p)
}

// FIXMOUTH2: the engage approach to the Zombie at (43,141) sat 9.4 from the exit (then 8.9): not walked to.
func TestMoveSkipsDenZombieApproach(t *testing.T) {
	h := newHarness(t)
	mv := &denMouth{Mover: h.s}
	h.f.mv = mv
	for _, p := range []input.Point{{X: 43, Y: 141}, {X: 42, Y: 140}} {
		if err := h.f.move(context.Background(), p); err != nil {
			t.Fatal(err)
		}
	}
	if len(mv.walks) != 0 {
		t.Fatalf("walked toward the mouth: %v", mv.walks)
	}
	if err := h.f.move(context.Background(), input.Point{X: 61, Y: 148}); err != nil || len(mv.walks) != 1 {
		t.Fatalf("a point 27 away must be walked to: %v %v", mv.walks, err)
	}
}

// FIXMOUTH2: evading from 3.2 subtiles of the exit is bounded: after mouthEvadeCap tries it stops and costs time.
func TestEvadeInsideMouthIsCapped(t *testing.T) {
	h := newHarness(t)
	mv := &denMouth{Mover: h.s}
	h.f.mv = mv
	from := input.Point{X: 36, Y: 138}
	t0 := h.s.Now()
	for i := 0; i < 10 && h.f.mayEvade(); i++ { // callers check mayEvade
		if err := h.f.evade(context.Background(), from, input.Point{X: 70, Y: 170}); err != nil {
			t.Fatal(err)
		}
	}
	if len(mv.walks) != mouthEvadeCap {
		t.Fatalf("%d walks, want the cap %d", len(mv.walks), mouthEvadeCap)
	}
	if h.f.mayEvade() {
		t.Fatal("evading must be off after the cap")
	}
	if h.s.Now().Sub(t0) <= 0 {
		t.Fatal("no time passed")
	}
}
