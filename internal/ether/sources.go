package ether

import (
	"encoding/json"
	"fmt"

	vaultdata "d2lod-bot/data/vault"
)

// Source is one earn rule (data/vault/ether.json). Exactly one amount style applies: Amount (flat), Difficulty
// (a table by difficulty), Rows (a table by sub key, a sub key not listed is refused) or Hook (the caller
// supplies the amount, for example the Quest board's `meta` column, at most MaxHook).
type Source struct {
	ID           string         `json:"id"`
	Label        string         `json:"label"`
	Amount       int            `json:"amount,omitempty"`
	Difficulty   map[string]int `json:"difficulty,omitempty"`
	Rows         map[string]int `json:"rows,omitempty"`
	Hook         bool           `json:"hook,omitempty"`
	MaxHook      int            `json:"maxHook,omitempty"`
	ByDifficulty bool           `json:"byDifficulty,omitempty"` // the difficulty is part of the key
	DailyCap     int            `json:"dailyCap,omitempty"`     // most ETHER this source pays per UTC day; 0 = no cap
}

// Sources is data/vault/ether.json.
type Sources struct {
	Currency string   `json:"currency"`
	Sources  []Source `json:"sources"`
}

// DefaultSources parses the embedded table.
func DefaultSources() (*Sources, error) { return ParseSources(vaultdata.Ether) }

// ParseSources validates an earn table.
func ParseSources(b []byte) (*Sources, error) {
	var s Sources
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("ether sources: %w", err)
	}
	if s.Currency == "" {
		return nil, fmt.Errorf("ether sources: no currency name")
	}
	seen := map[string]bool{}
	for _, r := range s.Sources {
		if r.ID == "" || seen[r.ID] {
			return nil, fmt.Errorf("ether sources: bad or duplicate id %q", r.ID)
		}
		seen[r.ID] = true
		n := 0
		if r.Amount > 0 {
			n++
		}
		if len(r.Difficulty) > 0 {
			n++
		}
		if len(r.Rows) > 0 {
			n++
		}
		if r.Hook {
			n++
			if r.MaxHook <= 0 {
				return nil, fmt.Errorf("ether sources: %q is a hook without maxHook", r.ID)
			}
		}
		if n != 1 || r.DailyCap < 0 {
			return nil, fmt.Errorf("ether sources: %q needs exactly one amount style", r.ID)
		}
		for k, v := range r.Difficulty {
			if v <= 0 {
				return nil, fmt.Errorf("ether sources: %q difficulty %q pays %d", r.ID, k, v)
			}
		}
		for k, v := range r.Rows {
			if v <= 0 {
				return nil, fmt.Errorf("ether sources: %q row %q pays %d", r.ID, k, v)
			}
		}
	}
	return &s, nil
}

func (s *Sources) find(id string) (*Source, bool) {
	for i := range s.Sources {
		if s.Sources[i].ID == id {
			return &s.Sources[i], true
		}
	}
	return nil, false
}
