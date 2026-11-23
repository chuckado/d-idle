package input

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
)

// CapsSchema is the schema string of data/input/caps/*.json.
const CapsSchema = "inputcaps/1"

// BackendCaps is what one backend build can do with input, in the style of the questspec capability sets.
// Verbs names the control-channel capabilities (the "caps" reply of the fork); "input.select" means "cmd select"
// really sets the skill on the chosen hand.
type BackendCaps struct {
	Schema  string   `json:"schema"`
	Backend string   `json:"backend"`
	Verbs   []string `json:"verbs"`
	Note    string   `json:"note,omitempty"`
}

// Has reports whether the set lists the capability.
func (c *BackendCaps) Has(name string) bool {
	if c == nil {
		return false
	}
	for _, v := range c.Verbs {
		if v == name {
			return true
		}
	}
	return false
}

// LoadBackendCaps parses one capability file; unknown fields are an error.
func LoadBackendCaps(data []byte) (*BackendCaps, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var c BackendCaps
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("inputcaps: %w", err)
	}
	if c.Schema != CapsSchema || c.Backend == "" {
		return nil, fmt.Errorf("inputcaps: bad schema or empty backend")
	}
	return &c, nil
}

// BackendCapsByName reads caps/<name>.json from fsys (data/input embeds them).
func BackendCapsByName(fsys fs.FS, name string) (*BackendCaps, error) {
	b, err := fs.ReadFile(fsys, "caps/"+name+".json")
	if err != nil {
		return nil, fmt.Errorf("inputcaps: unknown backend %q: %w", name, err)
	}
	return LoadBackendCaps(b)
}

// SkillSelector is the verb side of skill selection: put skill id on the left or right hand. The d2d controller
// implements it with "cmd select"; the answer is only a request, callers verify the hand from state.
type SkillSelector interface {
	SelectSkillID(id int, left bool) error
}

// CapSelect is the capability name of a fork build whose "cmd select" sets the skill.
const CapSelect = "input.select"
