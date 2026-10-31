package nav

import (
	"context"
	"errors"
	"testing"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/input"
)

// A pinned character (solid in the game, open on the map, as the drill's `debug pin on` ring) reaches no goal: a
// plain GoTo leaves by the Escape hook (town portal) long before the full ladder, and without a hook still ends in
// ErrStuck.
func TestPinnedGoToEscapesByHook(t *testing.T) {
	g := openField(1000, 1000, 200, 60)
	s := newSim(g, input.Point{X: 1010, Y: 1030})
	n := s.navigator()
	truth := g.Clone()
	blockRect(truth, d2map.Rect{X: 1004, Y: 1024, W: 5, H: 12}) // west wall
	blockRect(truth, d2map.Rect{X: 1013, Y: 1024, W: 5, H: 12}) // east wall
	blockRect(truth, d2map.Rect{X: 1004, Y: 1024, W: 14, H: 5}) // north
	blockRect(truth, d2map.Rect{X: 1004, Y: 1032, W: 14, H: 5}) // south
	s.truth, s.knownSrc = truth, g
	escaped := 0
	t0 := s.now
	n.Escape = func(ctx context.Context) error {
		escaped++
		s.px, s.py = 1100.5, 1030.5
		return nil
	}
	err := n.GoTo(ctxT(t), input.Point{X: 1150, Y: 1030})
	if !errors.Is(err, ErrLevelChanged) && err != nil || escaped != 1 {
		t.Fatalf("err=%v escaped=%d, want one escape", err, escaped)
	}
	if took := s.now.Sub(t0); took > 40*1e9 {
		t.Fatalf("escape took %v of game time, want a short ladder", took)
	}
	n.Escape = nil
	s.px, s.py = 1010.5, 1030.5
	if err := n.GoTo(ctxT(t), input.Point{X: 1150, Y: 1030}); !errors.Is(err, ErrStuck) {
		t.Fatalf("without a hook: %v, want ErrStuck", err)
	}
}
