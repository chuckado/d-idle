package leveler

import (
	"testing"

	"d2lod-bot/internal/data"
	"d2lod-bot/internal/merc"
)

// MercSlot is the vault compare toggle's merc rule: the slot a merc would wear the item in and whether his kind may.
// The made-up bases use the type code as the base type (IsA is an equality test for them).
func TestMercSlot(t *testing.T) {
	db := &data.DB{}
	base := func(kind data.ItemKind, typ string, two bool) *data.ItemBase {
		return &data.ItemBase{Kind: kind, Type: typ, TwoHanded: two, Name: typ, Code: typ}
	}
	helm, tors, shld := base(data.KindArmor, "helm", false), base(data.KindArmor, "tors", false), base(data.KindArmor, "shld", false)
	bow, spea, swor := base(data.KindWeapon, "bow", false), base(data.KindWeapon, "spea", true), base(data.KindWeapon, "swor", false)
	ring, boot := base(data.KindMisc, "ring", false), base(data.KindArmor, "boot", false)
	for _, tc := range []struct {
		label string
		class int
		b     *data.ItemBase
		key   string
		ok    bool
	}{
		{"rogue helm", merc.ClassRogue, helm, "mhead", true},
		{"desert armour", merc.ClassDesert, tors, "mtorso", true},
		{"rogue bow", merc.ClassRogue, bow, "mrhand", true},
		{"rogue sword", merc.ClassRogue, swor, "mrhand", false},
		{"desert spear", merc.ClassDesert, spea, "mrhand", true},
		{"iron wolf shield", merc.ClassIronWolf, shld, "mlhand", true},
		{"rogue shield", merc.ClassRogue, shld, "mlhand", false},
		{"barbarian one-hand sword", merc.ClassBarb, swor, "mrhand", true},
		{"ring: no merc slot", merc.ClassIronWolf, ring, "", false},
		{"boots: no merc slot", merc.ClassIronWolf, boot, "", false},
	} {
		key, ok := MercSlot(db, tc.class, tc.b)
		if key != tc.key || ok != tc.ok {
			t.Errorf("%s: got (%q, %v), want (%q, %v)", tc.label, key, ok, tc.key, tc.ok)
		}
	}
	if _, ok := MercSlot(nil, merc.ClassRogue, helm); ok {
		t.Error("a nil table must answer false")
	}
}
