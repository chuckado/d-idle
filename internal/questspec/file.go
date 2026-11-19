package questspec

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// File is one act's chains (data/questspec/actN.json). The chains inherit the file's schema.
type File struct {
	Schema string  `json:"schema"`
	Act    int     `json:"act"`
	Source string  `json:"source,omitempty"`
	Chains []Chain `json:"chains"`
}

// LoadFile parses an act file strictly and validates every chain; chain ids are unique within the file.
func LoadFile(data []byte) ([]*Chain, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var f File
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("questspec: %w", err)
	}
	if dec.More() {
		return nil, fmt.Errorf("questspec: trailing data")
	}
	if f.Schema != Schema {
		return nil, fmt.Errorf("questspec: file schema %q, want %q", f.Schema, Schema)
	}
	seen := map[string]bool{}
	var out []*Chain
	for i := range f.Chains {
		c := &f.Chains[i]
		c.Schema = f.Schema
		if seen[c.ID] {
			return nil, fmt.Errorf("questspec: duplicate chain %q", c.ID)
		}
		seen[c.ID] = true
		if err := c.Validate(); err != nil {
			return nil, fmt.Errorf("chain %s: %w", c.ID, err)
		}
		out = append(out, c)
	}
	return out, nil
}

// Allowed is one accepted finding of the shipped data: a known gap with an owner, so the CI check stays green while the
// gap is tracked, and fails when the entry goes stale (the finding is gone) or a NEW finding appears.
type Allowed struct {
	Chain string `json:"chain"`
	Caps  string `json:"caps"`
	Step  string `json:"step"`
	Rule  int    `json:"rule"`
	Owner string `json:"owner"`
	Why   string `json:"why"`
}

// AllowFile is data/questspec/allow.json.
type AllowFile struct {
	Schema  string    `json:"schema"`
	Allowed []Allowed `json:"allowed"`
}

// LoadAllow parses the allow list strictly.
func LoadAllow(data []byte) (*AllowFile, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var a AllowFile
	if err := dec.Decode(&a); err != nil {
		return nil, fmt.Errorf("questspec allow: %w", err)
	}
	if a.Schema != "questallow/1" {
		return nil, fmt.Errorf("questspec allow: schema %q", a.Schema)
	}
	for i, e := range a.Allowed {
		if e.Chain == "" || e.Caps == "" || e.Step == "" || e.Rule <= 0 || e.Owner == "" || e.Why == "" {
			return nil, fmt.Errorf("questspec allow: entry %d needs chain, caps, step, rule, owner and why", i)
		}
	}
	return &a, nil
}

// Filter splits findings of (chain, caps) into those the allow list does not cover, and marks the entries it used.
func (a *AllowFile) Filter(chain, caps string, fs []Finding, used map[int]bool) []Finding {
	var open []Finding
	for _, f := range fs {
		hit := false
		for i, e := range a.Allowed {
			if e.Chain == chain && e.Caps == caps && e.Step == f.Step && e.Rule == f.Rule {
				used[i] = true
				hit = true
			}
		}
		if !hit {
			open = append(open, f)
		}
	}
	return open
}
