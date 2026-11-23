package nav

import (
	"testing"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/input"
)

// nav.warpRead (card warptrigger): with the engine's trigger area of an exit the mouth gate refuses exactly the clicks
// that land on it (plus readTriggerMargin), not every click within the guessed 15 subtile bound. Same Tower Cellar 5
// walk as mouthgoal_test.go for the Paladin and the Sorceress.

// cellar5Area is a small click area around the stairs at (130,60): rows 58..62, x 128..132.
func cellar5Area() *d2map.Trigger {
	t := &d2map.Trigger{}
	for y := 58; y <= 62; y++ {
		t.Click = append(t.Click, d2map.TriggerRun{Y: y, X0: 128, X1: 132})
	}
	return t
}

func warpNav(t *testing.T, start input.Point, area *d2map.Trigger, knob bool) (*sim, *Navigator) {
	t.Helper()
	s, n := cellar5Nav(t, start)
	tuningOf(n).Nav.WarpRead = knob
	lm := n.w.Current()
	for i := range lm.Exits {
		lm.Exits[i].Trigger = area
	}
	return s, n
}

// Fails without the read area (the guessed bound refuses the waypoint 13.5 from the stairs, no goal set).
func TestReadAreaAllowsTheClickTheGuessedBoundRefused(t *testing.T) {
	for _, c := range cellar5Starts {
		s, n := warpNav(t, c.start, cellar5Area(), true)
		if err := n.in.ClickWorld(cellar5Edge); err != nil {
			t.Fatalf("%s: refused a click %.1f from the stairs and off the read area: %v", c.class, dist(cellar5Edge, mp(cellar5Stairs.Pos)), err)
		}
		if len(s.clicks) != 1 || s.clicks[0] != cellar5Edge {
			t.Fatalf("%s: clicks %v", c.class, s.clicks)
		}
	}
}

func TestReadAreaRefusesAClickOnIt(t *testing.T) {
	for _, c := range cellar5Starts {
		s, n := warpNav(t, c.start, cellar5Area(), true)
		for _, to := range []input.Point{{X: 130, Y: 60}, {X: 126, Y: 58}} { // on it; inside the margin
			if n.MouthClickOK(c.start, to) {
				t.Errorf("%s: MouthClickOK allows %v on the read area", c.class, to)
			}
			s.clicks = nil
			if err := n.in.ClickWorld(to); err == nil && len(s.clicks) > 0 && areasHit([]*d2map.Trigger{cellar5Area()}, s.clicks[len(s.clicks)-1]) {
				t.Errorf("%s: click %v reached the area (sent %v)", c.class, to, s.clicks)
			}
		}
		if !n.MouthClickOK(c.start, cellar5Edge) {
			t.Errorf("%s: MouthClickOK refuses a click off the read area", c.class)
		}
	}
}

// A click past the area is clipped back to the last cell the engine would not take.
func TestReadAreaClipsAClickBeyondIt(t *testing.T) {
	from, to := input.Point{X: 124, Y: 74}, input.Point{X: 130, Y: 60}
	q, ok := clickClear(from, to, []*d2map.Trigger{cellar5Area()})
	if !ok || areasHit([]*d2map.Trigger{cellar5Area()}, q) || dist(from, q) < gateMinClip {
		t.Fatalf("clip %v ok %v", q, ok)
	}
}

// With the knob off, or an exit without a read area, the guessed bound decides as before.
func TestNoReadAreaKeepsTheGuessedBound(t *testing.T) {
	for _, c := range cellar5Starts {
		for _, tc := range []struct {
			name string
			area *d2map.Trigger
			knob bool
		}{{"knob off", cellar5Area(), false}, {"no area", nil, true}} {
			s, n := warpNav(t, c.start, tc.area, tc.knob)
			if err := n.in.ClickWorld(cellar5Edge); err == nil && len(s.clicks) > 0 && s.clicks[len(s.clicks)-1] == cellar5Edge {
				t.Errorf("%s (%s): sent inside the guessed bound without a goal", c.class, tc.name)
			}
		}
	}
}

func TestTriggerClickNear(t *testing.T) {
	tr := cellar5Area()
	for _, c := range []struct {
		x, y, m int
		want    bool
	}{{130, 60, 0, true}, {127, 60, 0, false}, {127, 60, 1, true}, {135, 60, 2, false}, {130, 64, 2, true}, {130, 65, 2, false}} {
		if got := tr.ClickNear(c.x, c.y, c.m); got != c.want {
			t.Errorf("ClickNear(%d,%d,%d) = %v, want %v", c.x, c.y, c.m, got, c.want)
		}
	}
	if (*d2map.Trigger)(nil).ClickNear(1, 1, 5) {
		t.Error("nil trigger takes nothing")
	}
}
