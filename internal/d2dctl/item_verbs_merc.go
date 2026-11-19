package d2dctl

import (
	"context"
	"fmt"
)

// Mercenary item verbs of the private fork (docs/research/merc-gear.md sections 2.2 and 4; B9). The fork plans the
// change on a copy with the server's own mercenary slot rule and then sends the game's mercenary item command. Wire:
//
//	item merc-equip <id>        give backpack item id to the mercenary (the server's rule picks the slot)
//	item merc-unequip <slot>    take what he wears in slot (head, torso, rhand, lhand) into the backpack
//
// Both answer like the hero verbs (an ok line with slot=, displaced=, at=). U: this is the contract the fork half is
// built to; no fork build answers it yet, so "caps" does not list the two names and every caller keeps refusing with
// code "merc" (docs/research/merc-gear.md status lines).
const (
	CapItemMercEquip   = "item.merc-equip"
	CapItemMercUnequip = "item.merc-unequip"
)

// ItemErrMercType is the refusal of a mercenary whose type does not carry that kind of item; a dead mercenary answers
// ItemErrDead ("dead"). Both are rule refusals.
const ItemErrMercType = "merctype"

// MercSlots are the slot names of item merc-unequip.
var MercSlots = [...]string{"head", "torso", "rhand", "lhand"}

// ItemMercEquip gives backpack item id to the mercenary. The change is applied at the next tick: the caller proves it
// from a fresh snapshot.
func (c *Client) ItemMercEquip(ctx context.Context, id uint32) (ItemResult, error) {
	return c.itemVerb(ctx, fmt.Sprintf("item merc-equip %d", id))
}

// ItemMercUnequip takes the item the mercenary wears in slot (one of MercSlots) into the backpack.
func (c *Client) ItemMercUnequip(ctx context.Context, slot string) (ItemResult, error) {
	for _, s := range MercSlots {
		if s == slot {
			return c.itemVerb(ctx, "item merc-unequip "+slot)
		}
	}
	return ItemResult{}, &ItemRefused{Code: ItemErrSlot, Why: fmt.Sprintf("%q is not a mercenary slot", slot)}
}

// MercEquip is ItemMercEquip in the shape internal/invdesk uses; noop reports "he wears it already".
func (v ItemVerbs) MercEquip(ctx context.Context, id uint32) (bool, error) {
	r, err := v.C.ItemMercEquip(ctx, id)
	return r.Noop, err
}

// MercUnequip is ItemMercUnequip in the shape internal/invdesk uses.
func (v ItemVerbs) MercUnequip(ctx context.Context, slot string) (bool, error) {
	r, err := v.C.ItemMercUnequip(ctx, slot)
	return r.Noop, err
}
