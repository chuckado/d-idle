package vault

import (
	"encoding/json"
	"fmt"

	vaultdata "d2lod-bot/data/vault"
)

// Step is one purchasable capacity upgrade.
type Step struct {
	ID    string `json:"id"`
	Add   int    `json:"add"`
	Price int    `json:"price"`
}

// Tiers is data/vault/capacity.json.
type Tiers struct {
	Base       int    `json:"base"`
	Max        int    `json:"max"`
	PinPercent int    `json:"pinPercent"`
	Steps      []Step `json:"steps"`
}

// DefaultTiers parses the embedded table.
func DefaultTiers() (*Tiers, error) { return ParseTiers(vaultdata.Capacity) }

// ParseTiers validates a tier table.
func ParseTiers(b []byte) (*Tiers, error) {
	var t Tiers
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, fmt.Errorf("vault capacity table: %w", err)
	}
	if t.Base <= 0 || t.Max < t.Base || t.PinPercent < 0 || t.PinPercent > 100 {
		return nil, fmt.Errorf("vault capacity table: bad base/max/pinPercent")
	}
	seen := map[string]bool{}
	for _, s := range t.Steps {
		if s.ID == "" || seen[s.ID] || s.Add <= 0 || s.Price < 0 {
			return nil, fmt.Errorf("vault capacity table: bad step %q", s.ID)
		}
		seen[s.ID] = true
	}
	return &t, nil
}

// Capacity is base plus the owned steps, capped at Max. Unknown step ids are ignored (a step removed from the
// table does not crash an old account). override > 0 replaces the result (tests, support).
func (t *Tiers) Capacity(owned []string, override int) int {
	if override > 0 {
		return override
	}
	has := map[string]bool{}
	for _, id := range owned {
		has[id] = true
	}
	c := t.Base
	for _, s := range t.Steps {
		if has[s.ID] {
			c += s.Add
		}
	}
	return min(c, t.Max)
}

// Next is the first step not yet owned, in table order; ok is false when all are owned.
func (t *Tiers) Next(owned []string) (Step, bool) {
	has := map[string]bool{}
	for _, id := range owned {
		has[id] = true
	}
	for _, s := range t.Steps {
		if !has[s.ID] {
			return s, true
		}
	}
	return Step{}, false
}

// Account is the account-scoped hook the vault reads (V7 implements it over the ledger). The vault never writes
// the wallet itself.
type Account interface {
	OwnedSteps() ([]string, error)
}

// Wallet is the account meta-currency hook (ETHER, V7). Spend must be idempotent per key and return an error
// without changing anything when the balance is short.
type Wallet interface {
	Balance() (int, error)
	Spend(key string, amount int) error
}

// CapacityFrom reads the owned steps from acc and returns the capacity; a nil acc gives the base.
func (t *Tiers) CapacityFrom(acc Account, override int) (int, error) {
	if acc == nil {
		return t.Capacity(nil, override), nil
	}
	owned, err := acc.OwnedSteps()
	if err != nil {
		return 0, err
	}
	return t.Capacity(owned, override), nil
}
