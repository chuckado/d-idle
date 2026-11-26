package combat

import (
	"context"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
)

// T6 (Fable review 2026-11-25 3.3 item 2): the whole safety design rests on a running action returning soon after its
// context is cancelled. These tests cancel the fighter's actions mid-run on the simulator's virtual clock and bound the
// virtual time to the return. They cover the combat package; the nav, town and loot actions are listed in
// docs/safety-latency.md as not yet covered (they need their own simulators).

// cancelClock cancels ctx once the virtual time passes at, and remembers when.
type cancelClock struct {
	env.Clock
	cancel   context.CancelFunc
	at       time.Time
	cancelAt time.Time
}

func (c *cancelClock) Sleep(d time.Duration) {
	c.Clock.Sleep(d)
	if c.cancelAt.IsZero() && !c.Clock.Now().Before(c.at) {
		c.cancelAt = c.Clock.Now()
		c.cancel()
	}
}

// cancelLatencyBound is N: the virtual time an action may take from a cancel to its return. Two guard ticks of the
// default bot.watchMs is 100 ms; the bound is generous because one fight iteration may hold a cast of up to a second.
const cancelLatencyBound = 1500 * time.Millisecond

func TestFighterActionsReturnSoonAfterCancel(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Barbarian} { // build flexibility: the bound is not a Sorceress property
		for _, act := range []struct {
			name string
			run  func(h *harness, ctx context.Context) error
		}{
			{"Clear", func(h *harness, ctx context.Context) error { _, err := h.f.Clear(ctx, ClearOpts{}); return err }},
			{"Kill", func(h *harness, ctx context.Context) error { return h.f.Kill(ctx, 10) }},
		} {
			t.Run(class.String()+"/"+act.name, func(t *testing.T) {
				h := newHarness(t)
				h.s.class = class
				h.s.stock(itHP1, 0, 2)
				h.s.addMon(10, clsFallen, 130, 100, 1e9, 0, 0) // a monster that never dies and never hurts: the action runs until cancelled
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				cc := &cancelClock{Clock: h.s, cancel: cancel, at: h.s.Now().Add(5 * time.Second)}
				h.f.e.Clock = cc
				done := make(chan struct{})
				go func() { defer close(done); act.run(h, ctx) }()
				select {
				case <-done:
				case <-time.After(30 * time.Second): // a hang is the failure this test exists for; the bound itself is virtual
					t.Fatalf("%s did not return after the cancel", act.name)
				}
				if cc.cancelAt.IsZero() {
					t.Fatal("the action ended before the cancel: the scenario does not hold it")
				}
				if late := h.s.Now().Sub(cc.cancelAt); late > cancelLatencyBound {
					t.Fatalf("%s returned %v of virtual time after the cancel, bound %v", act.name, late, cancelLatencyBound)
				}
			})
		}
	}
}
