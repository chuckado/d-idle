package bot

import (
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
)

func recallRig(t *testing.T, class d2.Class, area int, on bool) *rig {
	g := newRig(t, func(c *Config, d *Deps, tu *env.Tuning) {
		if on {
			tu.UI.ManagementFast = 1
		}
	})
	g.w.set(func(w *world) { w.class, w.area = class, area })
	g.mover.exploreFn = blockExplore
	return g
}

// Recall now in the field walks home at the safe moment, says "arrived", and the goal loop resumes afterwards.
func TestRecallNowFieldGoesHomeAndResumes(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		g := recallRig(t, class, 2, true)
		g.run()
		g.r.Start(reach(10))
		g.waitFor("exploring", func() bool { return g.w.count("explore") > 0 })
		if v := g.r.RecallNow(); v.Error != "" || !v.Enabled {
			t.Fatalf("%v: refused: %+v", class, v)
		}
		g.waitFor("arrived", func() bool { return g.r.RecallView().State == RecallArrived })
		if g.w.count("totown") == 0 {
			t.Fatalf("%v: arrived without a walk home", class)
		}
		before := g.w.count("explore")
		g.waitFor("goal resumes", func() bool { return g.w.count("explore") > before })
		g.finish()
	}
}

// With the knob off the command is refused with the reason and nothing walks (the queue path stays as it was).
func TestRecallNowKnobOffIsRefused(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		g := recallRig(t, class, 2, false)
		g.run()
		g.r.Start(reach(10))
		g.waitFor("exploring", func() bool { return g.w.count("explore") > 0 })
		v := g.r.RecallNow()
		if v.Error == "" || v.Enabled || v.State != RecallIdle {
			t.Fatalf("%v: %+v", class, v)
		}
		time.Sleep(30 * time.Millisecond)
		if n := g.w.count("totown"); n != 0 {
			t.Fatalf("%v: walked home %d times with the knob off", class, n)
		}
		g.finish()
	}
}

// A hostile in reach holds the walk (the same safe moment as every player command): the fight comes first, the view says why.
func TestRecallNowWaitsForTheFight(t *testing.T) {
	g := recallRig(t, d2.Paladin, 2, true)
	g.w.addMonster(7, 10, 8, 0)
	g.run()
	g.r.Start(reach(10))
	g.r.RecallNow()
	g.waitFor("arrived", func() bool { return g.r.RecallView().State == RecallArrived })
	g.w.mu.Lock()
	calls := append([]string(nil), g.w.calls...)
	g.w.mu.Unlock()
	home, clear := -1, -1
	for i, c := range calls {
		if c == "totown" && home < 0 {
			home = i
		}
		if len(c) >= 5 && c[:5] == "clear" && clear < 0 {
			clear = i
		}
	}
	if clear < 0 || home < clear {
		t.Fatalf("went home at %d, fight cleared at %d: %v", home, clear, calls)
	}
	g.finish()
}

// A failed walk is reported as failed with a reason and retried; cancel withdraws a waiting command.
func TestRecallNowFailedWalkAndCancel(t *testing.T) {
	g := recallRig(t, d2.Sorceress, 2, true)
	g.town.toTownErr = errBoom
	g.run()
	g.r.Start(reach(10))
	g.waitFor("exploring", func() bool { return g.w.count("explore") > 0 })
	g.r.RecallNow()
	g.waitFor("tried", func() bool { return g.w.count("totown") > 0 })
	g.waitFor("failed", func() bool { return g.r.RecallView().State == RecallFailed })
	if v := g.r.RecallView(); v.Reason == "" || g.inTown() {
		t.Fatalf("%+v", v)
	}
	if v := g.r.RecallCancel(); v.State != RecallIdle {
		t.Fatalf("cancel: %+v", v)
	}
	g.finish()
}
