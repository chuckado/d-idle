package bot

import (
	"context"
	"sync"
	"testing"
	"time"

	"d2lod-bot/internal/boss"
	"d2lod-bot/internal/d2"
)

// A time-boxed farm counts its elapsed time on the env clock, not the wall (card pgclock, design B2): the first Farm call
// "takes" 90 virtual seconds (the fake clock steps, the wall does not move), so the resume asks for 10 minutes less 90 s.
// With the wall clock the resume asked for the whole 10 minutes, whatever the machine load made of it.
func TestFarmTimeBoxCountsEnvClockTime(t *testing.T) {
	for _, cls := range []d2.Class{d2.Sorceress, d2.Paladin} { // the default Sorceress and one other class
		g := newRig(t)
		g.w.class = cls
		var mu sync.Mutex
		var calls []FarmCall
		g.r.d.Task = &TaskDeps{Farm: func(ctx context.Context, c FarmCall) (boss.FarmReport, error) {
			mu.Lock()
			defer mu.Unlock()
			calls = append(calls, c)
			if len(calls) == 1 {
				g.clock.Sleep(90 * time.Second)
			}
			return boss.FarmReport{Stop: "interrupted"}, nil
		}}
		g.run()
		goal := runsGoal("Pindleskin", 0)
		goal.Params.UntilMins = 10
		if err := g.r.Start(goal); err != nil {
			t.Fatal(err)
		}
		g.waitFor("second farm call", func() bool { mu.Lock(); defer mu.Unlock(); return len(calls) >= 2 })
		mu.Lock()
		if calls[0].Duration != 10*time.Minute {
			t.Errorf("class %v: first call asks %v, want the whole 10m", cls, calls[0].Duration)
		}
		if want := 10*time.Minute - 90*time.Second; calls[1].Duration != want {
			t.Errorf("class %v: resume asks %v, want %v (90 virtual seconds were spent)", cls, calls[1].Duration, want)
		}
		mu.Unlock()
		g.finish()
	}
}
