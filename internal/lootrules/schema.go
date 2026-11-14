package lootrules

import (
	"encoding/json"
	"sort"
)

// FieldDef describes one condition field for the editor, which builds its pickers from this list (GET
// /api/lootrules/schema), so a new field needs no UI change.
type FieldDef struct {
	Name   string   `json:"name"`
	Kind   string   `json:"kind"` // number, bool, enum, list, text, need, who
	Ops    []string `json:"ops"`
	Label  string   `json:"label"`
	Unit   string   `json:"unit,omitempty"`
	Values []string `json:"values,omitempty"` // allowed values of an enum or list
	Hides  string   `json:"hides,omitempty"`  // where the field can be unreadable: ground, unidentified
}

var numOps = []string{"eq", "ne", "ge", "gt", "le", "lt"}

// Fields lists every field, sorted by name, plus the per-stat fields.
func Fields() []FieldDef {
	n := func(name, label, unit, hides string) FieldDef {
		return FieldDef{Name: name, Kind: "number", Ops: numOps, Label: label, Unit: unit, Hides: hides}
	}
	b := func(name, label, hides string) FieldDef {
		return FieldDef{Name: name, Kind: "bool", Ops: []string{"eq"}, Label: label, Hides: hides}
	}
	e := func(name, label string, vals ...string) FieldDef {
		return FieldDef{Name: name, Kind: "enum", Ops: []string{"eq", "ne", "in"}, Label: label, Values: vals}
	}
	fs := []FieldDef{
		e("quality", "Quality", "low", "normal", "superior", "magic", "set", "rare", "unique", "crafted"),
		{Name: "type", Kind: "list", Ops: []string{"in"}, Label: "Item type (with its parents)"},
		{Name: "group", Kind: "list", Ops: []string{"in"}, Label: "Group"},
		{Name: "code", Kind: "list", Ops: []string{"in"}, Label: "Base item code (globs)"},
		{Name: "name", Kind: "text", Ops: []string{"contains"}, Label: "Name contains"},
		n("ilvl", "Item level", "", "ground"), n("reqlvl", "Required level", "", ""), n("reqstr", "Required strength", "", ""),
		n("reqdex", "Required dexterity", "", ""), n("sockets", "Sockets", "", ""), n("maxsockets", "Most sockets possible", "", ""),
		n("value", "Sell value", "gold", "ground"), n("qty", "Quantity or pile size", "", ""), n("runeno", "Rune number", "", ""),
		n("grade", "Gem grade", "", ""), n("owned", "Already owned (pack and stash)", "", ""), n("durpct", "Durability", "%", ""),
		b("identified", "Identified", "ground"), b("ethereal", "Ethereal", "ground"), b("stackable", "Stackable", ""),
		b("quest", "Quest item", ""), b("broken", "Broken", ""),
		e("basetier", "Base tier", "normal", "exceptional", "elite"),
		{Name: "class_restricted", Kind: "list", Ops: []string{"in"}, Label: "Restricted to class"},
		{Name: "usable_by", Kind: "who", Ops: []string{"eq"}, Label: "Usable by (optionally after +level, +strength, +dexterity)", Values: []string{"hero", "merc"}},
		n("hero_usable_from", "Hero can wear it from level", "", ""), n("merc_usable_from", "Mercenary can wear it from level", "", ""),
		{Name: "need", Kind: "need", Ops: []string{"eq"}, Label: "The bot needs more of it", Values: sortedKeys(needKinds)},
		e("class", "Hero class"), e("build", "Build"), e("difficulty", "Difficulty", "Normal", "Nightmare", "Hell"),
		n("clvl", "Hero level", "", ""), n("gold", "Gold on hand", "gold", ""), n("act", "Act", "", ""),
		n("packfull", "Pack fullness", "%", ""), n("stashfull", "Stash fullness", "%", ""),
	}
	for s := range statNames {
		fs = append(fs, FieldDef{Name: "stat:" + s, Kind: "number", Ops: numOps, Label: "Property " + s, Hides: "unidentified"})
	}
	fs = append(fs, FieldDef{Name: "stat:ressum", Kind: "number", Ops: numOps, Label: "Sum of the four resistances", Hides: "unidentified"})
	sort.Slice(fs, func(i, j int) bool { return fs[i].Name < fs[j].Name })
	return fs
}

func sortedKeys(m map[string]bool) []string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

// ActionsFor lists the actions the editor may offer in a scope.
func ActionsFor(scope string) []string {
	switch scope {
	case PlaceGround:
		return []string{Pickup, Ignore}
	case PlacePack:
		return []string{Keep, KeepMerc, Stash, Sell, Drop, Identify, Ask, Ignore}
	case PlaceStash:
		return []string{Keep, Sell, Ignore}
	}
	return nil
}

// InvariantInfo is one protected row the editor shows above the list.
type InvariantInfo struct{ ID, Text string }

// Invariants are the floors no rule can override (data for the grey block of the editor).
func Invariants() []InvariantInfo {
	return []InvariantInfo{
		{"invariant:place", "Only items on the ground, in the backpack or in the stash are decided. Worn items, the belt, the cube and the cursor are never touched."},
		{"invariant:mark", "Items you marked keep or sell stay as you marked them."},
		{"invariant:start", "The starting gear stays."},
		{"invariant:quest", "Quest items are always picked up and kept."},
		{"invariant:supplies", "Tomes, scrolls and keys are never sold or dropped."},
		{"invariant:valuable", "Magic or better items are never dropped, and runes, gems, charms and jewels are never sold or dropped, unless a rule you confirmed says so."},
	}
}

func jsonOf(c Cond) (string, error) {
	b, err := json.Marshal(c)
	return string(b), err
}
