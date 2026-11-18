package leveler

import (
	"fmt"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/itemreq"
	"d2lod-bot/internal/merc"
)

// Judge is the view's access to the equip path's own rule: the item's requirement, the hero's item-level reasons
// (Candidate, the same list CheckPlayerEquip refuses on) and, when m is not nil, the mercenary's (merc.Allowed and
// merc.MeetsReq, the merc policy's own functions). u is the hero's unit, nil for an item only the mercenary wears
// (his id space is not in the snapshot); then only the merc side and the requirement are filled.
func (l *Leveler) Judge(s *d2.Snapshot, u *d2.Unit, it itemreq.Item, m *itemreq.Merc) itemreq.Judgement {
	var j itemreq.Judgement
	if l.e.Data == nil {
		return j
	}
	b := l.e.Data.ItemByCode(it.Code)
	if b == nil {
		return j
	}
	j.Wearable = b.Kind == data.KindWeapon || b.Kind == data.KindArmor ||
		l.e.Data.IsA(b.Type, "ring") || l.e.Data.IsA(b.Type, "amul")
	if u != nil && u.Item != nil && s != nil && s.Player != nil {
		c := l.cand(s.Player, *u)
		j.Req, j.Hero = c.Req, c.Reasons
	} else {
		j.Req = itemreq.Req{Lvl: b.LevelReq, Str: b.ReqStr, Dex: b.ReqDex}
		if rc := classOf(l.e.Data, b.Type, map[string]bool{}); rc != "" {
			j.Req.Class = []string{classDisplay(l.e.Data, rc)}
		}
	}
	if m != nil && j.Wearable {
		j.Merc = l.judgeMerc(b, it, *m, j.Req.Approx)
	}
	return j
}

// mercKind maps a base item to the kind merc.Allowed knows. Two-handed weapons, javelins, axes and pelts stay KindOther:
// the decompiled rule does not carry them (V for the Rogue's bow and the Desert Guard's spear and polearm; javelins,
// two-handed swords, axes and Druid pelts are U, docs/research/merc-gear.md 2.3, and are refused).
func mercKind(db *data.DB, b *data.ItemBase) (merc.Kind, merc.Slot) {
	switch {
	case db.IsA(b.Type, "helm") && !db.IsA(b.Type, "pelt"), db.IsA(b.Type, "circ"):
		return merc.KindHelm, merc.SlotHelm
	case db.IsA(b.Type, "tors"):
		return merc.KindArmor, merc.SlotArmor
	case db.IsA(b.Type, "shld"):
		return merc.KindShield, merc.SlotShield
	case b.Kind != data.KindWeapon || b.Stackable || db.IsA(b.Type, "jave"): // javelins and throwing weapons: U, refused
		return merc.KindOther, merc.SlotWeapon
	case db.IsA(b.Type, "bow") && !db.IsA(b.Type, "xbow"):
		return merc.KindBow, merc.SlotWeapon
	case db.IsA(b.Type, "spea"), db.IsA(b.Type, "pole"): // two-handed in the tables, and what the Desert Guard carries
		return merc.KindSpear, merc.SlotWeapon
	case b.TwoHanded: // two-handed swords and axes: U, refused
		return merc.KindOther, merc.SlotWeapon
	case db.IsA(b.Type, "swor"):
		return merc.KindSword1H, merc.SlotWeapon
	case db.IsA(b.Type, "axe"):
		return merc.KindAxe1H, merc.SlotWeapon
	}
	return merc.KindOther, merc.SlotWeapon
}

func (l *Leveler) judgeMerc(b *data.ItemBase, it itemreq.Item, m itemreq.Merc, approx bool) []itemreq.Reason {
	var rs []itemreq.Reason
	if !m.Alive {
		rs = append(rs, itemreq.Reason{Code: itemreq.Dead, Text: "the mercenary is dead: revive him first"})
	}
	kind, slot := mercKind(l.e.Data, b)
	mi := merc.Item{Name: b.Name, Base: b.Code, Kind: kind, Identified: it.Identified, Broken: it.Broken,
		ReqStr: b.ReqStr, ReqDex: b.ReqDex, ReqLevel: b.LevelReq}
	if !merc.Allowed(m.Class, slot, mi) {
		rs = append(rs, itemreq.Reason{Code: itemreq.MercType, Text: fmt.Sprintf("a %s cannot use that", merc.KindName(m.Class))})
		return rs
	}
	st := merc.MercState{Level: m.Level, Str: m.Str, Dex: m.Dex, StatsKnown: m.Str+m.Dex > 0}
	switch r := merc.MeetsReq(mi, st, true); r.Reason {
	case "ok":
	case "level":
		rs = append(rs, itemreq.Reason{Code: itemreq.Requirements, Stat: "lvl", Need: b.LevelReq, Have: m.Level, Approx: approx,
			Text: fmt.Sprintf("needs level %d (he is %d)", b.LevelReq, m.Level)})
	case "str":
		rs = append(rs, itemreq.Reason{Code: itemreq.Requirements, Stat: "str", Need: b.ReqStr, Have: m.Str, Approx: approx,
			Text: fmt.Sprintf("needs %d strength (he has %d)", b.ReqStr, m.Str)})
	case "dex":
		rs = append(rs, itemreq.Reason{Code: itemreq.Requirements, Stat: "dex", Need: b.ReqDex, Have: m.Dex, Approx: approx,
			Text: fmt.Sprintf("needs %d dexterity (he has %d)", b.ReqDex, m.Dex)})
	case "unidentified":
		rs = append(rs, itemreq.Reason{Code: itemreq.Identified, Text: "unidentified"})
	case "broken":
		rs = append(rs, itemreq.Reason{Code: itemreq.Broken, Text: "broken"})
	default: // "stats unknown": fails closed, as the merc policy does
		rs = append(rs, itemreq.Reason{Code: itemreq.Requirements, Approx: true, Text: "his strength and dexterity are not read yet"})
	}
	return rs
}
