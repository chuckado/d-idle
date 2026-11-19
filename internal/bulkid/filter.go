// Package bulkid plans and runs bulk identify (docs/design/bulk-identify-2026-11-18.md): which unidentified items are
// worth identifying, Cain trips, and the chunked moves between stash and pack. Everything here is class-free: the
// item types and qualities come from a data table.
package bulkid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"

	bulkiddata "d2lod-bot/data/bulkid"
)

// Item is what the planner knows about one item: only what the game's own tooltip shows of an unidentified item. It has
// no affix field, so nothing it feeds can leak a hidden roll.
type Item struct {
	ID      uint32   `json:"id"`
	Code    string   `json:"code"`
	Types   []string `json:"types"` // the base type chain (data.DB.IsA order), e.g. ring, jewelry
	Quality string   `json:"quality"`
	Ilvl    int      `json:"ilvl"`
	Value   int      `json:"value"` // estimated sell value (loot.SellValue lower bound)
	W       int      `json:"w,omitempty"`
	H       int      `json:"h,omitempty"`
	Place   string   `json:"place"` // PlacePack or PlaceStash
}

// Places an item can be in.
const (
	PlacePack  = "pack"
	PlaceStash = "stash"
)

// Rule is one filter row. The first rule that matches decides.
type Rule struct {
	ID       string   `json:"id"`
	Types    []string `json:"types"` // base type codes; "*" matches any
	Quality  []string `json:"quality"`
	MinIlvl  int      `json:"minIlvl"`
	Priority int      `json:"priority"`
}

// Filter is data/bulkid/filter.json.
type Filter struct {
	Schema  string `json:"schema"`
	Rules   []Rule `json:"rules"`
	Default bool   `json:"default"`
}

// DefaultFilter parses the embedded table.
func DefaultFilter() (*Filter, error) { return ParseFilter(bulkiddata.Filter) }

// ParseFilter validates a filter table; unknown fields and empty rules are errors.
func ParseFilter(b []byte) (*Filter, error) {
	var f Filter
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("bulkid filter: %w", err)
	}
	if f.Schema != "bulkid.filter/1" {
		return nil, fmt.Errorf("bulkid filter: schema %q", f.Schema)
	}
	seen := map[string]bool{}
	for _, r := range f.Rules {
		switch {
		case r.ID == "" || seen[r.ID]:
			return nil, fmt.Errorf("bulkid filter: rule id %q empty or repeated", r.ID)
		case len(r.Types) == 0 || len(r.Quality) == 0:
			return nil, fmt.Errorf("bulkid filter: rule %q needs types and quality", r.ID)
		}
		seen[r.ID] = true
	}
	return &f, nil
}

// Match reports the rule that selects it (identify) and its priority; ok false when no rule matches (the table default
// then decides, with priority 0).
func (f *Filter) Match(it Item) (id string, prio int, identify bool) {
	for _, r := range f.Rules {
		if !slices.Contains(r.Quality, it.Quality) || it.Ilvl < r.MinIlvl {
			continue
		}
		if slices.Contains(r.Types, "*") || anyIn(it.Types, r.Types) {
			return r.ID, r.Priority, true
		}
	}
	return "", 0, f.Default
}

func anyIn(have, want []string) bool {
	for _, h := range have {
		if slices.Contains(want, h) {
			return true
		}
	}
	return false
}
