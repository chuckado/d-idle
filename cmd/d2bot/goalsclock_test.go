package main

import (
	"testing"
	"time"

	"d2lod-bot/internal/goals"
)

type virtClock struct{ t time.Time }

func (c *virtClock) Now() time.Time        { return c.t }
func (c *virtClock) Sleep(d time.Duration) { c.t = c.t.Add(d) }

// The d2d path hands the goal manager the env clock (card pgclock). Before the change Options.Clock was nil and the manager
// ran on time.Now: this test fails then (the manager's own clock would read the wall, not the virtual date).
func TestD2dGoalsOptionsUseTheEnvClock(t *testing.T) {
	virt := time.Date(1970, 1, 1, 0, 1, 30, 0, time.UTC) // lockstep's base is the epoch: nothing a wall clock could return
	c := &virtClock{t: virt}
	opt := d2dGoalsOptions(nil, "", nil, c, nil)
	if opt.Clock == nil {
		t.Fatal("no clock: goals.New would default to time.Now")
	}
	if got := opt.Clock(); !got.Equal(virt) {
		t.Fatalf("clock reads %v, want the env clock's %v", got, virt)
	}
	c.Sleep(45 * time.Second)
	if got := opt.Clock(); !got.Equal(virt.Add(45 * time.Second)) {
		t.Fatalf("clock did not follow the env clock: %v", got)
	}
}

// A manager built from those options records the virtual date (two builds' goals: caster boss farm, melee boss farm).
func TestManagerBuiltOnTheEnvClockStampsVirtualTime(t *testing.T) {
	for _, area := range []string{"Mephisto", "Andariel"} {
		c := &virtClock{t: time.Date(1970, 1, 1, 0, 0, 5, 0, time.UTC)}
		m, err := goals.New(d2dGoalsOptions(stubRunner{}, "", nil, c, nil))
		if err != nil {
			t.Fatal(err)
		}
		g, err := m.Add(goals.Goal{Type: goals.FarmArea, Params: goals.Params{Area: area, UntilMins: 1}}, -1)
		if err != nil {
			t.Fatal(err)
		}
		if !g.CreatedAt.Equal(c.t) {
			t.Errorf("%s: goal stamped %v, want the virtual %v", area, g.CreatedAt, c.t)
		}
	}
}

type stubRunner struct{}

func (stubRunner) Start(goals.Goal) error     { return nil }
func (stubRunner) Pause() error               { return nil }
func (stubRunner) Resume() error              { return nil }
func (stubRunner) Stop() error                { return nil }
func (stubRunner) Status() goals.Status       { return goals.Status{} }
func (stubRunner) Events() <-chan goals.Event { return nil }
