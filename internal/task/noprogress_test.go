package task_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/task"
	"d2lod-bot/internal/task/sim"
)

// V lab 2026-11-23 (necro.act1-tools): "to-barracks ran" 125 times in 28 minutes. Each time the step finishes (the hero is
// in the Barracks) and a trip or a retreat undoes it, so no failure counter ever moved and the chunk never parked. The
// executor counts "ran" of one planned step with no other planned step running in between and parks the group at
// Policy.MaxRepeats with "no progress". Two classes (the rule is class generic), and a knob that moves the limit.
func TestRepeatedRanWithoutProgressParksTheGroup(t *testing.T) {
	for _, caster := range []bool{true, false} { // labels only: the executor never reads the class
		for _, limit := range []int{0, 5} { // 0 = the default 12
			g := sim.New(true)
			ex := task.NewExecutor(g, g)
			ex.Now = func() time.Time { return time.Unix(1000, 0) }
			ex.Policy.MaxRepeats = limit
			want := 12
			if limit > 0 {
				want = limit
			}
			at := false
			plan := fixedPlanner{task.Plan{Group: "tools", Nodes: []task.Node{{Step: task.Step{
				ID: "quest.tools/to-barracks", Budget: time.Second,
				Need: func(w task.World) (bool, task.Evidence) { return !at, task.EvMemory },
				Do:   func(context.Context, task.Actions) error { at = true; return nil },
			}}}}}
			var last task.Result
			n := 0
			for ; n < 3*want; n++ {
				at = false // a trip home or a retreat undid it
				r, err := ex.Tick(context.Background(), plan)
				if err != nil {
					t.Fatal(err)
				}
				last = r
				if r.Outcome == task.OutBlocked {
					break
				}
			}
			if last.Outcome != task.OutBlocked || n+1 != want || !strings.Contains(last.Attention, "no progress") {
				t.Errorf("caster %v limit %d: outcome %v after %d ticks, attention %q (want blocked at %d)", caster, limit, last.Outcome, n+1, last.Attention, want)
			}
			if why, ok := ex.Blocked("tools"); !ok || !strings.Contains(why, "no progress") {
				t.Errorf("caster %v limit %d: group not blocked with the reason: %q %v", caster, limit, why, ok)
			}
		}
	}
}
