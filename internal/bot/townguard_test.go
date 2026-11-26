package bot

import (
	"context"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
)

// stuckShop is a Townsman whose chores walk out of town (the portal was taken, the zone changed under the walk) and then
// keep walking toward an NPC that is not there, until their context is cancelled: the 02:51 death of 2026-11-26.
type stuckShop struct {
	*fakeTown
	g *rig
}

func (s *stuckShop) Chores(ctx context.Context, p ChorePlan) (ChoreReport, error) {
	s.w.call("chores %s", p.Reason)
	s.g.w.set(func(w *world) { w.area = 3 }) // Cold Plains
	s.g.w.addMonster(900, 1, 4, 0)
	s.g.w.addMonster(901, 1, -5, 3)
	for ctx.Err() == nil {
		s.g.clock.Sleep(100 * time.Millisecond)
	}
	return ChoreReport{}, ctx.Err()
}

// Scenario of the 02:51 death: a trip in its chores stage, the hero in a field with hostiles. The guard aborts the chores
// after the grace, nothing counts as a failure, and the trip starts again from the way home. Two classes: the rule is
// about the place, not the build.
func TestChoresOutsideTownAreAbortedAndTheTripGoesHome(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		t.Run(class.String(), func(t *testing.T) {
			g := newRig(t, func(c *Config, d *Deps, tu *env.Tuning) {
				d.Town = &stuckShop{fakeTown: &fakeTown{}, g: nil}
			})
			st := g.r.d.Town.(*stuckShop)
			st.g, st.fakeTown.w = g, g.w
			g.w.set(func(w *world) { w.class = class; w.state = d2.StateInGame; w.area = 1; w.life, w.maxL = 100, 100 })
			snap := g.r.poll()
			g.r.loop.trip = &trip{plan: g.r.fullChores("kit: test"), returnTo: 2, stage: 1}
			done := make(chan bool, 1)
			go func() { done <- g.r.advanceTrip(context.Background(), snap, g.r.control()) }()
			select {
			case <-done:
			case <-time.After(20 * time.Second):
				t.Fatal("the chores outside town were never aborted")
			}
			if g.r.loop.failures != 0 {
				t.Fatalf("the abort counted %d failures", g.r.loop.failures)
			}
			// the next step: not in town, stage reset, the way home
			snap = g.r.poll()
			g.r.advanceTrip(context.Background(), snap, g.r.control())
			if !strings.Contains(g.w.log(), "totown") {
				t.Fatalf("the trip did not restart from the way home: %s", g.w.log())
			}
		})
	}
}

func TestChoreWatchGrace(t *testing.T) {
	var c choreWatch
	t0 := time.Date(2026, 11, 26, 2, 51, 0, 0, time.UTC)
	field := &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Area: 3}}
	town := &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Area: 1}}
	if c.outside(field, t0) {
		t.Fatal("no chore running: nothing to abort")
	}
	c.begin()
	if c.outside(field, t0) || c.outside(field, t0.Add(choreGrace-time.Millisecond)) {
		t.Fatal("aborted inside the grace")
	}
	if !c.outside(field, t0.Add(choreGrace)) {
		t.Fatal("not aborted after the grace")
	}
	c.outside(town, t0.Add(time.Hour)) // back in town: the clock restarts
	if c.outside(field, t0.Add(time.Hour+time.Second)) {
		t.Fatal("grace did not restart")
	}
}
