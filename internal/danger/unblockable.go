package danger

import (
	_ "embed"
	"encoding/json"
)

//go:embed unblockable.json
var unblockableJSON []byte

// unblockable holds the shooter keys whose missile has a zero physical record, which game.exe's block roll skips
// (docs/research/decomp26-trace.md Q10 to Q13; the per-row list is docs/research/monster-missile-trace.md Q4).
var unblockable = func() map[string]bool {
	var f struct {
		Keys []string `json:"keys"`
	}
	if err := json.Unmarshal(unblockableJSON, &f); err != nil {
		panic(err)
	}
	m := map[string]bool{}
	for _, k := range f.Keys {
		m[k] = true
	}
	return m
}()
