// Package lootrules is the PROTOTYPE of the one item-decision rules engine (docs/lootrules.md): a versioned JSON rule
// set, compiled once, evaluated per item at the ground, pack and stash decision points. Not wired into the bot yet
// (cards L2 to L4). Pure: no game read, no clicks, no class or skill literal in code (class, build and item groups are
// data). The engine can only answer pickup, keep, stash, sell, drop, identify or ignore; there is no equip action, so a
// rule set cannot make an equipment decision (docs/gear-policy.md).
package lootrules

import "encoding/json"

// Schema is the format version this build reads and writes.
const Schema = "lootrules/1"

// Actions.
const (
	Pickup   = "pickup"   // ground: walk to it and take it
	Keep     = "keep"     // pack: stays where it is
	Stash    = "stash"    // pack -> stash at a town visit
	Sell     = "sell"     // pack -> vendor
	Drop     = "drop"     // pack -> ground (gated, see invariants)
	Identify = "identify" // pack: spend a scroll on it
	Ignore   = "ignore"   // do nothing
	Ask      = "ask"      // pack: raise a Needs-you question and leave the item alone until the player answers
	Vault    = "vault"    // ground only: peek the item and record it in the Ethereal Vault, do not pick it up (loot.vault, docs/vault.md)
	KeepMerc = "keepmerc" // keep for the mercenary: a keep that the sell chore and the tidy pass leave alone (the player equips him)
)

// Places an item can be in.
const (
	PlaceGround = "ground"
	PlacePack   = "pack"
	PlaceStash  = "stash"
	// PlaceSell is the vendor visit: a pack item asked "sell it or not" by the sell chore (wave 2b, loot.rulesEngine 3).
	// Its rules are the few sell-only decisions; an item no rule of this scope catches is judged by the ground rules, the
	// way the pickit judges today (loot.SellDecision).
	PlaceSell = "sell"
	// Everything else (equipment, belt, cube, cursor, vendor) is never decided: Decide answers ignore.
)

// RuleSet is the whole player-visible filter.
type RuleSet struct {
	Schema   string            `json:"schema"`
	Seq      int               `json:"seq,omitempty"` // change counter, owned by the store (ignored on PUT)
	Name     string            `json:"name"`
	Comment  string            `json:"comment,omitempty"`
	Limits   Limits            `json:"limits"`
	Groups   map[string]Group  `json:"groups,omitempty"`   // adds to the built-in groups
	Defaults map[string]string `json:"defaults,omitempty"` // scope -> action when no rule matches; absent = ignore
	Ack      []string          `json:"ack,omitempty"`      // the player's explicit acknowledgements ("drop-valuable")
	Preset   string            `json:"preset,omitempty"`   // the preset this set was made from ("starter", "class:<Name>"); "" or "custom" = hand-edited
	Options  *Options          `json:"options,omitempty"`  // the simple editor's knobs; with Preset the store rebuilds Rules from them on PUT
	Rules    []Rule            `json:"rules"`
}

// Limits are the safety valves.
type Limits struct {
	MaxActions  int `json:"maxActions"`            // stash plus drop (plus sell, identify) actions per pass
	StashMaxPct int `json:"stashMaxPct,omitempty"` // no stash action once the stash is this full
}

// Group names a set of base items: item types (ancestors count), type prefixes and base codes (path globs).
type Group struct {
	Types        []string `json:"types,omitempty"`
	TypePrefixes []string `json:"typePrefixes,omitempty"`
	Codes        []string `json:"codes,omitempty"`
}

// Rule is one line; rules are tried in order and the first match wins.
type Rule struct {
	ID        string   `json:"id"`
	Name      string   `json:"name,omitempty"`
	Off       bool     `json:"off,omitempty"`
	Scope     []string `json:"scope,omitempty"`   // ground, pack, stash; empty = ground and pack
	Classes   []string `json:"classes,omitempty"` // only for these hero classes (d2.Class.String()); empty = all
	Builds    []string `json:"builds,omitempty"`  // only for these builds; empty = all
	When      *Cond    `json:"when,omitempty"`    // nil = always
	Action    string   `json:"action"`
	Prio      int      `json:"prio,omitempty"`      // pickup order, higher first
	Tier      string   `json:"tier,omitempty"`      // S A B C: the pickit strictness knob filters ground rules by tier (kept so it can be translated 1:1)
	OnUnknown string   `json:"onUnknown,omitempty"` // "match": a rule whose condition is unreadable at this place matches anyway (take it, identify later)
	Pickup    *bool    `json:"pickup,omitempty"`    // a fate rule on the ground: take it? default true for keep and stash
	Reason    string   `json:"reason"`              // the text of the "why" tooltip
	Confirm   bool     `json:"confirm,omitempty"`   // the player confirmed this rule may drop or sell valuables
}

// Cond is a condition tree: exactly one of all, any, not or a leaf (f, op, v).
type Cond struct {
	All  []Cond          `json:"all,omitempty"`
	Any  []Cond          `json:"any,omitempty"`
	Not  *Cond           `json:"not,omitempty"`
	F    string          `json:"f,omitempty"`
	Plus *Plus           `json:"plus,omitempty"` // usable_by only: "usable after +N level, strength, dexterity"
	Op   string          `json:"op,omitempty"`
	V    json.RawMessage `json:"v,omitempty"`
}

// Hidden-field bits: what the decision point cannot read (docs/lootrules.md section 1).
const (
	HILvl uint8 = 1 << iota
	HIdentified
	HValue
	HStats
	HQuality
)

// Item is what a rule can see. The adapters (cards L2, L3) fill it from the snapshot, the d2d ground label or the pack.
type Item struct {
	Place      string
	Code, Name string
	Types      []string // the item type and its ancestors ("rin", "jewl", "weap", "swor")
	Quality    string   // low normal superior magic set rare unique crafted
	ILvl       int
	ReqLvl     int
	ReqStr     int
	ReqDex     int
	ClassOnly  string   // class-restricted item (orb, wand, pelt ...): the class name, "" = anyone
	MercTypes  []string // hireling types that can wear it (Hireling.txt plus item types); nil = unknown
	RuneNo     int      // rune number 1..33, 0 = not a rune
	Grade      int      // gem grade 1 chipped .. 5 perfect, 0 = not a gem
	BaseTier   string   // normal, exceptional or elite ("" = unknown)
	MaxSockets int      // most sockets the base can have
	Durability int      // current; MaxDur 0 = indestructible or unknown
	MaxDur     int
	Sockets    int
	Identified bool
	Ethereal   bool
	Stackable  bool
	Quest      bool
	StartItem  bool
	Qty        int            // gold pile size or stack count
	Value      int            // vendor sell value in gold
	Stats      map[string]int // display values by catalogue name (stats.go); only meaningful when readable
	Hidden     uint8
	Mark       string // the player's mark: keep, sell or ""
	ID         uint32
}

// Ctx is the situation.
type Ctx struct {
	Class, Build, Difficulty string
	CLvl, Gold               int
	// Hero and Merc are the wearers a requirement rule compares against. Merc nil = no mercenary. Level 0 = unread: a
	// requirement predicate then answers unknown (the item is judged again when the numbers arrive, see Fingerprint).
	Hero, Merc                *Actor
	Act                       int             // act of the area the hero is in (0 = unknown)
	Owned                     func(*Item) int // how many of this base the hero holds (pack plus stash); nil = 0
	PackFullPct, StashFullPct int
	IDTome                    bool // a Tome of Identify is in the pack (the field idtome)
	// Needs answers the "need" condition (belt, scroll, tome, key, gold, gem, charmroom needs); nil = every need is
	// satisfied, so a need rule never matches.
	Needs func(kind string, it *Item) bool
}

// Decision is the answer for one item.
type Decision struct {
	Action string
	Rule   string // rule id; "invariant:<name>" for the code floors; "default"
	Why    string
	Prio   int
	Tier   string
	Fate   string // the rule's own action when the ground turned a fate into pickup or ignore
	// Pending: a rule above the winner could not be judged (hidden field). Identify the item and ask again.
	Pending bool
}

// Plus is the "what if" of usable_by: the wearer gains these before the requirement check (level up, stat points).
type Plus struct {
	Lvl int `json:"lvl,omitempty"`
	Str int `json:"str,omitempty"`
	Dex int `json:"dex,omitempty"`
}

// Actor is a possible wearer: the hero (Class = class name) or the mercenary (Class = hireling type).
type Actor struct {
	Class    string
	Level    int
	Str, Dex int
}
