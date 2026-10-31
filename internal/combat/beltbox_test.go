package combat

import (
	"testing"

	"d2lod-bot/internal/d2"
)

// BELTFIX F7: d2d reports a belt potion's box 0..15 in GX with GY 0; Game.exe reports (column,row). Combat must see
// every box on both: column = box%4, the lowest box of a column is its front.
func TestBeltOfSeesEveryBox(t *testing.T) {
	cat := testCatalog()
	hp := func(box, gy int) d2.Unit {
		u := potion(itHP1, box)
		u.Item.GY = gy
		return u
	}
	// d2d: an HP potion in box 5 (column 1, row 1) is visible
	b := BeltOf(&d2.Inventory{Items: []d2.Unit{hp(5, 0)}}, cat)
	if b.Count(PotionHealing) != 1 || b.Cols[1].Count != 1 {
		t.Fatalf("d2d box 5: %+v", b)
	}
	// d2d: all sixteen boxes
	var all []d2.Unit
	for i := 0; i < 16; i++ {
		all = append(all, hp(i, 0))
	}
	if n := BeltOf(&d2.Inventory{Items: all}, cat).Count(PotionHealing); n != 16 {
		t.Fatalf("16 boxes: combat sees %d", n)
	}
	// Game.exe: (column 1, row 1) is the same box 5
	if c := BeltOf(&d2.Inventory{Items: []d2.Unit{hp(1, 1)}}, cat); c.Cols[1].Count != 1 {
		t.Fatalf("exe (1,1): %+v", c)
	}
	// the front potion of a column decides its kind: MP in box 1, HP above it in box 5
	mp := potion(itMP1, 1)
	m := BeltOf(&d2.Inventory{Items: []d2.Unit{hp(5, 0), mp}}, cat)
	if m.Cols[1].Kind != PotionMana || !m.Cols[1].Mixed {
		t.Fatalf("mixed column: %+v", m.Cols[1])
	}
}
