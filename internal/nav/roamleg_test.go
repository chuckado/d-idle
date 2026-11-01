package nav

import (
	"errors"
	"testing"

	"d2lod-bot/internal/input"
)

// failClicks fails every world click with err, like the mouth gate (ErrMouthGate) or an unreachable goal (ErrNoPath).
type failClicks struct {
	input.Controller
	clicks int
	err    error
}

func (f *failClicks) ClickWorld(input.Point) error {
	f.clicks++
	return f.err
}

func roamWith(t *testing.T, err error) (*failClicks, *Navigator, error) {
	t.Helper()
	_, n := newKnown(t, openField(2000, 2000, 120, 120), input.Point{X: 2010, Y: 2010}, true)
	f := &failClicks{Controller: n.in, err: err}
	n.in = f
	return f, n, n.Explore(ctxT(t), nil)
}

func failedTargets(n *Navigator) int {
	k := 0
	for _, b := range n.cov.failed {
		if b {
			k++
		}
	}
	return k
}

// ROAMLEG (V DRILLS21: 10 failed legs per episode). A held leg ends the episode after one leg and keeps the target.
func TestRoamMouthHoldEndsEpisodeWithoutFailingTargets(t *testing.T) {
	f, n, err := roamWith(t, ErrMouthGate)
	if !errors.Is(err, ErrRoamHeld) || !errors.Is(err, ErrExplored) {
		t.Fatalf("err %v, want ErrRoamHeld (matching ErrExplored)", err)
	}
	if n.cov.legs != 1 || n.RoamHeld() != 1 || failedTargets(n) != 0 {
		t.Fatalf("legs %d held %d failed %d, want 1 1 0 (clicks %d)", n.cov.legs, n.RoamHeld(), failedTargets(n), f.clicks)
	}
	// the next episode starts from the same coverage and tries the same target again
	if err := n.Explore(ctxT(t), nil); !errors.Is(err, ErrRoamHeld) || n.cov.legs != 2 || failedTargets(n) != 0 {
		t.Fatalf("second episode: err %v legs %d failed %d", err, n.cov.legs, failedTargets(n))
	}
}

// Other leg failures behave as before: the target is marked failed and the roam moves on until nothing is left.
func TestRoamNoPathStillFailsTheTarget(t *testing.T) {
	_, n, err := roamWith(t, ErrNoPath)
	if !errors.Is(err, ErrExplored) || errors.Is(err, ErrRoamHeld) {
		t.Fatalf("err %v, want plain ErrExplored", err)
	}
	if n.RoamHeld() != 0 || failedTargets(n) < 2 || n.cov.legs < 2 {
		t.Fatalf("held %d failed %d legs %d: a no-path leg must mark its target failed", n.RoamHeld(), failedTargets(n), n.cov.legs)
	}
	t.Logf("no-path roam: %d legs, %d failed targets (the old mouth-gate behaviour)", n.cov.legs, failedTargets(n))
}
