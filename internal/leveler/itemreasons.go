package leveler

import (
	"fmt"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/itemreq"
)

// stReqPct is item_req_percent ("requirements -x%"; S: standard ItemStatCost id, U on items in this lab).
const stReqPct = 91

// itemReasons is the validity rule of Candidate (and so of CheckPlayerEquip): every reason the item itself cannot be
// used by player p, in the order the rule tests them, plus the requirement the numbers come from. An empty list means
// nothing about the item stops the player; slots only says whether the base has a body slot at all. The wording keeps
// the old Why texts and adds the numbers (user decision 2026-11-14: "needs 47 strength (you have 38)").
func itemReasons(db *data.DB, p *d2.Player, u d2.Unit, b *data.ItemBase, st d2.Stats, slots []int) (itemreq.Req, []itemreq.Reason) {
	it := u.Item
	req := itemreq.Req{Lvl: b.LevelReq, Str: b.ReqStr, Dex: b.ReqDex}
	var rs []itemreq.Reason
	add := func(code, stat string, need, have any, text string, approx bool) {
		rs = append(rs, itemreq.Reason{Code: code, Stat: stat, Need: need, Have: have, Text: text, Approx: approx})
	}
	rc := classOf(db, b.Type, map[string]bool{})
	if rc != "" {
		req.Class = []string{classDisplay(db, rc)}
	}
	// Modifiers the bot cannot read exactly (U, each needs a probe, docs/inventory-ui.md "Requirement modifiers"): an
	// ethereal item's lower requirements, a "requirements -x%" affix and the level requirement an affix adds. The
	// numbers stay the base ones and Approx says so.
	if it.Ethereal() || st.Get(stReqPct) != 0 || (it.Quality >= d2.QualityMagic && it.ILvl > b.LevelReq) {
		req.Approx = true
	}
	if slots == nil {
		text := "not equipment for this class"
		if b.Kind == data.KindWeapon && b.Stackable {
			text = "throwing weapon (town owns ammo)"
		}
		add(itemreq.Slot, "", nil, nil, text, false)
	}
	if cl := db.ClassByName(p.Class.String()); cl != nil && rc != "" && rc != cl.Code {
		add(itemreq.Class, "", classDisplay(db, rc), cl.Name, classDisplay(db, rc)+" only", false)
	}
	if slots != nil && b.Kind == data.KindWeapon && b.Stackable {
		add(itemreq.Slot, "", nil, nil, "throwing weapon (town owns ammo)", false)
	}
	if it.Quality >= d2.QualityMagic && !worn(it) && !it.Identified() {
		// The game refuses to wear an unidentified magic or better item (VERIFIED LIVE: the equipfail logs show the town
		// chore logging "equip: skip Quilted Armor reason=unidentified" for item 46 two seconds before this planner
		// clicked the same item into the torso slot and the click was refused). The identify chore makes it usable.
		add(itemreq.Identified, "", nil, nil, "unidentified", false)
	}
	if it.Quality >= d2.QualityMagic && !worn(it) && p.Level < it.ILvl {
		// A magic, rare, set or unique item's required level is raised by its affixes and the game gives no way to read it
		// from memory. The item level is an upper bound, so the item counts as unusable below it (VERIFIED LIVE: a magic Cap
		// and a magic Wand, ilvl 7, were refused at level 2 although their bases need level 0 and 1; docs/leveler.md).
		add(itemreq.Requirements, "lvl", it.ILvl, p.Level,
			fmt.Sprintf("magic item level %d is above your level %d (the game may allow it)", it.ILvl, p.Level), true)
	}
	if p.Level < b.LevelReq {
		add(itemreq.Requirements, "lvl", b.LevelReq, p.Level, fmt.Sprintf("needs level %d (you are %d)", b.LevelReq, p.Level), req.Approx)
	}
	if p.Strength < b.ReqStr {
		add(itemreq.Requirements, "str", b.ReqStr, p.Strength, itemreq.NeedText(fmt.Sprintf("%d strength", b.ReqStr), p.Strength), req.Approx)
	}
	if p.Dexterity < b.ReqDex {
		add(itemreq.Requirements, "dex", b.ReqDex, p.Dexterity, itemreq.NeedText(fmt.Sprintf("%d dexterity", b.ReqDex), p.Dexterity), req.Approx)
	}
	if it.Flags&d2.FlagIdentified != 0 && st != nil && st.Get(stStr) < 0 && p.Strength+int(st.Get(stStr)) < 0 {
		add(itemreq.Requirements, "str", nil, p.Strength, "negative stats", false)
	}
	return req, rs
}

func classDisplay(db *data.DB, code string) string {
	for _, c := range db.Classes {
		if c.Code == code {
			return c.Name
		}
	}
	return code
}
