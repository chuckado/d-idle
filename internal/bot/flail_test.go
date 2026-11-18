package bot

import (
	"context"
	"testing"
	"time"
)

// flailFixture is the cycle from the sandbox log of 2026-11-18 18:13:40 to 18:14:20 (seconds since the first line,
// area id): Dark Wood 5, Rogue Encampment 1, Cold Plains 3. V: copied from the d2play docker log, "Entered ..." lines.
var flailFixture = []struct {
	dt   int
	area int
}{
	{0, 5}, {2, 1}, {1, 5}, {2, 3}, {2, 5}, {2, 1}, {1, 5}, {2, 3}, {2, 5}, {2, 1}, {1, 5}, {2, 3}, {2, 5}, {2, 1}, {1, 5},
}

func feedFlailFixture(g *rig) {
	for _, m := range flailFixture {
		g.clock.Sleep(time.Duration(m.dt) * time.Second)
		g.r.loop.flail.noteMove(g.clock.Now(), m.area, 0)
	}
}

func TestFlailHoldsInsteadOfGoingHome(t *testing.T) {
	g := newRig(t)
	g.w.set(func(w *world) { w.area = 5 })
	feedFlailFixture(g)
	g.r.poll()
	g.r.flailHold(context.Background(), g.r.latest(), levelGoalControl())
	if g.r.loop.flail.holdUntil.IsZero() {
		t.Fatal("the observed cycle did not start a hold")
	}
	if n := g.w.count("totown"); n != 0 {
		t.Fatalf("flailing bot went home %d times", n)
	}
	// still held a minute later, free after the hold
	g.clock.Sleep(time.Minute)
	g.r.poll()
	if _, _, holding := g.r.loop.flail.check(g.clock.Now(), 0, 6, 10*time.Minute, 5*time.Minute, g.r.areaName); !holding {
		t.Fatal("hold ended early")
	}
	g.clock.Sleep(5 * time.Minute)
	if _, _, holding := g.r.loop.flail.check(g.clock.Now(), 0, 6, 10*time.Minute, 5*time.Minute, g.r.areaName); holding {
		t.Fatal("hold never ended")
	}
}

// The control: with the guard off the same cycle sends the hero home (the loop of the log).
func TestFlailGuardOffGoesHome(t *testing.T) {
	g := newRig(t)
	g.r.e.Tune.Bot.FlailMoves = 0
	g.w.set(func(w *world) { w.area = 5 })
	feedFlailFixture(g)
	g.r.poll()
	g.r.noTarget(context.Background(), g.r.latest(), levelGoalControl())
	if n := g.w.count("totown"); n != 1 {
		t.Fatalf("control: %d trips", n)
	}
}

func TestFlailNeedsNoProgress(t *testing.T) {
	var g flailGuard
	now := time.Unix(1000, 0)
	nm := func(a int) string { return "a" }
	for i := 0; i < 8; i++ {
		g.noteMove(now.Add(time.Duration(i)*time.Second), i%2, uint64(i)) // a kill between every move
	}
	if _, started, _ := g.check(now.Add(9*time.Second), 8, 6, 10*time.Minute, 5*time.Minute, nm); started {
		t.Fatal("moves with kills in between are not flailing")
	}
	// slow moves outside the window are not flailing either
	var h flailGuard
	for i := 0; i < 8; i++ {
		h.noteMove(now.Add(time.Duration(i)*3*time.Minute), i%2, 0)
	}
	if _, started, _ := h.check(now.Add(24*time.Minute), 0, 6, 10*time.Minute, 5*time.Minute, nm); started {
		t.Fatal("slow moves are not flailing")
	}
}
