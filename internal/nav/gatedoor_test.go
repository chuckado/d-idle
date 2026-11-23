package nav

import (
	"context"
	"errors"
	"math"
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
	// the map lists the door only once the hero is near it (its room is up), like d2d's objects of the rooms laid
	s.onStep = func(s *sim) {
		if math.Hypot(s.px-3002, s.py-3057) < 60 && len(s.objects) == 0 {
			s.objects = append(s.objects, d2Unit(77, 24, 3002, 3057))
			s.levels[0].lm.Presets = []d2map.Preset{{Kind: d2map.PresetObject, ID: 24, Pos: d2map.Point{X: 3002, Y: 3057}}}
		}
	}
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

// The plan to the gate looked complete when it was picked (the cells behind the door were read as open) and the walk
// then found them shut: the walk ends at "the closest reachable point" and ExitTo used to return that error without
// looking for the door (V live, scenario act1.tools.malus, Monastery Gate to Outer Cloister). Must fail without
// doorAfterShortWalk. Both classes: the geometry knows no class, the navigator is built from the same world.
func TestExitToOpensTheDoorAfterAShortWalk(t *testing.T) {
	for _, cls := range []string{"Paladin", "Sorceress"} {
		t.Run(cls, func(t *testing.T) {
			s, cells := doorWorld(t)
			// the known grid shows the door cells open until the hero is near, then reads them shut like the truth
			base := s.onStep
			doorOpen := false
			s.revealRooms() // the first read already shows the cells open
			s.levels[0].known.SetBlock(cells, make([]uint16, cells.W*cells.H))
			s.onStep = func(s *sim) {
				base(s)
				near := math.Hypot(s.px-3002, s.py-3057) < 60
				l := s.levels[0]
				if _, ok := l.known.Flags(cells.X, cells.Y); !ok {
					return
				}
				if !near && !doorOpen {
					l.known.SetBlock(cells, make([]uint16, cells.W*cells.H))
				}
				if near {
					if !doorOpen {
						// a newly read cell: the planner rebuilds its field when the known count changes
						l.known.SetBlock(d2map.Rect{X: 2000, Y: 2000, W: 1, H: 1}, []uint16{0})
					}
					doorOpen = true
					// back to the truth: the cells are wall until the door is used
					l.known.SetBlock(cells, truthCells(s, cells))
				}
			}
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
				t.Fatalf("%s: exit: %v", cls, err)
			}
			if opened != 1 || s.cur().no != 2 {
				t.Fatalf("%s: door used %d times, level %d at %v", cls, opened, s.cur().no, s.pos())
			}
		})
	}
}

func truthCells(s *sim, r d2map.Rect) []uint16 {
	out := make([]uint16, 0, r.W*r.H)
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			f, _ := s.truth.Flags(x, y)
			out = append(out, f)
		}
	}
	return out
}

// A walk along the bottom of a level keeps off the strip that touches another level (V live: it stepped onto the strip
// and ended in the neighbour). Must fail without penalizeBorders.
func TestWalkKeepsOffTheBorderStripOfAnotherLevel(t *testing.T) {
	for _, cls := range []string{"Paladin", "Sorceress"} {
		t.Run(cls, func(t *testing.T) {
			g := openField(3000, 3000, 200, 80)
			blockRect(g, d2map.Rect{X: 3080, Y: 3000, W: 40, H: 56})
			l1 := &simLevel{no: 1, rect: d2map.Rect{X: 3000, Y: 3000, W: 200, H: 60}}
			l3 := &simLevel{no: 3, rect: d2map.Rect{X: 3000, Y: 3060, W: 200, H: 20}}
			s := newSim(g, input.Point{X: 3010, Y: 3058}, l1, l3)
			s.reveal = 300
			s.revealRooms()
			n := s.navigator()
			wps, err := n.PathTo(input.Point{X: 3190, Y: 3058})
			if err != nil {
				t.Fatal(err)
			}
			for _, w := range wps {
				if w.X > 3085 && w.X < 3115 && w.Y > 3057 {
					t.Fatalf("waypoint %v hugs the border strip: %v", w, wps)
				}
			}
		})
	}
}

// The door budget is per ExitTo call: the level is read again after a town trip with its doors shut. Must fail
// without the reset in ExitTo (the third visit used to refuse the door).
func TestExitToOpensTheSameDoorOnEveryVisit(t *testing.T) {
	for _, cls := range []string{"Paladin", "Sorceress"} {
		t.Run(cls, func(t *testing.T) {
			s, cells := doorWorld(t)
			s.onInteract = func(s *sim, tg input.Target) error {
				if tg.ID == 77 {
					openDoor(s, cells)
				}
				return nil
			}
			n := s.navigator()
			n.doorFn = func(class int) bool { return class == 24 }
			n.exploreFn = func(_ context.Context, _ func() bool, _ *input.Point) error { return ErrExplored }
			n.doorTried = map[d2map.Point]int{{X: 3002, Y: 3057}: maxDoorTries} // earlier visits used it up
			if err := n.ExitTo(ctxT(t), 2); err != nil {
				t.Fatalf("exit: %v", err)
			}
		})
	}
}

// The nearest reachable cell toward a cut-off gate must not lie on the strip toward another level: walking there
// changes the level (V live, scenario act1.tools.barracks: the Monastery Gate arrival pocket, the cell chosen was on
// the south strip and the hero ended in the Tamoe Highland). Must fail without the strip filter.
func TestApproachCellKeepsOffOtherLevelsStrip(t *testing.T) {
	for _, cls := range []string{"Paladin", "Sorceress"} {
		t.Run(cls, func(t *testing.T) {
			g := openField(3000, 3000, 200, 120)
			l1 := &simLevel{no: 1, rect: d2map.Rect{X: 3000, Y: 3000, W: 200, H: 120}}
			l3 := &simLevel{no: 3, rect: d2map.Rect{X: 3000, Y: 3120, W: 200, H: 40}}
			s := newSim(g, input.Point{X: 3100, Y: 3050}, l1, l3)
			s.reveal = 300
			s.revealRooms()
			n := s.navigator()
			o, err := n.observeWait(ctxT(t))
			if err != nil {
				t.Fatal(err)
			}
			cell, ok := n.nearestReachableTo(o, input.Point{X: 3100, Y: 3140}, 2)
			if !ok {
				t.Fatal("no cell")
			}
			if cell.Y >= 3117 {
				t.Fatalf("cell %v is on or beside the strip toward level 3", cell)
			}
		})
	}
}
