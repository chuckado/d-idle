package watch

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// vault.js in a stub DOM (node, skipped without it): the wallet chip, the Upgrade and Refund buttons with their reasons, the
// confirm sheet words, the read back after a press, and refusals that are never drawn as success (docs/ether.md).
func TestVaultEtherUIJS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	dir := t.TempDir()
	for n, name := range map[string]string{"vault.js": "vault.mjs", "toast.js": "toast.cjs", "vault-logic.js": "vault-logic.mjs", "vault.json": "vault.json"} {
		b, err := webFS.ReadFile("web/" + n)
		if err != nil {
			t.Fatal(err)
		}
		if n == "vault.js" {
			b = []byte(strings.Replace(string(b), "'/watch/vault-logic.js'", "'./vault-logic.mjs'", 1))
		}
		os.WriteFile(filepath.Join(dir, name), b, 0o600)
	}
	h := filepath.Join(dir, "h.mjs")
	hb, err := os.ReadFile("testdata/vault_ether.mjs")
	if err != nil {
		t.Fatal(err)
	}
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
