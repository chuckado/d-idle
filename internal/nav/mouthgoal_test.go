package nav

import (
	"context"
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
	{"Paladin", input.Point{X: 125, Y: 76}},
	{"Sorceress", input.Point{X: 124, Y: 74}}, // the second click of the logged walk
}

// Fails (ErrMouthGate, no click) without gateExtend.
func TestGateExtendsAWaypointInsideTheStairsBound(t *testing.T) {
	ex := mp(cellar5Stairs.Pos)
	for _, c := range cellar5Starts {
		_, n := cellar5Nav(t, c.start)
		n.goal, n.goalSet = cellar5Goal, true
		_, exits := n.gatedExits()
		if clickOKAll(c.start, cellar5Edge, exits) {
			t.Fatalf("%s: test geometry: the waypoint must be refused", c.class)
		}
		q, err := n.gateClick(n.in, cellar5Edge)
		if err != nil {
			t.Fatalf("%s: gate refused the waypoint: %v", c.class, err)
		}
		if q == cellar5Edge || !mouthClickOK(c.start, q, ex) || dist(q, cellar5Goal) >= dist(c.start, cellar5Goal) || dist(c.start, q) < gateUsefulClip {
			t.Fatalf("%s: sent %v from %v (%.1f from the stairs): not a safe click toward the goal", c.class, q, c.start, dist(q, ex))
		}
	}
}

// Whole walk: GoTo the Countess' stand makes progress west (clicks sent, no refusal) instead of failing at once.
func TestGoToFarGoalPastStairsBound(t *testing.T) {
	for _, c := range cellar5Starts {
		s, n := cellar5Nav(t, c.start)
		_ = n.GoTo(context.Background(), cellar5Goal) // the far end may not be reachable in the sim; the first leg is the point
		if got := n.gateRefused.Load(); got != 0 {
			t.Errorf("%s: %d refusals", c.class, got)
		}
		if len(s.clicks) == 0 {
			t.Fatalf("%s: no click sent", c.class)
		}
	}
}

// Without a walk goal (a swing, a step-away) nothing is extended and the old refusal stands.
func TestNoExtendWithoutGoal(t *testing.T) {
	_, n := cellar5Nav(t, input.Point{X: 125, Y: 76})
	lm, exits := n.gatedExits()
	if q, found := n.gateExtend(lm, input.Point{X: 125, Y: 76}, exits); found {
		t.Fatalf("extended to %v without a goal", q)
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
