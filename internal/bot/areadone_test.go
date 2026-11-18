package bot

import (
	"context"
	"testing"
	"time"
)

// Two "area done" decisions in the field within areaDoneGap send the hero home once, not every time (the Dark Wood
// ping-pong: a trip, and a town portal charge, about every 14 s for an hour).
func TestAreaDoneTripIsRateLimited(t *testing.T) {
	g := campaignRig(t, &fakeQuests{}, 8)
	g.w.set(func(w *world) { w.area = 5 })
	g.r.poll()
	c := levelGoalControl()
	g.r.noTarget(context.Background(), g.r.latest(), c)
	if n := g.w.count("totown"); n != 1 {
		t.Fatalf("first area done: %d trips", n)
	}
	g.w.set(func(w *world) { w.area = 5 })
	g.clock.Sleep(14 * time.Second)
	g.r.poll()
	g.r.noTarget(context.Background(), g.r.latest(), c)
	if n := g.w.count("totown"); n != 1 {
		t.Fatalf("second area done within the gap went home again: %d trips", n)
	}
	g.w.set(func(w *world) { w.area = 5 })
	g.clock.Sleep(areaDoneGap)
	g.r.poll()
	g.r.noTarget(context.Background(), g.r.latest(), c)
	if n := g.w.count("totown"); n != 2 {
		t.Fatalf("after the gap: %d trips", n)
	}
}
