package main

import (
	"context"
	"testing"
	"time"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/d2dctl/fake"
	"d2lod-bot/internal/d2denv"
)

// TestTickKeepsPacerCurrent is the F17 regression: the lab stepped the game with Client.Step, which the Pacer (the
// navigator's clock) does not see, so its virtual time lagged by every tick the drills had stepped. The first Sleep of the
// next walk caught up all at once: the walk's 2 minute deadline (taken before) expired after a few seconds.
func TestTickKeepsPacerCurrent(t *testing.T) {
	path, cleanup, err := fake.TempSocket()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	srv, err := fake.New(path, fake.Sample(), fake.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	c := d2dctl.New(path, d2dctl.Options{Timeout: 5 * time.Second})
	defer c.Close()
	ctx := context.Background()
	clk := d2denv.NewLateClock(ctx, c)
	if !clk.Detect(ctx) {
		t.Fatal("the fake is not in lockstep")
	}
	l := &lab{be: &d2denv.Backend{Client: c, Clock: clk}}
	clk.Sleep(100 * time.Millisecond) // warm the Pacer as the bot does
	for i := 0; i < 6; i++ {
		if err := l.Tick(ctx, 500); err != nil { // 20 s of game time each, 2 minutes in all
			t.Fatal(err)
		}
	}
	deadline := clk.Now().Add(2 * time.Minute) // a walk's deadline, taken at its start
	clk.Sleep(100 * time.Millisecond)          // the walk's first poll
	if clk.Now().After(deadline) {
		t.Fatalf("the walk's deadline expired after one poll: the Pacer lagged the stepped game time (now %v, deadline %v)", clk.Now(), deadline)
	}
}

func TestSyncPacerReportsLag(t *testing.T) {
	path, cleanup, err := fake.TempSocket()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	srv, err := fake.New(path, fake.Sample(), fake.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	c := d2dctl.New(path, d2dctl.Options{Timeout: 5 * time.Second})
	defer c.Close()
	ctx := context.Background()
	clk := d2denv.NewLateClock(ctx, c)
	clk.Detect(ctx)
	l := &lab{be: &d2denv.Backend{Client: c, Clock: clk}}
	clk.Sleep(100 * time.Millisecond)
	if _, err := c.Step(ctx, 1000); err != nil { // a direct step the Pacer does not see: 40 s
		t.Fatal(err)
	}
	if lag := l.syncPacer(); lag < 39*time.Second || lag > 41*time.Second {
		t.Fatalf("lag %v, want about 40s", lag)
	}
	if lag := l.syncPacer(); lag != 0 {
		t.Fatalf("second sync lag %v, want 0", lag)
	}
}
