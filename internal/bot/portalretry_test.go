package bot

import (
	"context"
	"fmt"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/town"
)

func TestPortalRetryDue(t *testing.T) {
	cut := fmt.Errorf("d2d portal: cancelled, nothing clicked: %w", town.ErrPortalCancelled)
	for _, c := range []struct {
		name    string
		err     error
		a       SafetyAction
		pending bool
		used    bool
		wait    int
		want    bool
	}{
		{"retreat cut the cast", cut, SafetyCrowdRetreat, false, false, 3, true},
		{"drink cut the cast", cut, SafetyDrinkLife, false, false, 3, true},
		{"second cut is not retried", cut, SafetyCrowdRetreat, false, true, 3, false},
		{"a pending trip retries itself", cut, SafetyCrowdRetreat, true, false, 3, false},
		{"chicken has its own way out", cut, SafetyChicken, false, false, 3, false},
		{"knob 0 is off", cut, SafetyCrowdRetreat, false, false, 0, false},
		{"another error", fmt.Errorf("walk failed"), SafetyCrowdRetreat, false, false, 3, false},
	} {
		if got := portalRetryDue(c.err, c.a, c.pending, c.used, c.wait); got != c.want {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
}

// A cast cancelled before the click by a crowd retreat queues one town trip from the same area, once (potion trip of the
// Search for Cain chain, docs/probes/provingground-pgfix4-2026-11-20.md); with the knob at 0 nothing is queued.
func TestPortalCancelRetry(t *testing.T) {
	for _, cl := range []d2.Class{d2.Paladin, d2.Sorceress} {
		for _, wait := range []int{3, 0} {
			t.Run(fmt.Sprintf("%v wait %d", cl, wait), func(t *testing.T) {
				g := newRig(t)
				g.r.e.Tune.Bot.PortalRetryWaitSec = wait
				g.w.set(func(w *world) { w.class = cl; w.state = d2.StateInGame; w.area = 38; w.life, w.maxL = 60, 100 })
				g.r.poll()
				cast := func(ctx context.Context) error {
					g.r.mu.Lock()
					cancel := g.r.cancelAct
					g.r.mu.Unlock()
					cancel(safetyInterrupt{SafetyCrowdRetreat})
					return fmt.Errorf("to town: %w", town.ErrPortalCancelled)
				}
				run := func() bool {
					return g.r.do(context.Background(), g.r.control(), StateResupplying, "to town", 0, "to town", cast)
				}
				if run() {
					t.Fatal("a cancelled cast reported success")
				}
				if got := g.r.loop.trip != nil; got != (wait > 0) {
					t.Fatalf("trip queued %v with the knob at %d", got, wait)
				}
				if wait == 0 {
					return
				}
				if g.r.loop.trip.returnTo != 38 || g.r.loop.trip.emergency {
					t.Fatalf("trip %+v, want a plain trip back to area 38", g.r.loop.trip)
				}
				g.r.loop.trip = nil
				run()
				if g.r.loop.trip != nil {
					t.Fatal("the second cancelled cast was retried again")
				}
			})
		}
	}
}
