package nav

import (
	"context"
	"errors"
	"testing"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/input"
)

// doorWorld: the gateWorld wall with a 14-cell border pocket at y 3050..3063 that a door object (class 24) at (3002,3057)
// shuts off from the town side. Using the door opens the cells in front of it.
func doorWorld(t *testing.T) (*sim, d2map.Rect) {
	t.Helper()
	g := openField(2800, 3000, 400, 120)
	blockRect(g, d2map.Rect{X: 2999, Y: 3000, W: 2, H: 50})
	blockRect(g, d2map.Rect{X: 2999, Y: 3064, W: 2, H: 56})
	doorCells := d2map.Rect{X: 3001, Y: 3048, W: 2, H: 18}
	blockRect(g, doorCells)
	l1 := &simLevel{no: 1, rect: d2map.Rect{X: 3000, Y: 3000, W: 200, H: 120}}
	l2 := &simLevel{no: 2, rect: d2map.Rect{X: 2800, Y: 3000, W: 200, H: 120}}
	s := newSim(g, input.Point{X: 3100, Y: 3090}, l1, l2)
	s.reveal = 300 // the whole level is in view, like a fully known d2d map
	s.objects = append(s.objects, d2Unit(77, 24, 3002, 3057))
	s.levels[0].lm.Presets = []d2map.Preset{{Kind: d2map.PresetObject, ID: 24, Pos: d2map.Point{X: 3002, Y: 3057}}}
	return s, doorCells
}

func openDoor(s *sim, cells d2map.Rect) {
	open := make([]uint16, cells.W*cells.H)
	s.truth.SetBlock(cells, open)
	for _, l := range s.levels {
		if _, ok := l.known.Flags(cells.X, cells.Y); ok {
			l.known.SetBlock(cells, open)
		}
	}
}

func TestExitToOpensTheDoorBesideACutOffGate(t *testing.T) {
	s, cells := doorWorld(t)
	opened := 0
	s.onInteract = func(s *sim, tg input.Target) error {
		if tg.ID == 77 {
			opened++
			openDoor(s, cells)
		}
		return nil
	}
	n := s.navigator()
	n.doorFn = func(class int) bool { return class == 24 }
	n.exploreFn = func(_ context.Context, _ func() bool, _ *input.Point) error { return ErrExplored }
	if err := n.ExitTo(ctxT(t), 2); err != nil {
		t.Fatalf("exit: %v", err)
	}
	if opened != 1 || s.cur().no != 2 {
		t.Fatalf("door used %d times, level %d at %v", opened, s.cur().no, s.pos())
	}
}

// Must fail without the door step: a door that is not a door class is never clicked, and the exit is refused.
func TestExitToLeavesAnObjectThatIsNoDoorAlone(t *testing.T) {
	s, cells := doorWorld(t)
	s.onInteract = func(s *sim, tg input.Target) error { openDoor(s, cells); return nil }
	n := s.navigator()
	n.doorFn = func(int) bool { return false }
	n.exploreFn = func(_ context.Context, _ func() bool, _ *input.Point) error { return ErrExplored }
	err := n.ExitTo(ctxT(t), 2)
	if err == nil || len(s.interacts) != 0 {
		t.Fatalf("err %v interacts %v", err, s.interacts)
	}
	if !errors.Is(err, ErrNoPath) && !errors.Is(err, ErrNoExit) && !errors.Is(err, ErrTimeout) {
		t.Logf("refused with %v", err)
	}
}
