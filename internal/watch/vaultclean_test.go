package watch

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The Vault page chrome (card vaultclean): vault.js in a stub DOM with fake timers (node, skipped without it).
func TestVaultCleanChromeJS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	dir := t.TempDir()
	for n, name := range map[string]string{"vault.js": "vault.mjs", "toast.js": "toast.cjs", "why.js": "why.cjs", "vault-logic.js": "vault-logic.mjs", "vault.json": "vault.json"} {
		b, err := webFS.ReadFile("web/" + n)
		if err != nil {
			t.Fatal(err)
		}
		if n == "vault.js" {
			b = []byte(strings.Replace(string(b), "'/watch/vault-logic.js'", "'./vault-logic.mjs'", 1))
		}
		os.WriteFile(filepath.Join(dir, name), b, 0o600)
	}
	hb, err := os.ReadFile("testdata/vault_clean.mjs")
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

// The grid sits in no frame and no fixed-height box: its rule and its ancestors' rules have no border, height or overflow box.
func TestVaultCleanGridHasNoFrame(t *testing.T) {
	b, err := webFS.ReadFile("web/vault.css")
	if err != nil {
		t.Fatal(err)
	}
	css := string(b)
	rule := func(sel string) string {
		m := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(sel) + ` \{([^}]*)\}`).FindStringSubmatch(css)
		if m == nil {
			t.Fatalf("no rule for %s", sel)
		}
		return m[1]
	}
	bad := regexp.MustCompile(`(^|[^-a-z])(border|height|block-size|max-height|max-block-size|overflow|overflow-y)\s*:`)
	for _, sel := range []string{".vt-grid", ".vt-main", ".vt-body"} {
		for _, d := range bad.FindAllStringSubmatch(rule(sel), -1) {
			if d[2] == "overflow" && strings.Contains(rule(sel), "overflow: visible") {
				continue
			}
			t.Errorf("%s has %s (a frame or a fixed height around the grid)", sel, d[2])
		}
	}
	if strings.Contains(css, ".vt-keeprow") || strings.Contains(css, ".vt-revealbar") || strings.Contains(css, ".vt-fold") {
		t.Error("vault.css still styles a band of the old layout")
	}
}
