package invdesk

import (
	"context"
	"errors"
	"strings"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/status"
)

// Synchronous equip for the vault's one-tap Equip (internal/vaultkeep, docs/vault.md "Equip"). The queue path (requestEquip,
// RunPending) stays the player's Inventory page; this is the same checks and the same executor (the verb, else the click
// path, both verified) called at once by the bot's own town step, ending in the same slot pin.

// FitSlots lists the slot keys that would accept pack item id now, in status.EquipSlots order, and the first refusal when
// none does.
func (d *Desk) FitSlots(id uint32) ([]string, error) {
	d.mu.Lock()
	eq := d.eq
	d.mu.Unlock()
	if eq == nil {
		return nil, refuse("off", "equipping is not available on this bot")
	}
	u, s, ok := d.unit(id)
	if !ok {
		return nil, refuse("item", "that item is not in the inventory any more")
	}
	if u.Item.Where != d2.ContInventory {
		return nil, refuse("place", "only backpack items can be equipped")
	}
	var fits []string
	var why error
	for _, k := range status.EquipSlots {
		_, err := eq.CheckPlayerEquip(s, id, SlotBody[k])
		switch {
		case err == nil:
			fits = append(fits, k)
		case why == nil || !isNoSlot(err):
			why = err
		}
	}
	if len(fits) == 0 {
		if why == nil {
			why = errors.New("it cannot be worn")
		}
		return nil, why
	}
	return fits, nil
}

// EquipNow wears pack item id in slot and pins the slot (the bot's own gear logic then leaves it alone). A shield that blocks
// a two-hander is taken off first (the player's unequip path); it lands in the backpack like the item that was worn.
func (d *Desk) EquipNow(ctx context.Context, id uint32, slot string) error {
	d.mu.Lock()
	eq, un := d.eq, d.un
	d.mu.Unlock()
	body, ok := SlotBody[slot]
	if !ok {
		return refuse("slot", "unknown slot %q", slot)
	}
	if eq == nil {
		return refuse("off", "equipping is not available on this bot")
	}
	_, s, found := d.unit(id)
	if !found {
		return refuse("item", "that item is not in the inventory any more")
	}
	name, err := eq.CheckPlayerEquip(s, id, body)
	if err != nil && strings.Contains(err.Error(), "cannot be worn together") && un != nil {
		other := map[int]int{d2.BodyRHand: d2.BodyLHand, d2.BodyLHand: d2.BodyRHand}[body]
		for _, x := range s.Inventory.Items {
			if x.Item != nil && x.Item.Where == d2.ContEquipment && x.Item.Body == other {
				if uerr := un.PlayerUnequip(ctx, x.ID); uerr != nil {
					return uerr
				}
				d.dropPinSlotBody(other)
				break
			}
		}
		name, err = "", nil
	}
	if err != nil {
		return refuse("requirements", "%v", err)
	}
	if err := eq.PlayerEquip(ctx, id, body); err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.pinLocked(Pin{Slot: slot, ID: id, Name: name, Since: d.c.Now().Unix()})
	return d.saveLocked()
}

func (d *Desk) dropPinSlotBody(body int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for k, b := range SlotBody {
		if b == body {
			d.dropPinLocked(k)
		}
	}
}
