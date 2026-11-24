package nav

import (
	"testing"
	"time"
)

// V live 2026-11-24 20:28Z, Catacombs 4: the way to Andariel passes within the stay radius of the stairs up; every fight
// slice began a walk there, stepped 20 tiles away and walked back, 30 times in three minutes with no attack. After two
// step-aways from one exit inside the window the walk goes on (the click gate still guards the trigger area); after the
// window the step-away is back. The rule reads no class, the geometry is the Den exit of the other mouth tests.
func TestMouthStepAwayLoopEnds(t *testing.T) {
	s, n, o := gateNav(t, denAt(5, 10))
	n.arrival = true
	_ = s
	for i := 0; i < mouthLoopMax; i++ {
		if _, ok := n.mouthStep(o); !ok {
			t.Fatalf("step-away %d refused", i+1)
		}
	}
	o.pos = denAt(8, 12) // back by the exit, 14 away: inside the stay radius
	if _, ok := n.mouthStep(o); ok {
		t.Fatal("a third step-away inside the window: the loop goes on")
	}
	n.clk.Sleep(mouthLoopWindow + time.Second)
	if _, ok := n.mouthStep(o); !ok {
		t.Fatal("after the window the step-away is back")
	}
}

// A bounce (the step fired the trigger) keeps its step-aways: the turned direction is not a loop.
func TestMouthStepAwayAfterABounceIsNotALoop(t *testing.T) {
	_, n, o := gateNav(t, denAt(5, 10))
	n.arrival = true
	for i := 0; i < mouthLoopMax; i++ {
		n.mouthStep(o)
	}
	n.noteMouthBounce()
	if _, ok := n.mouthStep(o); !ok {
		t.Fatal("a bounced exit must still step away")
	}
}
