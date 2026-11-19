package watch

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The page logic (filters, query string, sort, the multi-select, the meter words) in node.
func TestVaultLogicJS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	js, err := webFS.ReadFile("web/vault-logic.js")
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "vault-logic.mjs")
	if err := os.WriteFile(p, js, 0o600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(node, "testdata/vault_logic.mjs", p).CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
}

// The Vault tooltip is the inventory tooltip fed by vault rows: same text as the inventory draws, no property line for an
// unidentified row, a weapon and a jewel.
func TestVaultTipJS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	js, err := webFS.ReadFile("web/vault-logic.js")
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "vault-logic.mjs")
	if err := os.WriteFile(p, js, 0o600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(node, "testdata/vault_tip.mjs", p, "web/inventory-tip.js").CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
}

// The page, its scripts, styles and texts are served by the Vault handler, a POST on it is refused, and the texts name every
// refusal code the API can send (a code without a text would show the server's raw sentence).
func TestVaultPageServedAndTexts(t *testing.T) {
	h := NewVaultAPI().Handler()
	for path, ctype := range map[string]string{
		"/watch/vault": "text/html", "/watch/vault.js": "text/javascript", "/watch/vault-logic.js": "text/javascript", "/watch/vault.css": "text/css", "/watch/vault.json": "application/json",
	} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), ctype) || w.Body.Len() == 0 {
			t.Errorf("%s: %d %q", path, w.Code, w.Header().Get("Content-Type"))
		}
		if path == "/watch/vault" && !strings.Contains(w.Body.String(), "data-vault") {
			t.Error("the page has no data-vault body attribute (the script and the fit check look for it)")
		}
	}
	var tx struct {
		Refusals map[string]string `json:"refusals"`
		Kinds    []vaultKind       `json:"kinds"`
		Sorts    map[string]string `json:"sorts"`
	}
	b, err := webFS.ReadFile("web/vault.json")
	if err != nil || json.Unmarshal(b, &tx) != nil {
		t.Fatal("vault.json unreadable")
	}
	for _, code := range []string{"off", "state", "missing", "pincap", "confirm", "ids", "internal", "dead", "verbs", "stopped", "target", "noscroll", "pinned", "insufficient", "in-use", "out-of-order", "maxed", "nothing-to-refund", "wallet", "token", "origin", "content-type", "method", "body"} {
		if tx.Refusals[code] == "" {
			t.Errorf("no text for the refusal code %q", code)
		}
	}
	for _, s := range []string{"new", "old", "quality", "ilvl", "value", "name"} {
		if tx.Sorts[s] == "" {
			t.Errorf("no label for sort %q", s)
		}
	}
	if len(tx.Kinds) < 4 {
		t.Errorf("%d kinds", len(tx.Kinds))
	}
	// the three entry points to the page
	for file, want := range map[string]string{"web/shell.js": "vault-open", "web/panels/orders/panel.js": "/watch/vault", "web/panels/log/panel.js": "vault-link"} {
		src, err := webFS.ReadFile(file)
		if err != nil || !strings.Contains(string(src), want) {
			t.Errorf("%s has no %q entry", file, want)
		}
	}
	// the tooltip script finds the grid by this id; the page loads it and its style
	page, _ := webFS.ReadFile("web/vault.html")
	for _, want := range []string{"/watch/inventory-tip.css", "/watch/vault.js"} {
		if !strings.Contains(string(page), want) {
			t.Errorf("vault.html does not load %s", want)
		}
	}
	js, _ := webFS.ReadFile("web/vault.js")
	if !strings.Contains(string(js), "/watch/inventory-tip.js") { // loaded by the script once the grid exists (the tip wires #pack at load)
		t.Error("vault.js does not load the tooltip script")
	}
	if !strings.Contains(string(js), "grid.id = 'pack'") {
		t.Error("the grid is not id pack, so inventory-tip.js would not wire it")
	}
}

// vault.js parses as a module (the page itself runs in a browser: the fit pages press it, docs/vault.md).
func TestVaultPageScriptParses(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	js, err := webFS.ReadFile("web/vault.js")
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "vault.mjs")
	if err := os.WriteFile(p, js, 0o600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(node, "--check", p).CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
}
