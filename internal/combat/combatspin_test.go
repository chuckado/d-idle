package combat

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/input"
)

// mouthMover reports every point as inside a cave mouth's trigger area (the d2d mouth guard) and counts walks.
type mouthMover struct {
	Mover
	walks int
}

func (m *mouthMover) NearMouth(input.Point) bool { return true }
func (m *mouthMover) GoTo(ctx context.Context, p input.Point) error {
	m.walks++
	return m.Mover.GoTo(ctx, p)
}

// Live (m2early seed 2): a break-away (no potion, no portal, life falling) picked a flee point inside a cave mouth;
// move skipped it without letting time pass and abortFight reported it handled, so the fight loop re-picked the same
// point 40 times until the no-progress watchdog fired. A skipped move costs a tick, and the flee never picks such a
// point in the first place.
func TestFleeIntoCaveMouthDoesNotSpin(t *testing.T) {
	var buf bytes.Buffer
	h := newHarness(t)
	h.s.noTP = true
	h.s.life = 34
	h.s.healAmount = 0
	for i := 0; i < 4; i++ {
		h.s.addMon(uint32(10+i), clsFallen, 102+float64(i%3), 98+float64(i%4), 4000, 0, 0)
	}
	h.f.e.Log = slog.New(slog.NewTextHandler(&buf, nil))
	mv := &mouthMover{Mover: h.s}
	h.f.mv = mv
	_, _ = h.f.Clear(context.Background(), ClearOpts{Timeout: 20 * time.Second})
	if strings.Contains(buf.String(), "no progress, holding") {
		t.Fatal("the watchdog fired: a combat path iterated without time passing")
	}
	if strings.Contains(buf.String(), "move skipped") {
		t.Fatal("a point inside a cave mouth was offered to move")
	}
}

// A skipped move (point in a cave mouth) spends one poll tick, like a failed walk.
func TestSkippedMoveAdvancesTime(t *testing.T) {
	h := newHarness(t)
	mv := &mouthMover{Mover: h.s}
	h.f.mv = mv
	t0 := h.s.Now()
	for i := 0; i < 10; i++ {
		if err := h.f.move(context.Background(), input.Point{X: 90, Y: 90}); err != nil {
			t.Fatal(err)
		}
	}
	if mv.walks != 0 {
		t.Fatal("walked into a cave mouth")
	}
	if got := h.s.Now().Sub(t0); got < 10*ms(h.f.tune().PollMs) {
		t.Fatalf("10 skipped moves advanced only %v", got)
	}
}

// pointMouth is a mouth guard that only covers points within 9 subtiles of (255,145), like nav.NearMouth after a warp arrival.
type pointMouth struct {
	Mover
	walks []input.Point
}

func (m *pointMouth) NearMouth(p input.Point) bool {
	dx, dy := p.X-255, p.Y-145
	return dx*dx+dy*dy <= 81
}
func (m *pointMouth) GoTo(ctx context.Context, p input.Point) error {
	m.walks = append(m.walks, p)
	return m.Mover.GoTo(ctx, p)
}

// FIXMOUTH (Cave Level 1): the Clear approach toward a monster beside the mouth walked and fired the return trigger.
func TestMoveSkipsTargetsInsideAMouthOnly(t *testing.T) {
	h := newHarness(t)
	mv := &pointMouth{Mover: h.s}
	h.f.mv = mv
	if err := h.f.move(context.Background(), input.Point{X: 258, Y: 153}); err != nil {
		t.Fatal(err)
	}
	if len(mv.walks) != 0 {
		t.Fatalf("walked to a point in the mouth: %v", mv.walks)
	}
	if err := h.f.move(context.Background(), input.Point{X: 90, Y: 90}); err != nil {
		t.Fatal(err)
	}
	if len(mv.walks) != 1 {
		t.Fatalf("a point elsewhere must be walked to: %v", mv.walks)
	}
}
