package nav

import (
	"errors"
	"math"
	"testing"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/input"
)

var (
	denExit = d2map.Exit{Dest: 8, Pos: d2map.Point{X: 2048, Y: 2050}}
	denAt   = func(dx, dy int) input.Point { return input.Point{X: 2048 + dx, Y: 2050 + dy} }
)

func mp(p d2map.Point) input.Point { return input.Point{X: p.X, Y: p.Y} }

func mouthNav(t *testing.T, start input.Point) (*sim, *Navigator, *obs) {
	t.Helper()
	s, n := newKnown(t, openField(2000, 2000, 120, 120), start, true)
	o, err := n.observe()
	if err != nil {
		t.Fatal(err)
	}
	o.lm.Exits = []d2map.Exit{denExit}
	return s, n, o
}

// FIXMOUTH3, geometry V from DRILLS17 (Den): the arrival at 11.2 steps away, the step-away ends 28.7 from the exit and
// clears the one-shot arrival flag. NearMouth must still guard the Zombie at 10.0 (and the point 13.0 whose click fired).
func TestStayGuardOutlivesArrival(t *testing.T) {
	_, n, o := mouthNav(t, denAt(5, 10))
	n.arrival = true
	if _, ok := n.mouthStep(o); !ok {
		t.Fatal("arrival 11.2 from the exit must step away")
	}
	o.pos = denAt(14, 25) // 28.7 away
	if _, ok := n.mouthStep(o); ok {
		t.Fatal("clear of the exit: no step")
	}
	if n.arrival {
		t.Fatal("the one-shot flag ends here, as logged")
	}
	for _, p := range []input.Point{denAt(8, 6), denAt(7, 11), denAt(0, 15), denAt(-14, 0)} {
		if !n.NearMouth(p) {
			t.Errorf("%v (%.1f from the exit) must stay guarded after the arrival ended", p, dist(p, mp(denExit.Pos)))
		}
	}
	if n.NearMouth(denAt(0, 17)) || n.NearMouth(denAt(30, 30)) {
		t.Error("17 and 42 away must not be guarded")
	}
	// the guard belongs to the stay: another level has none
	n.stayLevel = o.level + 1
	if n.NearMouth(denAt(8, 6)) {
		t.Error("the guard of another level must not apply")
	}
}

// A level that never needed a step-away has no stay guard (the default behaviour is unchanged).
func TestNoStayGuardWithoutMouthContact(t *testing.T) {
	_, n, o := mouthNav(t, denAt(30, 30))
	if _, ok := n.mouthStep(o); ok {
		t.Fatal("42 away: no step")
	}
	if n.NearMouth(denAt(8, 6)) {
		t.Fatal("no guard without mouthGuard, avoid set, arrival or contact")
	}
}

// The wanted exit is neither stepped away from nor guarded, in both directions: ExitTo(8) from inside, and the other level's
// exit stays guarded while one is wanted.
func TestWantedExitIsNotGuarded(t *testing.T) {
	_, n, o := mouthNav(t, denAt(5, 10))
	n.arrival = true
	n.stayLevel = o.level
	n.wantExit = 8
	if _, ok := n.mouthStep(o); ok {
		t.Fatal("stepped away from the exit being taken")
	}
	if n.NearMouth(denAt(2, 2)) {
		t.Fatal("the wanted exit must not be guarded")
	}
	n.wantExit = 3 // wants another level: this exit is unwanted again
	if _, ok := n.mouthStep(o); !ok || !n.NearMouth(denAt(2, 2)) {
		t.Fatal("an exit to another level than the wanted one stays guarded")
	}
}

// Roam targets: the position 13 from the exit (DRILLS17 Cave Level 1) is dropped, as is 16; 18 and beyond stay.
func TestRoamTargetsFilteredAtTriggerRadius(t *testing.T) {
	exits := []d2map.Exit{denExit}
	pts := []input.Point{denAt(12, 5), denAt(10, 6), denAt(0, 15), denAt(-11, -9), denAt(0, 16), denAt(0, 18), denAt(25, 0)}
	got := dropNearExits(pts, exits)
	if len(got) != 2 || got[0] != denAt(0, 18) || got[1] != denAt(25, 0) {
		t.Fatalf("kept %v", got)
	}
}

func TestLeavesExit(t *testing.T) {
	e := denExit.Pos
	a := denAt(1, 3) // 3.2 from the exit
	for _, c := range []struct {
		name string
		to   input.Point
		want bool
	}{
		{"straight away", denAt(7, 22), true},
		{"toward the exit", denAt(-6, -10), false},
		{"past the exit", denAt(-3, -9), false},
		{"sideways keeps nearer", denAt(-12, -6), false},
	} {
		if got := leavesExit(a, c.to, e, math.Hypot(1, 3)); got != c.want {
			t.Errorf("%s: %v", c.name, got)
		}
	}
}

// FIXMOUTH3 proof case (DRILLS17 DW.m4 entry 3): start (256,148), exit (255,145), 3.2 away. The step-away heads away from the
// exit (vector exit to bot), several tiles in ONE click, no click lands near the trigger on the way.
func TestStepAwayFrom32HeadsAwayInOneClick(t *testing.T) {
	s, n, o := mouthNav(t, denAt(1, 3))
	n.arrival = true
	to, ok := n.mouthStep(o)
	if !ok {
		t.Fatal("no step")
	}
	if !leavesExit(o.pos, to, denExit.Pos, dist(o.pos, mp(denExit.Pos))) || dist(to, mp(denExit.Pos)) < 3.2+8 {
		t.Fatalf("step %v does not leave the exit", to)
	}
	n.walkLevel = o.level + 1
	_ = n.walk(ctxT(t), func(*obs) (input.Point, float64, error) { return to, 3, nil }, walkOpts{oneClick: true})
	if len(s.clicks) == 0 {
		t.Fatal("no click")
	}
	if len(s.clicks) > 2 {
		t.Errorf("%d clicks for a straight open path, want one long click", len(s.clicks))
	}
	if c := s.clicks[0]; dist(c, mp(denExit.Pos)) < 15 {
		t.Errorf("first click %v is %.1f from the exit", c, dist(c, mp(denExit.Pos)))
	}
}

// A wall behind the straight-away line makes the straight target unreachable: the opposite or flood choice is used and the
// path still never gets nearer than the start.
func TestStepAwayBlockedPicksSafeDirection(t *testing.T) {
	g := openField(2000, 2000, 120, 120)
	blockRect(g, d2map.Rect{X: 2040, Y: 2056, W: 30, H: 60}) // south of the exit
	_, n := newKnown(t, g, denAt(1, 3), true)
	o, err := n.observe()
	if err != nil {
		t.Fatal(err)
	}
	o.lm.Exits = []d2map.Exit{denExit}
	to, ok := n.mouthStep(o)
	if ok && !leavesExit(o.pos, to, denExit.Pos, dist(o.pos, mp(denExit.Pos))) {
		t.Fatalf("step %v crosses the exit", to)
	}
}

// Sim: an Explore sweep that starts inside the mouth never clicks within the trigger radius of the unwanted exit.
func TestExploreSweepNeverClicksInTheMouth(t *testing.T) {
	s, n, o := mouthNav(t, denAt(5, 10))
	n.walkLevel = o.level + 1
	err := n.Explore(ctxT(t), nil)
	if err != nil && !errors.Is(err, ErrExplored) {
		t.Logf("explore: %v", err)
	}
	if len(s.clicks) < 5 {
		t.Fatalf("only %d clicks: the sweep did not run", len(s.clicks))
	}
	for i, c := range s.clicks {
		if d := dist(c, mp(denExit.Pos)); d <= mouthApproachR && i > 0 {
			t.Fatalf("click %d at %v is %.1f from the exit; around: %v", i, c, d, s.clicks[max(0, i-6):min(len(s.clicks), i+3)])
		}
	}
}
