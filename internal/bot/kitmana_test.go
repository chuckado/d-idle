package bot

import (
	"testing"

	"d2lod-bot/internal/d2"
)

// V lab 2026-11-23 (necro scenarios): the kit asks kitMinMP (3) mana potions in the belt, a Necromancer from clvl 31 has one
// mana column and a Sorceress below it two, so the kit was never complete and the hero went home on every arrival
// ("mana potions 1/3"). The minimum is capped by the belt plan. Two builds. Must fail while the minimum is the raw knob.
func TestKitManaMinimumIsCappedByTheBeltPlan(t *testing.T) {
	for _, tc := range []struct {
		class       d2.Class
		level, mana int
	}{{d2.Necromancer, 40, 1}, {d2.Sorceress, 20, 2}} {
		g := newRig(t)
		g.w.set(func(w *world) { w.level, w.area, w.gold0, w.class = tc.level, 4, 900, tc.class })
		ids := []uint32{100, 101, 102, 103}
		g.w.belt(ids[:2]...)
		g.w.set(func(w *world) {
			for i := 0; i < tc.mana; i++ {
				w.items = append(w.items, d2.Unit{ID: uint32(200 + i), Type: d2.TypeItem, Class: 8, Item: &d2.Item{Where: d2.ContBelt}})
			}
			for i := 0; i < 2; i++ { // the pack holds the rest of the healing kit
				w.items = append(w.items, d2.Unit{ID: uint32(300 + i), Type: d2.TypeItem, Class: 1, Item: &d2.Item{Where: d2.ContInventory}})
			}
		})
		s, _ := g.w.snapshot()
		k, known := g.r.kitRead(s)
		if !known {
			t.Fatalf("%v: kit unknown", tc.class)
		}
		if k.minMP != tc.mana || k.MP != tc.mana {
			t.Errorf("%v clvl %d: minMP %d (want %d), belt mana %d", tc.class, tc.level, k.minMP, tc.mana, k.MP)
		}
		for _, m := range k.missing() {
			if len(m) > 4 && m[:4] == "mana" {
				t.Errorf("%v: kit still missing %q", tc.class, m)
			}
		}
	}
}
