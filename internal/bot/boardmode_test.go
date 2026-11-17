package bot

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/questboard"
)

var modeClasses = []d2.Class{d2.Sorceress, d2.Paladin}

// modeRig is boardRig for a class, with the mode stored and an explore hook that records whether a farm slice was running.
func modeRig(t *testing.T, cl d2.Class, on bool, mode string, area int) (*rig, *questboard.Exec, *atomic.Int32, *atomic.Int32) {
	t.Helper()
	g, x := boardRig(t, on)
	g.w.set(func(w *world) { w.class = cl; w.area = area })
	var farmSeen, trainSeen atomic.Int32
	g.mover.exploreFn = func(ctx context.Context, done func() bool) error {
		if g.r.boardFarming() {
			farmSeen.Add(1)
		} else {
			trainSeen.Add(1)
		}
		return nil
	}
	x.SetMode(mode)
	return g, x, &farmSeen, &trainSeen
}

func fieldWork(log string) bool {
	for _, k := range []string{"explore", "exit ", "clear r=", "wp ", "goto "} {
		if strings.Contains(log, k) {
			return true
		}
	}
	return false
}

// Idle: a hero standing in the field goes home and then does no field work and does not trip the stall detectors; the
// stored mode is read by the loop, not the planner.
func TestBoardModeIdleParksInTownNoFieldWork(t *testing.T) {
	for _, cl := range modeClasses {
		t.Run(cl.String(), func(t *testing.T) {
			g, _, farm, train := modeRig(t, cl, true, "idle", 2)
			g.run()
			defer g.finish()
			if err := g.r.Start(campaign()); err != nil {
				t.Fatal(err)
			}
			g.waitFor("home", func() bool { return g.w.count("totown") > 0 })
			g.waitFor("parked", func() bool { return g.r.parkedBoard.Load() })
			log := g.w.log()
			time.Sleep(150 * time.Millisecond)
			if g.w.log() != log && fieldWork(g.w.log()[len(log):]) {
				t.Fatalf("field work while idle: %s", g.w.log())
			}
			if fieldWork(g.w.log()) || farm.Load() != 0 || train.Load() != 0 {
				t.Fatalf("field work while idle: %s", g.w.log())
			}
			g.w.mu.Lock()
			area := g.w.area
			g.w.mu.Unlock()
			if area != 1 {
				t.Fatalf("idle hero is in area %d, want the town", area)
			}
			// hours of parking: neither the state watchdog nor the campaign stall ladder fires
			g.clock.Sleep(3 * time.Hour)
			if why := g.r.stuckCheck(g.r.e.Clock.Now()); why != "" {
				t.Fatalf("watchdog fired while parked on purpose: %s", why)
			}
			if g.stalled() != nil {
				t.Fatal("campaign stall exit while idle")
			}
		})
	}
}

// Train and Farm both work the field; only a farm slice farms (the area choice reads boardFarming).
func TestBoardModeTrainAndFarmSlices(t *testing.T) {
	for _, cl := range modeClasses {
		for _, mode := range []string{"train", "farm"} {
			t.Run(cl.String()+"/"+mode, func(t *testing.T) {
				g, _, farm, train := modeRig(t, cl, true, mode, 1)
				g.run()
				defer g.finish()
				if err := g.r.Start(campaign()); err != nil {
					t.Fatal(err)
				}
				g.waitFor("field work", func() bool { return farm.Load()+train.Load() > 0 })
				if mode == "train" && farm.Load() != 0 {
					t.Fatal("a train slice farmed")
				}
				if mode == "farm" && train.Load() != 0 {
					t.Fatal("a farm slice ran as training")
				}
				if g.r.parkedBoard.Load() {
					t.Fatal("parked while working the field")
				}
			})
		}
	}
}

// A mode switch mid-run: field work stops at the next safe moment and the hero parks; back to train it leaves again.
func TestBoardModeSwitchMidRun(t *testing.T) {
	for _, cl := range modeClasses {
		t.Run(cl.String(), func(t *testing.T) {
			g, x, farm, train := modeRig(t, cl, true, "train", 1)
			g.run()
			defer g.finish()
			if err := g.r.Start(campaign()); err != nil {
				t.Fatal(err)
			}
			g.waitFor("training", func() bool { return train.Load() > 0 })
			x.SetMode("idle")
			g.waitFor("parked", func() bool { return g.r.parkedBoard.Load() })
			time.Sleep(50 * time.Millisecond)
			n := train.Load()
			time.Sleep(150 * time.Millisecond)
			if train.Load() != n {
				t.Fatal("field work after the switch to idle")
			}
			x.SetMode("farm")
			g.waitFor("farming", func() bool { return farm.Load() > 0 })
		})
	}
}

// A hostile near holds a mode change back: the hero finishes the fight first. Driven without the loop (one tick each).
func TestBoardModeChangeWaitsForASafeMoment(t *testing.T) {
	for _, cl := range modeClasses {
		t.Run(cl.String(), func(t *testing.T) {
			g, x, _, _ := modeRig(t, cl, true, "train", 2)
			g.r.poll()
			if m := g.r.boardMode(g.r.latest()); m != "train" {
				t.Fatalf("first tick: %s", m)
			}
			g.w.addMonster(500, 1, 5, 5)
			g.r.poll()
			x.SetMode("idle")
			if g.r.safeToPause(g.r.latest()) {
				t.Skip("the fake hostile is not inside the clear radius")
			}
			if m := g.r.boardMode(g.r.latest()); m != "train" {
				t.Fatalf("mode switched with a hostile near: %s", m)
			}
			g.w.set(func(w *world) { w.mons = nil })
			g.r.poll()
			if m := g.r.boardMode(g.r.latest()); m != "idle" {
				t.Fatalf("mode did not switch once safe: %s", m)
			}
		})
	}
}

// An accepted chunk comes first whatever the mode; after the hand-in the hero returns to the mode (idle: no field work).
func TestBoardModeChunkPrecedenceThenBack(t *testing.T) {
	for _, cl := range modeClasses {
		t.Run(cl.String(), func(t *testing.T) {
			g, x, _, train := modeRig(t, cl, true, "idle", 1)
			g.run()
			defer g.finish()
			if err := g.r.Start(campaign()); err != nil {
				t.Fatal(err)
			}
			g.waitFor("parked", func() bool { return g.r.parkedBoard.Load() })
			if err := x.Accept("den_of_evil", "clear_den"); err != nil {
				t.Fatal(err)
			}
			g.waitFor("the chunk ran", func() bool { d, _ := x.Next(); return d.Kind == questboard.DecAwait })
			if g.r.parkedBoard.Load() && g.w.count("explore") == 0 {
				t.Fatal("the chunk did not run in idle mode")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			if _, err := x.HandIn(ctx, "den_of_evil", "clear_den"); err != nil {
				t.Fatal(err)
			}
			g.waitFor("parked again", func() bool { d, _ := x.Next(); return d.Kind == questboard.DecIdle && g.r.parkedBoard.Load() })
			log := g.w.log()
			time.Sleep(150 * time.Millisecond)
			if g.w.log() != log && fieldWork(g.w.log()[len(log):]) {
				t.Fatalf("field work after the hand-in in idle mode: %s", g.w.log())
			}
			_ = train
		})
	}
}

// Knob 0: the stored mode changes nothing; the planner plays the Den as before and nothing is ever parked.
func TestBoardModeKnobOffIsTheOldPlanner(t *testing.T) {
	g, _, _, _ := modeRig(t, d2.Sorceress, false, "idle", 1)
	g.run()
	defer g.finish()
	if err := g.r.Start(campaign()); err != nil {
		t.Fatal(err)
	}
	g.waitFor("reward talk", func() bool { return g.w.count("talk Akara/reward") > 0 })
	if g.r.parkedBoard.Load() || g.r.boardFarming() {
		t.Fatal("board mode state set with the knob off")
	}
}

// The meter data: words per mode, the levels to go for the pinned target, the running quest first.
func TestQuestMeterWords(t *testing.T) {
	for _, cl := range modeClasses {
		t.Run(cl.String(), func(t *testing.T) {
			g, x, _, _ := modeRig(t, cl, true, "train", 1)
			g.run()
			defer g.finish()
			g.waitFor("snapshot", func() bool { return g.r.QuestClass() != "" })
			st, err := x.State()
			if err != nil {
				t.Fatal(err)
			}
			m := g.r.QuestMeter("train", st)
			if m == nil || m.Mode != "train" || m.Line == "" || m.Target == "" {
				t.Fatalf("meter: %+v", m)
			}
			if m := g.r.QuestMeter("idle", st); !strings.Contains(m.Line, "Waiting in town") {
				t.Fatalf("meter: %+v", m)
			}
			if m := g.r.QuestMeter("farm", st); !strings.Contains(m.Line, "Farming") {
				t.Fatalf("meter: %+v", m)
			}
		})
	}
}
