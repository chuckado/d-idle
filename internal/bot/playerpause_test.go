package bot

import (
	"path/filepath"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
)

func pauseRig(t *testing.T, class d2.Class, area int) *rig {
	g := newRig(t, func(c *Config, d *Deps, tu *env.Tuning) {})
	g.w.set(func(w *world) { w.class, w.area = class, area })
	g.mover.exploreFn = blockExplore
	return g
}

func (g *rig) pauseState() string { return g.r.PlayerPauseView().State }

func (g *rig) inTown() bool {
	g.w.mu.Lock()
	defer g.w.mu.Unlock()
	return g.w.area == 1
}

// Pause in the field goes home first and then stops; the state is never "paused" while the hero stands outside town;
// Resume continues the plan.
func TestPlayerPauseFieldGoesHomeThenStops(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		g := pauseRig(t, class, 2)
		g.run()
		g.r.Start(reach(10))
		g.waitFor("exploring", func() bool { return g.w.count("explore") > 0 })
		g.r.PlayerPause(true)
		g.waitFor("paused", func() bool {
			if g.pauseState() == PausePaused && !g.inTown() {
				t.Errorf("%v: paused outside town", class)
			}
			return g.pauseState() == PausePaused
		})
		if !g.inTown() || g.w.count("totown") == 0 {
			t.Fatalf("%v: paused without going home", class)
		}
		g.waitFor("runner paused", func() bool { return g.r.State() == StatePaused })
		before := g.w.count("explore")
		g.r.PlayerPause(false)
		if g.pauseState() != PauseRunning {
			t.Fatalf("%v: not running after resume", class)
		}
		g.waitFor("plan continues", func() bool { return g.r.State() != StatePaused && g.w.count("explore") > before })
		g.finish()
	}
}

// In town the pause is immediate and needs no walk.
func TestPlayerPauseInTownStopsAtOnce(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		g := pauseRig(t, class, 1)
		g.run()
		g.r.Start(reach(10))
		g.r.PlayerPause(true)
		g.waitFor("paused", func() bool { return g.pauseState() == PausePaused })
		if n := g.w.count("totown"); n != 0 {
			t.Fatalf("%v: walked home from town: %d", class, n)
		}
		g.finish()
	}
}

// A hostile in reach holds the walk home (the same safe moment as the errand trip); the pack is fought first.
func TestPlayerPauseWaitsForTheFight(t *testing.T) {
	g := pauseRig(t, d2.Sorceress, 2)
	g.w.addMonster(7, 10, 8, 0)
	g.run()
	g.r.Start(reach(10))
	g.r.PlayerPause(true)
	g.waitFor("paused", func() bool { return g.pauseState() == PausePaused })
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

// A failed walk home does not pause the hero in the field; it stays "going_home" and retries.
func TestPlayerPauseFailedWalkStaysGoingHome(t *testing.T) {
	g := pauseRig(t, d2.Paladin, 2)
	g.town.toTownErr = errBoom
	g.run()
	g.r.Start(reach(10))
	g.waitFor("exploring", func() bool { return g.w.count("explore") > 0 })
	g.r.PlayerPause(true)
	g.waitFor("tried", func() bool { return g.w.count("totown") > 0 })
	time.Sleep(20 * time.Millisecond)
	if g.pauseState() != PauseGoingHome || g.inTown() {
		t.Fatalf("state %s, in town %v", g.pauseState(), g.inTown())
	}
	g.finish()
}

// A restart stays paused: the request is read back from the file and the hero goes home and stops again.
func TestPlayerPauseSurvivesARestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "playerpause.json")
	g := pauseRig(t, d2.Sorceress, 1)
	g.r.SetPausePath(path)
	g.r.PlayerPause(true)
	g.run()
	g.r.Start(reach(10))
	g.waitFor("paused", func() bool { return g.pauseState() == PausePaused })
	g.finish()

	g2 := pauseRig(t, d2.Paladin, 2)
	g2.r.SetPausePath(path)
	if g2.pauseState() != PauseGoingHome {
		t.Fatalf("after restart: %s", g2.pauseState())
	}
	g2.run()
	g2.r.Start(reach(10))
	g2.waitFor("paused again", func() bool { return g2.pauseState() == PausePaused })
	if !g2.inTown() {
		t.Fatal("paused outside town after restart")
	}
	g2.r.PlayerPause(false)
	g2.finish()

	g3 := pauseRig(t, d2.Sorceress, 2)
	g3.r.SetPausePath(path)
	if g3.pauseState() != PauseRunning {
		t.Fatalf("resume was not persisted: %s", g3.pauseState())
	}
}
