// Package itemreq carries the structured answer of the equip path's validity rule: what an item requires and why a
// wearer (the hero or the mercenary) cannot use it, with numbers. The rule itself lives in internal/leveler
// (Candidate, CheckPlayerEquip) and internal/merc (Allowed, MeetsReq); this package only gives their answer a shape the
// inventory view (internal/invdesk) can print without a second copy of the rule (docs/inventory-ui.md, "Requirements").
package itemreq

import "fmt"

// Reason codes. The first seven are the equip path's own vocabulary (docs/design/item-verbs.md section 5); broken and
// merctype were added for the view.
const (
	Class        = "class"        // another class's item
	Requirements = "requirements" // level, strength or dexterity (Stat says which)
	Identified   = "identified"   // an unidentified magic or better item
	TwoHand      = "twohand"      // a two-hander and a shield cannot be worn together
	Slot         = "slot"         // no slot takes it
	NoSpace      = "nospace"      // nowhere to put the swapped-out item
	Dead         = "dead"         // the mercenary is dead: no gear change
	Broken       = "broken"       // durability 0 (a warning: the game still lets it be worn)
	MercType     = "merctype"     // the mercenary's type does not carry that kind
)

// Reason is one refusal. Need and Have are numbers for a requirement and names for a class; Stat is lvl, str or dex
// for code requirements. Approx says Need is the base number and the real one may differ (see Req.Approx). Warn marks
// a note that does not change the verdict.
type Reason struct {
	Code   string `json:"code"`
	Stat   string `json:"stat,omitempty"`
	Need   any    `json:"need,omitempty"`
	Have   any    `json:"have,omitempty"`
	Text   string `json:"text"`
	Approx bool   `json:"approx,omitempty"`
	Warn   bool   `json:"warn,omitempty"`
}

// Req is what an item itself requires after the modifiers the bot can read. Approx is set when a modifier that the
// bot cannot read exactly applies (ethereal, a requirements percentage, the level requirement of an affix): the
// numbers are then the base ones.
type Req struct {
	Lvl    int      `json:"lvl,omitempty"`
	Str    int      `json:"str,omitempty"`
	Dex    int      `json:"dex,omitempty"`
	Class  []string `json:"class,omitempty"`
	Approx bool     `json:"approx,omitempty"`
}

// Side is the verdict for one wearer.
type Side struct {
	OK      bool     `json:"ok"`
	Reasons []Reason `json:"reasons"`
}

// Usable is the hero's and, when a mercenary is owned, the mercenary's verdict.
type Usable struct {
	Hero *Side `json:"hero,omitempty"`
	Merc *Side `json:"merc,omitempty"`
}

// Merc is what the judge needs to know about the mercenary.
type Merc struct {
	Class    int
	Level    int
	Str, Dex int
	Alive    bool
}

// Item is what the judge needs of an item besides the unit: the base code and the flags the view already holds, so a
// mercenary's worn item (which has no unit of ours) is judged by the same code.
type Item struct {
	Code       string
	Identified bool
	Broken     bool
}

// Judgement is the item-level answer: the requirement, the hero's item-level reasons (empty = nothing about the item
// itself stops him) and the mercenary's (nil when no mercenary was asked about). Wearable is false for potions,
// scrolls, gems and the like: they get no verdict at all.
type Judgement struct {
	Wearable bool
	Req      Req
	Hero     []Reason
	Merc     []Reason
}

// Error is a refusal that carries its structured reasons; errors.Is and errors.As see through it to Err.
type Error struct {
	Err     error
	Reasons []Reason
}

func (e *Error) Error() string { return e.Err.Error() }
func (e *Error) Unwrap() error { return e.Err }

// NeedText is the text of a requirement reason, "needs 47 strength (you have 38)".
func NeedText(what string, have int) string {
	return fmt.Sprintf("needs %s (you have %d)", what, have)
}
