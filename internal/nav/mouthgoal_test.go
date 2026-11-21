package nav

import (
	"context"
	"errors"
	"testing"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/input"
)

// Tower Cellar Level 5 on d2d (V, job 20261121-005917, Paladin 11): the stairs from level 24 arrive at (135,67); the
// Countess' read stand is (121,70), 14.2 from them, inside mouthTriggerR; the hero walks there from (125,76). The nav
// has no class data: the Paladin and the Sorceress differ only in where the walk starts (the Sorceress teleports to the
// stairs' side after a town portal, here 5 subtiles from the stairs).
var (
	cellar5Stairs = d2map.Exit{Dest: 24, Pos: d2map.Point{X: 135, Y: 67}}
	cellar5Goal   = input.Point{X: 121, Y: 70}
)

func cellar5Nav(t *testing.T, start input.Point) (*sim, *Navigator) {
	t.Helper()
	s, n := newKnown(t, openField(60, 20, 120, 120), start, true)
	o, err := n.observe()
	if err != nil {
		t.Fatal(err)
	}
	o.lm.Exits = []d2map.Exit{cellar5Stairs}
	n.stayLevel = o.level
	return s, n
}

// The walk to a goal that lies inside the bound of the arrival stairs must not fail with ErrMouthGate, and no click may
// come nearer to the stairs than the closer of its ends. Fails (ErrMouthGate on the first click) without goalPocketOK.
func TestGoToGoalInsideArrivalStairsBound(t *testing.T) {
	for _, c := range []struct {
		class string
		start input.Point
	}{
		{"Paladin", input.Point{X: 125, Y: 76}},
		{"Sorceress", input.Point{X: 131, Y: 71}},
	} {
		s, n := cellar5Nav(t, c.start)
		ex := mp(cellar5Stairs.Pos)
		if err := n.GoTo(context.Background(), cellar5Goal); err != nil {
			t.Fatalf("%s: GoTo %v from %v: %v", c.class, cellar5Goal, c.start, err)
		}
		if len(s.clicks) == 0 {
			t.Fatalf("%s: no click sent", c.class)
		}
		from := c.start
		for _, q := range s.clicks {
			if !goalPocketOK(from, q, ex, cellar5Goal) && !mouthClickOK(from, q, ex) {
				t.Errorf("%s: click %v from %v approaches the stairs", c.class, q, from)
			}
			from = q
		}
		if got := n.gateRefused.Load(); got != 0 {
			t.Errorf("%s: %d refusals", c.class, got)
		}
	}
}

// The same click outside a GoTo (a swing, a step-away) keeps the old rule: its target inside the bound is refused.
func TestClickInsideBoundStillRefusedWithoutGoal(t *testing.T) {
	_, n := cellar5Nav(t, input.Point{X: 125, Y: 76})
	if err := n.in.ClickWorld(cellar5Goal); err == nil || !errors.Is(err, ErrMouthGate) {
		// the gate may rewrite it (a clip); it must never send the goal itself
		if n.MouthClickOK(input.Point{X: 125, Y: 76}, cellar5Goal) {
			t.Fatal("MouthClickOK (the Fighter's check) must still refuse a target inside the bound")
		}
	}
}

// A goal pocket never lets a click move toward the exit: the straight line from the goal's side through the stairs is
// refused even though the goal itself lies inside the bound.
func TestGoalPocketRefusesApproach(t *testing.T) {
	ex := mp(cellar5Stairs.Pos)
	if goalPocketOK(input.Point{X: 125, Y: 76}, input.Point{X: 145, Y: 60}, ex, cellar5Goal) {
		t.Error("a click passing over the stairs must be refused")
	}
	if !goalPocketOK(input.Point{X: 125, Y: 76}, cellar5Goal, ex, cellar5Goal) {
		t.Error("the Countess walk must pass")
	}
	if goalPocketOK(cellar5Goal, input.Point{X: 134, Y: 67}, ex, cellar5Goal) {
		t.Error("walking up to the stairs must be refused")
	}
}

// A click whose straight line grazes the trigger area is rerouted around it (a waypoint), not refused, when a clear
// route exists. The hero stands 17 from the stairs, the clip would be 2 long (a crawl), so the gate detours.
func TestGrazingClickDetoursAroundTrigger(t *testing.T) {
	from, to := input.Point{X: 118, Y: 67}, input.Point{X: 160, Y: 67}
	_, n := cellar5Nav(t, from)
	lm, exits := n.gatedExits()
	if len(exits) != 1 {
		t.Fatalf("exits %v", exits)
	}
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
	if got, err := n.gateClick(n.in, to); err != nil || got == to {
		t.Fatalf("gateClick: %v %v", got, err)
	}
}
