// Package delivery is the loader and validator of the skill delivery object of the capability table
// (data/capabilities/capabilities.json, schema didle.cap/2; docs/design/fable-review-2026-11-21/combat-engine-b-grammar.md).
// It is a leaf package: it knows no class, no skill and no engine code. Two things are generated from the table
// and never hand-edited (card dlv-data): the combat package's skill view (classdata/skills.json), power's skillKinds,
// and every skill's tier.
package delivery

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Schema is the capability table schema this package reads.
const Schema = "didle.cap/2"

// Closed value sets (grammar rule 2: an unknown value is a load error that names the skill).
var (
	purposes = set("damage", "upkeep", "mark", "move", "utility", "passive")
	hands    = set("", "left", "right")
	targets  = set("unit", "point", "corpse", "self", "none")
	presses  = set("tap", "repeat", "channel")
	stances  = set("", "any", "range", "contact", "insidePack", "behindArmy")
	observes = set("state", "minions", "aura", "traps", "clock", "none")
	powers   = set("", "melee", "missile", "spell", "summon")
	drivers  = set("attack", "buff", "aura", "summon", "move", "passive", "none")
)

func set(v ...string) map[string]bool {
	m := make(map[string]bool, len(v))
	for _, s := range v {
		m[s] = true
	}
	return m
}

// Tiers: what the engine can do with a skill (Q2 of the combat plan). They are computed, never typed.
const (
	Supported    = "supported"
	Experimental = "experimental"
	Unsupported  = "unsupported"
)

// Delivery is the subset of the grammar the engine reads today. Use is nil for a skill nothing casts.
type Delivery struct {
	Purpose string   `json:"purpose"`
	Use     *Use     `json:"use,omitempty"`
	Reach   *Reach   `json:"reach,omitempty"`
	Observe *Observe `json:"observe,omitempty"`
	// Power is how internal/power estimates the skill (melee, missile, spell, summon); "" when it does not.
	Power string `json:"power,omitempty"`
	// Driver names the engine loop that can play the skill; "none" when there is none. A passive needs no driver.
	Driver string `json:"driver"`
}

// Use says how the skill is cast.
type Use struct {
	Hand   string `json:"hand,omitempty"`
	Target string `json:"target"`
	Press  string `json:"press"`
}

// Reach is the range in subtiles and the stance the hero takes.
type Reach struct {
	Range  int    `json:"range"`
	Stance string `json:"stance,omitempty"`
}

// Observe says how the engine sees that the skill is up.
type Observe struct {
	By string `json:"by"`
}

// Entry is one row of the table, reduced to what this package reads.
type Entry struct {
	Class    string    `json:"class"`
	Reason   string    `json:"reason"`
	Delivery *Delivery `json:"delivery"`
}

// Table is the capability table as this package sees it. Generic holds the engine's basics (Attack, Kick), which
// belong to no class.
type Table struct {
	Schema  string           `json:"schema"`
	Skills  map[string]Entry `json:"skills"`
	Generic map[string]Entry `json:"generic"`
}

// Proven is the evidence file: the skills with a green Proving Ground row on the backend in use.
type Proven struct {
	Schema string `json:"schema"`
	Rows   []struct {
		Skill    string `json:"skill"`
		Scenario string `json:"scenario"`
		Backend  string `json:"backend"`
	} `json:"rows"`
}

// Load parses and validates the table.
func Load(b []byte) (*Table, error) {
	var t Table
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, fmt.Errorf("delivery: capabilities.json: %w", err)
	}
	if t.Schema != Schema {
		return nil, fmt.Errorf("delivery: capabilities.json schema %q, want %q", t.Schema, Schema)
	}
	for _, m := range []map[string]Entry{t.Skills, t.Generic} {
		for n, e := range m {
			if e.Delivery == nil {
				return nil, fmt.Errorf("delivery: skill %q has no delivery object", n)
			}
			if err := e.Delivery.Validate(); err != nil {
				return nil, fmt.Errorf("delivery: skill %q: %w", n, err)
			}
		}
	}
	return &t, nil
}

// Validate checks every field against its closed set and the cross rules.
func (d *Delivery) Validate() error {
	chk := func(field, v string, ok map[string]bool) error {
		if !ok[v] {
			return fmt.Errorf("%s %q is not in its closed set", field, v)
		}
		return nil
	}
	if err := firstErr(chk("purpose", d.Purpose, purposes), chk("power", d.Power, powers), chk("driver", d.Driver, drivers)); err != nil {
		return err
	}
	if d.Observe != nil {
		if err := chk("observe.by", d.Observe.By, observes); err != nil {
			return err
		}
	}
	if d.Use == nil {
		if d.Reach != nil {
			return fmt.Errorf("reach without use")
		}
		return nil
	}
	if d.Reach == nil {
		return fmt.Errorf("a castable skill needs a reach")
	}
	if err := firstErr(chk("use.hand", d.Use.Hand, hands), chk("use.target", d.Use.Target, targets), chk("use.press", d.Use.Press, presses), chk("reach.stance", d.Reach.Stance, stances)); err != nil {
		return err
	}
	if _, err := d.CombatKind(); err != nil {
		return err
	}
	return nil
}

func firstErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

// CombatKind is the aiming kind the combat package knows (bolt, ground, self, buff, melee, aura, summon), derived
// from the grammar. The rules are the inverse of what the old hand-typed table said; a row the rules cannot place
// is an error.
func (d *Delivery) CombatKind() (string, error) {
	if d.Use == nil {
		return "", fmt.Errorf("not castable")
	}
	by := ""
	if d.Observe != nil {
		by = d.Observe.By
	}
	switch d.Purpose {
	case "upkeep", "mark", "passive":
		switch by {
		case "aura":
			return "aura", nil
		case "minions":
			return "summon", nil
		}
		return "buff", nil
	case "utility":
		if d.Use.Target == "self" || d.Use.Target == "none" {
			return "buff", nil
		}
	}
	switch d.Use.Target {
	case "self", "none":
		return "self", nil
	case "point":
		return "ground", nil
	case "unit":
		if d.Reach.Stance == "contact" {
			return "melee", nil
		}
		return "bolt", nil
	}
	return "", fmt.Errorf("target %q has no combat kind yet", d.Use.Target)
}

// TierOf computes the tier of a row (Q2). unsupported: no driver; supported: a passive (nothing to deliver) or a
// green Proving Ground row on the backend in use; experimental: a valid delivery and a driver, no such row yet.
func (d *Delivery) TierOf(proven bool) string {
	switch {
	case d.Driver == "passive":
		return Supported
	case d.Driver == "none":
		return Unsupported
	case proven:
		return Supported
	}
	return Experimental
}

// ProvenSkills returns the set of skills with a green row on the backend.
func ProvenSkills(b []byte, backend string) (map[string]bool, error) {
	var p Proven
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("delivery: proven.json: %w", err)
	}
	m := map[string]bool{}
	for _, r := range p.Rows {
		if r.Skill == "" || r.Scenario == "" {
			return nil, fmt.Errorf("delivery: proven.json: a row needs skill and scenario")
		}
		if r.Backend == backend {
			m[r.Skill] = true
		}
	}
	return m, nil
}

// All returns every entry, class skills and generics, with the generic flag.
func (t *Table) All() map[string]Entry {
	out := make(map[string]Entry, len(t.Skills)+len(t.Generic))
	for n, e := range t.Generic {
		out[n] = e
	}
	for n, e := range t.Skills {
		out[n] = e
	}
	return out
}

// CombatSkill is one row of the combat package's skill view (classdata/skills.json).
type CombatSkill struct {
	Name  string `json:"name"`
	Kind  string `json:"kind"`
	Range int    `json:"range"`
	Hand  string `json:"hand,omitempty"`
	Hold  bool   `json:"hold,omitempty"`
}

// CombatView is the generated content of classdata/skills.json, sorted by name.
func (t *Table) CombatView() ([]CombatSkill, error) {
	var out []CombatSkill
	for n, e := range t.All() {
		d := e.Delivery
		if d.Use == nil {
			continue
		}
		k, err := d.CombatKind()
		if err != nil {
			return nil, fmt.Errorf("delivery: skill %q: %w", n, err)
		}
		out = append(out, CombatSkill{Name: n, Kind: k, Range: d.Reach.Range, Hand: d.Use.Hand, Hold: d.Use.Press == "repeat"})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// PowerKinds is the generated skillKinds map of internal/power/profiles.json.
func (t *Table) PowerKinds() map[string]string {
	m := map[string]string{}
	for n, e := range t.All() {
		if e.Delivery.Power != "" {
			m[n] = e.Delivery.Power
		}
	}
	return m
}

// Tiers computes the tier of every class skill; proven is the set from ProvenSkills.
func (t *Table) Tiers(proven map[string]bool) map[string]string {
	m := make(map[string]string, len(t.Skills))
	for n, e := range t.Skills {
		m[n] = e.Delivery.TierOf(proven[n])
	}
	return m
}

// NeedsReason reports whether a tier requires a reason text.
func NeedsReason(tier string) bool { return strings.TrimSpace(tier) != Supported }
