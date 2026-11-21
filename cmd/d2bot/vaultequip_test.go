package main

import (
	"context"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/invdesk"
)

type freshReader struct {
	d2.StateReader
	s *d2.Snapshot
}

func (f freshReader) Snapshot() (*d2.Snapshot, error) { return f.s, nil }

type okEquip struct{ id uint32 }

func (okEquip) CheckPlayerEquip(*d2.Snapshot, uint32, int) (string, error) { return "item", nil }
func (o *okEquip) PlayerEquip(_ context.Context, id uint32, _ int) error   { o.id = id; return nil }

func packSnap(class d2.Class, ids ...uint32) *d2.Snapshot {
	inv := &d2.Inventory{}
	for _, id := range ids {
		inv.Items = append(inv.Items, d2.Unit{ID: id, Type: d2.TypeItem, Item: &d2.Item{Where: d2.ContInventory}})
	}
	return &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Class: class}, Inventory: inv}
}

// F2 (pgvault live): the runner's snapshot predates the Keep that created the unit in the same tick; the equip hand must
// read the fresh one, else the desk refuses with "not in the inventory any more". Two classes, same path; the same test
// without the hold fails.
func TestEquipHandSeesTheUnitTheKeepJustMade(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		for _, slot := range []string{"head", "lring"} {
			ref := &runnerRef{}
			stale := packSnap(class, 1)
			eq := &okEquip{}
			desk, err := invdesk.New(invdesk.Config{Snap: ref.deskSnap}) // no runner: the stale view lists nothing
			if err != nil {
				t.Fatal(err)
			}
			desk.SetEquipper(eq)
			ref.inv = desk
			e := &env.Env{Read: freshReader{s: packSnap(class, 1, 77)}}
			h := equipHand{ref: ref, e: e, snap: func() *d2.Snapshot { return stale }}
			if _, err := desk.FitSlots(77); err == nil {
				t.Fatal("precondition: the stale snapshot must not list unit 77")
			}
			if _, err := h.Pick(77, slot); err != nil && slot == "head" {
				t.Fatalf("%v Pick: %v", class, err)
			}
			if _, err := h.Wear(context.Background(), 77, slot); err != nil {
				t.Fatalf("%v Wear %s: %v", class, slot, err)
			}
			if eq.id != 77 {
				t.Fatalf("%v: equipped %d, want 77", class, eq.id)
			}
			if ref.hold.Load() != nil {
				t.Fatal("the held snapshot leaked")
			}
		}
	}
}

// appearsLater lists unit 77 only after the bot clock stepped three times, like a created item under d2d's lockstep.
type appearsLater struct {
	freshReader
	class d2.Class
	steps int
}

func (a *appearsLater) Sleep(time.Duration) { a.steps++ }
func (a *appearsLater) Now() time.Time      { return time.Unix(1, 0) }
func (a *appearsLater) Snapshot() (*d2.Snapshot, error) {
	if a.steps >= 3 {
		return packSnap(a.class, 1, 77), nil
	}
	return packSnap(a.class, 1), nil
}

// F2, second cause: the Keep's create answers with the id before the game lists the item. The hand waits on the bot clock
// for the unit instead of refusing "not in the inventory any more". Two classes.
func TestEquipHandWaitsForTheCreatedUnit(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		w := &appearsLater{class: class}
		ref := &runnerRef{}
		desk, err := invdesk.New(invdesk.Config{Snap: ref.deskSnap})
		if err != nil {
			t.Fatal(err)
		}
		desk.SetEquipper(&okEquip{})
		ref.inv = desk
		h := equipHand{ref: ref, e: &env.Env{Clock: w, Read: w}, snap: func() *d2.Snapshot { return packSnap(class, 1) }}
		slot, err := h.Pick(77, "")
		if err != nil || slot == "" {
			t.Fatalf("%v: Pick %q %v (clock steps %d)", class, slot, err, w.steps)
		}
	}
}
