package invdesk

import (
	"errors"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/itemreq"
	"d2lod-bot/internal/status"
)

// Judge is the equip path's structured answer (leveler.Leveler.Judge): the item's requirement and the item-level
// reasons for the hero and, when a mercenary is given, for him. The Equipper from Config may implement it; without it
// the view carries no req and no usable (docs/inventory-ui.md, "Requirements and usability").
type Judge interface {
	Judge(s *d2.Snapshot, u *d2.Unit, it itemreq.Item, m *itemreq.Merc) itemreq.Judgement
}

// usableInput is what the view knows when it decorates one item.
type usableInput struct {
	judge   Judge
	snap    *d2.Snapshot
	unit    *d2.Unit // the hero's unit; nil for an item only the mercenary wears
	item    status.Item
	place   string // pack, worn, stash, belt, cube, mercworn
	merc    *itemreq.Merc
	canWear int   // pack items: how many slots take it now (from CheckPlayerEquip, the same rule)
	wearWhy error // pack items with no slot: the refusal that explains it
	hasWhy  bool  // wearWhy was computed (the equip check ran)
}

// fillUsable sets iv.Req and iv.Usable. The hero's verdict for a backpack item IS CheckPlayerEquip's: ok exactly when
// a slot takes it. For every other place it is the item-level list of the same rule (Candidate). The mercenary's is
// merc.Allowed and merc.MeetsReq. Items that are no equipment (potions, gems, charms) get neither.
func fillUsable(iv *ItemView, in usableInput) {
	if in.judge == nil {
		return
	}
	j := in.judge.Judge(in.snap, in.unit, itemreq.Item{Code: in.item.Code, Identified: in.item.Identified, Broken: broken(in.item)}, in.merc)
	if !j.Wearable {
		return
	}
	req := j.Req
	iv.Req = &req
	us := &itemreq.Usable{}
	if in.unit != nil {
		us.Hero = heroSide(j, in)
	}
	if in.merc != nil {
		s := itemreq.Side{OK: true, Reasons: []itemreq.Reason{}}
		for _, r := range j.Merc {
			s.Reasons = append(s.Reasons, r)
			if !r.Warn {
				s.OK = false
			}
		}
		us.Merc = &s
	}
	iv.Usable = us
}

func heroSide(j itemreq.Judgement, in usableInput) *itemreq.Side {
	s := &itemreq.Side{Reasons: []itemreq.Reason{}}
	switch {
	case in.place == "pack" && in.hasWhy:
		s.OK = in.canWear > 0
		if !s.OK {
			s.Reasons = append(s.Reasons, j.Hero...)
			if len(s.Reasons) == 0 { // the item itself is fine: the slot rule refused (twohand, slot)
				s.Reasons = append(s.Reasons, slotReasons(in.wearWhy)...)
			}
		}
	default:
		s.OK = len(j.Hero) == 0
		s.Reasons = append(s.Reasons, j.Hero...)
	}
	if broken(in.item) {
		s.Reasons = append(s.Reasons, itemreq.Reason{Code: itemreq.Broken, Text: "broken: it does nothing until repaired", Warn: true})
	}
	return s
}

// slotReasons takes the structured reasons out of a CheckPlayerEquip refusal, or makes one from its text.
func slotReasons(err error) []itemreq.Reason {
	var e *itemreq.Error
	if errors.As(err, &e) {
		return e.Reasons
	}
	if err == nil {
		return nil
	}
	return []itemreq.Reason{{Code: itemreq.Slot, Text: err.Error()}}
}

func broken(it status.Item) bool { return it.MaxDur > 0 && it.Dur == 0 }

// mercFacts is the mercenary the judge measures against; nil when none is employed.
func mercFacts(m *status.MercInv) *itemreq.Merc {
	if m == nil || !m.Hired {
		return nil
	}
	return &itemreq.Merc{Class: m.Class, Level: m.Level, Str: m.Str, Dex: m.Dex, Alive: m.Alive}
}

// Judge passes the structured answer through from the click path, like CheckPlayerEquip: one rule for both paths.
func (v *VerbEquipper) Judge(s *d2.Snapshot, u *d2.Unit, it itemreq.Item, m *itemreq.Merc) itemreq.Judgement {
	if j, ok := v.Click.(Judge); ok {
		return j.Judge(s, u, it, m)
	}
	return itemreq.Judgement{}
}

// stashBlock is the item-level refusal (class, requirements, identified) of a stash item the hero could not wear once
// it is out of the stash; empty when nothing about the item stops him (or the equipper cannot judge). The same list
// the view's usable.hero reports, so Apply and the page agree.
func stashBlock(eq Equipper, s *d2.Snapshot, it status.Item) []itemreq.Reason {
	j, ok := eq.(Judge)
	if !ok || s == nil || s.Inventory == nil {
		return nil
	}
	for _, u := range s.Inventory.Items {
		if u.ID == it.ID {
			u := u
			// a stash item is judged as a pack item: the unidentified and item-level bounds apply to it the same way
			return j.Judge(s, &u, itemreq.Item{Code: it.Code, Identified: it.Identified}, nil).Hero
		}
	}
	return nil
}
