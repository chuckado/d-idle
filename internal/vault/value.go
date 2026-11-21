package vault

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	vaultdata "d2lod-bot/data/vault"
)

// ValueTable is data/vault/value.json: the stable trim order. It is deliberately independent of the loot rules
// (those answer "pick up?", this answers "which record goes first when the vault is full").
type ValueTable struct {
	Quality           map[string]int `json:"quality"`
	Ilvl              int            `json:"ilvl"`
	Socket            int            `json:"socket"`
	Eth               int            `json:"eth"`
	ProtectQuality    []int          `json:"protectQuality"`
	ProtectCodePrefix []string       `json:"protectCodePrefix"`
}

// DefaultValues parses the embedded table.
func DefaultValues() (*ValueTable, error) { return ParseValues(vaultdata.Value) }

// ParseValues validates a value table.
func ParseValues(b []byte) (*ValueTable, error) {
	var v ValueTable
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("vault value table: %w", err)
	}
	if len(v.Quality) == 0 {
		return nil, fmt.Errorf("vault value table: no quality weights")
	}
	return &v, nil
}

// Of is the value of an item: the same input always gives the same number.
func (v *ValueTable) Of(it *Item) int {
	n := v.Quality[strconv.Itoa(it.Quality)]*1000 + it.Ilvl*v.Ilvl + it.Sockets*v.Socket
	if it.Eth {
		n += v.Eth
	}
	return n
}

// Protected items are trimmed only when nothing else is left and the caller allows it.
func (v *ValueTable) Protected(it *Item) bool {
	for _, q := range v.ProtectQuality {
		if it.Quality == q {
			return true
		}
	}
	for _, p := range v.ProtectCodePrefix {
		if strings.HasPrefix(it.Code, p) {
			return true
		}
	}
	return false
}
