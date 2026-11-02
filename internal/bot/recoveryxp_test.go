package bot

import (
	"testing"

	"d2lod-bot/internal/d2"
)

func TestRecoveryXPModel(t *testing.T) {
	// V-d2d means at clvl 12: Cold Plains under 3, Stony Field about 15, Dark Wood about 35.
	if x := areaXPPerKill(12, 2); x >= 3 {
		t.Fatalf("cold plains %.1f", x)
	}
	if x := areaXPPerKill(12, 4); x < 12 || x > 20 {
		t.Fatalf("stony %.1f", x)
	}
	if x := areaXPPerKill(12, 5); x < 28 || x > 40 {
		t.Fatalf("dark wood %.1f", x)
	}
	if xpLevelPct(30, 36) != 100*30.0/36 || xpLevelPct(5, 15) != 1.95 || xpLevelPct(5, 5) != 100 {
		t.Fatal("penalty table")
	}
}

func xpPick(g *rig, level int, ex map[int]bool) (int, bool) {
	g.w.set(func(w *world) { w.level = level })
	s, _ := g.w.snapshot()
	return g.r.recoveryXPArea(s, ex)
}

func TestRecoveryXPAreaByLevel(t *testing.T) {
	g := newRig(t)
	if a, ok := xpPick(g, 14, nil); !ok || (a != 5 && a != 4 && a != 6) {
		t.Fatalf("clvl 14 picked %d %v", a, ok)
	}
	if a, _ := xpPick(g, 14, nil); a == 3 || a == 2 {
		t.Fatalf("clvl 14 must not farm %d", a)
	}
	if a, ok := xpPick(g, 6, nil); !ok || (a != 2 && a != 3 && a != 9) {
		t.Fatalf("clvl 6 picked %d", a)
	}
	// the cap: with +0 a clvl 6 character takes nothing above monster level 6
	g.e.Tune.Resource.RecoveryMaxAreaAbove = 0
	if a, _ := xpPick(g, 9, nil); a == 6 {
		t.Fatalf("clvl 9 capped at +0 picked Black Marsh")
	}
}

func TestRecoveryXPAreaExcluded(t *testing.T) {
	g := newRig(t)
	a, _ := xpPick(g, 14, nil)
	b, ok := xpPick(g, 14, map[int]bool{a: true})
	if !ok || b == a {
		t.Fatalf("excluded %d still picked %d", a, b)
	}
	all := map[int]bool{}
	for id := range recoveryXPTier {
		all[id] = true
	}
	if _, ok := xpPick(g, 14, all); ok {
		t.Fatal("everything excluded must report nothing")
	}
}

func TestRecoveryAreaBannedFallsThrough(t *testing.T) {
	g := newRig(t)
	g.w.set(func(w *world) { w.level = 14 })
	g.w.belt()
	s, _ := g.w.snapshot()
	for id := range recoveryXPTier {
		g.r.avoid[id] = g.clock.Now().Add(20 * 60e9)
	}
	if a, ok := g.r.recoveryArea(s); ok {
		t.Fatalf("all banned must leave no area (noidle wait), got %d", a)
	}
}

func TestRecoveryAreaUsesXPPick(t *testing.T) {
	g := newRig(t)
	g.w.set(func(w *world) { w.level = 14 })
	g.w.set(func(w *world) {
		w.items = append(w.items, d2.Unit{ID: 900, Type: d2.TypeItem, Class: 5, Item: &d2.Item{Where: d2.ContInventory}})
	})
	s, _ := g.w.snapshot()
	if !g.r.resInput(s).HasTP {
		t.Fatal("rig has no town portal")
	}
	a, ok := g.r.recoveryArea(s)
	if !ok || a == 2 || a == 3 {
		t.Fatalf("clvl 14 recovery area %d %v", a, ok)
	}
}

// XPGAP: the early-entry knob lets a recovering clvl 3 character farm Cold Plains (plan enter 5), 0 keeps the plan levels.
func TestRecoveryXPEarlyEntry(t *testing.T) {
	g := newRig(t)
	if a, ok := xpPick(g, 3, nil); !ok || a != 2 {
		t.Fatalf("knob 0, clvl 3 picked %d %v, want Blood Moor", a, ok)
	}
	g.e.Tune.Resource.RecoveryXPEarly = 2
	if a, ok := xpPick(g, 3, nil); !ok || a != 3 {
		t.Fatalf("knob 2, clvl 3 picked %d %v, want Cold Plains", a, ok)
	}
	if a, ok := xpPick(g, 2, nil); !ok || a != 2 {
		t.Fatalf("knob 2, clvl 2 picked %d %v, want Blood Moor", a, ok)
	}
}

// monster-stats-trace B1: the table follows the real plain-monster means (22.0 Blood Moor, 26.1 Den, ...) and ranks
// Cave 1 > Cave 2 > Cold Plains, which the old 10 x level formula tied.
func TestAreaXPTable(t *testing.T) {
	for _, tc := range []struct {
		area, clvl int
		want       float64
	}{
		{2, 5, 22.0}, {8, 5, 26.1}, {3, 5, 25.6}, {9, 5, 35.2}, {13, 5, 27.2}, {4, 5, 37.3}, {5, 10, 47.0}, {6, 10, 52.0},
		{2, 10, 5.2}, // trace 4.9 at clvl 10 (per unit mix): within 10 percent
		{44, 15, 215.8 * 0.95},
	} {
		got := areaXP(tc.clvl, tc.area, 0, d2.Normal)
		if got < tc.want*0.9 || got > tc.want*1.1 {
			t.Errorf("area %d clvl %d: %.1f want about %.1f", tc.area, tc.clvl, got, tc.want)
		}
	}
	if !(areaXP(5, 9, 2, d2.Normal) > areaXP(5, 13, 2, d2.Normal) && areaXP(5, 13, 2, d2.Normal) > areaXP(5, 3, 2, d2.Normal)) {
		t.Error("order Cave 1 > Cave 2 > Cold Plains")
	}
	if areaXP(5, 2, 7, d2.Nightmare) != areaXPPerKill(5, 7) || areaXP(5, 999, 4, d2.Normal) != areaXPPerKill(5, 4) {
		t.Error("fallback for harder difficulties and unknown areas")
	}
	for _, id := range []int{8, 13} {
		if normalMLvl[id] != map[int]int{8: 1, 13: 2}[id] {
			t.Errorf("fallback level %d", id)
		}
	}
}
