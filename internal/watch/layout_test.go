package watch

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The shipped layouts (web/layout.json, the default, and the editor's three presets) pass the strict check against
// the real page: known ids, nothing overlapping, spans inside the columns.
func TestShippedLayoutsAreValid(t *testing.T) {
	known, err := LayoutKnownIDs()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"web/layout.json", "web/layouts/mockup.json", "web/layouts/live.json", "web/layouts/command.json",
		"web/layouts/classic.json", "web/layouts/simple.json", "web/layouts/standard.json", "web/layouts/stream.json", "web/layouts/debug.json"} {
		b, err := webFS.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		l, err := ParseLayout(b)
		if err == nil {
			err = l.Check(known)
		}
		if err != nil {
			t.Errorf("%s: %v", f, err)
		}
		for _, bp := range []string{"phone", "tablet", "laptop", "wide"} {
			if len(l.Breakpoints[bp].Panels) == 0 {
				t.Errorf("%s: no panels at %s", f, bp)
			}
		}
	}
	// the default is Standard, byte for byte; classic is the pre-redesign default, frozen (docs/ui/redesign-2026-11-14/D-layouts.md section 4)
	def, _ := webFS.ReadFile("web/layout.json")
	std, _ := webFS.ReadFile("web/layouts/standard.json")
	if !bytes.Equal(def, std) {
		t.Error("web/layout.json is not web/layouts/standard.json")
	}
	classic, _ := webFS.ReadFile("web/layouts/classic.json")
	if sum := fmt.Sprintf("%x", sha256.Sum256(classic)); sum != classicLayoutSHA256 {
		t.Errorf("web/layouts/classic.json changed (sha256 %s): it is the layout of before the redesign and is only ever added to", sum)
	}
	// a preset names every panel of the main area (placed or switched off) and every card, so nothing is left to the packer;
	// an opt-in panel is named only where the preset wants it
	ms, _ := Manifests()
	lc, _ := shellLayoutConf()
	for _, f := range []string{"web/layout.json", "web/layouts/classic.json", "web/layouts/simple.json", "web/layouts/standard.json", "web/layouts/stream.json", "web/layouts/debug.json"} {
		if f == "web/layouts/stream.json" || f == "web/layouts/simple.json" {
			continue // short presets by design: the rest is left to the packer, which never mounts an opt-in panel it does not name
		}
		b, _ := webFS.ReadFile(f)
		l, _ := ParseLayout(b)
		named := map[string]bool{}
		for _, p := range l.Breakpoints["wide"].Panels {
			named[p.ID] = true
		}
		for _, m := range ms {
			if m.Band != "status" && m.MergedInto == "" && !m.OptIn && !named[m.ID] {
				t.Errorf("%s does not name the panel %s", f, m.ID)
			}
		}
		if f != "web/layouts/classic.json" {
			for _, id := range lc.Compat {
				if !named[id] {
					t.Errorf("%s does not name the card %s", f, id)
				}
			}
		}
	}
}

// sha256 of web/layouts/classic.json, which is web/layout.json as it was before the redesign (origin/main at 2026-11-14).
const classicLayoutSHA256 = "2e8e5c1447908cd9536800e0ecc2cc818eab94fc855a6a5f91cddf9bfc866fa0"

// Check catches what the loader lets through.
func TestLayoutCheck(t *testing.T) {
	mk := func(panels ...LayoutPanel) LayoutManifest {
		return LayoutManifest{Schema: layoutSchema, Cols: 12, Breakpoints: map[string]LayoutBreakpoint{"wide": {Cols: 12, Panels: panels}}}
	}
	a := LayoutPanel{ID: "a", Col: 1, Row: 1, ColSpan: 6, RowSpan: 3}
	cases := []struct {
		name string
		l    LayoutManifest
		want string
	}{
		{"ok", mk(a, LayoutPanel{ID: "b", Col: 7, Row: 1, ColSpan: 6, RowSpan: 3, Order: 1}), ""},
		{"overlap", mk(a, LayoutPanel{ID: "b", Col: 6, Row: 3, ColSpan: 2, RowSpan: 1, Order: 1}), "overlaps"},
		{"collapsed takes one row", mk(LayoutPanel{ID: "a", Col: 1, Row: 1, ColSpan: 6, RowSpan: 3, Collapsed: true}, LayoutPanel{ID: "b", Col: 1, Row: 2, ColSpan: 6, RowSpan: 1, Order: 1}), ""},
		{"twice", mk(a, LayoutPanel{ID: "a", Col: 7, Row: 1, ColSpan: 6, RowSpan: 3, Order: 1}), "twice"},
		{"order", mk(a, LayoutPanel{ID: "b", Col: 7, Row: 1, ColSpan: 6, RowSpan: 3}), "order 0"},
		{"unknown id", mk(LayoutPanel{ID: "zz", Col: 1, Row: 1, ColSpan: 1, RowSpan: 1}), "unknown panel id"},
	}
	known := func(id string) bool { return id == "a" || id == "b" }
	for _, c := range cases {
		err := c.l.Check(known)
		if c.want == "" && err != nil || c.want != "" && (err == nil || !strings.Contains(err.Error(), c.want)) {
			t.Errorf("%s: %v, want %q", c.name, err, c.want)
		}
	}
	bad := mk(a)
	bad.Breakpoints["huge"] = LayoutBreakpoint{}
	if err := bad.Check(nil); err == nil {
		t.Error("an unknown breakpoint name passed")
	}
}

// GET /api/layout serves the default; POST stores a per-profile copy that survives a restart; reset restores the
// default. The route has the gates of the other answer routes and the viewer never writes.
func TestLayoutRoutes(t *testing.T) {
	r := newRig(t, "paladin")
	path := filepath.Join(t.TempDir(), "state", "layout-human.json")
	r.sim.Hub.SetLayoutPath(path)
	def, _ := webFS.ReadFile("web/layout.json")
	if w := r.get(LayoutRoute, nil); w.Code != 200 || !bytes.Equal(w.Body.Bytes(), def) {
		t.Fatalf("GET: %d, body is not web/layout.json", w.Code)
	}
	if w := r.postRaw(r.view, LayoutRoute, `{"reset":true}`, nil); w.Code != 405 {
		t.Errorf("the viewer accepted a POST: %d", w.Code)
	}
	hdr := map[string]string{"Origin": origin}
	// the editor's raw export (it has overlapping entries and names panels this page lacks) is accepted unchanged
	raw, err := os.ReadFile("testdata/compat/layout-v1-command.json")
	if err != nil {
		t.Fatal(err)
	}
	w := r.postRaw(r.ans, LayoutRoute, `{"layout":`+string(raw)+`}`, hdr)
	if w.Code != 200 {
		t.Fatalf("POST: %d %s", w.Code, w.Body.String())
	}
	var rep struct {
		OK     bool           `json:"ok"`
		Layout LayoutManifest `json:"layout"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &rep); err != nil || !rep.OK || rep.Layout.Breakpoints["wide"].Panels[0].ID != "video" {
		t.Fatalf("reply %s: %v", w.Body.String(), err)
	}
	if b, err := os.ReadFile(path); err != nil || !bytes.Equal(bytes.TrimSpace(b), bytes.TrimSpace(raw)) {
		t.Errorf("the stored copy is not the posted file: %v", err)
	}
	// a fresh hub with the same profile path serves the stored layout; another profile still gets the default
	r2 := newRig(t, "paladin")
	r2.sim.Hub.SetLayoutPath(path)
	if w := r2.get(LayoutRoute, nil); !bytes.Equal(bytes.TrimSpace(w.Body.Bytes()), bytes.TrimSpace(raw)) {
		t.Error("the stored layout did not survive a restart")
	}
	r3 := newRig(t, "paladin")
	r3.sim.Hub.SetLayoutPath(filepath.Join(filepath.Dir(path), "layout-other.json"))
	if w := r3.get(LayoutRoute, nil); !bytes.Equal(w.Body.Bytes(), def) {
		t.Error("another profile got this profile's layout")
	}
	for name, c := range map[string]struct {
		body string
		hdr  map[string]string
		code int
	}{
		"foreign origin": {`{"reset":true}`, map[string]string{"Origin": "http://evil.example"}, 403},
		"later schema":   {`{"layout":{"schema":"watch.layout/2"}}`, hdr, 400},
		"outside grid":   {`{"layout":{"schema":"watch.layout/1","cols":12,"breakpoints":{"wide":{"cols":12,"panels":[{"id":"video","col":9,"row":1,"colSpan":6,"rowSpan":1}]}}}}`, hdr, 400},
		"empty":          {`{}`, hdr, 400},
		"unknown field":  {`{"layout":null,"x":1}`, hdr, 400},
		"too large":      {`{"layout":{"schema":"watch.layout/1","pad":"` + strings.Repeat("x", layoutMaxBody) + `"}}`, hdr, 400},
		"not json":       {`{"reset":true}`, map[string]string{"Origin": origin, "Content-Type": "text/plain"}, 415},
	} {
		if w := r.postRaw(r.ans, LayoutRoute, c.body, c.hdr); w.Code != c.code {
			t.Errorf("%s: %d, want %d (%s)", name, w.Code, c.code, w.Body.String())
		}
	}
	if w := r.get(LayoutRoute, nil); !bytes.Equal(bytes.TrimSpace(w.Body.Bytes()), bytes.TrimSpace(raw)) {
		t.Error("a refused POST changed the layout")
	}
	if w := r.postRaw(r.ans, LayoutRoute, `{"reset":true}`, hdr); w.Code != 200 {
		t.Fatalf("reset: %d %s", w.Code, w.Body.String())
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("reset left the stored copy")
	}
	if w := r.get(LayoutRoute, nil); !bytes.Equal(w.Body.Bytes(), def) {
		t.Error("reset did not restore the default")
	}
	// a token-guarded answer route refuses a POST without it
	guarded := r.sim.Hub.AnswerHandler(AnswerOptions{Token: "tok", Origins: []string{origin}})
	if w := r.postRaw(guarded, LayoutRoute, `{"reset":true}`, hdr); w.Code != 401 {
		t.Errorf("no token: %d", w.Code)
	}
	// a stored file that no longer parses is ignored, not served
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, []byte(`{"schema":"watch.layout/9"}`), 0o644)
	r.sim.Hub.SetLayoutPath(path)
	if w := r.get(LayoutRoute, nil); !bytes.Equal(w.Body.Bytes(), def) {
		t.Error("an unreadable stored layout was served")
	}
}

// /watch/panels.json carries the id mapping of the layout, and the layout assets are served.
func TestLayoutAssetsAndConf(t *testing.T) {
	r := newRig(t, "sorceress")
	var pv PanelsView
	r.get("/watch/panels.json", &pv)
	if pv.Layout.Aliases["whyline"] != "why" || !containsStr(pv.Layout.Compat, "orders-map") || !containsStr(pv.Layout.Reserved, "risk") {
		t.Errorf("layout conf: %+v", pv.Layout)
	}
	page, _ := webFS.ReadFile("web/watch.html")
	for _, id := range pv.Layout.Compat {
		if !bytes.Contains(page, []byte(`data-lid="`+id+`"`)) {
			t.Errorf("watch.html has no card with data-lid=%q", id)
		}
	}
	for a, to := range pv.Layout.Aliases {
		if _, err := webFS.ReadFile("web/panels/" + to + "/panel.json"); err != nil && !containsStr(pv.Layout.Compat, to) {
			t.Errorf("alias %s points at %s, which is not on the page", a, to)
		}
	}
	for p, ctype := range map[string]string{"/watch/layout.js": "text/javascript", "/watch/layoutedit.js": "text/javascript", "/watch/layouts/mockup.json": "application/json",
		"/watch/layouts/live.json": "application/json", "/watch/layouts/command.json": "application/json", "/watch/layouts/classic.json": "application/json",
		"/watch/layouts/simple.json": "application/json", "/watch/layouts/standard.json": "application/json", "/watch/layouts/stream.json": "application/json",
		"/watch/layouts/debug.json": "application/json"} {
		w := httptest.NewRecorder()
		r.view.ServeHTTP(w, httptest.NewRequest("GET", p, nil))
		if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), ctype) {
			t.Errorf("%s: %d %s", p, w.Code, w.Header().Get("Content-Type"))
		}
	}
	if w := r.get("/watch/layouts/other.json", nil); w.Code != 404 {
		t.Errorf("an unknown preset: %d", w.Code)
	}
}

// The layout scripts keep the page rules: the editor is loaded only on demand, nothing injects styles (the page CSP is
// style-src 'self'), focus never scrolls the page, the editor has no timer or fetch of its own, text is textContent.
func TestLayoutJSPatterns(t *testing.T) {
	page, _ := webFS.ReadFile("web/watch.html")
	if bytes.Contains(page, []byte("layoutedit.js")) || !bytes.Contains(page, []byte(`id="btn-layout"`)) {
		t.Error("watch.html must offer Edit layout and must not load layoutedit.js itself")
	}
	shell, _ := webFS.ReadFile("web/shell.js")
	for _, want := range []string{"import('/watch/layoutedit.js')", "from '/watch/layout.js'", "'/api/layout'", "l.schema === 'watch.layout/1'"} {
		if !bytes.Contains(shell, []byte(want)) {
			t.Errorf("shell.js lacks %s", want)
		}
	}
	// which layout loads (S1 review R2 to R4): a missing X-Layout-Source header reads as stored, a new hero has ticked nothing and
	// is pinned to Simple, the overlay strip ignores the layout's status list
	for _, want := range []string{"srv.source !== 'default'", "!(o.body.steps || []).some((x) => x.done)", "pref(PREF, 'simple')", "selectPanels(list, sel && sel.layout, html.dataset.mode)"} {
		if !bytes.Contains(shell, []byte(want)) {
			t.Errorf("shell.js lacks %s", want)
		}
	}
	bad := regexp.MustCompile(`createElement\(\s*['"]style['"]|setAttribute\(\s*['"]style['"]|insertAdjacentHTML|innerHTML|outerHTML|cssText|<style|\sstyle\s*=`)
	own := regexp.MustCompile(`\b(setInterval|setTimeout|fetch|EventSource|XMLHttpRequest)\s*\(`)
	focus := regexp.MustCompile(`\.focus\(([^)]*)\)`)
	for _, name := range []string{"layout.js", "layoutedit.js"} {
		b, err := webFS.ReadFile("web/" + name)
		if err != nil {
			t.Fatal(err)
		}
		if m := bad.Find(b); m != nil {
			t.Errorf("web/%s: %q is blocked by the page CSP or is an HTML sink", name, m)
		}
		if m := own.Find(b); m != nil {
			t.Errorf("web/%s: %q (the shell owns timers and requests)", name, m)
		}
		for _, m := range focus.FindAllSubmatch(b, -1) {
			if !bytes.Contains(m[1], []byte("preventScroll")) {
				t.Errorf("web/%s: %s scrolls the page", name, m[0])
			}
		}
	}
}

// The layout logic in node: the editor's raw exports (compat fixtures) and the shipped files plan without overlap at
// every width, editor ids land on the real panels.
func TestLayoutPlanJS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not available")
	}
	args := []string{"testdata/layout_plan.js", "web/layout.js", "testdata/compat/layout-v1-mockup.json", "testdata/compat/layout-v1-live.json",
		"testdata/compat/layout-v1-command.json", "web/layouts/classic.json", "web/layouts/mockup.json", "web/layouts/live.json", "web/layouts/command.json"}
	if out, err := exec.Command(node, args...).CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
}

// Every layout fixture of v1 (files the editor exported) still loads; fixtures are only added.
func TestCompatLayoutFixturesLoad(t *testing.T) {
	files, _ := filepath.Glob("testdata/compat/layout-v1-*.json")
	if len(files) < 3 {
		t.Fatalf("layout fixtures: %v", files)
	}
	for _, f := range files {
		b, _ := os.ReadFile(f)
		if l, err := ParseLayout(b); err != nil || len(l.Breakpoints["wide"].Panels) == 0 {
			t.Errorf("%s no longer loads: %v", f, err)
		}
	}
}

// The editor's move in node: the minimap-sized panel drops exactly into free cells (hidden panels reserve nothing),
// overlap pushes down and cascades, a preview never changes the layout it started from, a drop back is a no-op.
func TestLayoutMoveJS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not available")
	}
	if out, err := exec.Command(node, "testdata/layout_move.js", "web/layout.js", "web/layout.json").CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
}

// The live preview of the editor: it is computed from the committed cells every time (so going back is free), it is
// dropped on Escape, on a release outside the grid and on pointercancel, Enter commits a keyboard preview, the slide
// honours reduced motion, and nothing but Save reaches the stored layout. The normal page never loads any of it.
func TestLayoutEditPreviewContract(t *testing.T) {
	b, err := webFS.ReadFile("web/layoutedit.js")
	if err != nil {
		t.Fatal(err)
	}
	js := string(b)
	for _, want := range []string{"moveItem(base, id, locked(id, rect), cols)", "prefers-reduced-motion: reduce", "k === 'Escape'", "k === 'Enter'", "'pointercancel', cancel",
		"if (keep) commit(next); else revert();", "main.dataset.preview = id", "delete main.dataset.preview"} {
		if !strings.Contains(js, want) {
			t.Errorf("layoutedit.js lacks %q", want)
		}
	}
	if n := strings.Count(js, "api.save("); n != 1 {
		t.Errorf("api.save is called from %d places; only the Save button may store the layout", n)
	}
	if strings.Contains(js, "api.preview(") {
		t.Error("a preview must not go through the shell's layout object")
	}
	css, _ := webFS.ReadFile("web/shell.css")
	for _, want := range []string{`[data-flip="go"] { transition: transform var(--d-dur-2) var(--d-ease); }`, ".d-main[data-preview]"} {
		if !strings.Contains(string(css), want) {
			t.Errorf("shell.css lacks %q", want)
		}
	}
	for _, sel := range regexp.MustCompile(`(?m)^\s*(\S[^{]*(data-flip|data-preview|data-editcell)[^{]*)\{`).FindAllStringSubmatch(string(css), -1) {
		// .d-editframe exists only while the editor runs
		if !strings.Contains(sel[1], "[data-editing]") && !strings.Contains(sel[1], "[data-preview]") && !strings.Contains(sel[1], ".d-editframe") {
			t.Errorf("an edit-mode rule applies to the normal page: %s", sel[1])
		}
	}
}

// The editor driven in a stub DOM (node): a pointer drag previews live and pushes panels down, a drop commits, Escape,
// a release outside the grid and pointercancel revert, the keyboard previews until Enter, Save posts the unchanged
// export format, Done leaves no trace.
func TestLayoutEditJS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not available")
	}
	out, err := exec.Command(node, "testdata/layout_edit.js", "web/layout.js", "web/layoutedit.js").CombinedOutput()
	if err != nil || !strings.Contains(string(out), "ok") {
		t.Fatalf("%v\n%s", err, out)
	}
}

// Every card on /watch is a layout target: each pre-shell <section> carries a data-lid that shell.json lists, each
// panel outside the top band is placed by the default layout, and the editor frames them all the same way (the frame
// test is testdata/layout_edit.js). What is not a card stays out on purpose: the top band and the strips above the grid.
func TestEveryWatchCardIsALayoutTarget(t *testing.T) {
	page, _ := webFS.ReadFile("web/watch.html")
	lc, err := shellLayoutConf()
	if err != nil {
		t.Fatal(err)
	}
	secs := regexp.MustCompile(`<section\b[^>]*>`).FindAll(page, -1)
	if len(secs) < 10 {
		t.Fatalf("only %d sections found on the page", len(secs))
	}
	lid := regexp.MustCompile(`data-lid="([a-z0-9-]+)"`)
	seen := map[string]bool{}
	for _, s := range secs {
		m := lid.FindSubmatch(s)
		switch {
		case m == nil:
			t.Errorf("a card without a layout id cannot be edited: %s", s)
		case !containsStr(lc.Compat, string(m[1])):
			t.Errorf("data-lid %q is not in the compat list of shell.json", m[1])
		case seen[string(m[1])]:
			t.Errorf("data-lid %q twice", m[1])
		default:
			seen[string(m[1])] = true
		}
	}
	for _, id := range lc.Compat {
		if !seen[id] {
			t.Errorf("shell.json lists the card %q, the page has none", id)
		}
	}
	js, _ := webFS.ReadFile("web/shell.js")
	if !strings.Contains(string(js), `querySelectorAll('.d-panel[data-panel], [data-lid]')`) {
		t.Error("shell.js must take every panel root and every data-lid card as a layout target")
	}
	ed, _ := webFS.ReadFile("web/layoutedit.js")
	if regexp.MustCompile(`\bn\.(append|appendChild|prepend|insertBefore)\(`).Match(ed) {
		t.Error("the editor must not put its handle inside a panel: a card that redraws itself wipes it")
	}
}

// The game picture always fits its panel: contained and centred in a frame that is 4:3 of the width when the cell
// is as tall as its content and the whole body when the layout gives the cell a height; the card head is never
// squeezed; the panel tells the editor its shape.
func TestVideoPictureFitsItsPanel(t *testing.T) {
	css, _ := webFS.ReadFile("web/panels/video/panel.css")
	for _, want := range []string{"aspect-ratio: 4 / 3", "object-fit: contain", "object-position: center", "min-block-size: 0"} {
		if !strings.Contains(string(css), want) {
			t.Errorf("video panel.css lacks %q", want)
		}
	}
	// the card's height comes from its width: the panel declares fit, so its rows are measured, never the stored span
	sh, _ := webFS.ReadFile("web/shell.js")
	for _, want := range []string{"if (m.fit) root.dataset.fit = '1'", "function measureFit()", "contentHeight(n, getComputedStyle(n))", "new ResizeObserver(scheduleFit)", "'fullscreenchange', scheduleFit"} {
		if !strings.Contains(string(sh), want) {
			t.Errorf("shell.js lacks %q", want)
		}
	}
	js, _ := webFS.ReadFile("web/panels/video/panel.js")
	if !strings.Contains(string(js), "root.dataset.aspect = '4/3'") {
		t.Error("the video panel must declare its aspect for the editor")
	}
	// the default layout gives the picture rows that match 4:3 at the width of each breakpoint (within one row)
	// (Standard's picture is a fit panel too: its stored span is not used, so the rule is for the layouts that predate fit)
	raw, _ := webFS.ReadFile("web/layouts/classic.json")
	l, err := ParseLayout(raw)
	if err != nil {
		t.Fatal(err)
	}
	for bp, width := range map[string]float64{"laptop": 1280, "wide": 1920} {
		for _, p := range l.Breakpoints[bp].Panels {
			if p.ID != "video" {
				continue
			}
			colw := (width - 24 - 11*12) / 12
			need := (62 + (float64(p.ColSpan)*colw+float64(p.ColSpan-1)*12-24)*3/4 + 12) / 60
			if d := float64(p.RowSpan) - need; d < -0.01 || d > 1 {
				t.Errorf("%s: video spans %d rows, 4:3 needs %.1f", bp, p.RowSpan, need)
			}
		}
	}
}

// hidden: true on an entry is additive: old files (no key) load as before, a hidden entry reserves nothing.
func TestLayoutHiddenEntry(t *testing.T) {
	l, err := ParseLayout([]byte(`{"schema":"watch.layout/1","cols":12,"breakpoints":{"wide":{"cols":12,"panels":[
		{"id":"a","col":1,"row":1,"colSpan":6,"rowSpan":3,"order":0,"hidden":true},{"id":"b","col":1,"row":1,"colSpan":6,"rowSpan":3,"order":1}]}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if !l.Breakpoints["wide"].Panels[0].Hidden || l.Breakpoints["wide"].Panels[1].Hidden {
		t.Errorf("hidden not read: %+v", l.Breakpoints["wide"].Panels)
	}
	if err := l.Check(nil); err != nil {
		t.Errorf("a hidden entry must not count as an overlap: %v", err)
	}
	b, _ := json.Marshal(l.Breakpoints["wide"].Panels[1])
	if strings.Contains(string(b), "hidden") {
		t.Errorf("a shown entry must not grow a hidden key: %s", b)
	}
}

// Fit content: every panel of the main area and every pre-shell card declares it, so no card keeps a stored span
// taller than what it shows; the manifest key is additive (a manifest without it loads, with the stored spans).
func TestPanelsFitTheirContent(t *testing.T) {
	ms, err := Manifests()
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range ms {
		if (m.Band == "side" || m.Band == "stage") && !m.Fit {
			t.Errorf("panel %s does not fit its content", m.ID)
		}
		if m.Band == "status" && m.Fit {
			t.Errorf("panel %s is in the top band; fit is for the grid", m.ID)
		}
	}
	lc, _ := shellLayoutConf()
	for _, id := range lc.Compat {
		if !containsStr(lc.Fit, id) {
			t.Errorf("the card %s does not fit its content (shell.json layout.fit)", id)
		}
	}
	old, _ := os.ReadFile("testdata/compat/panels-v1/video.json")
	if m, err := parseManifest(old, "video"); err != nil || m.Fit {
		t.Errorf("a v1 manifest must load without fit: %v %+v", err, m)
	}
	var pv PanelsView
	newRig(t, "paladin").get("/watch/panels.json", &pv)
	if len(pv.Layout.Fit) == 0 {
		t.Error("/watch/panels.json does not carry the fit list")
	}
	fit := 0
	for _, m := range pv.Panels {
		if m.Fit {
			fit++
		}
	}
	if fit == 0 {
		t.Error("/watch/panels.json does not carry the fit flag")
	}
}

// The fit in node: a video panel 600 px wide whose stored span is far too tall renders with no empty area.
func TestLayoutFitJS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not available")
	}
	if out, err := exec.Command(node, "testdata/layout_fit.js", "web/layout.js").CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
}

// S1 in node: old layouts keep their cells (the golden is what the shell before S1 planned), Standard and Simple resolve against
// today's panel list with every fallback in its cell, an opt-in panel stays unmounted unless named, a stand-in saves as its tile.
func TestLayoutS1JS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not available")
	}
	out, err := exec.Command(node, "testdata/layout_s1.js", "check", "web/layout.js", "testdata/layout_cells_s1.json", "web", "testdata/compat").CombinedOutput()
	if err != nil || !strings.Contains(string(out), "ok") {
		t.Fatalf("%v\n%s", err, out)
	}
}

// The compat fixtures and the old default (classic) still parse to the same placed ids and cells as before S1: the Go loader's side
// of TestLayoutS1JS (the golden holds the node plan; here the raw entries are compared with the file).
func TestOldLayoutsKeepTheirEntries(t *testing.T) {
	for _, f := range []string{"testdata/compat/layout-v1-mockup.json", "testdata/compat/layout-v1-live.json", "testdata/compat/layout-v1-command.json"} {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		l, err := ParseLayout(b)
		if err != nil || len(l.Status) != 0 {
			t.Errorf("%s: %v status %v", f, err, l.Status)
		}
		for bp, v := range l.Breakpoints {
			for _, p := range v.Panels {
				if p.Fallback != "" {
					t.Errorf("%s %s: %s has a fallback", f, bp, p.ID)
				}
			}
		}
	}
}

// The layout keys of S1: status entries are ids or {id, fallback}, an entry may name a fallback, both round-trip, and an
// unknown id passes only with a (known) fallback.
func TestLayoutStatusAndFallback(t *testing.T) {
	src := `{"schema":"watch.layout/1","cols":12,"status":["a",{"id":"later","fallback":"b"}],"breakpoints":{"wide":{"cols":12,"panels":[
		{"id":"later","col":1,"row":1,"colSpan":6,"rowSpan":3,"order":0,"fallback":"b"},{"id":"b","col":1,"row":1,"colSpan":6,"rowSpan":3,"order":1,"hidden":true}]}}}`
	l, err := ParseLayout([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Status) != 2 || l.Status[0] != (LayoutStatus{ID: "a"}) || l.Status[1] != (LayoutStatus{ID: "later", Fallback: "b"}) || l.Breakpoints["wide"].Panels[0].Fallback != "b" {
		t.Fatalf("read: %+v", l)
	}
	known := func(id string) bool { return id == "a" || id == "b" }
	if err := l.Check(known); err != nil {
		t.Errorf("an unbuilt id with a known fallback: %v", err)
	}
	if b, _ := json.Marshal(l.Status); string(b) != `["a",{"id":"later","fallback":"b"}]` {
		t.Errorf("status marshals as %s", b)
	}
	l.Breakpoints["wide"].Panels[0].Fallback = "zz"
	if err := l.Check(known); err == nil || !strings.Contains(err.Error(), "fallback") {
		t.Errorf("unknown fallback: %v", err)
	}
	l.Breakpoints["wide"].Panels[0].Fallback = ""
	if err := l.Check(known); err == nil {
		t.Error("an unknown id without a fallback passed")
	}
	if _, err := ParseLayout([]byte(`{"schema":"watch.layout/1","cols":12,"status":[7],"breakpoints":{}}`)); err == nil {
		t.Error("a status entry that is a number was accepted")
	}
}

// GET /api/layout says whether the layout is the hero's stored copy (the shell starts a new hero on Simple only when it is not),
// and the panel manifest key optIn reaches /watch/panels.json.
func TestLayoutSourceHeader(t *testing.T) {
	r := newRig(t, "paladin")
	path := filepath.Join(t.TempDir(), "layout-human.json")
	r.sim.Hub.SetLayoutPath(path)
	get := func() string {
		w := httptest.NewRecorder()
		r.view.ServeHTTP(w, httptest.NewRequest("GET", LayoutRoute, nil))
		return w.Header().Get("X-Layout-Source")
	}
	if got := get(); got != "default" {
		t.Errorf("no stored layout: %q", got)
	}
	hdr := map[string]string{"Origin": origin}
	raw, _ := os.ReadFile("testdata/compat/layout-v1-live.json")
	if w := r.postRaw(r.ans, LayoutRoute, `{"layout":`+string(raw)+`}`, hdr); w.Code != 200 {
		t.Fatalf("POST: %d", w.Code)
	}
	if got := get(); got != "stored" {
		t.Errorf("after a save: %q", got)
	}
	if w := r.postRaw(r.ans, LayoutRoute, `{"reset":true}`, hdr); w.Code != 200 {
		t.Fatalf("reset: %d", w.Code)
	}
	if got := get(); got != "default" {
		t.Errorf("after a reset: %q", got)
	}
	m, err := parseManifest([]byte(`{"schema":"watch.panel/1","id":"x","band":"side","optIn":true}`), "x")
	if err != nil || !m.OptIn {
		t.Errorf("optIn: %+v %v", m, err)
	}
	if ms, _ := Manifests(); len(ms) > 0 {
		for _, m := range ms {
			if m.OptIn {
				t.Errorf("%s: no panel of today is opt-in", m.ID)
			}
		}
	}
}

// The new tokens of the visual system (redesign E.2) exist, as aliases over the themed values, so the light theme follows.
func TestRedesignTokens(t *testing.T) {
	b, err := os.ReadFile("web/tokens.css")
	if err != nil {
		t.Fatal(err)
	}
	css := string(b)
	for _, tok := range []string{"--d-ok:", "--d-danger:", "--d-muted:", "--d-state-bg-ok:", "--d-state-bg-warn:", "--d-state-bg-danger:", "--d-state-bg-info:", "--d-hit: 44px", "--d-row: 36px"} {
		if !strings.Contains(css, tok) {
			t.Errorf("tokens.css lacks %s", tok)
		}
	}
	for _, alias := range []string{"--d-ok: var(--d-good)", "--d-danger: var(--d-bad)", "--d-muted: var(--d-text-3)"} {
		if !strings.Contains(css, alias) {
			t.Errorf("tokens.css lacks the alias %s", alias)
		}
	}
	// no hex of its own in the aliases: the light theme redefines the base colours and the aliases resolve on the same element
	if regexp.MustCompile(`--d-(ok|danger|muted|state-bg-[a-z]+):[^;]*#`).MatchString(css) {
		t.Error("a new token holds a colour literal; it must follow the themed value")
	}
}
