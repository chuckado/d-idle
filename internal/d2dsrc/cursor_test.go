package d2dsrc

import (
	"testing"

	"d2lod-bot/internal/d2"
)

func TestCursorFromHeldCode(t *testing.T) {
	s := &Source{}
	mk := func(held string, items ...d2.Unit) *builder {
		return &builder{s: s, rp: newRep(&s.trk), snap: &d2.Snapshot{Inventory: &d2.Inventory{Items: items}}, held: held}
	}
	u := d2.Unit{ID: 6, Class: -1, Item: &d2.Item{Where: d2.ContInventory}}
	mk("", u).cursor() // remember the pack
	b := mk("tsc")     // item 6 gone, held
	b.cursor()
	if c := b.snap.Inventory.Cursor; c == nil || c.ID != 6 || c.Item.Where != d2.ContCursor {
		t.Fatalf("cursor %+v", c)
	}
	b = mk("tsc") // still held next read
	b.cursor()
	if b.snap.Inventory.Cursor == nil {
		t.Fatal("cursor lost while held")
	}
	b = mk("", u) // put back
	b.cursor()
	if b.snap.Inventory.Cursor != nil {
		t.Fatal("cursor set after the put")
	}
}

// The item list can lose the unit one read before "held" shows its code (Questy 2026-11-26: the equip click worked, the cursor
// never showed). Two items stand for two bases: a vanished unit stays a cursor candidate for a few reads, the latest one first,
// and an item gone for longer than that is not taken for the cursor item.
func TestCursorSurvivesTheHeldLag(t *testing.T) {
	for _, id := range []uint32{5, 9} {
		s := &Source{}
		mk := func(held string, items ...d2.Unit) *builder {
			return &builder{s: s, rp: newRep(&s.trk), snap: &d2.Snapshot{Inventory: &d2.Inventory{Items: items}}, held: held}
		}
		u := d2.Unit{ID: id, Class: -1, Item: &d2.Item{Where: d2.ContInventory}}
		mk("", u).cursor()
		mk("").cursor() // the list lost the unit, held is still empty
		b := mk("lea")
		b.cursor()
		if c := b.snap.Inventory.Cursor; c == nil || c.ID != id {
			t.Fatalf("id %d: cursor %+v after the held lag", id, c)
		}
		// an item that left the list long ago is not the held one
		s2 := &Source{}
		mk2 := func(held string, items ...d2.Unit) *builder {
			return &builder{s: s2, rp: newRep(&s2.trk), snap: &d2.Snapshot{Inventory: &d2.Inventory{Items: items}}, held: held}
		}
		mk2("", u).cursor()
		for i := 0; i < cursorMemory+2; i++ {
			mk2("").cursor()
		}
		b = mk2("lea")
		b.cursor()
		if b.snap.Inventory.Cursor != nil {
			t.Fatalf("id %d: a long gone unit became the cursor item", id)
		}
	}
}
