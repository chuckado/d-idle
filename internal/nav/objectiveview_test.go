package nav

import (
	"testing"

	"d2lod-bot/internal/input"
)

// ObjectiveView is the goal of the walk in progress, per the Paladin and Sorceress walks of the Tower Cellar 5 fixture.
func TestObjectiveViewIsTheWalkGoal(t *testing.T) {
	for _, c := range cellar5Starts {
		_, n := cellar5Nav(t, c.start)
		if _, _, ok := n.ObjectiveView(); ok {
			t.Fatalf("%s: an objective with no walk goal", c.class)
		}
		n.goal, n.goalSet = cellar5Goal, true
		level, p, ok := n.ObjectiveView()
		if !ok || p != cellar5Goal || level != n.w.Current().No {
			t.Fatalf("%s: objective %v %v %v", c.class, level, p, ok)
		}
		n.goalSet = false
		if _, _, ok := n.ObjectiveView(); ok {
			t.Fatalf("%s: the objective outlived the walk", c.class)
		}
	}
}

func TestForgetCoverageRunsTheNewGameHook(t *testing.T) {
	_, n := cellar5Nav(t, input.Point{X: 124, Y: 74})
	calls := 0
	n.SetOnNewGame(func() { calls++ })
	n.ForgetCoverage()
	n.ForgetCoverage()
	if calls != 2 {
		t.Fatalf("hook ran %d times", calls)
	}
}
