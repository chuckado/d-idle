package combat

import (
	"errors"
	"testing"

	"d2lod-bot/internal/d2"
)

func packWith(units ...d2.Unit) *d2.Snapshot {
	return &d2.Snapshot{Inventory: &d2.Inventory{Items: units}}
}

func pItem(id, class int, addr uint32) d2.Unit {
	return d2.Unit{ID: uint32(id), Addr: addr, Type: d2.TypeItem, Class: class, Item: &d2.Item{Where: d2.ContInventory}}
}

const itTome = 530

func pickCat() MapCatalog {
	return MapCatalog{Items: map[int]ItemClass{itTSC: {Portal: true}, itTome: {Portal: true, Tome: true}}}
}

// F4/F5: an empty tome (no stat list, d2d: qty 0) with two scrolls behind it: the pick is a scroll, never the tome.
func TestPortalSourceSkipsEmptyTome(t *testing.T) {
	stats := func(addr uint32) (d2.Stats, error) {
		if addr == 1 {
			return d2.Stats{}, nil // the empty tome
		}
		return nil, errors.New("scroll has no stat list")
	}
	s := packWith(pItem(1, itTome, 1), pItem(2, itTSC, 2), pItem(3, itTSC, 3))
	if u := PortalSource(s, pickCat(), stats); u == nil || u.ID != 2 {
		t.Fatalf("pick %v, want the first scroll (id 2)", u)
	}
}

func TestPortalSourceNoneForEmptyTomeAlone(t *testing.T) {
	stats := func(uint32) (d2.Stats, error) { return d2.Stats{}, nil }
	if u := PortalSource(packWith(pItem(1, itTome, 1)), pickCat(), stats); u != nil {
		t.Fatalf("empty tome picked: %v", u)
	}
}

func TestPortalSourcePrefersTomeWithCharges(t *testing.T) {
	stats := func(uint32) (d2.Stats, error) { return d2.Stats{statQuantity: 4}, nil }
	if u := PortalSource(packWith(pItem(1, itTome, 1), pItem(2, itTSC, 2)), pickCat(), stats); u == nil || u.ID != 1 {
		t.Fatalf("pick %v, want the tome with charges", u)
	}
}

// The combat picker and the town portal pick are the same function: they agree on every pack.
func TestFighterAndPortalSourceAgree(t *testing.T) {
	h, g := tomeHarness(t, -1)
	s, _ := g.snap()
	if h.f.portalItem(s) != PortalSource(s, h.f.cat, h.f.stats()) {
		t.Fatal("picks differ")
	}
}
