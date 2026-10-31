package d2dctl

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"
)

// waits-audit: the NPC approach deadline is read from Now, so under lockstep it is game time. A fake clock that only the
// pause advances stands in for the Pacer: a menu that never opens ends the approach after about the timeout of virtual
// time, with no real sleeping.
func TestApproachNPCDeadlineIsGameTime(t *testing.T) {
	oldPause, oldNow := Pause, Now
	t.Cleanup(func() { Pause, Now = oldPause, oldNow })
	virt := time.Unix(0, 0)
	pauses := 0
	Pause = func(ctx context.Context, d time.Duration) error { virt = virt.Add(d); pauses++; return ctx.Err() }
	Now = func() time.Time { return virt }
	_, c := startStub(t, map[string]string{
		"npcs":        "Kashya\t64\t-148\t1\t7\t22.9\t11.5\nok\n",
		"debug level": "ok 1 30.00 12.00 80 80 0 0\n",
	}, nil)
	w := time.Now()
	err := c.ApproachNPC(context.Background(), 7, 6*time.Second)
	if err == nil || !strings.Contains(err.Error(), "not reached") {
		t.Fatalf("want a timeout, got %v", err)
	}
	el := virt.Sub(time.Unix(0, 0))
	if el < 6*time.Second || el > 20*time.Second || pauses == 0 {
		t.Fatalf("virtual time %v after %d pauses: the deadline must end within one round of 6 s", el, pauses)
	}
	if time.Since(w) > 5*time.Second {
		t.Fatalf("took %v of wall time", time.Since(w))
	}
}

// Real time: Now is the wall clock unless a lockstep run installs another.
func TestNowDefaultsToWall(t *testing.T) {
	if d := time.Since(Now()); d < -time.Second || d > time.Second {
		t.Fatalf("default Now is %v off the wall clock", d)
	}
}

// DRILLS5 H.pause: a long camp walk (30 s of game time, the player moving the whole time) must not trip the 15 s base
// timeout: the budget grows with the distance to the NPC.
func TestApproachNPCBudgetGrowsWithDistance(t *testing.T) {
	oldPause, oldNow := Pause, Now
	t.Cleanup(func() { Pause, Now = oldPause, oldNow })
	virt := time.Unix(0, 0)
	Pause = func(ctx context.Context, d time.Duration) error { virt = virt.Add(d); return ctx.Err() }
	Now = func() time.Time { return virt }
	_, c := startStub(t, map[string]string{"npcs": "Kashya\t64\t-148\t1\t7\t50.0\t20.0\nok\n"}, func(line string, s *stub) {
		el := virt.Sub(time.Unix(0, 0)).Seconds()
		s.rep["debug level"] = "ok 1 " + itoa(int(20+el)) + ".00 20.00 80 80 0 0\n"
		if el >= 30 {
			s.rep["menu"] = "Kashya\t265\t39\nbounty\t265\t72\nok\n"
		}
	})
	if err := c.ApproachNPC(context.Background(), 7, 15*time.Second); err != nil {
		t.Fatalf("a 30 s walk of a 30-cell approach failed: %v", err)
	}
}

// A player that stopped at an object is side-stepped after about a second, not after the whole 3 s round.
func TestApproachNPCStallDetectedEarly(t *testing.T) {
	oldPause, oldNow := Pause, Now
	t.Cleanup(func() { Pause, Now = oldPause, oldNow })
	virt := time.Unix(0, 0)
	moveAt := time.Duration(-1)
	Pause = func(ctx context.Context, d time.Duration) error { virt = virt.Add(d); return ctx.Err() }
	Now = func() time.Time { return virt }
	x := 25.5
	_, c := startStub(t, map[string]string{"npcs": "Kashya\t64\t-148\t1\t7\t40.0\t22.5\nok\n"}, func(line string, s *stub) {
		if strings.HasPrefix(line, "cmd move") {
			moveAt = virt.Sub(time.Unix(0, 0))
			x = 30
		}
		s.rep["debug level"] = "ok 1 " + strconv.FormatFloat(x, 'f', 2, 64) + " 22.50 80 80 0 0\n"
		if x > 29 && line == "cmd interact 7" {
			s.rep["menu"] = "Kashya\t265\t39\nbounty\t265\t72\nok\n"
		}
	})
	if err := c.ApproachNPC(context.Background(), 7, 15*time.Second); err != nil {
		t.Fatal(err)
	}
	if moveAt < 0 || moveAt > 2*time.Second {
		t.Fatalf("side step at %v of game time, want within 2 s", moveAt)
	}
}
