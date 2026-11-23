package task_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"d2lod-bot/internal/task"
	"d2lod-bot/internal/task/sim"
)

// drive ticks the Normal Act 1 campaign of a clvl 8 character with a full kit and returns the sim log.
func driveGear(t *testing.T, mut func(w *task.World)) *sim.Game {
	t.Helper()
	g := sim.New(true)
	g.SetLevel(8)
	g.XPStep = 0 // leveling does not change the picture
	g.W.Potions, g.W.HasTP, g.W.InTown = 10, true, true
	g.MarkDone(task.QDen)
	mut(&g.W)
	ex := task.NewExecutor(g, g)
	c := task.NormalAct1(task.Act1Options{SkipTower: true})
	for i := 0; i < 40; i++ {
		if _, err := ex.Tick(context.Background(), c); err != nil {
			t.Fatal(err)
		}
	}
	return g
}

func fieldTrips(g *sim.Game) []string {
	var out []string
	for _, l := range g.Log {
		var a int
		if _, err := fmt.Sscanf(l, "travel:%d", &a); err == nil && !(a == 1) {
			out = append(out, l)
		}
	}
	return out
}

// V: Soaky 2026-10-29. A gated (naked / destitute) character never gets a field travel, a clear or a boss kill, whatever
// quest is next; the healthy twin runs the same campaign into the field as before.
func TestGatedCharacterStaysOutOfTheField(t *testing.T) {
	healthy := driveGear(t, func(w *task.World) {})
	if len(fieldTrips(healthy)) == 0 {
		t.Fatalf("healthy clvl 8 never left town: %v", healthy.Log)
	}
	gated := driveGear(t, func(w *task.World) {
		w.Gated, w.GatedWhy, w.SafeArea = true, "no weapon and no body armor", 2
	})
	for _, l := range gated.Log {
		var a int
		if _, err := fmt.Sscanf(l, "travel:%d", &a); err == nil && a != 2 {
			t.Fatalf("gated character travelled: %v", gated.Log)
		}
		if len(l) > 5 && (l[:5] == "clear" || l[:4] == "kill") {
			t.Fatalf("gated character fought: %v", gated.Log)
		}
	}
}

func TestBannedAreaIsNotEntered(t *testing.T) {
	g := driveGear(t, func(w *task.World) {
		w.Banned = map[task.Area]bool{task.Barracks: true, task.BurialGrounds: true, task.DarkWood: true}
	})
	for _, a := range []task.Area{task.Barracks, task.BurialGrounds} {
		if g.Calls[fmt.Sprintf("travel:%d", a)] > 0 {
			t.Fatalf("travelled to banned area %d: %v", a, g.Log)
		}
	}
}

// G1 attempt 4: a gated character looped talk-charsi retry/skipped and 'Go to the Barracks' not ready. It must not start
// the Tools quest at all: no Charsi greeting, no waypoint trip. A healthy twin still talks to her.
func TestGatedCharacterDoesNotStartToolsQuest(t *testing.T) {
	gated := driveGear(t, func(w *task.World) {
		w.Gated, w.GatedWhy, w.SafeArea = true, "no weapon and no body armor", 2
	})
	for _, l := range gated.Log {
		if strings.HasPrefix(l, "talk") || strings.HasPrefix(l, "wp") {
			t.Fatalf("gated character started a quest step: %v", gated.Log)
		}
	}
	healthy := driveGear(t, func(w *task.World) {})
	talked := false
	for _, l := range healthy.Log {
		talked = talked || strings.HasPrefix(l, "talk")
	}
	if !talked {
		t.Fatalf("healthy character never talked: %v", healthy.Log)
	}
}

// The Horadric Malus comes from a click on the Malus stand, not from the Smith's death (V fork code and lab 2026-11-23): the
// chunk kills the Smith, clicks the stand, picks the Malus up, and only then hands it in. A melee hero (imbue useful) and a
// caster (no imbue use: the Smith still dies and the stand is still clicked when the chunk runs). Must fail while the step is
// a bare pick-up (the stand click never happens and the pick-up finds nothing).
func TestMalusComesFromTheStand(t *testing.T) {
	for _, tc := range []struct {
		name  string
		melee bool
	}{{"paladin", true}, {"necromancer", false}} {
		g := sim.New(true)
		g.SetLevel(28)
		g.XPStep = 0
		g.W.Potions, g.W.HasTP, g.W.InTown, g.W.Caster = 10, true, true, !tc.melee
		g.W.ImbueUseful = tc.melee
		g.W.RouteMinLevel, g.W.RouteMinLevelMerc = 0, 0
		g.MarkDone(task.QDen)
		ex := task.NewExecutor(g, g)
		c := task.NormalAct1(task.Act1Options{SkipTower: true})
		for i := 0; i < 60; i++ {
			if _, err := ex.Tick(context.Background(), c); err != nil {
				t.Fatal(err)
			}
		}
		idx := func(prefix string) int {
			for i, l := range g.Log {
				if strings.HasPrefix(l, prefix) {
					return i
				}
			}
			return -1
		}
		kill, stand := idx("kill:The Smith"), idx("interact:Malus stand")
		pick := -1 // the pick-up after the click (one may come first: the Malus may already lie on the floor)
		for i, l := range g.Log {
			if i > stand && stand >= 0 && strings.HasPrefix(l, "pickup:Horadric Malus") {
				pick = i
				break
			}
		}
		if kill < 0 || stand < kill || pick < stand {
			t.Errorf("%s: kill %d, stand %d, pick-up %d (log %v)", tc.name, kill, stand, pick, g.Log)
		}
	}
}
