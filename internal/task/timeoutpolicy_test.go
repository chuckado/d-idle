package task_test

import (
	"context"
	"testing"
	"time"

	"d2lod-bot/internal/task"
	"d2lod-bot/internal/task/sim"
)

// A step that keeps timing out while the exploration climbs gets one extension, then is given up with a reason; one that
// explored nothing is given up at once. Two classes: the rule is class-free, the walk only differs in area.
func TestStepTimeoutPolicyExtendsOnceThenParks(t *testing.T) {
	for _, tc := range []struct {
		class string
		area  task.Area
	}{{"Paladin", task.DarkWood}, {"Sorceress", task.ColdPlains}} {
		for _, progress := range []bool{true, false} {
			t.Run(tc.class, func(t *testing.T) {
				g := sim.New(true)
				ex := task.NewExecutor(g, g)
				pct, runs := 4, 0
				ex.Hooks.Progress = func() (int, bool) { return pct, true }
				var seen []int
				ex.Hooks.OnTimeout = func(_ task.Step, n, p int, _ bool) { seen = append(seen, n) }
				to := task.Wrap(task.KindTimeout, context.DeadlineExceeded)
				step := task.StepWaypoint("wp-test", tc.area, func(task.World) bool { return false })
				step.Budget = time.Second
				step.Do = func(context.Context, task.Actions) error {
					runs++
					if progress {
						pct += 10
					}
					return to
				}
				p := fixedPlanner{task.Plan{Group: "x", Nodes: []task.Node{{Step: step}}}}
				var last task.Result
				for i := 0; i < 6; i++ {
					last, _ = ex.Tick(context.Background(), p)
					if last.Outcome == task.OutSkipped {
						break
					}
				}
				want := 2
				if !progress {
					want = 1
				}
				if runs != want || last.Outcome != task.OutSkipped || !ex.Skipped("wp-test") {
					t.Fatalf("runs %d (want %d) outcome %v", runs, want, last.Outcome)
				}
				if last.Attention == "" {
					t.Fatalf("no visible reason: %+v", last)
				}
				if len(seen) != want {
					t.Fatalf("OnTimeout calls %v", seen)
				}
			})
		}
	}
}
