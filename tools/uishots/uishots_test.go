package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writePNG(t *testing.T, p string, w, h, dots int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < dots; i++ {
		img.Set(i%w, i/w, color.RGBA{255, 255, 255, 255})
	}
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	png.Encode(f, img)
}

func TestDiffDirs(t *testing.T) {
	g, n := t.TempDir(), t.TempDir()
	writePNG(t, filepath.Join(g, "a.png"), 100, 100, 0)
	writePNG(t, filepath.Join(n, "a.png"), 100, 100, 30) // 0.3 percent changed
	writePNG(t, filepath.Join(n, "b.png"), 10, 10, 0)    // no golden: new
	if res, err := diffDirs(g, n, 0.005); err != nil || len(res) != 2 || !strings.HasPrefix(res[1], "NEW") {
		t.Fatalf("%v %v", res, err)
	}
	if _, err := diffDirs(g, n, 0.001); err == nil {
		t.Fatal("0.3 percent passed a 0.1 percent limit")
	}
	writePNG(t, filepath.Join(n, "a.png"), 90, 100, 0)
	if _, err := diffDirs(g, n, 0.5); err == nil {
		t.Fatal("a size change passed")
	}
}

func TestCheckDump(t *testing.T) {
	saved := pagePanels
	defer func() { pagePanels = saved }()
	pagePanels = map[string][]string{"watch-": {"hero-status", "why", "video", "map"}}
	good := `<div class="d-shell"><div data-panel="hero-status" data-state="ready"></div><div data-state="ready" data-panel="why"></div>` +
		`<section class="d-card d-panel" data-panel="video" data-size="m" data-state="ready"></section><section data-panel="map" data-state="ready"></section></div>`
	if bad := checkDump("watch-1280-dark.html", good); len(bad) != 0 {
		t.Fatal(bad)
	}
	bad := checkDump("watch-1280-dark.html", strings.Replace(strings.Replace(good, `"map" data-state="ready"`, `"map" data-state="loading"`, 1), `data-panel="video"`, `data-panel="video2"`, 1))
	if len(bad) != 2 {
		t.Fatalf("want a loading map and a missing video, got %v", bad)
	}
	if bad := checkDump("watch-x.html", strings.Replace(good, `"whyline"`, `"whyline" data-x`, 1)+`<p data-panel="gear" data-state="error"></p>`); len(bad) != 1 {
		t.Fatalf("error state: %v", bad)
	}
}

func TestFixtures(t *testing.T) {
	s := sensesFixture()
	m := s["map"].(map[string]any)
	n := 0
	rle := m["rle"].([]int)
	for i := 1; i < len(rle); i += 2 {
		n += rle[i]
	}
	if n != m["w"].(int)*m["h"].(int) {
		t.Fatalf("rle covers %d cells", n)
	}
	if b := frame(); len(b) < 1000 || b[0] != 0xff || b[1] != 0xd8 {
		t.Fatal("frame is not a JPEG")
	}
}

func TestCheckConsole(t *testing.T) {
	log := `[1:2:INFO:CONSOLE(12)] "watch layout: no panel for risk (ignored)", source: http://x/watch/shell.js (12)
[1:2:INFO:CONSOLE:7] "Uncaught (in promise) ReferenceError: x is not defined", source: http://x/watch/shell.js (7)
[1:2:INFO:CONSOLE(40)] "Uncaught TypeError: Cannot read properties of null (reading 'x')", source: http://x/watch/panels/gear/panel.js (40)
[1:2:INFO:CONSOLE(0)] "Failed to load resource: the server responded with a status of 404 (Not Found)", source: http://x/api/events (0)`
	bad := checkConsole("watch-1440-dark.html", log)
	if len(bad) != 2 || !strings.Contains(bad[0], "ReferenceError") || !strings.Contains(bad[1], "Uncaught TypeError") {
		t.Fatalf("got %v", bad)
	}
}

func TestCheckVideoFit(t *testing.T) {
	page := func(rows string) string {
		return `<div class="d-shell"><section class="d-card d-panel" data-panel="video" data-size="m" data-state="ready" data-fit="1"` + rows + `></section></div>`
	}
	saved := pagePanels
	defer func() { pagePanels = saved }()
	pagePanels = map[string][]string{}
	for rows, bad := range map[string]bool{` data-rows="13"`: false, ` data-rows="12"`: false, ` data-rows="17"`: true, ` data-rows="9"`: true, ``: true} {
		got := checkDump("watch-1920-dark.html", page(rows))
		if (len(got) > 0) != bad {
			t.Errorf("rows %q: %v", rows, got)
		}
	}
	if got := checkDump("watch-390-dark.html", page(``)); len(got) != 0 {
		t.Errorf("a flow breakpoint has no fitted rows: %v", got)
	}
}

func fitDump(name, json string) (string, string) {
	return name, `<html><body><pre id="fit-result">` + strings.ReplaceAll(json, "&", "&amp;") + `</pre></body></html>`
}

// The fit check judges the numbers fit.js measures: a cockpit that fills its viewport passes; a page that scrolls, a
// panel that overflows its box and dead space each fail; the deliberately overflowing fixture must fail, and a pass there
// is itself reported.
// The Vault is one scrolling document since card vaultdesktop (2026-11-21): a page taller than the window passes, a page wider fails.
func TestFitVaultScrollsButNotSideways(t *testing.T) {
	vault := `"vault":{"meter":true,"cells":120}`
	base := `{"vw":1440,"vh":900,"scrollH":4000,"scrollW":1440,"overflow":[],"slots":[],"covered":0,"hole":0,"holeAt":[0,0],` + vault + `}`
	for _, c := range []struct{ json, want string }{
		{base, ""},
		{strings.Replace(base, `"scrollW":1440`, `"scrollW":1500`, 1), "sideways"},
	} {
		dir := t.TempDir()
		n, body := fitDump("fit-vault-pop-1440x900", c.json)
		if err := os.WriteFile(filepath.Join(dir, n+".html"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		lines, ok, err := fitDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if c.want == "" && !ok || c.want != "" && (ok || !strings.Contains(lines[0], c.want)) {
			t.Errorf("%q: ok=%v %v, want %q", c.json, ok, lines, c.want)
		}
	}
}

func TestFitJudgeAndDir(t *testing.T) {
	okJSON := `{"vw":1440,"vh":900,"scrollH":900,"scrollW":1440,"overflow":[],"slots":[{"slot":"band"},{"slot":"stage"}],"covered":96,"hole":8,"holeAt":[0,44]}`
	cases := []struct{ name, json, want string }{
		{"fit-cockpit-1440x900", okJSON, ""},
		{"fit-cockpit-1440x900", strings.Replace(okJSON, `"scrollH":900`, `"scrollH":1700`, 1), "the page scrolls"},
		{"fit-cockpit-1440x900", strings.Replace(okJSON, `"scrollW":1440`, `"scrollW":1500`, 1), "sideways"},
		{"fit-cockpit-1440x900", strings.Replace(okJSON, `"overflow":[]`, `"overflow":[{"el":"panel orders","dh":40,"dw":0}]`, 1), "panel orders overflows"},
		{"fit-cockpit-1440x900", strings.Replace(okJSON, `"hole":8`, `"hole":120`, 1), "dead space"},
		{"fit-cockpit-1440x900", strings.Replace(okJSON, `"slots":[{"slot":"band"},{"slot":"stage"}]`, `"slots":[]`, 1), "no [data-cockpit-slot]"},
		{"fit-cockpit-1920x1080", okJSON, "the viewport is 1440x900, not 1920x1080"},
		{"fit-cockpit-1440x900", `{"error":"boom"}`, "the measurement failed"},
		{"fit-cockpit-1440x900", ``, "no measurement"},
		{"fit-cockpit-1440x900", strings.Replace(okJSON, `"overflow":[]`, `"outside":[{"el":"slot strip","right":1440,"bottom":1100}]`, 1), "ends outside the window"},
		{"fit-cockpit-1440x900", strings.Replace(okJSON, `"overflow":[]`, `"empty":["rail"]`, 1), "slot rail holds no visible panel"},
		{"fit-cockpit-1440x900", strings.Replace(okJSON, `"overflow":[]`, `"unready":["orders:loading"]`, 1), "is not ready"},
		{"fit-cockpit-menu-1440x900", strings.Replace(okJSON, `"overflow":[]`, `"clicked":"#btn-more","hits":{"#menu-more":false,"#btn-layout":false}`, 1), "is not hittable"},
	}
	for _, c := range cases {
		dir := t.TempDir()
		n, body := fitDump(c.name, c.json)
		if err := os.WriteFile(filepath.Join(dir, n+".html"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		lines, ok, err := fitDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if c.want == "" && (!ok || !strings.HasPrefix(lines[0], "ok")) || c.want != "" && (ok || !strings.Contains(lines[0], c.want)) {
			t.Errorf("%s: ok=%v %v, want %q", c.name, ok, lines, c.want)
		}
	}
	// the overflow fixture: failing is the expected result, passing is the failure
	dir := t.TempDir()
	bad := strings.Replace(okJSON, `"overflow":[]`, `"overflow":[{"el":"panel fx-overflow","dh":800,"dw":0}]`, 1)
	for n, j := range map[string]string{"fit-overflow-fixture-1440x900": bad, "fit-overflow-classic-1440x900": strings.Replace(okJSON, `"scrollH":900`, `"scrollH":2600`, 1)} {
		n, body := fitDump(n, j)
		os.WriteFile(filepath.Join(dir, n+".html"), []byte(body), 0o644)
	}
	if lines, ok, _ := fitDir(dir); !ok || len(lines) != 2 || !strings.Contains(lines[0], "failed as it must") {
		t.Errorf("a failing overflow fixture is the expected result: %v", lines)
	}
	n, body := fitDump("fit-overflow-fixture-1440x900", okJSON)
	os.WriteFile(filepath.Join(dir, n+".html"), []byte(body), 0o644)
	if lines, ok, _ := fitDir(dir); ok || !strings.Contains(strings.Join(lines, "\n"), "must fail with") {
		t.Errorf("an overflow fixture that passes must fail the run: %v", lines)
	}
	// R2: a must-fail page counts only for the reason it was built for; a broken measurement, a viewport mismatch and a
	// different verdict fail the run
	for _, c := range []struct{ name, json, want string }{
		{"fit-overflow-fixture-1440x900", `{"error":"boom"}`, "the measurement failed"},
		{"fit-overflow-fixture-1440x900", ``, "no measurement"},
		{"fit-overflow-fixture-1440x900", strings.Replace(bad, `"vw":1440`, `"vw":1000`, 1), "the viewport is"},
		{"fit-overflow-fixture-1440x900", strings.Replace(okJSON, `"scrollH":900`, `"scrollH":1700`, 1), "must fail with"},
		{"fit-blank-rail-1440x900", bad, "must fail with"},
		{"fit-blank-rail-1440x900", strings.Replace(okJSON, `"overflow":[]`, `"empty":["rail"]`, 1), "failed as it must"},
		{"fit-tall-frame-1440x900", strings.Replace(okJSON, `"overflow":[]`, `"outside":[{"el":"slot strip","right":1,"bottom":2000}]`, 1), "failed as it must"},
	} {
		d := t.TempDir()
		n, body := fitDump(c.name, c.json)
		os.WriteFile(filepath.Join(d, n+".html"), []byte(body), 0o644)
		lines, ok, _ := fitDir(d)
		isOK := strings.Contains(c.want, "failed as it must")
		if ok != isOK || !strings.Contains(strings.Join(lines, "\n"), c.want) {
			t.Errorf("%s %q: ok=%v %v, want %q", c.name, c.json[:min(len(c.json), 30)], ok, lines, c.want)
		}
	}
	if _, _, err := fitDir(t.TempDir()); err == nil {
		t.Error("an empty directory is not a pass")
	}
}

// The navigation verdicts (card uinav2): each way the shared navigation can be wrong is named, and a good one is silent.
func TestFitNavJudge(t *testing.T) {
	good := func() *fitNav {
		return &fitNav{Labels: []string{"Live", "Gear", "Skills", "Vault", "More"}, Current: 1, Inside: true, Bar: true}
	}
	if bad := good().judge(390); len(bad) != 0 {
		t.Fatalf("a good bar is judged: %v", bad)
	}
	wide := good()
	wide.Bar = false
	if bad := wide.judge(1366); len(bad) != 0 {
		t.Fatalf("a good band is judged: %v", bad)
	}
	for name, c := range map[string]struct {
		mut  func(*fitNav)
		vw   int
		want string
	}{
		"missing":   {func(n *fitNav) { n.Missing = true }, 390, "no navigation"},
		"labels":    {func(n *fitNav) { n.Labels = []string{"Live", "Gear"} }, 390, "want Live,Gear,Skills,Vault,More"},
		"two":       {func(n *fitNav) { n.Current = 2 }, 390, "marks 2 destinations"},
		"none":      {func(n *fitNav) { n.Current = 0 }, 390, "marks 0 destinations"},
		"bar wide":  {func(n *fitNav) {}, 1366, "bar=true"},
		"band thin": {func(n *fitNav) { n.Bar = false }, 390, "bar=false"},
		"outside":   {func(n *fitNav) { n.Inside = false }, 390, "not inside the window"},
		"tiny":      {func(n *fitNav) { n.Tiny = []string{"More 75x20"} }, 390, "too small"},
		"covers":    {func(n *fitNav) { n.Covers = []string{"main ends at 844"} }, 390, "covers content"},
	} {
		n := good()
		c.mut(n)
		found := false
		for _, b := range n.judge(c.vw) {
			found = found || strings.Contains(b, c.want)
		}
		if !found {
			t.Errorf("%s: no verdict containing %q in %v", name, c.want, n.judge(c.vw))
		}
	}
	var none *fitNav
	if bad := none.judge(390); bad != nil {
		t.Errorf("a page without it is not judged here: %v", bad)
	}
	if bad := (fitResult{Flow: true, VW: 390}).judge(); len(bad) != 1 || !strings.Contains(bad[0], "no navigation") {
		t.Errorf("a navigation page with none: %v", bad)
	}
}

// The Build tab verdicts (card buildcards): one column on a phone, a grid above, no scroll box, no sideways page, 44 px buttons,
// the text equivalent of every star row.
func TestFitBuildsJudge(t *testing.T) {
	good := &fitBuilds{Cards: 5, Cols: 1}
	if bad := good.judge(390, 390); len(bad) != 0 {
		t.Errorf("a phone list of one column is a fit: %v", bad)
	}
	grid := &fitBuilds{Cards: 5, Cols: 3}
	if bad := grid.judge(1366, 1366); len(bad) != 0 {
		t.Errorf("a wide grid is a fit: %v", bad)
	}
	for name, c := range map[string]struct {
		b      *fitBuilds
		vw, sw int
		want   string
	}{
		"grid on a phone": {&fitBuilds{Cards: 5, Cols: 2}, 390, 390, "want 1 on a phone"},
		"one wide column": {&fitBuilds{Cards: 5, Cols: 1}, 1366, 1366, "want a grid"},
		"sideways":        {&fitBuilds{Cards: 5, Cols: 1}, 390, 520, "scrolls sideways"},
		"scroll box":      {&fitBuilds{Cards: 5, Cols: 1, Boxes: []string{"ul.x"}}, 390, 390, "scroll box"},
		"small":           {&fitBuilds{Cards: 5, Cols: 1, Small: []string{"bc-details 30"}}, 390, 390, "too small"},
		"no text":         {&fitBuilds{Cards: 5, Cols: 1, Notext: []string{"(none)"}}, 390, 390, "text equivalent"},
		"no cards":        {&fitBuilds{}, 390, 390, "no card"},
	} {
		found := false
		for _, b := range c.b.judge(c.vw, c.sw) {
			found = found || strings.Contains(b, c.want)
		}
		if !found {
			t.Errorf("%s: no verdict containing %q", name, c.want)
		}
	}
	// an open detail takes the whole row: the column count is not judged then
	if bad := (&fitBuilds{Cards: 5, Open: 1, Cols: 1}).judge(1366, 1366); len(bad) != 0 {
		t.Errorf("an open detail on a wide window: %v", bad)
	}
	var none *fitBuilds
	if none.judge(390, 390) != nil {
		t.Error("a page without the list is not judged here")
	}
}
