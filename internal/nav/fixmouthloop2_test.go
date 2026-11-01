package nav

import (
	"errors"
	"testing"

	"d2lod-bot/internal/input"
)

// heldClicks fails every world click the way the mouth gate does after its rewrite cap.
type heldClicks struct {
	input.Controller
	clicks int
}

func (h *heldClicks) ClickWorld(input.Point) error {
	h.clicks++
	return ErrMouthGate
}

// FIXMOUTHLOOP2 (V DRILLS20): a step-away whose clicks are held must end the outer walk with ErrMouthGate after one
// attempt, not re-fire the step-away until the walk timeout.
func TestHeldStepAwayEndsTheWalkOnce(t *testing.T) {
	_, n, _ := gateNav(t, denAt(2, 6)) // inside the bound of the unwanted exit
	h := &heldClicks{Controller: n.in}
	n.in = h
	err := n.GoTo(ctxT(t), denAt(5, 33))
	t.Logf("held clicks %d, err %v", h.clicks, err)
	if !errors.Is(err, ErrMouthGate) {
		t.Fatalf("err %v, want ErrMouthGate", err)
	}
	if h.clicks != 1 {
		t.Fatalf("%d clicks: the step-away was not a single attempt", h.clicks)
	}
}

// A step-away that works still ends the walk normally and reaches the goal.
func TestStepAwayStillCompletes(t *testing.T) {
	s, n, _ := gateNav(t, denAt(2, 6))
	if err := n.GoTo(ctxT(t), denAt(5, 33)); err != nil {
		t.Fatal(err)
	}
	if dist(s.pos(), denAt(5, 33)) > 5 {
		t.Fatalf("ended at %v", s.pos())
	}
}

// A wanted exit gets no step-away at all.
func TestWantedExitNoStepAway(t *testing.T) {
	_, n, _ := gateNav(t, denAt(2, 6))
	n.wantExit = denExit.Dest
	h := &heldClicks{Controller: n.in}
	n.in = h
	_ = n.GoTo(ctxT(t), denAt(5, 33))
	if h.clicks > 1 { // the walk's own click only
		t.Fatalf("%d clicks for a wanted exit", h.clicks)
	}
}
