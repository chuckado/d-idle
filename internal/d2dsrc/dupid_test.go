package d2dsrc

import (
	"reflect"
	"testing"

	"d2lod-bot/internal/d2"
)

// The cap worn by an Equip order and the next picked potion once both carried unit id 4 (card recallsorc): the reader names
// the shared id so the log shows it, for a Paladin and a Sorceress alike.
func TestDuplicateCarriedIDs(t *testing.T) {
	for _, cls := range []d2.Class{d2.Paladin, d2.Sorceress} {
		inv := []d2.Unit{
			{ID: 1, Item: &d2.Item{Where: d2.ContEquipment}},
			{ID: 4, Item: &d2.Item{Where: d2.ContEquipment, Body: d2.BodyHead}},
			{ID: 4, Item: &d2.Item{Where: d2.ContInventory}},
			{ID: 0xffffffff, Item: &d2.Item{Where: d2.ContInventory}},
			{ID: 0xffffffff, Item: &d2.Item{Where: d2.ContInventory}},
		}
		if got := DuplicateCarriedIDs(inv); !reflect.DeepEqual(got, []uint32{4}) {
			t.Errorf("class %v: duplicates %v, want [4]", cls, got)
		}
		if got := DuplicateCarriedIDs(inv[:2]); len(got) != 0 {
			t.Errorf("class %v: unique ids reported as %v", cls, got)
		}
	}
}
