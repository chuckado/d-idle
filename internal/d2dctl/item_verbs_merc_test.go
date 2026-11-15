package d2dctl

import (
	"context"
	"errors"
	"testing"
)

// The merc verbs against a stub server, for two hireling types: the fork answers the type refusal per mercenary.
func TestMercVerbs(t *testing.T) {
	ctx := context.Background()
	s, c := startStub(t, map[string]string{
		"caps":                    "caps=item.equip,item.merc-equip,item.merc-unequip\nok\n",
		"item merc-equip 5":       "equip id=5 slot=1 displaced=-1 at=-1,-1\nok\n",
		"item merc-equip 6":       "noop slot=1\nok\n",
		"item merc-equip 7":       "err merctype a Desert Guard does not carry a bow\n",
		"item merc-equip 8":       "err merctype a Rogue Scout does not carry a spear\n",
		"item merc-equip 9":       "err dead the mercenary is dead\n",
		"item merc-equip 10":      "err requirements strength 40 needed, he has 31\n",
		"item merc-equip 11":      "err lab verbs off: item needs d2d --lab-verbs\n",
		"item merc-unequip head":  "unequip id=5 slot=1 displaced=-1 at=2,0\nok\n",
		"item merc-unequip torso": "err nospace no room in the backpack\n",
	}, nil)
	caps, _ := c.Caps(ctx)
	if !HasCap(caps, CapItemMercEquip) || !HasCap(caps, CapItemMercUnequip) {
		t.Fatalf("caps %v", caps)
	}
	if r, err := c.ItemMercEquip(ctx, 5); err != nil || r.Noop || r.Slot != 1 {
		t.Fatalf("equip %+v %v", r, err)
	}
	if noop, err := (ItemVerbs{C: c}).MercEquip(ctx, 6); err != nil || !noop {
		t.Fatalf("noop %v %v", noop, err)
	}
	for id, code := range map[uint32]string{7: ItemErrMercType, 8: ItemErrMercType, 9: ItemErrDead, 10: ItemErrRequirements} {
		_, err := c.ItemMercEquip(ctx, id)
		var ref *ItemRefused
		if !errors.As(err, &ref) || ref.Code != code || !ref.RuleRefusal() {
			t.Errorf("id %d: want code %s, got %v", id, code, err)
		}
	}
	_, err := c.ItemMercEquip(ctx, 11) // lab verbs off is not a rule refusal: the caller refuses with code merc
	if err == nil || errors.As(err, new(*ItemRefused)) {
		t.Errorf("lab verbs off: %v", err)
	}
	if r, err := c.ItemMercUnequip(ctx, "head"); err != nil || r.AtCol != 2 {
		t.Fatalf("unequip %+v %v", r, err)
	}
	if _, err := c.ItemMercUnequip(ctx, "torso"); err == nil || !errors.As(err, new(*ItemRefused)) {
		t.Errorf("nospace: %v", err)
	}
	n := len(s.seen)
	if _, err := c.ItemMercUnequip(ctx, "neck"); err == nil || len(s.seen) != n { // a bad slot never reaches the wire
		t.Errorf("bad slot: %v, sent %d more lines", err, len(s.seen)-n)
	}
}
