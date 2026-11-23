package watch

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Sell all on the Vault page (card sellall): vault.js in a stub DOM with fake timers (node, skipped without it).
func TestVaultSellAllJS(t *testing.T) {
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
	hb, err := os.ReadFile("testdata/vault_sellall.mjs")
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
