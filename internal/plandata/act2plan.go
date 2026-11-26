package plandata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Act2Slice is one farm slice of the Act 2 plan (the field names of levelplan/3 slices, plus the life gate).
type Act2Slice struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Areas      []int  `json:"areas"`
	Entry      int    `json:"entry"`
	Exit       int    `json:"exit"`
	NoMercAdd  int    `json:"no_merc_add"`
	WayIn      string `json:"way_in"`
	ViaWP      int    `json:"via_wp"`
	MinMaxLife int    `json:"min_max_life"`
	Why        string `json:"why"`
}

// Act2Class is the plan of one class; Archetype (melee or caster) is the key of the admission rows.
type Act2Class struct {
	Archetype string      `json:"archetype"`
	Build     string      `json:"build"`
	Tag       string      `json:"tag"`
	Slices    []Act2Slice `json:"slices"`
}

// Act2Plan is act2plan.json.
type Act2Plan struct {
	Schema     string `json:"schema"`
	Difficulty string `json:"difficulty"`
	Doc        string `json:"doc"`
	Note       string `json:"note"`
	Kit        struct {
		Healing int `json:"healing"`
		Portal  int `json:"portal"`
	} `json:"kit"`
	Merc struct {
		NPC          string            `json:"npc"`
		Act          int               `json:"act"`
		HireFromClvl int               `json:"hireFromClvl"`
		KeepThrough  string            `json:"keepThrough"`
		Aura         map[string]string `json:"aura"`
		ReviveAlways bool              `json:"reviveAlways"`
		Note         string            `json:"note"`
	} `json:"merc"`
	Duriel struct {
		Areas      []int          `json:"areas"`
		MinLevel   map[string]int `json:"minLevel"`
		MinMaxLife map[string]int `json:"minMaxLife"`
		Why        string         `json:"why"`
	} `json:"duriel"`
	Classes map[string]*Act2Class `json:"classes"`
}

// LoadAct2Plan reads the embedded act2plan.json (strict: an unknown field is an error). It does not validate the content
// against the areas table; the tests do that.
func LoadAct2Plan() (*Act2Plan, error) {
	b, err := FS.ReadFile("act2plan.json")
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var raw struct {
		Act2Plan
		Tags json.RawMessage `json:"tags"`
	}
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("act2plan.json: %w", err)
	}
	return &raw.Act2Plan, nil
}

// For is the plan of class; an unknown class is an error naming it, never another class.
func (p *Act2Plan) For(class string) (*Act2Class, error) {
	if c, ok := p.Classes[strings.ToLower(class)]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("act2plan: no plan for class %q", class)
}
