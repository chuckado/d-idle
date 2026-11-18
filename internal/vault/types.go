// Package vault is the offline store of the Ethereal Vault (docs/design/loot-vault-2026-11-16.md, slice V2):
// an append-only event log per hero, content-hash dedupe, a state machine, data-driven capacity and trim,
// an in-memory index rebuilt by replay, an archive for everything that leaves the vault, and export/import.
// Nothing here talks to the game; the bot wiring is slice V3.
package vault

import (
	"errors"
	"time"
)

// State of a record. Pin and identified are flags on the record (Rec.Pin, Item.Identified), not states:
// the design's "pinned" and "revealed" are views of those flags (Rec.Status).
type State string

const (
	Recorded  State = "recorded"
	KeepQ     State = "keepq"     // Keep clicked, waiting for a town visit
	Kept      State = "kept"      // created in the game, final
	SellQ     State = "sellq"     // marked for the sell errand
	Sold      State = "sold"      // final
	Discarded State = "discarded" // final, archived
	Trimmed   State = "trimmed"   // final, archived
)

// Stages of a reveal or sale errand (slice V6). Each is logged BEFORE the game step it names, so a restart can tell what
// may exist in the hero.
const (
	StageRevBegin   = "rb" // reveal: the temporary copy is about to be created (it may or may not exist)
	StageRevMade    = "rc" // reveal: the copy exists in the hero and is not yet identified in the vault
	StageRevCleanup = "rt" // reveal: the identified bytes are stored; the temporary copy still has to leave the hero
	StageSellBegin  = "sb" // sale: the copy is about to be created in the pack
	StageSellMade   = "sc" // sale: the copy exists; the sale is under way
)

// Keep targets.
const (
	WhereStash = "stash"
	WhereInv   = "inv"
)

// Live states count against capacity.
func (s State) Live() bool { return s == Recorded || s == KeepQ || s == SellQ }

// Found says where a drop was seen.
type Found struct {
	Time         time.Time `json:"t"`
	Zone         string    `json:"zone,omitempty"`
	Area         string    `json:"area,omitempty"`
	MonsterClass string    `json:"monster,omitempty"`
}

// Item is what the caller records: the item bytes plus metadata. Raw is the only truth; the rest is derived by
// the caller (item codec, V1) and stored for search and display. d2d items have uid 0 (V0 finding), so the
// content hash of Raw is the identity.
type Item struct {
	ID         string   `json:"id"`
	Hero       string   `json:"hero"`
	Raw        string   `json:"raw"` // hex of the item bytes
	Code       string   `json:"code"`
	Name       string   `json:"name,omitempty"`
	Quality    int      `json:"q"`
	Ilvl       int      `json:"ilvl"`
	Identified bool     `json:"identified"`
	Chain      []string `json:"chain,omitempty"` // base type chain, most specific first, e.g. swor, weap
	Sockets    int      `json:"sockets,omitempty"`
	Eth        bool     `json:"eth,omitempty"`
	Found      Found    `json:"found"`
	Value      int      `json:"value"`
}

// Rec is a record as the index holds it.
type Rec struct {
	Item
	Hash     string    `json:"hash"`
	State    State     `json:"state"`
	Pin      bool      `json:"pin,omitempty"`
	InFlight bool      `json:"inflight,omitempty"` // a keep create was started and not confirmed
	Attempts int       `json:"attempts,omitempty"`
	Where    string    `json:"where,omitempty"`   // keep target: "" or stash, inv
	Reason   string    `json:"reason,omitempty"`  // last failure or trim/discard reason
	Gold     int       `json:"gold,omitempty"`    // sale price once sold
	RevQ     bool      `json:"revq,omitempty"`    // Reveal asked: the item is identified at the next town visit (V6)
	Stage    string    `json:"stage,omitempty"`   // in-flight step of a reveal or sale (Stage* constants), "" when idle
	GoldPre  int       `json:"goldpre,omitempty"` // purse before the sale create, for the read-back after a restart
	Updated  time.Time `json:"updated"`
	seq      int       // position in insertion order
}

// Status is the one-word view the UI shows.
func (r *Rec) Status() string {
	switch {
	case r.State == Recorded && r.Pin:
		return "pinned"
	case r.State == Recorded && r.Identified:
		return "revealed"
	}
	return string(r.State)
}

var (
	ErrDuplicate    = errors.New("vault: identical item already in the vault")
	ErrOverCapacity = errors.New("vault: over capacity, new records are blocked")
	ErrFull         = errors.New("vault: full and nothing may be trimmed")
	ErrNotFound     = errors.New("vault: no such record")
	ErrState        = errors.New("vault: transition not allowed from this state")
	ErrPinCap       = errors.New("vault: pin limit reached")
	ErrInvalid      = errors.New("vault: invalid item")
	ErrTooLarge     = errors.New("vault: import too large")
	ErrPinned       = errors.New("vault: pinned items are never sold")
)
