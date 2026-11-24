package watch

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The Vault item sheet on a phone: the selected cell clears the bottom sheet, the sheet carries the tooltip lines, and one
// button cycles Compare off, hero, merc (testdata/vault_compare.mjs).
func TestVaultCompareJS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	dir := t.TempDir()
	for n, name := range map[string]string{"vault.js": "vault.mjs", "toast.js": "toast.cjs", "why.js": "why.cjs", "vault-logic.js": "vault-logic.mjs", "vault.json": "vault.json", "inventory-tip.js": "inventory-tip.cjs"} {
		b, err := webFS.ReadFile("web/" + n)
		if err != nil {
			t.Fatal(err)
		}
		if n == "vault.js" {
			b = []byte(strings.Replace(string(b), "'/watch/vault-logic.js'", "'./vault-logic.mjs'", 1))
		}
		os.WriteFile(filepath.Join(dir, name), b, 0o600)
	}
	hb, err := os.ReadFile("testdata/vault_compare.mjs")
	if err != nil {
		t.Fatal(err)
	}
	h := filepath.Join(dir, "h.mjs")
	os.WriteFile(h, hb, 0o600)
	b, err := exec.Command(node, h).CombinedOutput()
	if err != nil {
		t.Fatalf("node: %v\n%s", err, b)
	}
	var out map[string]any
	if err := json.Unmarshal(b[strings.Index(string(b), "{"):], &out); err != nil {
		t.Fatalf("%v\n%s", err, b)
	}
	if fails, _ := out["fails"].([]any); len(fails) > 0 {
		t.Fatalf("checks failed: %v", fails)
	}
}

// The defect behind "I cannot see the tooltip": the floating tip sits under the bottom item sheet (z-index) for any cell in
// the lower part of a phone screen. The page hides the floating tip below the docked layout and the sheet carries its lines;
// this pins the two facts the fix rests on, so removing either fails the test.
func TestVaultPhoneTipCannotHideUnderSheet(t *testing.T) {
	css, err := webFS.ReadFile("web/vault.css")
	if err != nil {
		t.Fatal(err)
	}
	tipCSS, _ := webFS.ReadFile("web/inventory-tip.css")
	if !strings.Contains(string(tipCSS), "--d-z-sheet") || !strings.Contains(string(css), "z-index: var(--d-z-banner)") {
		t.Skip("the stacking order changed: re-derive the defect")
	}
	if !strings.Contains(string(css), "@media (max-width: 899px) { [data-vault] .inv-tip { display: none; } }") {
		t.Error("below 900px the floating tooltip must be hidden: the bottom item sheet (z-index above it) covers it for a cell in the lower half")
	}
	js, _ := webFS.ReadFile("web/vault.js")
	for _, want := range []string{"clearDelta(", "keepInView(it.id)", "d-compare", "'vt-compare'"} {
		if !strings.Contains(string(js), want) {
			t.Errorf("vault.js lost %q", want)
		}
	}
}
