package watch

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type navCfg struct {
	Destinations []struct {
		ID, Label, Href string
		Match           []string
		Sheet           bool
	}
	More  []struct{ ID, Label, Href string }
	Areas map[string][]struct{ ID, Label, Href string }
}

func loadNav(t *testing.T) navCfg {
	t.Helper()
	b, err := webFS.ReadFile("web/nav.json")
	if err != nil {
		t.Fatal(err)
	}
	var c navCfg
	if err := json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	return c
}

// The page-to-page links the UI review asks for (docs/design/fable-review-2026-11-22, matrix B1): the one navigation list reaches
// every page from every page. The five destinations are the bar; More holds the rest; the Skills area has its three sub-tabs.
func TestPageLinkMatrix(t *testing.T) {
	c := loadNav(t)
	if len(c.Destinations) != 5 {
		t.Fatalf("%d destinations, want the five of review B3", len(c.Destinations))
	}
	var got []string
	for _, d := range c.Destinations {
		got = append(got, d.Label)
	}
	if strings.Join(got, ",") != "Live,Gear,Skills,Vault,More" {
		t.Errorf("destinations %v", got)
	}
	links := map[string]bool{}
	for _, d := range c.Destinations {
		links[d.Href] = true
	}
	for _, m := range c.More {
		links[m.Href] = true
	}
	for _, a := range c.Areas {
		for _, s := range a {
			links[s.Href] = true
		}
	}
	for _, want := range []string{"/watch?live=1", "/watch/inventory", "/watch/skills", "/watch/points", "/watch/points?tab=build", "/watch/vault",
		"/watch/characters", "/watch/return", "/watch/hall", "/watch/policy", "/watch?live=1&drawer=onboarding"} {
		if !links[want] {
			t.Errorf("nav.json has no link to %s", want)
		}
	}
	if links["/watch"] || links["/watch/pacts"] {
		t.Error("nav.json links the bare /watch (it can land on the Return screen) or Pacts (stays unlinked, user decision 2026-11-22)")
	}
}

// Every page template loads the shared navigation and writes no destination links of its own (docs/ui-design-system.md "Navigation").
// A page that is not in this list does not exist for the user; a new page is added here and fails until it loads nav.js and nav.css.
func TestNavEveryPage(t *testing.T) {
	live, _ := webFS.ReadFile("web/watch.html")
	vault, _ := webFS.ReadFile("web/vault.html")
	chars, _ := webFS.ReadFile("web/characters.html")
	pages := []struct{ name, path, src string }{
		{"live", "/watch", string(live)},
		{"inventory", "/watch/inventory", charPage},
		{"skills", "/watch/skills", skillsPage},
		{"points", "/watch/points", pointsPage},
		{"vault", "/watch/vault", string(vault)},
		{"characters", "/watch/characters", string(chars)},
		{"return", "/watch/return", string(RenderReturn(ReturnData{}, false))},
		{"policy", "/watch/policy", string(RenderPolicy(PolicyView{}))},
		{"hall", "/watch/hall", string(RenderHall(&HallView{}))},
	}
	c := loadNav(t)
	hasSettings := false
	for _, m := range c.More {
		hasSettings = hasSettings || m.ID == "settings"
	}
	if !hasSettings {
		t.Error("nav.json More has no Settings entry (rule revoking and alert settings live there; every page below loads nav.js, so every page reaches it)")
	}
	for _, p := range pages {
		for _, want := range []string{`<script src="/watch/nav.js"></script>`, `href="/watch/nav.css"`} {
			if !strings.Contains(p.src, want) {
				t.Errorf("%s: the page does not load %s", p.name, want)
			}
		}
		// the old per-page headers' links are gone: the band link list, the Inventory phone list, a hand-written Back
		for _, bad := range []string{`aria-label="Other pages"`, `cv-morelist`, `>Live screen</a>`, `>Back</a>`, `id="vaultlink"`} {
			if strings.Contains(p.src, bad) {
				t.Errorf("%s: still carries %q", p.name, bad)
			}
		}
		n := 0
		for _, d := range c.Destinations {
			for _, m := range d.Match {
				if m == p.path {
					n++
				}
			}
		}
		if n != 1 {
			t.Errorf("%s (%s): matched by %d destinations in nav.json, want exactly 1 (one aria-current)", p.name, p.path, n)
		}
	}
	for _, f := range []string{"web/vault.js", "web/characters.js"} {
		b, _ := webFS.ReadFile(f)
		for _, bad := range []string{"backHref", "'/watch/inventory'", "'/watch/skills'", "'/watch?live=1'"} {
			if bytes.Contains(b, []byte(bad)) {
				t.Errorf("%s: builds its own page link %s (the shared navigation is the way to other pages)", f, bad)
			}
		}
	}
	// the page of every route the nav lists is served by a route this package knows
	for _, d := range c.Destinations {
		if d.Sheet {
			continue
		}
		if d.Href == "" || !strings.HasPrefix(d.Href, "/watch") {
			t.Errorf("destination %s: href %q", d.ID, d.Href)
		}
	}
}

// web/nav.js in a stub DOM (node, skipped without it): the address picks the destination, five destinations and one aria-current on the
// bar and in the band on every page, the More sheet's focus in, Escape and focus back, the Skills sub-tabs.
func TestNavJS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	dir := t.TempDir()
	for src, dst := range map[string]string{"web/nav.js": "nav.cjs", "web/nav.json": "nav.json"} {
		b, err := webFS.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		os.WriteFile(filepath.Join(dir, dst), b, 0o600)
	}
	for _, f := range []string{"nav_dom.cjs", "nav_test.cjs"} {
		b, err := os.ReadFile(filepath.Join("testdata", f))
		if err != nil {
			t.Fatal(err)
		}
		os.WriteFile(filepath.Join(dir, f), b, 0o600)
	}
	b, err := exec.Command(node, filepath.Join(dir, "nav_test.cjs"), dir).CombinedOutput()
	if err != nil {
		t.Fatalf("node: %v\n%s", err, b)
	}
	var out struct{ Fails []string }
	if err := json.Unmarshal(b[strings.Index(string(b), "{"):], &out); err != nil {
		t.Fatalf("%v\n%s", err, b)
	}
	for _, f := range out.Fails {
		t.Error(f)
	}
}
