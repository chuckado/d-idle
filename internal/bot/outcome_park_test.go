package bot

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"d2lod-bot/internal/goals"
	"d2lod-bot/internal/task"
)

// A parked quest step and a held journey step are typed refusals handled once: one refusal line per change, no spin
// guard warnings, no failure. Two classes (staleClasses); knob off keeps the spin guard. Must fail while the holds
// pause through spinGuard.
func TestJourneyHoldAndParkAreTypedRefusals(t *testing.T) {
	for _, tc := range staleClasses {
		t.Run(tc.name, func(t *testing.T) {
			g := richRig(t, tc.class, 8)
			buf := &lockedBuf{}
			g.r.e.Log = slog.New(slog.NewTextHandler(buf, nil))
			g.r.Start(goals.Goal{ID: "x", Type: goals.Idle})
			ctx := context.Background()
			c := g.r.control()
			c.spec = levelGoalControl().spec
			ex := task.NewExecutor(taskEnv{g.r}, taskActs{g.r, c})
			why := "area not allowed now: Blood Moor (2): skipped"
			for i := 0; i < 12; i++ {
				if !g.r.journeyBlocked(ctx, g.r.latest(), c, ex, why) {
					t.Fatal("the hold must take the tick")
				}
			}
			out := buf.String()
			if strings.Count(out, "step refused") != 1 || strings.Contains(out, "backing off") || g.r.loop.failures != 0 {
				t.Fatalf("log:\n%s", out)
			}
			// the park of a not-found step travels as the typed refusal through planningBlock
			g.r.loop.task.ex = ex
			g.r.loop.task.stepTitle = "Take the bark"
			res := notFound("quest.search_for_cain/tree")
			g.r.planningBlock(ctx, res)
			var ref *Refused
			if err := g.r.planningBlock(ctx, res); !errors.As(err, &ref) || ref.Need != "parked" || ref.Wake == "" {
				t.Fatalf("park: %v", err)
			}
		})
	}
}
