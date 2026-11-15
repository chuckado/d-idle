package bot

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
)

func resetRig(t *testing.T, class d2.Class, area int, cycler bool) *rig {
	g := newRig(t, func(c *Config, d *Deps, tu *env.Tuning) {})
	if cycler {
		g.r.d.Cycler = fakeCycler{g.w}
	}
	g.w.set(func(w *world) { w.class, w.area = class, area })
	g.mover.exploreFn = blockExplore
	return g
}

func (g *rig) resetState() string { return g.r.ResetView().State }

// A soft reset in the field leaves the game once, the same goal carries on afterwards, and the last result says done.
func TestPlayerResetLeavesAndRejoinsThenResumes(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		g := resetRig(t, class, 2, true)
		g.run()
		g.r.Start(reach(10))
		g.waitFor("exploring", func() bool { return g.w.count("explore") > 0 })
		if v := g.r.PlayerReset(true); v.State != ResetResetting || v.Already {
			t.Fatalf("%v: press: %+v", class, v)
		}
		g.waitFor("done", func() bool { v := g.r.ResetView(); return v.State == ResetIdle && v.Last != nil })
		if n := g.w.count("newgame normal"); n != 1 {
			t.Fatalf("%v: new games: %d (%v)", class, n, g.w.calls)
		}
		if l := g.r.ResetView().Last; l.Result != "done" || l.Hard {
			t.Fatalf("%v: last: %+v", class, l)
		}
		before := g.w.count("explore")
		g.waitFor("plan resumes", func() bool { return g.w.count("explore") > before })
		if st := g.r.Status(); st.GoalID == "" {
			t.Fatalf("%v: goal gone after the reset", class)
		}
		g.finish()
	}
}

// A second press while one runs changes nothing and says so.
func TestPlayerResetDoublePressIsIdempotent(t *testing.T) {
	g := resetRig(t, d2.Sorceress, 1, true)
	first := g.r.PlayerReset(false)
	second := g.r.PlayerReset(true)
	if !second.Already || second.State != ResetResetting || !second.Since.Equal(*first.Since) {
		t.Fatalf("second press: %+v vs %+v", second, first)
	}
	g.run()
	g.r.Start(reach(10))
	g.waitFor("done", func() bool { return g.r.ResetView().Last != nil })
	if n := g.w.count("newgame normal"); n != 1 {
		t.Fatalf("new games: %d", n)
	}
	g.finish()
}

// A backend without a Cycler cannot leave the game: the reset ends failed with the reason (the wire falls back to hard).
func TestPlayerResetWithoutCyclerFails(t *testing.T) {
	g := resetRig(t, d2.Paladin, 1, false)
	if g.r.CanSoftReset() {
		t.Fatal("CanSoftReset without a cycler")
	}
	g.r.PlayerReset(false)
	g.run()
	g.r.Start(reach(10))
	g.waitFor("failed", func() bool { return g.resetState() == ResetFailed })
	if g.r.ResetView().Reason == "" {
		t.Fatal("no reason")
	}
	g.r.PlayerReset(false) // a failed reset can be pressed again
	if g.resetState() != ResetResetting {
		t.Fatalf("after the second press: %s", g.resetState())
	}
	g.finish()
}

// A paused hero is reset and stays paused (the pause is the player's, the reset does not clear it).
func TestPlayerResetKeepsThePause(t *testing.T) {
	g := resetRig(t, d2.Sorceress, 1, true)
	g.run()
	g.r.Start(reach(10))
	g.r.PlayerPause(true)
	g.waitFor("paused", func() bool { return g.pauseState() == PausePaused })
	g.r.PlayerReset(true)
	g.waitFor("done", func() bool { return g.r.ResetView().Last != nil })
	if g.pauseState() != PausePaused {
		t.Fatalf("pause after reset: %s", g.pauseState())
	}
	g.finish()
}

// A hard reset the old process recorded shows as rejoining in the new one and ends when the hero is in a game; the
// last result survives in the file.
func TestPlayerResetHardSurvivesTheRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "playerreset.json")
	g := resetRig(t, d2.Sorceress, 2, true)
	g.r.SetResetPath(path)
	if v := g.r.BeginHardReset(); v.State != ResetResetting || !v.Hard {
		t.Fatalf("hard: %+v", v)
	}
	if again := g.r.BeginHardReset(); !again.Already {
		t.Fatalf("second hard press: %+v", again)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}

	g2 := resetRig(t, d2.Paladin, 1, true)
	g2.r.SetResetPath(path)
	if v := g2.r.ResetView(); v.State != ResetRejoining || !v.Hard {
		t.Fatalf("after restart: %+v", v)
	}
	g2.run()
	g2.r.Start(reach(10))
	g2.waitFor("done", func() bool { return g2.r.ResetView().State == ResetIdle })
	if n := g2.w.count("newgame normal"); n != 0 {
		t.Fatalf("a hard reset must not leave the game again: %d", n)
	}
	g2.finish()

	g3 := resetRig(t, d2.Sorceress, 1, true)
	g3.r.SetResetPath(path)
	if v := g3.r.ResetView(); v.State != ResetIdle || v.Last == nil || v.Last.Result != "done" || !v.Last.Hard {
		t.Fatalf("third process: %+v", v)
	}
}

// Without force the reset waits while a hostile is near, but never longer than resetPatience.
func TestPlayerResetPatienceIsBounded(t *testing.T) {
	if resetPatience > 30*time.Second {
		t.Fatalf("a logout is an escape: patience %v is too long", resetPatience)
	}
}
