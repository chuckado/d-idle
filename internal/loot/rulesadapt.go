package loot

import (
	"fmt"
	"strconv"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/lootrules"
	"d2lod-bot/internal/merc"
)

// The adapter that fills lootrules.Item, Ctx and the two Actors from the real sources (docs/lootrules.md section 1,
// wave 1: shadow only, nothing here decides). A field the source cannot read is flagged in Item.Hidden or left at its
// zero value with the flag; it is never invented.

var ruleQuality = map[int]string{d2.QualityLow: "low", d2.QualityNormal: "normal", d2.QualitySuperior: "superior",
	d2.QualityMagic: "magic", d2.QualitySet: "set", d2.QualityRare: "rare", d2.QualityUnique: "unique", d2.QualityCrafted: "crafted"}

// ruleStats maps the catalogue names of internal/lootrules/stats.go to item stat ids and the right shift. S: the ids
// of pickit_starter.json (life 7 shifted 8, resists 39 41 43 45, FCR 105, FHR 99, MF 80, all skills 127, class skills 83,
// skill tab 188, sockets 194) and the player stat list (str 0, enr 1, dex 2, vit 3); U: mana (9, shift 8), frw 96,
// ias 93, gf 79, ed 17.
var ruleStats = []struct {
	name       string
	id, shift  int
	registered bool // false: id from memory of the item stat cost table, U
}{
	{"life", 7, 8, true}, {"mana", 9, 8, false}, {"str", 0, 0, true}, {"dex", 2, 0, true}, {"vit", 3, 0, true}, {"enr", 1, 0, true},
	{"fireres", 39, 0, true}, {"lightres", 41, 0, true}, {"coldres", 43, 0, true}, {"poisres", 45, 0, true},
	{"fcr", 105, 0, true}, {"fhr", 99, 0, true}, {"frw", 96, 0, false}, {"ias", 93, 0, false}, {"mf", 80, 0, true}, {"gf", 79, 0, false},
	{"allskills", 127, 0, true}, {"classskills", 83, 0, true}, {"skilltab", 188, 0, true}, {"sockets", 194, 0, true}, {"ed", 17, 0, false},
}

const (
	statDurability    = 72 // U: item_durability / maxdurability stat ids of the stat cost table
	statMaxDurability = 73
	statQuantityRule  = 70 // V in this package (statQuantity): the tome and stack count
)

// ItemOpts is what the caller knows beyond the unit.
type ItemOpts struct {
	Place  string   // lootrules.PlaceGround, PlacePack or PlaceStash
	Stats  d2.Stats // the unit's stat list; nil = unreadable
	Mark   string   // the player's mark: keep, sell or ""
	MaxBuy int      // vendor cap for the value estimate (MaxBuyOf)
	// LabelOnly: the unit comes from a d2d ground label (docs/d2d-loot.md): no item level, no identified flag, the
	// quality is a guess and a magic-or-better value is only a bound.
	LabelOnly bool
	// NeverSell: the town's own never-sell list (the Cube, keys, quest-like items) counts as a quest item, as tidy does.
	NeverSell bool
}

// RulesAdapter builds engine inputs from the looter's data and snapshot readers.
type RulesAdapter struct {
	l       *Looter
	classes map[string]string // class restriction code -> class name
}

// Rules returns the adapter of this looter (nil Data: every call reports an error).
func (l *Looter) Rules() *RulesAdapter {
	a := &RulesAdapter{l: l, classes: map[string]string{}}
	if l.e.Data != nil {
		for _, c := range l.e.Data.Classes {
			a.classes[c.Code] = c.Name
		}
	}
	return a
}

// Item fills the rule item of u. An unknown base row is an error (the caller leaves the item undecided).
func (a *RulesAdapter) Item(u d2.Unit, o ItemOpts) (lootrules.Item, error) {
	db := a.l.e.Data
	if db == nil || u.Item == nil {
		return lootrules.Item{}, fmt.Errorf("loot: no item data for unit %d", u.ID)
	}
	b := a.l.base(u.Class)
	if b == nil {
		return lootrules.Item{}, fmt.Errorf("loot: unit %d has unknown base class %d", u.ID, u.Class)
	}
	chain := merc.TypeChain(db, b.Type)
	if b.Type2 != "" {
		for _, t := range merc.TypeChain(db, b.Type2) {
			if !has(chain, t) {
				chain = append(chain, t)
			}
		}
	}
	it := lootrules.Item{
		Place: o.Place, Code: b.Code, Name: b.Display, Types: chain, Quality: ruleQuality[u.Item.Quality],
		ILvl: u.Item.ILvl, ReqLvl: b.LevelReq, ReqStr: b.ReqStr, ReqDex: b.ReqDex, ClassOnly: a.classOnly(chain),
		RuneNo: runeNo(b.Code), Grade: gemGrade(b), BaseTier: baseTier(b), MaxSockets: b.Sockets,
		Identified: u.Item.Identified(), Ethereal: u.Item.Ethereal(), Stackable: b.Stackable, Quest: b.Quest || o.NeverSell || QuestCode(b.Code),
		StartItem: u.Item.Flags&d2.FlagStartItem != 0, Mark: o.Mark, ID: u.ID,
		Value: SellValue(db, b, u.Item, o.MaxBuy),
	}
	if it.Quality == "" {
		it.Hidden |= lootrules.HQuality
	}
	it.MercTypes = merc.WearTypes(chain, b.TwoHanded)
	if o.LabelOnly {
		it.Hidden |= lootrules.HILvl | lootrules.HIdentified
		it.ILvl, it.Identified = 0, false
		if magicPlus(it.Quality) {
			it.Hidden |= lootrules.HValue
		}
	}
	if o.Stats == nil {
		it.Hidden |= lootrules.HStats
		return it, nil
	}
	it.Stats = map[string]int{}
	for _, s := range ruleStats {
		if o.Stats.Has(s.id) {
			it.Stats[s.name] = int(o.Stats.Get(s.id)) >> uint(s.shift)
		}
	}
	it.Sockets = it.Stats["sockets"]
	if b.Type == "gold" {
		it.Qty = int(o.Stats.Get(d2.StatGold))
	} else if o.Stats.Has(statQuantityRule) {
		it.Qty = int(o.Stats.Get(statQuantityRule))
	}
	if o.Stats.Has(statMaxDurability) {
		it.MaxDur, it.Durability = int(o.Stats.Get(statMaxDurability)), int(o.Stats.Get(statDurability))
	}
	return it, nil
}

func magicPlus(q string) bool {
	switch q {
	case "magic", "set", "rare", "unique", "crafted":
		return true
	}
	return false
}

func has(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}

// classOnly is the class name of the first class restriction on the type chain ("" = anyone).
func (a *RulesAdapter) classOnly(chain []string) string {
	for _, t := range chain {
		if ty := a.l.e.Data.ItemType(t); ty != nil && ty.Class != "" {
			return a.classes[ty.Class]
		}
	}
	return ""
}

// runeNo is the rune number of the codes r01 to r33, 0 for anything else (S: the code order is the rune order).
func runeNo(code string) int {
	if len(code) != 3 || code[0] != 'r' {
		return 0
	}
	n, err := strconv.Atoi(code[1:])
	if err != nil || n < 1 || n > 33 {
		return 0
	}
	return n
}

// gemGrade is 1 chipped to 5 perfect from the code's second letter (U: c f s or u z or l p; skulls differ in the
// last two); 0 for a base that is not a gem or skull.
func gemGrade(b *data.ItemBase) int {
	if len(b.Code) != 3 || !(b.Code[0] == 'g' || b.Code[:2] == "sk") {
		return 0
	}
	if b.Code[0] == 'g' {
		return map[byte]int{'c': 1, 'f': 2, 's': 3, 'z': 4, 'p': 5}[b.Code[1]]
	}
	return map[byte]int{'c': 1, 'f': 2, 'u': 3, 'l': 4, 'z': 5}[b.Code[2]]
}

// baseTier compares the base code with the NormCode, UberCode and UltraCode columns ("" = unknown or not a tiered base).
func baseTier(b *data.ItemBase) string {
	switch b.Code {
	case "":
		return ""
	case b.NormCode:
		return "normal"
	case b.UberCode:
		return "exceptional"
	case b.UltraCode:
		return "elite"
	}
	return ""
}

// ActorsOf reads the hero and the mercenary of a snapshot. Merc nil = none hired. A merc whose hireling type is not in
// Hireling.txt (or whose stats are not read) comes with Level 0, so a requirement rule judges UNKNOWN.
func (a *RulesAdapter) ActorsOf(s *d2.Snapshot) (hero, mc *lootrules.Actor) {
	p := s.Player
	hero = &lootrules.Actor{Class: p.Class.String(), Level: p.Level, Str: p.Strength, Dex: p.Dexterity}
	if s.Merc == nil || !s.Merc.Hired {
		return hero, nil
	}
	mc = &lootrules.Actor{}
	if h := a.l.e.Data.HirelingByID(s.Merc.Type); h != nil {
		mc.Class = merc.KindName(h.Class)
	}
	if mc.Class != "" && s.Merc.Level > 0 && s.Merc.Str+s.Merc.Dex > 0 {
		mc.Level, mc.Str, mc.Dex = s.Merc.Level, s.Merc.Str, s.Merc.Dex
	}
	return hero, mc
}

// Ctx builds the rule context of a snapshot: hero and merc actors, level, gold, difficulty, act, the pack and stash
// fill the caller measured, the owned count and the needs the pickit has today (needsOf; the empty-slot need from the
// naked-slot flags). build is the leveler build name ("" when unknown).
func (a *RulesAdapter) Ctx(s *d2.Snapshot, build string, packPct, stashPct int) *lootrules.Ctx {
	hero, mc := a.ActorsOf(s)
	lc := a.l.ctxFrom(s)
	n := needsOf{l: a.l, s: s}
	db := a.l.e.Data
	cx := &lootrules.Ctx{
		Class: hero.Class, Build: build, Difficulty: s.Player.Difficulty.String(), CLvl: s.Player.Level, Gold: s.Player.Gold,
		Hero: hero, Merc: mc, Act: data.ActOfArea(s.Player.Area), PackFullPct: packPct, StashFullPct: stashPct,
	}
	cx.Needs = func(kind string, it *lootrules.Item) bool {
		b := db.ItemByCode(it.Code)
		if b == nil {
			return false
		}
		if kind == "emptyslot" {
			return a.l.fillsEmptySlot(b, &d2.Item{}, lc)
		}
		return n.Want(kind, b)
	}
	owned := map[string]int{}
	if s.Inventory != nil {
		for _, c := range []d2.Container{d2.ContInventory, d2.ContStash} {
			for _, u := range s.Inventory.Container(c) {
				if b := a.l.base(u.Class); b != nil {
					owned[b.Code]++
				}
			}
		}
	}
	cx.Owned = func(it *lootrules.Item) int { return owned[it.Code] }
	if s.Inventory != nil {
		for _, u := range s.Inventory.Container(d2.ContInventory) {
			if b := a.l.base(u.Class); b != nil && b.Code == "ibk" {
				cx.IDTome = true
			}
		}
	}
	return cx
}
