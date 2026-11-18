package invdesk

import (
	"context"
	"errors"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/itemreq"
	"d2lod-bot/internal/status"
)

// Mercenary item verbs (B9, docs/research/merc-gear.md section 4): the player gives an item to the mercenary or takes
// one off through the fork's merc verbs, on the fast lane only. They are never queued: the bot's own executors (the click
// path) know nothing about his panel, so a request the lane cannot run now is refused with its reason instead of waiting.
// Nothing here decides what he wears: every call is the player's explicit request from the page (docs/gear-policy.md).
//
// Without the capability (a backend that does not list item.merc-equip, the knobs off, the lane closed) the desk answers
// exactly as before: code "merc", "the bot cannot change the mercenary's gear yet".

// Capability names of the merc verbs (d2dctl.CapItemMercEquip, CapItemMercUnequip).
const (
	CapItemMercEquip   = "item.merc-equip"
	CapItemMercUnequip = "item.merc-unequip"
)

// MercVerbs is what the lane needs of the backend for the mercenary. The slot of an unequip is a hireling slot name.
type MercVerbs interface {
	MercEquip(ctx context.Context, id uint32) (noop bool, err error)
	MercUnequip(ctx context.Context, slot string) (noop bool, err error)
}

// mercSlotName maps the page's merc slot keys to the fork's slot names.
var mercSlotName = map[string]string{"mhead": "head", "mtorso": "torso", "mrhand": "rhand", "mlhand": "lhand"}

var errNoMercVerbs = errors.New("this backend does not offer the mercenary item verbs")

// MercEquip and MercUnequip hold the channel like every other verb, so a lane request cannot interleave with a bot verb.
// A backend whose client does not implement MercVerbs answers an error (the lane checks the capability first).
func (s Serial) MercEquip(ctx context.Context, id uint32) (bool, error) {
	mv, ok := s.V.(MercVerbs)
	if !ok {
		return false, errNoMercVerbs
	}
	rel, err := s.hold(ctx)
	if err != nil {
		return false, err
	}
	defer rel()
	return mv.MercEquip(ctx, id)
}

func (s Serial) MercUnequip(ctx context.Context, slot string) (bool, error) {
	mv, ok := s.V.(MercVerbs)
	if !ok {
		return false, errNoMercVerbs
	}
	rel, err := s.hold(ctx)
	if err != nil {
		return false, err
	}
	defer rel()
	return mv.MercUnequip(ctx, slot)
}

var _ MercVerbs = Serial{}

// mercEquip gives backpack item id to the mercenary (slot "" or one of mhead, mtorso, mrhand, mlhand: the server's rule
// picks the slot, a named slot is checked against it after the fact). unequip: id 0, slot names what comes off.
func (d *Desk) mercVerb(equip bool, id uint32, slot string) (Request, error) {
	d.mu.Lock()
	f, eq := d.fast, d.eq
	d.mu.Unlock()
	old := func() (Request, error) { // nothing changes while the verbs are not on offer
		var data status.InventoryData
		if d.c.Inventory != nil {
			data = d.c.Inventory()
		}
		return Request{}, refuse("merc", "%s", d.mercWhy(data))
	}
	if f == nil || f.Verbs == nil || f.Instant == nil || !f.Instant() || f.On == nil || !f.On() {
		return old()
	}
	mv, ok := f.Verbs.(MercVerbs)
	if !ok {
		return old()
	}
	capName := CapItemMercUnequip
	if equip {
		capName = CapItemMercEquip
	}
	wait := f.Wait
	if wait <= 0 {
		wait = DefaultFastWait
	}
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	if !f.capSet(ctx)[capName] {
		return old()
	}

	// From here the verbs are on offer: every refusal says why.
	var data status.InventoryData
	if d.c.Inventory != nil {
		data = d.c.Inventory()
	}
	m := data.Merc
	switch {
	case m == nil || !m.Hired:
		return Request{}, refuse("merc", "no mercenary is hired")
	case !m.Alive:
		return Request{}, refuse(itemreq.Dead, "the mercenary is dead: revive him first, the game does not change a dead mercenary's gear")
	}
	s := d.snap()
	if why := f.why(s); why != "" {
		return Request{}, refuse("busy", "%s", why)
	}
	if !s.Player.InTown() {
		return Request{}, refuse("town", "the mercenary's gear changes in town")
	}
	if slot != "" {
		if _, ok := mercSlotName[slot]; !ok {
			return Request{}, refuse("slot", "unknown mercenary slot %q", slot)
		}
	}

	r := Request{ID: id, Kind: KindEquip, Target: TargetMerc, Slot: slot, State: StateDoing, At: d.c.Now().Unix(), How: HowFast}
	var run func(ctx context.Context) error
	if equip {
		u, _, ok := d.unit(id)
		if !ok {
			return Request{}, refuse("item", "that item is not in the inventory any more")
		}
		if u.Item.Where != d2.ContInventory {
			return Request{}, refuse("place", "only backpack items can be given to the mercenary: move it to the backpack first")
		}
		it, ok := packItem(data, id)
		if !ok {
			return Request{}, refuse("item", "that item is not in the inventory any more")
		}
		r.Name = itemName(it)
		if err := d.mercLegal(eq, s, u, it, m); err != nil {
			return Request{}, err
		}
		run = func(ctx context.Context) error {
			noop, err := mv.MercEquip(ctx, id)
			if err != nil || noop {
				return err
			}
			return f.settle(ctx, func(s *d2.Snapshot) bool { return mercWears(s, id) })
		}
	} else {
		it := m.Equipment[slot]
		if slot == "" || it == nil {
			return Request{}, refuse("slot", "nothing is worn there")
		}
		r.Kind, r.ID, r.Name = KindUnequip, it.ID, itemName(*it)
		run = func(ctx context.Context) error {
			noop, err := mv.MercUnequip(ctx, mercSlotName[slot])
			if err != nil || noop {
				return err
			}
			return f.settle(ctx, func(s *d2.Snapshot) bool {
				u, ok := findUnit(s, it.ID)
				return ok && u.Item != nil && u.Item.Where == d2.ContInventory
			})
		}
	}

	d.mu.Lock()
	if d.doingLocked() {
		d.mu.Unlock()
		return Request{}, refuse("busy", "an action is running")
	}
	d.mu.Unlock()

	var sent bool
	err := run(context.WithValue(ctx, sentKey{}, &sent))
	code, why, final := classify(err)
	if err != nil && !final { // the verb did not run (or is safe to repeat): there is no queue to fall back to
		code, why = "merc", "the mercenary item verb did not run ("+why+"): try again"
		if sent {
			code, why = "timeout", "no answer in time and the change may have happened; check what he wears and ask again"
		}
		f.dropCaps()
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if err != nil {
		r.State, r.Reason, r.Code = StateFailed, why, code
		d.recentLocked(r)
		_ = d.saveLocked()
		return r, refuse(code, "%s", why)
	}
	r.State = StateDone
	d.recentLocked(r)
	_ = d.saveLocked()
	return r, nil
}

// mercLegal is the page's own verdict for the item and this mercenary (usable.merc: merc.Allowed and merc.MeetsReq), so a
// request the page marks red is refused here with the same words. An Equipper that cannot judge leaves the verdict to the
// fork's rule (it refuses an illegal item itself).
func (d *Desk) mercLegal(eq Equipper, s *d2.Snapshot, u d2.Unit, it status.Item, m *status.MercInv) error {
	j, ok := eq.(Judge)
	if !ok {
		return nil
	}
	jd := j.Judge(s, nil, itemreq.Item{Code: it.Code, Identified: it.Identified, Broken: broken(it)}, mercFacts(m))
	if !jd.Wearable {
		return refuse(itemreq.Slot, "that is not equipment: a mercenary cannot wear it")
	}
	for _, r := range jd.Merc {
		if !r.Warn {
			return refuse(r.Code, "%s", r.Text)
		}
	}
	return nil
}

// mercWears: the mercenary's list shows the item (or, when his gear cannot be read, the item left the backpack).
func mercWears(s *d2.Snapshot, id uint32) bool {
	if s == nil {
		return false
	}
	if s.Merc != nil && s.Merc.GearKnown {
		for _, g := range s.Merc.Gear {
			if g.ID == id {
				return true
			}
		}
		return false
	}
	u, ok := findUnit(s, id)
	return !ok || u.Item == nil || u.Item.Where != d2.ContInventory
}

func packItem(data status.InventoryData, id uint32) (status.Item, bool) {
	for _, it := range data.Inventory {
		if it.ID == id {
			return it, true
		}
	}
	return status.Item{}, false
}
