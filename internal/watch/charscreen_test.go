package watch

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

// The character canvas (card uic2, docs/ui/redesign-2026-11-14/briefs/C2-character-canvas.md): the node DOM goldens
// (testdata/charscreen), the page markup the older scripts and the fit check depend on, the routes, and the rule that the
// page decides nothing about requirements. The fit check itself needs a browser: tools/uishots.sh --fit.

func TestCharScreenDOMGoldens(t *testing.T) {
	runNode(t, "testdata/charscreen/canvas_dom.js", "web/inventory-canvas.js")
	runNode(t, "testdata/charscreen/merc_canvas_dom.js", "web/inventory-merc.js")
}

// Every id the scripts of the cards page write to is on the canvas page too, so those scripts run unchanged; the canvas
// has the two zones and the bar that the fit check measures.
func TestCharPageKeepsTheIdsAndFitsTheCheck(t *testing.T) {
	for _, id := range []string{"equip", "equipmeta", "pack", "packmeta", "belt", "stash", "stashmeta", "note", "detail", "queue", "repair", "msg", "gearprompt",
		"errands", "stashgold", "merc-doll", "gearsets", "btn-theme", "t-queue"} {
		if !strings.Contains(charPage, `id="`+id+`"`) {
			t.Errorf("canvas page lacks id %q", id)
		}
	}
	for _, want := range []string{`data-cockpit-slot="who"`, `data-cockpit-slot="store"`, `data-cockpit-panel data-panel="char-who"`, `data-cockpit-panel data-panel="char-store"`,
		`data-cockpit-slot="band"`, `href="/watch/inventory-canvas.css"`, `src="/watch/inventory-canvas.js"`, `src="/watch/inventory.js"`, `src="/watch/inventory-sets.js"`,
		`id="cv-q"`, `id="cv-sell"`, `id="cv-repair"`, `id="cv-sel"`} {
		if !strings.Contains(charPage, want) {
			t.Errorf("canvas page lacks %s", want)
		}
	}
	if strings.Contains(charPage, "<style") || strings.Contains(charPage, "style=") {
		t.Error("canvas page has inline styles")
	}
	// the cards a canvas has no use for: no Repair card, no Item actions form, no legend line
	for _, gone := range []string{`class="d-legend"`, `inv-repair`, `inv-queue`} {
		if strings.Contains(charPage, gone) {
			t.Errorf("canvas page still has %s", gone)
		}
	}
	// the scripts load in the order the old page used, the canvas last (it reads what they draw)
	last := -1
	for _, s := range []string{"inventory.js", "inventory-tip.js", "inventory-actions.js", "inventory-merc.js", "inventory-drag.js", "inventory-sets.js", "inventory-canvas.js"} {
		i := strings.Index(charPage, `src="/watch/`+s+`"`)
		if i < last {
			t.Errorf("script %s is out of order", s)
		}
		last = i
	}
}

func TestCharPageRoutes(t *testing.T) {
	h := NewInvAPI().Handler()
	body := func(p string) string {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, p, nil))
		if w.Code != 200 {
			t.Fatalf("GET %s: %d", p, w.Code)
		}
		return w.Body.String()
	}
	if got := body("/watch/inventory"); got != charPage {
		t.Error("/watch/inventory is not the canvas")
	}
	if got := body("/watch/inventory?view=classic"); got != invPage {
		t.Error("?view=classic is not the cards page")
	}
	for path, ctype := range map[string]string{"/watch/inventory-canvas.js": "text/javascript", "/watch/inventory-canvas.css": "text/css"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), ctype) || w.Body.Len() < 500 {
			t.Errorf("%s: %d %q", path, w.Code, w.Header().Get("Content-Type"))
		}
	}
	found := map[string]bool{}
	for _, p := range InvRoutes {
		found[p] = true
	}
	for _, p := range []string{"/watch/inventory-canvas.js", "/watch/inventory-canvas.css"} {
		if !found[p] {
			t.Errorf("InvRoutes lacks %s", p)
		}
	}
}

// No rule in the page: requirement arithmetic, class names and level or stat literals belong to the server (C1's `usable`).
func TestCanvasScriptsDecideNothing(t *testing.T) {
	arith := regexp.MustCompile(`\b(need|have|req|lvl|str|dex)\w*\s*(<|>|<=|>=|-)\s*[\w.]*\b(lvl|str|dex|level|strength|dexterity|need|have)\b|\.(need|have)\b\s*[<>]|\.req\b|class\s*(===|==|!==|indexOf)`)
	classes := regexp.MustCompile(`(?i)\b(paladin|sorceress|amazon|necromancer|barbarian|druid|assassin)\b`) // the merc script names mercenary kinds (Barbarian is one) in its display table
	for _, f := range []string{"inventory-canvas.js", "inventory-merc.js"} {
		b, err := webFS.ReadFile("web/" + f)
		if err != nil {
			t.Fatal(err)
		}
		if m := arith.FindString(string(b)); m != "" {
			t.Errorf("web/%s computes a requirement: %q", f, m)
		}
		if m := classes.FindString(string(b)); m != "" && f == "inventory-canvas.js" {
			t.Errorf("web/%s lists a class: %q", f, m)
		}
	}
	bad := regexp.MustCompile(`scrollIntoView|scrollTo\(|scrollBy\(|\.scrollTop\s*=|innerHTML|outerHTML|insertAdjacentHTML|cssText|setInterval|setTimeout|(?i)\bpin(ned)?\b|unpin`)
	b, _ := webFS.ReadFile("web/inventory-canvas.js")
	if m := bad.FindString(string(b)); m != "" {
		t.Errorf("web/inventory-canvas.js matches %q", m)
	}
	for _, m := range regexp.MustCompile(`\.focus\([^)]*\)`).FindAllString(string(b), -1) {
		if !strings.Contains(m, "preventScroll: true") {
			t.Errorf("focus call without preventScroll: %s", m)
		}
	}
}

func TestCanvasCSSRules(t *testing.T) {
	css, err := webFS.ReadFile("web/inventory-canvas.css")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(css), "@layer components") || regexp.MustCompile(`@media|!important|#[0-9a-fA-F]{3,6}\b|rgb\(|hsl\(`).Match(css) {
		t.Error("inventory-canvas.css: needs its layer, no media query, !important or colour literal")
	}
}
