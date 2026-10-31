package main

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
)

type portalReader struct {
	d2.StateReader
	snap *d2.Snapshot
}

func (r portalReader) Snapshot() (*d2.Snapshot, error)    { return r.snap, nil }
func (r portalReader) UnitStats(uint32) (d2.Stats, error) { return d2.Stats{}, nil } // qty absent: empty tome

// F4: with an empty tome and no scroll the portal verb reports no source without a click (In is nil: a click panics)
// and the caller's fallback treats it as no portal at once.
func TestD2DPortalEmptyTomeNoSourceNoClick(t *testing.T) {
	cat := combat.MapCatalog{Items: map[int]combat.ItemClass{1: {Portal: true, Tome: true}}}
	s := &d2.Snapshot{Player: &d2.Player{}, Inventory: &d2.Inventory{Items: []d2.Unit{
		{ID: 5, Type: d2.TypeItem, Class: 1, Item: &d2.Item{Where: d2.ContInventory}},
	}}}
	e := &env.Env{Read: portalReader{snap: s}, Log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	err := d2dPortal(e, nil, cat)(t.Context())
	if !errors.Is(err, combat.ErrNoPortal) {
		t.Fatalf("err %v, want ErrNoPortal", err)
	}
}
