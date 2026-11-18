package main

import (
	"testing"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/merc"
)

func TestSupplyLightTwoClasses(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		base := glanceKit{Class: class, Level: 20, Diff: d2.Normal, Slots: 8, Portal: true}
		// the plan's own numbers, whatever the class: fill the belt to exactly the plan, then take one away
		full := base
		full.Heal, full.Mana = 1000, 1000
		if l := supplyLight(full); l.Light != "ok" || l.Word != "Stocked" {
			t.Errorf("%v full: %+v", class, l)
		}
		low := base
		low.Heal, low.Mana = 1, 1000
		if l := supplyLight(low); l.Light != "warn" || l.Word != "Low" {
			t.Errorf("%v low: %+v", class, l)
		}
		none := base
		none.Mana = 1000
		if l := supplyLight(none); l.Light != "bad" || l.Word != "No potions" {
			t.Errorf("%v no potions: %+v", class, l)
		}
		np := full
		np.Portal = false
		if l := supplyLight(np); l.Light != "bad" || l.Word != "No portal" {
			t.Errorf("%v no portal: %+v", class, l)
		}
	}
}

func TestSupplyLightDetailNamesPlan(t *testing.T) {
	a := supplyLight(glanceKit{Class: d2.Sorceress, Level: 20, Slots: 4, Portal: true, Heal: 1, Mana: 1})
	b := supplyLight(glanceKit{Class: d2.Sorceress, Level: 20, Slots: 12, Portal: true, Heal: 1, Mana: 1})
	if a.Detail == b.Detail {
		t.Errorf("belt size not in the plan numbers: %q", a.Detail)
	}
}

func TestGlanceSuppliesUnknown(t *testing.T) {
	cat := combat.MapCatalog{}
	for name, s := range map[string]*d2.Snapshot{"nil": nil, "no player": {}, "no inventory": {Player: &d2.Player{}}} {
		if l, ok := glanceSupplies(func() *d2.Snapshot { return s }, cat, nil, "")(); ok {
			t.Errorf("%s: want unknown, got %+v", name, l)
		}
	}
}

func TestGlanceSuppliesFromSnapshot(t *testing.T) {
	cat := combat.MapCatalog{Items: map[int]combat.ItemClass{1: {Portal: true}, 2: {Potion: combat.PotionHealing, Tier: 1}}}
	inv := &d2.Inventory{Items: []d2.Unit{
		{Class: 1, Item: &d2.Item{Where: d2.ContInventory}},
		{Class: 2, Item: &d2.Item{Where: d2.ContBelt, GX: 0}},
	}}
	s := &d2.Snapshot{Player: &d2.Player{Class: d2.Sorceress, Level: 10}, Inventory: inv}
	l, ok := glanceSupplies(func() *d2.Snapshot { return s }, cat, nil, "")()
	if !ok || l.Light != "warn" {
		t.Fatalf("one potion of a bigger plan: %+v %v", l, ok)
	}
}

func TestGlanceMerc(t *testing.T) {
	w := &mercWire{}
	for state, want := range map[string]string{"alive": "ok", "dead": "bad", "none": "warn"} {
		l, ok := glanceMerc(func() (merc.PlayerView, *mercWire) { return merc.PlayerView{State: state}, w })()
		if !ok || l.Light != want {
			t.Errorf("%s: %+v %v", state, l, ok)
		}
	}
	if _, ok := glanceMerc(func() (merc.PlayerView, *mercWire) { return merc.PlayerView{State: "alive"}, nil })(); ok {
		t.Error("no wiring must be unknown")
	}
	if _, ok := glanceMerc(func() (merc.PlayerView, *mercWire) { return merc.PlayerView{State: "off"}, w })(); ok {
		t.Error("unknown state must be unknown")
	}
}
