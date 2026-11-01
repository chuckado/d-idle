package combat

import (
	"context"
	"testing"
	"time"

	"d2lod-bot/internal/input"
	"d2lod-bot/internal/nav"
)

// heldMover fails every walk the way the mouth gate does after its rewrite cap.
type heldMover struct {
	Mover
	walks int
	hold  bool
}

func (m *heldMover) GoTo(ctx context.Context, p input.Point) error {
	m.walks++
	if m.hold {
		return nav.ErrMouthGate
	}
	return m.Mover.GoTo(ctx, p)
}

func runHeld(t *testing.T, hold bool) (*heldMover, *harness) {
	h := newHarness(t)
	// out of range and never walking: every tick wants an approach; each target is dropped after MaxApproaches, a
	// pack of ten keeps the loop going for the whole run (the Den: 117 restarts in one cycle)
	for i := 0; i < 10; i++ {
		h.s.addMon(uint32(10+i), clsFallen, 125+float64(i), 100, 4000, 0, 0)
	}
	mv := &heldMover{Mover: h.s, hold: hold}
	h.f.mv = mv
	_, _ = h.f.Clear(context.Background(), ClearOpts{Timeout: 60 * time.Second})
	return mv, h
}

// FIXMOUTHLOOP: a held walk is not restarted from the same spot on every engage tick.
func TestMouthHoldBacksOff(t *testing.T) {
	mv, h := runHeld(t, true)
	// 60 s: 3 holds 3 s apart, then a 30 s hold, then one retry: far below one walk per tick
	t.Logf("walks %d skips %d", mv.walks, h.f.mouthHold.skips)
	if mv.walks > 8 {
		t.Fatalf("%d walks in 60 s under a held gate", mv.walks)
	}
	if mv.walks < 1 || h.f.mouthHold.skips == 0 {
		t.Fatalf("walks %d skips %d: the back-off never engaged", mv.walks, h.f.mouthHold.skips)
	}
}

// Without a hold the same fight walks on every approach tick (the baseline the back-off replaces).
func TestMouthHoldOffByDefault(t *testing.T) {
	mv, h := runHeld(t, false)
	if h.f.mouthHold.skips != 0 || mv.walks == 0 {
		t.Fatalf("walks %d skips %d without any hold", mv.walks, h.f.mouthHold.skips)
	}
}
