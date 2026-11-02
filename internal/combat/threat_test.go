package combat

import (
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/env"
)

// Vel 8 / 24 are the trace's shaman fire and Skeleton Archer arrow rows (docs/research/monster-missile-trace.md Q6:
// 0.43 s and 0.14 s at 4 subtiles).
func testThreats() ThreatTable {
	return ThreatTable{
		22: ThreatFromRow(&data.Missile{ID: 22, Name: "shafire1", Vel: 8}),
		30: ThreatFromRow(&data.Missile{ID: 30, Name: "skbowarrow1", Vel: 24, ToHit: true, MaxDamage: 6}),
	}
}

func TestThreatSpeedAndVerdict(t *testing.T) {
	tb := testThreats()
	if s := tb[30].Speed; s < 28 || s > 28.3 {
		t.Fatalf("arrow speed %v", s)
	}
	if !tb[30].Blockable || tb[22].Blockable {
		t.Fatal("blockable proxy")
	}
	for _, c := range []struct {
		id      int
		dist    float64
		shooter bool
		want    Verdict
	}{
		{22, 4, false, VerdictDodge},      // 0.43 s at 4 subtiles beats 0.4
		{30, 4, false, VerdictIgnore},     // 0.14 s
		{30, 4, true, VerdictKillShooter}, // same with a known shooter
		{30, 12, true, VerdictDodge},      // 0.43 s from further out
		{99, 1, true, VerdictDodge},       // zero speed: the old rule
	} {
		th := tb[c.id]
		if got := th.Classify(c.dist, 0.4, c.shooter); got != c.want {
			t.Errorf("%+v: %v", c, got)
		}
	}
}

func TestRankIncoming(t *testing.T) {
	snap := &d2.Snapshot{Player: &d2.Player{ID: 1, X: 100, Y: 100},
		Monsters: []d2.Unit{{ID: 50, Class: 7}},
		Missiles: []d2.Unit{
			{Class: 22, OwnerType: 1, OwnerID: 60, X: 104, Y: 100}, // owner not in view: slow, dodge
			{Class: 30, OwnerType: 1, OwnerID: 50, X: 102, Y: 100}, // arrow from a visible archer
			{Class: 30, OwnerType: 1, OwnerID: 50, X: 101, Y: 100}, // second arrow, same shooter once
			{Class: 30, OwnerType: 0, OwnerID: 1, X: 101, Y: 101},  // our own
			{Class: 77, OwnerType: 1, OwnerID: 61, X: 100, Y: 103}, // unknown row: dodge
			{Class: 22, OwnerType: 1, OwnerID: 60, X: 130, Y: 100}, // far
		}}
	in := RankIncoming(snap, testThreats(), 4.5, 0.4)
	if len(in.Dodge) != 2 || len(in.Shooters) != 1 || in.Shooters[0] != 50 {
		t.Fatalf("%+v", in)
	}
	if got := RankIncoming(snap, nil, 4.5, 0.4); len(got.Dodge) != 4 || len(got.Shooters) != 0 {
		t.Fatalf("nil table keeps the old rule: %+v", got)
	}
}

func TestShooterPeakAndKillFirst(t *testing.T) {
	mage := &data.Monster{Key: "skmage_fire", Ranged: true, Killable: true}
	mage.Elem[0] = data.MonElem{Min: 4, Max: 129}
	archer := &data.Monster{Key: "sk_archer", Ranged: true, Killable: true}
	archer.A1Max = [3]int{6, 27, 68}
	melee := &data.Monster{Key: "zombie", Killable: true, A1Max: [3]int{9, 9, 9}}
	if p, u := ShooterPeak(mage, d2.Hell); p != 129 || !u {
		t.Fatalf("mage %d %v", p, u)
	}
	if p, u := ShooterPeak(archer, d2.Hell); p != 68 || u {
		t.Fatalf("archer %d %v", p, u)
	}
	if p, _ := ShooterPeak(melee, d2.Hell); p != 0 {
		t.Fatal("melee has no shooter peak")
	}
	cat := MapCatalog{Monsters: map[int]*data.Monster{1: mage, 2: archer, 3: melee}}
	snap := &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{ID: 1, Difficulty: d2.Hell, Area: 5, X: 100, Y: 100},
		Monsters: []d2.Unit{{ID: 10, Class: 1, Mode: 1, X: 110, Y: 100}, {ID: 11, Class: 2, Mode: 1, X: 105, Y: 100}, {ID: 12, Class: 3, Mode: 1, X: 102, Y: 100}}}
	for _, c := range []struct {
		min  int
		want [3]bool
	}{{0, [3]bool{}}, {100, [3]bool{true, false, false}}, {130, [3]bool{}}} {
		hs := Hostiles(snap, HostileCtx{Cat: cat, Diff: d2.Hell, ShooterKillMin: c.min})
		var got [3]bool
		for _, h := range hs {
			got[h.Unit.ID-10] = h.KillFast
		}
		if got != c.want {
			t.Errorf("min %d: %v", c.min, got)
		}
	}
}

func TestThreatKillsFastShooterFirst(t *testing.T) {
	tbl := threatTable
	threatTable = testThreats()
	defer func() { threatTable = tbl }()
	cat := MapCatalog{Monsters: map[int]*data.Monster{7: {Key: "sk_archer", Killable: true, Ranged: true}, 3: {Key: "zombie", Killable: true}}}
	snap := &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{ID: 1, X: 100, Y: 100, Area: 5},
		Monsters: []d2.Unit{{ID: 10, Class: 7, Mode: 1, X: 112, Y: 100}, {ID: 11, Class: 3, Mode: 1, X: 102, Y: 100}},
		Missiles: []d2.Unit{{Class: 30, OwnerType: 1, OwnerID: 10, X: 103, Y: 100}}}
	for _, react := range []float64{0, 0.4} {
		hs := Hostiles(snap, HostileCtx{Cat: cat, ThreatReactSec: react})
		fast := map[uint32]bool{}
		for _, h := range hs {
			fast[h.Unit.ID] = h.KillFast
		}
		if fast[10] != (react > 0) || fast[11] {
			t.Errorf("react %v: %v", react, fast)
		}
	}
}

func TestRunnerAndStander(t *testing.T) {
	for ai, want := range map[string]bool{"Zombie": true, "Brute": true, "BloodHawk": true, "Fallen": false, "Skeleton": false, "Goatman": false} {
		if got := Runs(&data.Monster{AI: ai}); got != want {
			t.Errorf("%s runs %v", ai, got)
		}
	}
	if !Runs(&data.Monster{AI: "Other", Flying: true}) || Runs(nil) {
		t.Fatal("flyer / nil")
	}
	if !Stands(&data.Monster{AI: "Fallen"}) || Stands(&data.Monster{AI: "Goatman"}) || Stands(nil) {
		t.Fatal("stander")
	}
	hs := []Target{{Dist: 8, Mon: &data.Monster{AI: "Brute"}}, {Dist: 2, Mon: &data.Monster{AI: "Skeleton"}}}
	if !anyRunner(hs, 10) || anyRunner(hs, 5) {
		t.Fatal("anyRunner distance")
	}
}

func TestRunnerCloseKnob(t *testing.T) {
	cat := MapCatalog{Monsters: map[int]*data.Monster{3: {Key: "zombie1", AI: "Zombie", Killable: true}}}
	snap := &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{ID: 1, X: 100, Y: 100, Area: 5},
		Monsters: []d2.Unit{{ID: 11, Class: 3, Mode: 1, X: 102, Y: 100}}}
	for _, on := range []int{0, 1} {
		h := newHarness(nil, func(tn *env.Tuning) { tn.Combat.RunnerGuard = on })
		h.f.SetCatalog(cat)
		if got := h.f.runnerClose(snap, snap.Player, 4); got != (on == 1) {
			t.Errorf("guard %d: %v", on, got)
		}
	}
}
