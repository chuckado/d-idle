package bot

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Leaving an emptied area (KILLFIX): off by default, on after EmptyAreaSec without a kill, restarted by a kill, by
// a pause longer than 90 s and by another area; a quest area is never left this way.
func TestEmptyAreaLeave(t *testing.T) {
	g := newRig(t)
	r := g.r
	r.loop.cleared = map[int]bool{}
	if r.emptyAreaLeave(2) || r.loop.cleared[2] {
		t.Fatal("off by default")
	}
	r.cfg.EmptyAreaSec = 100
	step := func(d time.Duration) bool { g.clock.Sleep(d); return r.emptyAreaLeave(2) }
	if step(0) || step(50*time.Second) {
		t.Fatal("too early")
	}
	r.obMu.Lock()
	r.ob.kills++
	r.obMu.Unlock()
	if step(40*time.Second) || step(50*time.Second) || step(40*time.Second) { // kill restarted the count: 90 s since
		t.Fatal("a kill must restart the count")
	}
	if step(time.Hour) { // a long pause (town trip, new game) restarts it too
		t.Fatal("a pause must restart the count")
	}
	if step(50*time.Second) || !step(55*time.Second) || !r.loop.cleared[2] {
		t.Fatalf("105 s without a kill must clear the area: %v", r.loop.cleared)
	}
	r.loop.cleared = map[int]bool{}
	r.loop.emptySince = time.Time{}
	run := LevelRun{Area: 2}
	r.loop.run = &run
	step(0)
	if step(200 * time.Second) {
		t.Fatal("a quest area is not left")
	}
}

// A roam that keeps failing in a quiet area (the g1 soak: "explore: nav: timeout: roaming for 5m0s" in Cold Plains and
// Blood Moor in turn, nothing ever marked cleared) ends the area after exploreGiveUp failures, counted per area so a
// visit to another area in between does not restart them; a kill does, a quest area never gives up.
func TestExploreGaveUp(t *testing.T) {
	g := newRig(t)
	r := g.r
	ctx := context.Background()
	fail := errors.New("nav: timeout: roaming")
	for _, c := range []struct {
		name  string
		steps []int // areas of successive failures
		kill  bool  // a kill after the first failure
		want  bool  // area 2 cleared at the end
	}{
		{"two failures", []int{2, 2}, false, true},
		{"ping-pong between areas", []int{2, 3, 2, 3}, false, true},
		{"a kill restarts", []int{2, 2}, true, false},
		{"one failure", []int{2}, false, false},
	} {
		r.loop.cleared = map[int]bool{}
		r.loop.exploreFails = nil
		for i, a := range c.steps {
			if i == 1 && c.kill {
				r.obMu.Lock()
				r.ob.kills++
				r.obMu.Unlock()
			}
			r.exploreGaveUp(ctx, a, fail)
		}
		if r.loop.cleared[2] != c.want {
			t.Errorf("%s: cleared[2] = %v, want %v", c.name, r.loop.cleared[2], c.want)
		}
	}
	r.loop.cleared, r.loop.exploreFails = map[int]bool{}, nil
	for _, k := range []RunKind{RunQuest, RunBoss} {
		r.loop.cleared, r.loop.exploreFails = map[int]bool{}, nil
		r.loop.run = &LevelRun{Area: 2, Kind: k}
		r.exploreGaveUp(ctx, 2, fail)
		if r.exploreGaveUp(ctx, 2, fail) || r.loop.cleared[2] {
			t.Fatalf("a %s run keeps roaming", k)
		}
	}
	// the planner sets a clear run for every leveling step: it must still give up (Gate 1: 20 of 22 timeouts)
	r.loop.cleared, r.loop.exploreFails = map[int]bool{}, nil
	r.loop.run = &LevelRun{Area: 2, Kind: RunClear}
	r.exploreGaveUp(ctx, 2, fail)
	if !r.exploreGaveUp(ctx, 2, fail) || !r.loop.cleared[2] {
		t.Fatal("a clear run in the failing area gives up after two failures")
	}
}

// The 17:27 window of Gate 1: Blood Moor and Cold Plains alternating about every 35 s wall (7 virtual minutes a cycle),
// no kill, no explore timeout. Both become cleared; a kill, a third area or a quest run prevents it.
func TestPingPongGuard(t *testing.T) {
	for _, c := range []struct {
		name  string
		areas []int
		kill  int // index before which a kill happens, -1 none
		run   *LevelRun
		step  time.Duration
		want  bool
	}{
		{"replay", []int{2, 3, 2, 3, 2, 3, 2, 3}, -1, &LevelRun{Area: 3, Kind: RunClear}, 150 * time.Second, true},
		{"a kill restarts", []int{2, 3, 2, 3, 2, 3}, 3, nil, 150 * time.Second, false},
		{"third area", []int{2, 3, 2, 3, 4, 3, 2, 3}, -1, nil, 150 * time.Second, false},
		{"too fast", []int{2, 3, 2, 3, 2}, -1, nil, 10 * time.Second, false},
		{"quest run", []int{2, 3, 2, 3, 2, 3, 2, 3}, -1, &LevelRun{Area: 3, Kind: RunQuest}, 150 * time.Second, false},
	} {
		g := newRig(t)
		r := g.r
		r.loop.cleared = map[int]bool{}
		r.loop.run = c.run
		for i, a := range c.areas {
			if i == c.kill {
				r.obMu.Lock()
				r.ob.kills++
				r.obMu.Unlock()
			}
			g.clock.Sleep(c.step)
			r.noteArea(a)
		}
		if got := r.loop.cleared[2] && r.loop.cleared[3]; got != c.want {
			t.Errorf("%s: both cleared = %v, want %v", c.name, got, c.want)
		}
	}
}
