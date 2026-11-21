package nav

import (
	"testing"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/input"
)

// Tower Cellar Level 5 on d2d (V, job 20261121-005917, Paladin 11, K1 fork): the stairs from level 24 list at (130,60),
// the hero stands at (125,76), the Countess' read stand is (21,70), 100 subtiles west and outside the known map, so
// the planner's next waypoint is the known grid's edge (121,70), 13.5 from the stairs, and the gate refused it for good
// ("click {121 70}: no path: the click would cross the trigger area of an unwanted exit"). The nav has no class data:
// the Paladin and the Sorceress differ only in where the walk starts.
var (
	cellar5Stairs = d2map.Exit{Dest: 24, Pos: d2map.Point{X: 130, Y: 60}}
	cellar5Goal   = input.Point{X: 21, Y: 70}
	cellar5Edge   = input.Point{X: 121, Y: 70}
)

// cellar5Nav: an open level whose known map starts at x=120 (the Countess' side is not up yet).
func cellar5Nav(t *testing.T, start input.Point) (*sim, *Navigator) {
	t.Helper()
	s, n := newKnown(t, openField(20, 20, 200, 120), start, true)
	o, err := n.observe()
	if err != nil {
		t.Fatal(err)
	}
	o.lm.Exits = []d2map.Exit{cellar5Stairs}
	n.stayLevel = o.level
	return s, n
}

var cellar5Starts = []struct {
	class string
	start input.Point
}{
	{"Paladin", input.Point{X: 124, Y: 74}}, // the second click of the logged walk
	{"Sorceress", input.Point{X: 121, Y: 73}},
}

// Fails (ErrMouthGate, no click) without the goal rule in gateClick.
func TestGoToWaypointInsideStairsBoundGoesThroughWhenGoalBeyond(t *testing.T) {
	for _, c := range cellar5Starts {
		s, n := cellar5Nav(t, c.start)
		n.goal, n.goalSet = cellar5Goal, true
		_, exits := n.gatedExits()
		if clickOKAll(c.start, cellar5Edge, exits) {
			t.Fatalf("%s: test geometry: the waypoint must fail the 15 subtile bound", c.class)
		}
		if err := n.in.ClickWorld(cellar5Edge); err != nil {
			t.Fatalf("%s: gate refused the waypoint: %v", c.class, err)
		}
		if len(s.clicks) != 1 || s.clicks[0] != cellar5Edge {
			t.Fatalf("%s: clicks %v", c.class, s.clicks)
		}
	}
}

// The core stays protected: a waypoint within mouthCoreR of the stairs, or a segment over them, is refused even with a goal.
func TestGoalDoesNotOpenTheCore(t *testing.T) {
	for _, c := range cellar5Starts {
		s, n := cellar5Nav(t, c.start)
		n.goal, n.goalSet = cellar5Goal, true
		for _, to := range []input.Point{{X: 128, Y: 64}, {X: 135, Y: 55}} {
			if err := n.in.ClickWorld(to); err == nil && len(s.clicks) > 0 && s.clicks[len(s.clicks)-1] == to {
				t.Errorf("%s: click %v (%.1f from the stairs) went through", c.class, to, dist(to, mp(cellar5Stairs.Pos)))
			}
		}
	}
}

// Without a walk goal (a swing, a step-away) nothing changes: the click inside the bound is refused.
func TestNoGoalNoPass(t *testing.T) {
	for _, c := range cellar5Starts {
		s, n := cellar5Nav(t, c.start)
		if err := n.in.ClickWorld(cellar5Edge); err == nil && len(s.clicks) > 0 && s.clicks[len(s.clicks)-1] == cellar5Edge {
			t.Errorf("%s: sent without a goal", c.class)
		}
	}
}

// A click whose straight line grazes the trigger area is rerouted around it (a waypoint), not refused, when a clear
// route exists. The hero stands 17 from the stairs, the clip would be 2 long (a crawl), so the gate detours.
func TestGrazingClickDetoursAroundTrigger(t *testing.T) {
	from, to := input.Point{X: 113, Y: 60}, input.Point{X: 160, Y: 60}
	_, n := cellar5Nav(t, from)
	lm, exits := n.gatedExits()
	if clickOKAll(from, to, exits) {
		t.Fatal("test geometry: the straight click must cross the trigger area")
	}
	q, how := n.gateRewrite(lm, from, to, exits)
	if how != "detour" {
		t.Fatalf("how %q (%v), want a detour waypoint", how, q)
	}
	if !clickOKAll(from, q, exits) {
		t.Fatalf("waypoint %v is not safe", q)
	}
}
