package nav

import (
	"testing"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/input"
)

// Sewers Level 2 on d2d (V, scenario drill.act2.radament Paladin 11, nav.warpRead and fullGrid on, job 20261123-005221):
// the hero arrives at (262,134), 11 subtiles from the stairs back to level 47 at (255,125); the stairs down to level 49
// lie at (135,25), far west. Every click of the walk toward them starts within the 15 subtile bound of the arrival stairs
// and the gate held for good ("click {259 138}: no path: the click would cross the trigger area of an unwanted exit",
// 100 holds, the board parked). The walk to a gate has the gate as its goal, like a GoTo: a click that stays outside the
// 10 subtile core goes through. The geometry knows no class; the two heroes only differ in where the walk starts.
func TestExitWalkIsAWalkWithAGoal(t *testing.T) {
	back := d2map.Exit{Dest: 47, Pos: d2map.Point{X: 255, Y: 125}}
	down := d2map.Exit{Dest: 49, Pos: d2map.Point{X: 135, Y: 25}}
	for _, c := range []struct {
		class string
		start input.Point
	}{{"Paladin", input.Point{X: 264, Y: 137}}, {"Sorceress", input.Point{X: 263, Y: 138}}} {
		t.Run(c.class, func(t *testing.T) {
			s, n := newKnown(t, openField(100, 0, 220, 200), c.start, true)
			o, err := n.observe()
			if err != nil {
				t.Fatal(err)
			}
			o.lm.Exits = []d2map.Exit{back, down}
			n.stayLevel = o.level
			n.wantExit = down.Dest
			wo := exitWalkOpts(gate{at: mp(down.Pos), push: mp(down.Pos)})
			if wo.goal == nil || *wo.goal != mp(down.Pos) {
				t.Fatalf("%s: the exit walk carries no goal: %+v", c.class, wo)
			}
			click := input.Point{X: 259, Y: 138}
			if err := n.in.ClickWorld(click); err == nil {
				t.Fatalf("%s: without a goal the click went through (the log shows the gate holding it)", c.class)
			}
			n.goal, n.goalSet = *wo.goal, true
			s.clicks = nil
			_, exits := n.gatedExits()
			if clickOKAll(c.start, click, exits) {
				t.Fatalf("%s: test geometry: the click must fail the 15 subtile bound", c.class)
			}
			if err := n.in.ClickWorld(click); err != nil || len(s.clicks) != 1 || s.clicks[0] != click {
				t.Fatalf("%s: the click toward the stairs was refused: %v %v", c.class, err, s.clicks)
			}
		})
	}
}

// After the rewrite cap the click of a walk with a goal goes through when it stays outside the core (the Sewers case: the
// rewrites shuttled the hero, then the cap held every click for good); without a goal the cap still holds.
func TestRewriteCapDoesNotHoldAWalkWithAGoalOutsideTheCore(t *testing.T) {
	back := d2map.Exit{Dest: 47, Pos: d2map.Point{X: 255, Y: 125}}
	for _, c := range []struct {
		class string
		start input.Point
	}{{"Paladin", input.Point{X: 264, Y: 137}}, {"Sorceress", input.Point{X: 263, Y: 138}}} {
		s, n := newKnown(t, openField(100, 0, 220, 200), c.start, true)
		o, err := n.observe()
		if err != nil {
			t.Fatal(err)
		}
		o.lm.Exits = []d2map.Exit{back}
		n.stayLevel = o.level
		click := input.Point{X: 259, Y: 138}
		n.gateRewrites = gateMaxRewrites
		if err := n.in.ClickWorld(click); err == nil {
			t.Fatalf("%s: without a goal the capped gate must hold", c.class)
		}
		n.goal, n.goalSet = input.Point{X: 135, Y: 25}, true
		s.clicks = nil
		n.gateRewrites = gateMaxRewrites
		if err := n.in.ClickWorld(click); err != nil || len(s.clicks) != 1 {
			t.Fatalf("%s: the capped gate held a click with a goal beyond: %v %v", c.class, err, s.clicks)
		}
		n.gateRewrites = gateMaxRewrites
		if err := n.in.ClickWorld(input.Point{X: 256, Y: 129}); err == nil {
			t.Fatalf("%s: a click inside the core went through", c.class)
		}
	}
}

// The door walk and the strip walk of an exit leg carry their target as the goal too (see goalOpts).
func TestGoalOptsCarryTheTarget(t *testing.T) {
	for _, p := range []input.Point{{X: 121, Y: 180}, {X: 20, Y: 25}} {
		for _, pursuit := range []bool{true, false} {
			wo := goalOpts(p, pursuit)
			if wo.goal == nil || *wo.goal != p || wo.pursuit != pursuit {
				t.Errorf("goalOpts(%v, %v) = %+v", p, pursuit, wo)
			}
		}
	}
}
