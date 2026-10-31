package d2denv

import (
	"context"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2dctl"
)

// waits-audit: Install under lockstep puts a stepping Pause and the virtual Now into d2dctl; on a free-running d2d the
// pause is a wall sleep that sends no step and Now is the wall clock.
func TestInstallLockstepAndWall(t *testing.T) {
	oldP, oldN := d2dctl.Pause, d2dctl.Now
	t.Cleanup(func() { d2dctl.Pause, d2dctl.Now = oldP, oldN })
	srv, p := startFake(t) // lockstep fake (TestPauseStepsUnderLockstep)
	c := d2dctl.New(p, d2dctl.Options{})
	t.Cleanup(func() { c.Close() })
	steps := func() (n int) {
		for _, l := range srv.Lines() {
			if strings.HasPrefix(l, "step ") {
				n++
			}
		}
		return
	}
	if !NewLateClock(context.Background(), c).Install(context.Background()) {
		t.Fatal("fake d2d is not lockstep")
	}
	t0, before, w := d2dctl.Now(), steps(), time.Now()
	if err := d2dctl.Pause(context.Background(), 4*time.Second); err != nil {
		t.Fatal(err)
	}
	if steps() == before || time.Since(w) > time.Second {
		t.Fatalf("lockstep pause sent %d steps in %v", steps()-before, time.Since(w))
	}
	if dv := d2dctl.Now().Sub(t0); dv < 3900*time.Millisecond || dv > 4500*time.Millisecond {
		t.Fatalf("virtual Now advanced %v for a 4 s pause", dv)
	}

	// free-running: a LateClock that never found lockstep (Detect not called) is the wall clock
	wall := NewLateClock(context.Background(), c)
	d2dctl.Pause, d2dctl.Now = wall.Pause, wall.Now
	before, w = steps(), time.Now()
	if err := d2dctl.Pause(context.Background(), 120*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if steps() != before || time.Since(w) < 100*time.Millisecond {
		t.Fatalf("wall pause: %d steps, %v", steps()-before, time.Since(w))
	}
	if d := time.Since(d2dctl.Now()); d < -time.Second || d > time.Second {
		t.Fatalf("wall Now is %v off", d)
	}
}
