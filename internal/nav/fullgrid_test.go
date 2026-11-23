package nav

import (
	"errors"
	"testing"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/input"
)

// nav.fullGrid (card knownedge): a level grid that is measured whole is planned on alone; a partial one keeps the
// optimistic walk and the reveal path. Both starts are the Tower Cellar 5 walk of the Paladin and the Sorceress
// (cellar5Starts: the nav has no class data, they differ in where the walk starts).

func wholeNav(t *testing.T, start input.Point, fullGrid bool) *Navigator {
	t.Helper()
	_, n := newKnown(t, openField(20, 20, 200, 120), start, true)
	tuningOf(n).Nav.FullGrid = fullGrid
	return n
}

// partialNav: the same level, but only the ground near the hero is known (like a level whose rooms are not up).
func partialNav(t *testing.T, start input.Point, fullGrid bool) *Navigator {
	t.Helper()
	s := newSim(openField(20, 20, 200, 120), start)
	s.Update()
	n := s.navigator()
	n.w = knownSim{s}
	tuningOf(n).Nav.FullGrid = fullGrid
	return n
}

func TestFullyKnownIsMeasuredWithFullGrid(t *testing.T) {
	for _, c := range cellar5Starts {
		if !wholeNav(t, c.start, true).fullyKnown() {
			t.Errorf("%s: a whole grid is not fully known with nav.fullGrid", c.class)
		}
		if partialNav(t, c.start, true).fullyKnown() {
			t.Errorf("%s: a partial grid counts as fully known with nav.fullGrid (the world source's blanket true was trusted)", c.class)
		}
		if !partialNav(t, c.start, false).fullyKnown() {
			t.Errorf("%s: nav.fullGrid 0 must keep the source's blanket answer", c.class)
		}
	}
}

// A goal beyond a whole grid is refused with the goal named; it used to be walked toward as an optimistic "edge".
func TestGoalOutsideWholeGridIsRefused(t *testing.T) {
	goal := input.Point{X: 5, Y: 70} // west of the level rectangle
	for _, c := range cellar5Starts {
		n := wholeNav(t, c.start, true)
		o, err := n.observe()
		if err != nil {
			t.Fatal(err)
		}
		_, err = n.plan(o.lm, c.start, goal, 3, nil, 0)
		if !errors.Is(err, ErrNoPath) {
			t.Errorf("%s: err = %v, want ErrNoPath for a goal outside the whole grid", c.class, err)
		}
	}
}

// With a partial grid the optimistic walk toward the unknown goal stays (nothing known leads there yet).
func TestPartialGridStillWalksOptimistically(t *testing.T) {
	goal := input.Point{X: 200, Y: 120}
	for _, c := range cellar5Starts {
		n := partialNav(t, c.start, true)
		o, err := n.observe()
		if err != nil {
			t.Fatal(err)
		}
		if gridWhole(o.lm) {
			t.Fatalf("%s: test geometry: the grid must be partial", c.class)
		}
		res, err := n.plan(o.lm, c.start, goal, 3, nil, 0)
		if err != nil || !res.optimistic {
			t.Errorf("%s: plan %v optimistic %v, want an optimistic path", c.class, err, res.optimistic)
		}
	}
}

func TestGridWholeNeedsEveryCell(t *testing.T) {
	g := openField(20, 20, 10, 10)
	lm := &d2map.LevelMap{Rect: d2map.Rect{X: 20, Y: 20, W: 10, H: 10}, Grid: g}
	if !gridWhole(lm) {
		t.Fatal("a fully known rectangle is whole")
	}
	lm.Rect.W = 11
	if gridWhole(lm) {
		t.Fatal("a rectangle larger than the known cells is not whole")
	}
	if gridWhole(nil) || gridWhole(&d2map.LevelMap{}) {
		t.Fatal("no map is not whole")
	}
}
