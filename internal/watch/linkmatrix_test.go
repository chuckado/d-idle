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

// The page-to-page links the UI review asks for (docs/design/fable-review-2026-11-22, section E1, matrix B1): every page is
// reached from the pages it belongs next to, and every way back to the live screen is /watch?live=1 (plain /watch can land
// on the Return screen). Asserted on the served page templates and scripts, so a page that loses a link fails here.
func TestPageLinkMatrix(t *testing.T) {
	href := func(src, h string) bool { return strings.Contains(src, `href="`+h+`"`) }
	nav := func(src string) string { // the band's Other pages nav
		i := strings.Index(src, `aria-label="Other pages"`)
		if i < 0 {
			return ""
		}
		return src[i : i+strings.Index(src[i:], "</nav>")]
	}
	for _, c := range []struct {
		name, src string
		band      []string // links in the band (desktop)
		phone     []string // links in the phone list ("..."), or "" when the band keeps them
	}{
		{"inventory", charPage, []string{"/watch?live=1", "/watch/skills", "/watch/points", "/watch/vault", "/watch/policy"},
			[]string{"/watch/inventory", "/watch/skills", "/watch/points", "/watch/vault", "/watch/policy", "/watch/return"}},
		{"skills", skillsPage, []string{"/watch?live=1", "/watch/inventory", "/watch/points", "/watch/policy"}, nil},
		{"points", pointsPage, []string{"/watch?live=1", "/watch/skills", "/watch/inventory"}, nil},
	} {
		band := nav(c.src)
		for _, h := range c.band {
			if !href(band, h) {
				t.Errorf("%s: the band has no link to %s", c.name, h)
			}
		}
		if len(c.phone) > 0 {
			i := strings.Index(c.src, `id="cv-morelist"`)
			list := c.src[i : i+strings.Index(c.src[i:], "</div>")]
			for _, h := range c.phone {
				if !href(list, h) {
					t.Errorf("%s: the phone list has no link to %s", c.name, h)
				}
			}
		}
	}

	vjs, _ := webFS.ReadFile("web/vault.js")
	vj := string(vjs)
	for _, want := range []string{"d.backHref || '/watch?live=1'", "'/watch/inventory'", "'/watch/skills'", "backHref: back"} {
		if !strings.Contains(vj, want) {
			t.Errorf("vault.js: missing %q", want)
		}
	}
	cjs, _ := webFS.ReadFile("web/characters.js")
	if !strings.Contains(string(cjs), "d.backHref || '/watch?live=1'") {
		t.Error("characters.js: Back does not default to the live cockpit")
	}
	// no back target of these two pages may be the bare /watch
	bare := regexp.MustCompile(`'/watch'(\s*[:;)]|\s*$)`)
	for n, src := range map[string]string{"vault.js": vj, "characters.js": string(cjs)} {
		if bare.MatchString(src) {
			t.Errorf("%s: a back target is the bare '/watch' (use /watch?live=1)", n)
		}
	}
}

// The Vault header in a stub DOM (node, skipped without it): Back to the live cockpit, Inventory, Skills.
func TestVaultHeaderLinksJS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	dir := t.TempDir()
	for n, name := range map[string]string{"vault.js": "vault.mjs", "vault-logic.js": "vault-logic.mjs", "vault.json": "vault.json"} {
		b, err := webFS.ReadFile("web/" + n)
		if err != nil {
			t.Fatal(err)
		}
		if n == "vault.js" {
			b = []byte(strings.Replace(string(b), "'/watch/vault-logic.js'", "'./vault-logic.mjs'", 1))
		}
		os.WriteFile(filepath.Join(dir, name), b, 0o600)
	}
	hb, err := os.ReadFile("testdata/vault_nav.mjs")
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
