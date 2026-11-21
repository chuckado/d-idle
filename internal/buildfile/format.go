// Package buildfile is an small, shareable build file ("didle.build/1"): ordered skill and
// stat goals that expand deterministically into the per-level rows the leveler consumes (leveler.Row). Not wired
// into the plan tables yet; see README.md.
package buildfile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Schema is the only accepted value of Build.Schema.
const Schema = "didle.build/1"

// MaxClvl is the last character level a file can plan.
const MaxClvl = 99

// Stat names in the order of leveler's attribute indexes (Str, Dex, Vit, Ene).
var StatNames = [4]string{"str", "dex", "vit", "ene"}

// Build is one build file. Main-attack hints are deliberately NOT part of it: the engine derives them from the
// allocation (the skill with the most hard points in the current phase), so a friend cannot ship a hint that
// disagrees with the points.
type Build struct {
	Schema string `json:"schema"`
	Name   string `json:"name"`
	Class  string `json:"class"` // lower case full name: sorceress, paladin, ...
	Author string `json:"author,omitempty"`
	Notes  string `json:"notes,omitempty"`
	Source string `json:"source,omitempty"`

	// Skills are ordered goals ("raise Skill to N hard points"). Prerequisites are added automatically.
	Skills []Goal     `json:"skills"`
	Stats  []StatGoal `json:"stats,omitempty"`
	// Phases is a leveling plan that needs no respec: each phase's goals are only open from FromClvl, and goals still
	// unmet after UntilClvl are dropped (points cannot be taken back without a respec).
	Phases []Phase `json:"phases,omitempty"`
	// RespecAt, when set, refunds everything at that clvl; Skills is then the post-respec build and Phases the
	// pre-respec leveling. Expand only computes both row sets, nothing is executed.
	RespecAt *Respec `json:"respec_at,omitempty"`
	// Quest lists extra points the player takes (Den of Evil, Radament, Izual, Lam Esen). Empty = none assumed.
	Quest []QuestPoint `json:"quest_points,omitempty"`
}

// Goal raises Skill to To hard points.
type Goal struct {
	Skill string `json:"skill"`
	To    int    `json:"to"`
	When  *When  `json:"when,omitempty"`
	// Hold (kolbot "satisfy"): later goals wait until this one is complete, even if this one is not spendable yet.
	Hold bool `json:"hold,omitempty"`
}

// When gates a goal: not before Clvl, and not before After holds.
type When struct {
	Clvl  int    `json:"clvl,omitempty"`
	After *After `json:"after,omitempty"`
}

// After is "Skill has at least Level hard points".
type After struct {
	Skill string `json:"skill"`
	Level int    `json:"level"`
}

// StatTarget is a number (absolute attribute value, as the character panel shows it without items) or "rest".
type StatTarget struct {
	N    int
	Rest bool
}

func (t *StatTarget) UnmarshalJSON(b []byte) error {
	var s string
	if json.Unmarshal(b, &s) == nil {
		if s != "rest" {
			return fmt.Errorf(`stat target %q: only "rest" or a number`, s)
		}
		*t = StatTarget{Rest: true}
		return nil
	}
	var n int
	if err := json.Unmarshal(b, &n); err != nil {
		return fmt.Errorf(`stat target: want a whole number or "rest"`)
	}
	*t = StatTarget{N: n}
	return nil
}

func (t StatTarget) MarshalJSON() ([]byte, error) {
	if t.Rest {
		return []byte(`"rest"`), nil
	}
	return json.Marshal(t.N)
}

// StatGoal raises a stat to To. Per, when set, caps the points per level this goal takes while it is open
// (the rest flows to the next goal): "3 of every 5 into Str, 2 into Vit" is two goals with per 3 and per 2.
type StatGoal struct {
	Stat string     `json:"stat"`
	To   StatTarget `json:"to"`
	Per  int        `json:"per,omitempty"`
	When *When      `json:"when,omitempty"` // only Clvl is honoured for stats
	// ReqFor is a comment ("Sacrifice wants 60 Str"); never evaluated.
	ReqFor string `json:"req_for,omitempty"`
}

// Phase is a leveling phase without respec.
type Phase struct {
	Name      string `json:"name"`
	FromClvl  int    `json:"from_clvl"`
	UntilClvl int    `json:"until_clvl,omitempty"`
	Skills    []Goal `json:"skills"`
}

// Respec marks the level a respec happens.
type Respec struct {
	Clvl int `json:"clvl"`
}

// QuestPoint is extra skill and stat points granted when the character reaches Clvl (quest done around then).
type QuestPoint struct {
	Clvl  int `json:"clvl"`
	Skill int `json:"skill,omitempty"`
	Stat  int `json:"stat,omitempty"`
}

// Parse decodes a build file strictly: unknown fields, trailing data and a wrong schema are errors. It does not
// validate against game data (see Validate).
func Parse(raw []byte) (*Build, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var b Build
	if err := dec.Decode(&b); err != nil {
		return nil, fmt.Errorf("buildfile: %w", err)
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("buildfile: trailing data after the build object")
	}
	if b.Schema != Schema {
		return nil, fmt.Errorf("buildfile: schema %q, want %q", b.Schema, Schema)
	}
	b.Class = strings.ToLower(b.Class)
	return &b, nil
}

// Marshal writes the canonical indented form.
func Marshal(b *Build) ([]byte, error) { return json.MarshalIndent(b, "", " ") }

func statIndex(s string) int {
	for i, n := range StatNames {
		if n == s {
			return i
		}
	}
	return -1
}
