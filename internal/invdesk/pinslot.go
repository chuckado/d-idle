package invdesk

// PinSlot pins a slot to the unit now worn there, for an item the direct equip verbs put on the body (no pack step, so no
// EquipNow): the gear logic must not undo the player's pick (the pin has no expiry, as after EquipNow).
func (d *Desk) PinSlot(slot string, id uint32, name string) error {
	if _, ok := SlotBody[slot]; !ok {
		return refuse("slot", "unknown slot %q", slot)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.pinLocked(Pin{Slot: slot, ID: id, Name: name, Since: d.c.Now().Unix()})
	return d.saveLocked()
}
