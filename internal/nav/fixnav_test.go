package nav

import (
	"testing"
	"time"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/input"
)

// FIXNAV: a step-away that ended in another level (the trigger fired) turns the direction on the next try.
func TestMouthStepTurnsAfterABounce(t *testing.T) {
	_, n := newKnown(t, openField(2000, 2000, 120, 120), input.Point{X: 2050, Y: 2060}, true)
	o, err := n.observe()
	if err != nil {
		t.Fatal(err)
	}
	o.lm.Exits = []d2map.Exit{{Dest: 8, Pos: d2map.Point{X: 2050, Y: 2056}}} // the warp is north: away is south
	first, ok := n.mouthStep(o)
	if !ok || first.Y <= o.pos.Y+10 || abs(first.X-o.pos.X) > 3 {
		t.Fatalf("first step %v %v: want straight south", first, ok)
	}
	n.noteMouthBounce()
	second, ok := n.mouthStep(o)
	if !ok || second == first || abs(second.X-o.pos.X) < 10 {
		t.Fatalf("second step %v after a bounce: want a sideways turn, first was %v", second, first)
	}
	n.noteMouthBounce()
	third, _ := n.mouthStep(o)
	if third == first || third == second {
		t.Fatalf("third step %v repeats an earlier one", third)
	}
	n.clk = &shiftClock{n.clk, mouthFailTTL + time.Second}
	if again, _ := n.mouthStep(o); again != first {
		t.Fatalf("after the memory expired the step is %v, want %v", again, first)
	}
}

func TestKnownWaypointRelTowns(t *testing.T) {
	for lvl, want := range map[int]input.Point{75: {X: 159, Y: 49}, 103: {X: 44, Y: 19}, 109: {X: 114, Y: 69}} {
		if got := knownWaypointRel[lvl]; got != want {
			t.Errorf("level %d: %v, want %v", lvl, got, want)
		}
	}
	if _, ok := knownWaypointPos[75]; ok {
		t.Error("the act-absolute table must stay Game.exe-only for these towns")
	}
}

type shiftClock struct {
	env.Clock
	d time.Duration
}

func (c *shiftClock) Now() time.Time { return c.Clock.Now().Add(c.d) }
