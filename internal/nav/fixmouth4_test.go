package nav

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/input"
)

// Geometry of the DRILLS16 to 18 logs, relative to the listed exit position (the Den exit (35,135) is denExit here).
var (
	d18Target   = denAt(2, 15) // Fighter.move target (37,150): 15.1
	d18StepFrom = denAt(2, 13) // step-away start (37,148): 13.2
	d18StepTo   = denAt(5, 33) // step-away goal (40,168)
	d18Swung    = denAt(1, 3)  // after the swings (36,138): 3.2
	d18Replan   = denAt(2, 5)  // the replan's first click (37,140): 5.4, fired
)

func gateNav(t *testing.T, start input.Point) (*sim, *Navigator, *obs) {
	t.Helper()
	s, n, o := mouthNav(t, start)
	n.stayLevel = o.level
	return s, n, o
}

func TestMouthClickOKGeometry(t *testing.T) {
	ex := mp(denExit.Pos)
	for _, c := range []struct {
		name     string
		from, to input.Point
		want     bool
	}{
		{"DRILLS18 move target 15.1 from outside", denAt(6, 16), d18Target, true},
		{"DRILLS18 step-away from 13.2", d18StepFrom, d18StepTo, true},
		{"DRILLS18 replan click 5.4 from 3.2 (fired)", d18Swung, d18Replan, false},
		{"DRILLS18 move target from inside the bound, end 15.1", d18StepFrom, d18Target, true},
		{"DRILLS17 Den click 13.0 (fired)", denAt(14, 25), denAt(7, 11), false},
		{"DRILLS17 Cave click 11.7 (fired)", denAt(12, 5), denAt(10, 6), false},
		{"DRILLS17 step-away second waypoint 12.1 (fired)", denAt(1, 3), denAt(5, 11), false},
		{"DRILLS16 click 16.2 (did not fire), segment outside", denAt(10, 16), denAt(0, 17), true},
		{"segment through the disc to a far point", denAt(-20, 0), denAt(20, 0), false},
		{"outside, segment grazing at 14", denAt(-20, 14), denAt(20, 14), false},
		{"inside, toward the exit then out past it", d18Swung, denAt(-10, -30), false},
		{"inside, straight out far", d18Swung, denAt(7, 22), true},
	} {
		if got := mouthClickOK(c.from, c.to, ex); got != c.want {
			t.Errorf("%s: from %v (%.1f) to %v (%.1f): got %v", c.name, c.from, dist(c.from, ex), c.to, dist(c.to, ex), got)
		}
	}
}

// The DRILLS18 bounce: the character stands at 3.2 after the swings and the replanned step-away clicks (37,140), 5.4
// from the exit. The gate must send a click along the away vector instead.
func TestGateRewritesDrills18ReplanClick(t *testing.T) {
	s, n, _ := gateNav(t, d18Swung)
	if err := n.in.ClickWorld(d18Replan); err != nil {
		t.Fatal(err)
	}
	if len(s.clicks) != 1 {
		t.Fatalf("clicks %v", s.clicks)
	}
	c := s.clicks[0]
	if c == d18Replan || !mouthClickOK(d18Swung, c, mp(denExit.Pos)) {
		t.Fatalf("sent %v (%.1f from the exit)", c, dist(c, mp(denExit.Pos)))
	}
}

func TestGateClipsAGrazingClick(t *testing.T) {
	s, n, _ := gateNav(t, denAt(-30, 10))
	if err := n.in.ClickWorld(denAt(30, 10)); err != nil {
		t.Fatal(err)
	}
	c := s.clicks[0]
	if !mouthClickOK(denAt(-30, 10), c, mp(denExit.Pos)) || dist(denAt(-30, 10), c) < gateMinClip {
		t.Fatalf("clip %v", c)
	}
	if c.X >= denExit.Pos.X {
		t.Fatalf("clip %v went past the exit", c)
	}
}

// From the boundary with no room to clip, and after the rewrite cap, the gate holds: no click and ErrMouthGate (which is
// an ErrNoPath, so roam and explore drop the target).
func TestGateHoldsAfterRewriteCap(t *testing.T) {
	s, n, _ := gateNav(t, d18Swung)
	for i := 0; i < gateMaxRewrites; i++ {
		if err := n.in.ClickWorld(d18Replan); err != nil {
			t.Fatalf("rewrite %d: %v", i, err)
		}
	}
	err := n.in.ClickWorld(d18Replan)
	if !errors.Is(err, ErrMouthGate) || !errors.Is(err, ErrNoPath) {
		t.Fatalf("after the cap: %v", err)
	}
	if len(s.clicks) != gateMaxRewrites {
		t.Fatalf("%d clicks sent, want %d", len(s.clicks), gateMaxRewrites)
	}
	s2, n2, _ := gateNav(t, denAt(0, 16))
	if err := n2.in.ClickWorld(denAt(0, 10)); !errors.Is(err, ErrMouthGate) || len(s2.clicks) != 0 {
		t.Fatalf("no room to clip: err %v clicks %v", err, s2.clicks)
	}
}

// Property: random positions and click targets in a level with one unwanted exit; every click that reaches the game
// keeps the invariant while the stay holds.
func TestGateProperty(t *testing.T) {
	s, n, _ := gateNav(t, denAt(0, 30))
	rng := rand.New(rand.NewSource(4))
	ex := mp(denExit.Pos)
	sent, rewritten, held := 0, 0, 0
	for i := 0; i < 2000; i++ {
		from := denAt(rng.Intn(80)-40, rng.Intn(80)-40)
		if dist(from, ex) < 0.5 {
			continue
		}
		to := denAt(rng.Intn(80)-40, rng.Intn(80)-40)
		s.px, s.py = float64(from.X)+0.5, float64(from.Y)+0.5
		n.gateRewrites = 0
		before := len(s.clicks)
		err := n.in.ClickWorld(to)
		if err != nil {
			if !errors.Is(err, ErrMouthGate) {
				t.Fatal(err)
			}
			held++
			continue
		}
		if len(s.clicks) != before+1 {
			t.Fatalf("one ClickWorld sent %d clicks", len(s.clicks)-before)
		}
		c := s.clicks[len(s.clicks)-1]
		if !mouthClickOK(from, c, ex) {
			t.Fatalf("case %d: from %v to %v sent %v (%.1f from the exit)", i, from, to, c, dist(c, ex))
		}
		sent++
		if c != to {
			rewritten++
		}
	}
	if sent < 1000 || rewritten == 0 || held == 0 {
		t.Fatalf("degenerate: sent %d rewritten %d held %d", sent, rewritten, held)
	}
}

func TestGateIgnoresWantedExit(t *testing.T) {
	s, n, _ := gateNav(t, d18Swung)
	n.wantExit = denExit.Dest
	if err := n.in.ClickWorld(d18Replan); err != nil || s.clicks[0] != d18Replan {
		t.Fatalf("a wanted exit must not be gated: %v %v", err, s.clicks)
	}
	if !n.MouthClickOK(d18Swung, mp(denExit.Pos)) {
		t.Fatal("MouthClickOK must ignore the wanted exit")
	}
	n.wantExit = 3 // another level is wanted: this exit is gated again
	if n.MouthClickOK(d18Swung, d18Replan) {
		t.Fatal("an exit to a level other than the wanted one stays gated")
	}
}

func TestGateOffWithoutStay(t *testing.T) {
	s, n, _ := mouthNav(t, d18Swung)
	if err := n.in.ClickWorld(d18Replan); err != nil || s.clicks[0] != d18Replan {
		t.Fatalf("no stay, no arrival: the click goes as it is: %v %v", err, s.clicks)
	}
	n.arrival = true
	if n.MouthClickOK(d18Swung, d18Replan) {
		t.Fatal("the arrival flag gates as the stay does")
	}
}

func TestGateOffInTown(t *testing.T) {
	g := openField(2000, 2000, 120, 120)
	s := newSim(g, d18Swung, &simLevel{no: 1, rect: g.Bounds()})
	s.reveal = 1000
	s.Update()
	n := s.navigator()
	n.w = knownSim{s}
	n.w.Current().Exits = []d2map.Exit{denExit}
	n.stayLevel, n.arrival = 1, true
	if err := n.in.ClickWorld(d18Replan); err != nil || s.clicks[0] != d18Replan {
		t.Fatalf("towns are never gated: %v %v", err, s.clicks)
	}
}

// A step-away that replans from 3.2 and whose new path heads to the exit picks a new away goal from there.
func TestStepAwayReplanRetargets(t *testing.T) {
	_, n, o := gateNav(t, d18Swung)
	n.mouthLast, n.mouthLastOK = mouthKey{o.level, denExit.Pos.X, denExit.Pos.Y}, true
	fl := &fleeState{exit: mp(denExit.Pos), goal: denAt(-10, -30)}
	res := planResult{wps: []input.Point{d18Swung, d18Replan, denAt(-5, -10)}}
	if !n.fleeRetarget(o, res, fl) {
		t.Fatal("a replan toward the exit must retarget")
	}
	if !leavesExit(o.pos, fl.goal, denExit.Pos, dist(o.pos, mp(denExit.Pos))) {
		t.Fatalf("new goal %v does not leave the exit", fl.goal)
	}
	res = planResult{wps: []input.Point{d18Swung, denAt(3, 9), fl.goal}}
	if n.fleeRetarget(o, res, fl) {
		t.Fatal("a replan that heads away keeps its goal")
	}
	fl.retargets = fleeRetargets
	if n.fleeRetarget(o, planResult{wps: []input.Point{d18Replan}}, fl) {
		t.Fatal("retargets are bounded")
	}
}

// Every world click nav sends goes through the gate: New wraps the controller.
func TestEveryNavClickGoesThroughTheGate(t *testing.T) {
	_, n, _ := mouthNav(t, d18Swung)
	if _, ok := n.in.(mouthGate); !ok {
		t.Fatalf("n.in is %T, not the mouth gate", n.in)
	}
}

// S: an Explore sweep in a level whose stay guard holds, with a pack beside the mouth and a wall that corners the
// character at the mouth. The walk stalls in the pack and
// yields; the "fight" drags the character to 3.2 from the exit (the DRILLS18 swings) and leaves the step-away blocked by
// a monster for a while. No click that reaches the game breaks the invariant.
func TestSweepWithPackAndBlockedStepAwayNeverCrosses(t *testing.T) {
	g := openField(2000, 2000, 120, 120)
	blockRect(g, d2map.Rect{X: 2036, Y: 2056, W: 26, H: 2}) // the corner: the way out from 3.2 runs along it
	s, n := newKnown(t, g, d18StepFrom, true)
	o, err := n.observe()
	if err != nil {
		t.Fatal(err)
	}
	o.lm.Exits = []d2map.Exit{denExit}
	n.stayLevel = o.level
	ex := mp(denExit.Pos)
	type click struct{ from, to input.Point }
	var clicks []click
	s.onClickWorld = func(s *sim, p input.Point) { clicks = append(clicks, click{s.pos(), p}) }
	s.solidUnits = true
	s.units = []d2.Unit{
		{ID: 7, Type: d2.TypeMonster, X: d18StepFrom.X + 1, Y: d18StepFrom.Y + 4, Mode: d2.MonModeStand},
		{ID: 8, Type: d2.TypeMonster, X: d18StepFrom.X - 1, Y: d18StepFrom.Y + 4, Mode: d2.MonModeStand},
		{ID: 9, Type: d2.TypeMonster, X: d18StepFrom.X, Y: d18StepFrom.Y + 5, Mode: d2.MonModeStand},
	}
	yields := 0
	n.Yield = func(ctx context.Context, near []d2.Unit) error {
		yields++
		if yields == 1 {
			s.px, s.py = float64(d18Swung.X)+0.5, float64(d18Swung.Y)+0.5 // the swings walked us to the mouth
			s.target = nil
			return nil
		}
		s.units = nil // the pack is dead
		return nil
	}
	err = n.Explore(ctxT(t), nil)
	if err != nil && !errors.Is(err, ErrExplored) {
		t.Logf("explore: %v", err)
	}
	if len(clicks) < 5 {
		t.Fatalf("only %d clicks: the sweep did not run", len(clicks))
	}
	if yields == 0 {
		t.Fatal("the pack never stalled the walk: the scenario did not happen")
	}
	for i, c := range clicks {
		if !mouthClickOK(c.from, c.to, ex) {
			var b strings.Builder
			for _, x := range clicks[max(0, i-4):min(len(clicks), i+2)] {
				fmt.Fprintf(&b, "%v->%v ", x.from, x.to)
			}
			t.Fatalf("click %d from %v (%.1f) to %v (%.1f) breaks the invariant; around: %s", i, c.from, dist(c.from, ex), c.to, dist(c.to, ex), b.String())
		}
	}
}

// S: the DRILLS18 corner. A wall just south of the character at 3.2 from the exit: the replanned way out first runs along
// the wall, inside the bound. The gate rewrites or holds; no click breaks the invariant and the sweep ends.
func TestCorneredAtTheMouthNeverCrosses(t *testing.T) {
	g := openField(2000, 2000, 120, 120)
	blockRect(g, d2map.Rect{X: 2036, Y: 2056, W: 26, H: 2})
	s, n := newKnown(t, g, d18Swung, true)
	o, err := n.observe()
	if err != nil {
		t.Fatal(err)
	}
	o.lm.Exits = []d2map.Exit{denExit}
	n.stayLevel = o.level
	ex := mp(denExit.Pos)
	type click struct{ from, to input.Point }
	var clicks []click
	s.onClickWorld = func(s *sim, p input.Point) { clicks = append(clicks, click{s.pos(), p}) }
	_ = n.GoTo(ctxT(t), denAt(5, 33))
	for i, c := range clicks {
		if !mouthClickOK(c.from, c.to, ex) {
			t.Fatalf("click %d from %v (%.1f) to %v (%.1f) breaks the invariant", i, c.from, dist(c.from, ex), c.to, dist(c.to, ex))
		}
	}
	if len(clicks) == 0 {
		t.Fatal("no click at all")
	}
}
