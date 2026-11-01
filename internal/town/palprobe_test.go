package town_test

import (
	"context"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/gear/geartest"
	"d2lod-bot/internal/town"
)

// PALPROBE part B (c): town behaviour for a Paladin on the sim rig (real item tables, so they skip without an
// install; the worker host runs them). Prices and vendor stock are the rig's, not the game's (U on d2d).

// A Paladin with an empty weapon hand and a two-hander in the shop buys the one-hander: a shield build never rebuys a
// two-hander (leveler.Prefs.WantsShield, town/rebuy.go).
func TestPalProbeRebuyNeverBuysTwoHander(t *testing.T) {
	geartest.Need(t)
	r := earlyRig(t, d2.Paladin, 4)
	r.W.Stock(r.Charsi, "2hs") // two-handed sword
	r.W.Stock(r.Charsi, "ssd") // short sword
	rep, err := r.T.Rebuy(context.Background(), 600)
	if err != nil {
		t.Fatal(err)
	}
	s, _ := r.W.Snapshot()
	w := s.Inventory.Equipped(d2.BodyRHand)
	if w == nil {
		t.Fatalf("a Paladin below clvl 15 needs a weapon and bought none: %+v", rep)
	}
	if w.Class == codeClass(r, "2hs") {
		t.Fatalf("a two-hander was bought over the shield slot: %+v", rep)
	}
}

// Repair is class-agnostic: a worn weapon at 28 percent durability is repaired for a melee class in the same visit
// (a weapon wears every swing, a staff almost never; the policy reads every worn item).
func TestPalProbeRepairTriggersForMeleeWeapon(t *testing.T) {
	r := audit(t, 200, 5, true)
	r.W.Player.Class = d2.Paladin
	r.W.RepairBtn = town.RepairAll
	for _, u := range r.W.Items {
		if u.Item.Where == d2.ContEquipment && u.Item.Body == d2.BodyRHand {
			r.W.Stats[u.Addr] = d2.Stats{72: 28, 73: 100}
		}
	}
	rep := run(t, r)
	if !rep.Repaired {
		t.Fatalf("a worn weapon at 28%% was not repaired: %+v", rep.Steps)
	}
}
