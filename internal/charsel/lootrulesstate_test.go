package charsel

import (
	"os"
	"path/filepath"
	"testing"
)

// The loot rules file follows the character: switching parks one hero's set and the next hero starts without one (so the
// store falls back to its class preset).
func TestSwapStateParksTheLootRulesOfTheHero(t *testing.T) {
	inst := t.TempDir()
	dir := filepath.Join(inst, "state", "ladder")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"lootrules-human.json", "lootrules-human.json.bak"} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("a"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := SwapState(inst, "Alpha", "Beta"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "lootrules-human.json")); !os.IsNotExist(err) {
		t.Fatal("the new hero inherited the old hero's loot rules")
	}
	if err := SwapState(inst, "Beta", "Alpha"); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"lootrules-human.json", "lootrules-human.json.bak"} {
		if b, err := os.ReadFile(filepath.Join(dir, n)); err != nil || string(b) != "a" {
			t.Errorf("%s not back after the switch: %v", n, err)
		}
	}
}
