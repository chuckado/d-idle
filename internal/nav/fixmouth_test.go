package nav

import (
	"testing"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/input"
)

// FIXMOUTH, geometry V from the DRILLS14 logs: d2d arrives (+5,+10) from the listed exit position, 11.2 subtiles.
func TestMouthStepArrivalRadius(t *testing.T) {
	exit := []d2map.Exit{{Dest: 8, Pos: d2map.Point{X: 2048, Y: 2050}}}
	for _, c := range []struct {
		name    string
		at      input.Point
		arrival bool
		want    bool
	}{
		{"arrival at the drill offset", input.Point{X: 2053, Y: 2060}, true, true},
		{"same cell, not an arrival", input.Point{X: 2053, Y: 2060}, false, false},
		{"arrival 20 away", input.Point{X: 2048, Y: 2070}, true, false},
		{"arrival 16.2 away", input.Point{X: 2048 + 3, Y: 2050 + 16}, true, false},
		{"plain walk within mouthR", input.Point{X: 2053, Y: 2050}, false, true},
	} {
		_, n := newKnown(t, openField(2000, 2000, 120, 120), c.at, true)
		o, err := n.observe()
		if err != nil {
			t.Fatal(err)
		}
		o.lm.Exits = exit
		n.arrival = c.arrival
		to, ok := n.mouthStep(o)
		if ok != c.want {
			t.Errorf("%s: ok = %v, want %v (step %v)", c.name, ok, c.want, to)
		}
	}
}

// The first roam walk after a warp arrival starts with the step away, not with a click toward the coverage point.
func TestRoamFirstWalkAfterArrivalStepsAway(t *testing.T) {
	s, n := newKnown(t, openField(2000, 2000, 120, 120), input.Point{X: 2053, Y: 2060}, true)
	o, err := n.observe()
	if err != nil {
		t.Fatal(err)
	}
	o.lm.Exits = []d2map.Exit{{Dest: 8, Pos: d2map.Point{X: 2048, Y: 2050}}}
	n.walkLevel = o.level + 1 // the last walk was in another level: this is the arrival
	_ = n.Explore(ctxT(t), func() bool { return len(s.clicks) >= 1 })
	if len(s.clicks) == 0 {
		t.Fatal("no click")
	}
	c := s.clicks[0]
	if d := dist(c, input.Point{X: 2048, Y: 2050}); d < 16 {
		t.Fatalf("first click %v is %.1f from the exit: not a step away", c, d)
	}
	if n.arrival {
		t.Fatal("arrival must be consumed by the first walk")
	}
}

func TestNearMouthAnswersAfterArrival(t *testing.T) {
	_, n := newKnown(t, openField(2000, 2000, 120, 120), input.Point{X: 2053, Y: 2060}, true)
	o, err := n.observe()
	if err != nil {
		t.Fatal(err)
	}
	o.lm.Exits = []d2map.Exit{{Dest: 8, Pos: d2map.Point{X: 2048, Y: 2050}}}
	p := input.Point{X: 2053, Y: 2056}
	if n.NearMouth(p) {
		t.Fatal("off without guard or arrival")
	}
	n.arrival = true
	if !n.NearMouth(p) || n.NearMouth(input.Point{X: 2100, Y: 2100}) {
		t.Fatal("after an arrival, within mouthR only")
	}
}
