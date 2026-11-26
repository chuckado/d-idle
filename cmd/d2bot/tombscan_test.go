package main

import (
	"context"
	"testing"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/quest"
)

// The scan visits the tombs in order from the Canyon and stops in the one with the orifice; a hero already in a tomb with the
// orifice stays. Must fail while the Canyon is searched for the orifice (V drill.act2.duriel: "object class 152 not found
// (the whole level was walked)" after 12 minutes in the Canyon).
func TestFindTrueTombStopsAtTheTombWithTheOrifice(t *testing.T) {
	for _, truth := range []int{66, 69, 72} {
		cur := canyonOfMagi
		var visits []int
		exit := func(_ context.Context, to int) error { cur = to; visits = append(visits, to); return nil }
		err := findTrueTomb(context.Background(), func() int { return cur }, exit, func() bool { return cur == truth }, nil)
		if err != nil || cur != truth {
			t.Fatalf("truth %d: err %v ended in %d", truth, err, cur)
		}
		want := 2*(truth-66) + 1 // each wrong tomb is entered and left, then the right one
		if len(visits) != want {
			t.Fatalf("truth %d: visits %v", truth, visits)
		}
	}
	cur := 68
	err := findTrueTomb(context.Background(), func() int { return cur }, func(context.Context, int) error { t.Fatal("moved"); return nil }, func() bool { return true }, nil)
	if err != nil {
		t.Fatalf("already in the tomb: %v", err)
	}
	cur = 40
	if err := findTrueTomb(context.Background(), func() int { return cur }, func(context.Context, int) error { return nil }, func() bool { return false }, nil); err == nil {
		t.Fatal("a scan from town was accepted")
	}
}

// The orifice shows only after the tomb is explored: the scan explores each tomb before it moves on.
func TestFindTrueTombExploresEachTomb(t *testing.T) {
	cur := canyonOfMagi
	built := map[int]bool{}
	explored := 0
	exit := func(_ context.Context, to int) error { cur = to; return nil }
	has := func() bool { return cur == 68 && built[68] }
	explore := func(_ context.Context, stop func() bool) error { explored++; built[cur] = true; return nil }
	if err := findTrueTomb(context.Background(), func() int { return cur }, exit, has, explore); err != nil || cur != 68 {
		t.Fatalf("err %v in %d", err, cur)
	}
	if explored != 3 {
		t.Fatalf("explored %d tombs, want 3 (66, 67, 68)", explored)
	}
}

// The orifice preset of the level map is the point the walk goes to (the unit exists only once the room is up).
func TestOrificePresetIsTheWalkTarget(t *testing.T) {
	lm := &d2map.LevelMap{Presets: []d2map.Preset{
		{Kind: d2map.PresetObject, ID: 1, Pos: d2map.Point{X: 5, Y: 5}},
		{Kind: d2map.PresetObject, ID: quest.ObjOrifice, Pos: d2map.Point{X: 58, Y: 218}},
	}}
	if at, ok := orificePreset(lm); !ok || at.X != 58 || at.Y != 218 {
		t.Fatalf("%v %v", at, ok)
	}
	if _, ok := orificePreset(&d2map.LevelMap{}); ok {
		t.Fatal("an orifice from nowhere")
	}
}
