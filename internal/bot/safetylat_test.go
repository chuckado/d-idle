package bot

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
)

// A safety verdict is stamped by the guard and closed when the input leaves: the latency is the whole wait, the time the
// running action took to notice the cancel included (Fable review P3 a).
func TestSafetyLatencyIsMeasuredFromVerdictToInput(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Barbarian} {
		for _, c := range []struct {
			name string
			a    SafetyAction
		}{{"drink", SafetyDrinkLife}, {"chicken", SafetyChicken}} {
			t.Run(class.String()+"/"+c.name, func(t *testing.T) {
				g := newRig(t, func(cfg *Config, d *Deps, tu *env.Tuning) { tu.Bot.ReactionMs = 0 })
				g.w.set(func(w *world) { w.state = d2.StateInGame; w.area = 2; w.life, w.maxL = 30, 100; w.class = class })
				g.fight.safety = func(*d2.Snapshot) SafetyAction { return c.a }
				ctx, cancel := context.WithCancelCause(context.Background())
				defer cancel(nil)
				if !g.r.guardOnce(ctx, cancel, StateClearing) {
					t.Fatal("the guard did not cancel")
				}
				var si safetyInterrupt
				if !errors.As(context.Cause(ctx), &si) {
					t.Fatal("no safety cause")
				}
				g.clock.Sleep(400 * time.Millisecond) // the running action takes this long to return
				g.r.handleSafety(context.Background(), si.Action)
				g.r.obMu.Lock()
				l := g.r.ob.sl
				kv := g.r.ob.sl.minuteKV()
				g.r.obMu.Unlock()
				if l.lastMs < 400 || l.lastMs > 500 || l.n != 1 || l.slow != 1 {
					t.Fatalf("latency %d ms n=%d slow=%d, want about 400 ms, one sample, slow", l.lastMs, l.n, l.slow)
				}
				if len(kv) == 0 || kv[0] != "safetyN" {
					t.Fatalf("minute line has no safety keys: %v", kv)
				}
				if again := (&safetyLat{}).minuteKV(); again != nil {
					t.Fatalf("a quiet minute must add nothing: %v", again)
				}
			})
		}
	}
}

func TestSafetyLatencyWithoutVerdictIsNotCounted(t *testing.T) {
	g := newRig(t)
	g.r.noteSent(SafetyDrinkLife, g.clock.Now(), false) // a hazard or paused-fight send has no guard verdict
	if g.r.ob.sl.n != 0 {
		t.Fatal("counted a send that had no verdict")
	}
}

func TestAutopsyNamesTheSafetyLatency(t *testing.T) {
	au := &Autopsy{State: "clearing", LastSafety: "chicken"}
	ms := 1840
	au.SafetyLatMs = &ms
	if s := au.summarize("Fallen", "Cold Plains"); !strings.Contains(s, "1840 ms after its verdict") {
		t.Fatalf("summary does not name the latency: %s", s)
	}
}
