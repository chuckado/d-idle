package bot

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/goals"
	"d2lod-bot/internal/task"
)

// The non-board quest runner: a quest that is not ready and a ban wait are typed refusals (planningBlock), handled once
// by do, never a failure; knob off: the old nil. Two classes. Must fail while planningBlock returns nil for them.
func TestPlanningBlockRefusalsAreTyped(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Barbarian} {
		notReady := task.Result{Outcome: task.OutBlocked, StepID: "andariel", Kind: task.KindDanger,
			Err: task.Errorf(task.KindDanger, "not ready: clvl 8 < 12")}
		banned := task.Result{Outcome: task.OutBlocked, StepID: "level.8", Kind: task.KindNotFound,
			Err: task.Errorf(task.KindNotFound, "no eligible reachable leveling area")}
		for _, c := range []struct {
			name, need string
			res        task.Result
			ban        bool
		}{{"not ready", "too_weak", notReady, false}, {"ban wait", "area_banned", banned, true}} {
			g := soakyRig(t, 1)
			g.w.set(func(w *world) { w.class = class })
			buf := &lockedBuf{}
			g.r.e.Log = slog.New(slog.NewTextHandler(buf, nil))
			g.r.Start(goals.Goal{ID: "x", Type: goals.Idle})
			if c.ban {
				g.r.avoid[2] = g.clock.Now().Add(20 * 60 * 1e9)
			}
			var ref *Refused
			err := g.r.planningBlock(context.Background(), c.res)
			if !errors.As(err, &ref) || ref.Need != c.need || ref.Wake == "" {
				t.Fatalf("%v %s: %v", class, c.name, err)
			}
			for i := 0; i < 10; i++ {
				g.r.do(context.Background(), g.r.control(), StateTraveling, "x", 0, "task step", func(context.Context) error {
					return g.r.planningBlock(context.Background(), c.res)
				})
			}
			if g.r.loop.failures != 0 || strings.Contains(buf.String(), "action failed") || strings.Count(buf.String(), "step refused") != 1 {
				t.Fatalf("%v %s: failures %d log:\n%s", class, c.name, g.r.loop.failures, buf.String())
			}
			g.r.e.Tune.Bot.TypedOutcome = 0
			if err := g.r.planningBlock(context.Background(), c.res); err != nil {
				t.Fatalf("%v %s knob off: %v", class, c.name, err)
			}
		}
	}
}
