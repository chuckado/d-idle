package bot

import (
	"testing"

	"d2lod-bot/internal/d2"
)

type tomeCatalog struct{ fakeCatalog }

func (c tomeCatalog) Item(class int) (ItemMeta, bool) {
	if class == 6 {
		return ItemMeta{Name: "Tome of Town Portal", GoalType: "book", W: 1, H: 2}, true
	}
	return c.fakeCatalog.Item(class)
}

type statsReader struct {
	d2.StateReader
	st map[uint32]d2.Stats
}

func (r statsReader) UnitStats(a uint32) (d2.Stats, error) {
	if s, ok := r.st[a]; ok {
		return s, nil
	}
	return nil, d2.ErrUnsupported
}

func hasTP(t *testing.T, charges int, scrolls int) bool {
	g := newRig(t)
	g.r.d.Catalog = tomeCatalog{}
	st := map[uint32]d2.Stats{}
	if charges > 0 {
		st[100] = d2.Stats{70: int32(charges)}
	}
	g.r.e.Read = statsReader{st: st}
	g.w.set(func(w *world) {
		w.items = append(w.items, d2.Unit{ID: 900, Addr: 100, Type: d2.TypeItem, Class: 6, Item: &d2.Item{Where: d2.ContInventory}})
		for i := 0; i < scrolls; i++ {
			w.items = append(w.items, d2.Unit{ID: uint32(901 + i), Addr: uint32(200 + i), Type: d2.TypeItem, Class: 5, Item: &d2.Item{Where: d2.ContInventory}})
		}
	})
	s, _ := g.w.snapshot()
	return g.r.resInput(s).HasTP
}

// F5: an empty Tome of Town Portal is not a portal; a tome with charges or a scroll is.
func TestHasTPCountsCharges(t *testing.T) {
	if hasTP(t, 0, 0) {
		t.Error("empty tome counted as a portal")
	}
	if !hasTP(t, 0, 2) {
		t.Error("empty tome plus scrolls: scrolls are a portal source")
	}
	if !hasTP(t, 3, 0) {
		t.Error("tome with 3 charges is a portal source")
	}
}
