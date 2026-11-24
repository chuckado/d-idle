package plandata

import (
	"strings"
	"testing"
)

func TestBossLevelsRows(t *testing.T) {
	for _, c := range []string{"paladin", "sorceress", "necromancer", "amazon", "druid", "assassin", "barbarian"} {
		r, err := BossLevels(GateAndariel, c)
		if err != nil {
			t.Fatalf("%s: %v", c, err)
		}
		if r.RecommendedMerc > r.Recommended || r.FloorMerc > r.Floor || r.Floor > r.Recommended || r.FloorMerc > r.RecommendedMerc || r.FloorMerc <= 0 {
			t.Errorf("%s: inconsistent row %+v", c, r)
		}
	}
	// the user's recalibration of 2026-11-24 (a human kills her at clvl 9 to 10): floors 11 / 10, recommended 14 / 13, every class
	for _, c := range []string{"paladin", "sorceress", "necromancer"} {
		if r, _ := BossLevels(GateAndariel, c); r != (BossLevel{14, 13, 11, 10}) {
			t.Errorf("%s row %+v, want 14 / 13, floors 11 / 10", c, r)
		}
	}
}

func TestBossLevelsUnknownNamesIt(t *testing.T) {
	if _, err := BossLevels(GateAndariel, "warlock"); err == nil || !strings.Contains(err.Error(), "warlock") {
		t.Errorf("unknown class: %v", err)
	}
	if _, err := BossLevels("mephisto", "paladin"); err == nil || !strings.Contains(err.Error(), "mephisto") {
		t.Errorf("unknown gate: %v", err)
	}
	if LowestFloor(GateAndariel) != 10 {
		t.Errorf("lowest floor %d", LowestFloor(GateAndariel))
	}
}
