package bot

import (
	"testing"

	"d2lod-bot/internal/d2"
)

func TestSheetOf(t *testing.T) {
	// nothing read: every value absent, never a zero shown as a value
	h := sheetOf(&d2.Player{})
	if h.HasAttrs || h.HasDefense || h.HasDamage {
		t.Fatalf("empty player: %+v", h)
	}
	p := &d2.Player{Strength: 38, Dexterity: 25, Vitality: 30, Energy: 20, Stats: d2.Stats{31: 214, 21: 4, 22: 9}}
	h = sheetOf(p)
	if !h.HasAttrs || h.Str != 38 || h.Energy != 20 || !h.HasDefense || h.Defense != 214 || !h.HasDamage || h.DmgMin != 4 || h.DmgMax != 9 {
		t.Fatalf("full player: %+v", h)
	}
	// a stat list with max damage 0 carries no damage
	if sheetOf(&d2.Player{Stats: d2.Stats{21: 0, 22: 0}}).HasDamage {
		t.Fatal("zero damage reported")
	}
}
