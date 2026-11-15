package bot

import (
	"context"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/goals"
)

func TestFarmResetPays(t *testing.T) {
	// 4000 XP in 10 minutes, a 1 minute overhead: a reset yields 4000/11 = 364 per minute
	if p, f := FarmResetPays(4000, 10, 300, 1); !p || f < 363 || f > 365 {
		t.Fatalf("a zone at 300 XP/min must be reset: %v %.0f", p, f)
	}
	if p, _ := FarmResetPays(4000, 10, 380, 1); p {
		t.Fatal("a zone still paying more than a fresh one is not reset")
	}
	if p, _ := FarmResetPays(4000, 10, 300, 8); p {
		t.Fatal("a long walk back keeps the hero in the zone: 4000/18 = 222")
	}
	if p, _ := FarmResetPays(0, 10, 0, 1); p {
		t.Fatal("no XP seen: no reset")
	}
}

// farmRig is a hero in the field with a roaming mover at pct coverage, a leveling slice and the loop on.
func farmLoopRig(t *testing.T, cl d2.Class, pct int) *rig {
	t.Helper()
	g := ladderRig(t, cl, 6)
	g.w.set(func(w *world) { w.area = 2; w.life, w.maxL = 100, 100; w.level = 5; w.exp = 5000 })
	g.r.d.Mover = covMover{fakeMover: g.mover, v: pct, m: 100}
	g.r.e.Tune.Bot.FarmReset, g.r.e.Tune.Bot.FarmResetPct, g.r.e.Tune.Bot.FarmResetCooldownMin = 1, 85, 5
	g.r.loop.levelSlice = true
	g.r.poll()
	g.r.farmNote(2)
	return g
}

// run advances 10 s steps; gain is the XP of each step.
func (g *rig) farmAdvance(d time.Duration, gain uint32) {
	for t := time.Duration(0); t < d; t += 10 * time.Second {
		g.clock.Sleep(10 * time.Second)
		g.w.set(func(w *world) { w.exp += gain })
		g.r.poll()
		g.r.farmNote(2)
	}
}

func TestFarmResetDueAndGuards(t *testing.T) {
	for _, cl := range []d2.Class{d2.Sorceress, d2.Paladin} {
		t.Run(cl.String(), func(t *testing.T) {
			g := farmLoopRig(t, cl, 90)
			g.farmAdvance(6*time.Minute, 60) // 6 XP/s early
			g.farmAdvance(4*time.Minute, 3)  // the zone dries up: 0.3 XP/s
			if !g.r.farmResetDue(2) {
				t.Fatal("90% covered, rate fell below the cycle average: reset")
			}
			g.r.loop.farm.pending = 0

			g.r.e.Tune.Bot.FarmResetPct = 95
			if g.r.farmResetDue(2) {
				t.Fatal("below the knob's coverage")
			}
			g.r.e.Tune.Bot.FarmResetPct = 85
			g.r.e.Tune.Bot.FarmReset = 0
			if g.r.farmResetDue(2) {
				t.Fatal("knob off")
			}
			g.r.e.Tune.Bot.FarmReset = 1
			g.r.loop.run = &LevelRun{Area: 2, Kind: RunQuest}
			if g.r.farmResetDue(2) {
				t.Fatal("a quest run keeps roaming")
			}
			g.r.loop.run = nil
			g.r.loop.farm.lastAt = g.clock.Now().Add(-time.Minute)
			if g.r.farmResetDue(2) {
				t.Fatal("cooldown")
			}
			g.r.loop.farm.lastAt = time.Time{}
			g.r.loop.recoveredAt = g.clock.Now()
			if g.r.farmResetDue(2) {
				t.Fatal("within 180 s of a recovery")
			}
			g.r.loop.recoveredAt = time.Time{}
			g.r.loop.levelSlice = false
			if g.r.farmResetDue(2) {
				t.Fatal("not a leveling slice")
			}
		})
	}
}

func TestFarmResetKeepsRichZone(t *testing.T) {
	g := farmLoopRig(t, d2.Sorceress, 90)
	g.farmAdvance(10*time.Minute, 60) // still 6 XP/s: no reset pays
	if g.r.farmResetDue(2) {
		t.Fatal("a zone that still pays full rate must not reset")
	}
}

// The reset goes home, takes a fresh game, forgets the cleared areas and sends the planner back to the same zone.
func TestFarmResetStepAndReturn(t *testing.T) {
	g := farmLoopRig(t, d2.Sorceress, 90)
	g.farmAdvance(6*time.Minute, 60)
	g.farmAdvance(4*time.Minute, 3)
	if !g.r.farmResetDue(2) {
		t.Fatal("not due")
	}
	g.r.loop.cleared = map[int]bool{3: true}
	s := g.r.latest()
	if done, _ := g.r.farmResetStep(context.Background(), s); !done {
		t.Fatal("the step must act")
	}
	g.w.set(func(w *world) { w.area = 1 })
	g.r.poll()
	if done, err := g.r.farmResetStep(context.Background(), g.r.latest()); !done || err != nil {
		t.Fatalf("fresh game: %v %v", done, err)
	}
	if g.w.count("newgame") == 0 || len(g.r.loop.cleared) != 0 || g.r.loop.farm.pending != 0 {
		t.Fatalf("no fresh game or the old one's state kept: %d %v", g.w.count("newgame"), g.r.loop.cleared)
	}
	if g.r.loop.farm.returnTo != 2 || g.r.loop.farm.lastAt.IsZero() {
		t.Fatal("the zone to return to is not recorded")
	}
	c := control{spec: goalSpec{kind: goals.ReachLevel}}
	if got := g.r.farmReturn(g.r.latest(), c, 9); got != 2 {
		t.Fatalf("the planner's pick 9 must be replaced by the zone 2, got %d", got)
	}
	g.r.loop.cleared[2] = true
	if got := g.r.farmReturn(g.r.latest(), c, 9); got != 9 {
		t.Fatalf("a cleared zone is not returned to, got %d", got)
	}
}
