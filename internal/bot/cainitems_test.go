package bot

import (
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/task"
)

type cainCatalog struct{ fakeCatalog }

// The two chain items by code with wrong or localised display names.
func (cainCatalog) Item(class int) (ItemMeta, bool) {
	switch class {
	case 901:
		return ItemMeta{Name: "Parchemin d'Inifuss", Code: "bks"}, true
	case 902:
		return ItemMeta{Name: "Cle des cairns", Code: "bkd"}, true
	case 903:
		return ItemMeta{Name: "Figurine de jade", Code: "j34"}, true
	case 904:
		return ItemMeta{Name: "Oiseau dore", Code: "g34"}, true
	}
	return fakeCatalog{}.Item(class)
}

// Quest items are read by base code, not display name (RC3 of the 2026-11-17 Search for Cain diagnosis). The check
// is class-free: it holds for any hero's pack.
func TestScanQuestItemsByCode(t *testing.T) {
	r := &Runner{d: Deps{Catalog: cainCatalog{}}}
	for _, class := range []string{"Paladin", "Sorceress"} {
		pack := func(cls ...int) *d2.Snapshot {
			var it []d2.Unit
			for _, c := range cls {
				it = append(it, d2.Unit{Class: c, Item: &d2.Item{Where: d2.ContInventory}})
			}
			return &d2.Snapshot{Inventory: &d2.Inventory{Items: it}}
		}
		out := map[string]int{}
		r.scanQuestItems(pack(901), out)
		if out[task.ItemBark] != 1 || out[task.ItemCainSc] != 0 {
			t.Fatalf("%s: bks must read as the bark: %v", class, out)
		}
		out = map[string]int{}
		r.scanQuestItems(pack(902, 2), out)
		if out[task.ItemCainSc] != 1 || out[task.ItemBark] != 0 {
			t.Fatalf("%s: bkd must read as the key: %v", class, out)
		}
	}
}

// The Act 3 Golden Bird items are read by code too: a localised or differently named Jade Figurine must not read as absent
// (drill.act3s.golden_bird "not startable: no Jade Figurine in the pack" with the figurine in the pack). Two builds.
func TestScanQuestItemsReadsTheGoldenBirdByCode(t *testing.T) {
	r := &Runner{d: Deps{Catalog: cainCatalog{}}}
	for _, class := range []string{"Paladin", "Sorceress"} {
		snap := &d2.Snapshot{Inventory: &d2.Inventory{Items: []d2.Unit{
			{Class: 903, Item: &d2.Item{Where: d2.ContInventory}}, {Class: 904, Item: &d2.Item{Where: d2.ContInventory}}}}}
		out := map[string]int{}
		r.scanQuestItems(snap, out)
		if out[task.ItemJade] != 1 || out[task.ItemGolden] != 1 {
			t.Fatalf("%s: %v", class, out)
		}
	}
}
