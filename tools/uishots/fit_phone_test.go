package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The phone cockpit grows with its content and scrolls as a page (web/utilities.css); the fit rule no longer calls that a defect
// (card uifit1) but keeps judging everything else. Each row is a page that must pass or fail with the words given.
func TestFitPhoneCockpitScrolls(t *testing.T) {
	ph := `{"vw":390,"vh":844,"scrollH":@S,"scrollW":390,"overflow":[],"slots":[{"slot":"band"},{"slot":"stage"},{"slot":"phoneTabs"}],"covered":85,"hole":@Q,"holeAt":[368,104]@X}`
	mk := func(h, hole, x string) string {
		return strings.NewReplacer("@S", h, "@Q", hole, "@X", x).Replace(ph)
	}
	cases := []struct{ name, json, want string }{
		{"a long phone cockpit scrolls by design", mk("1176", "22", `,"outside":[{"el":"panel orders","right":367,"bottom":1112}]`), ""},
		{"a phone page that is NOT taller than the window keeps its dead-space rule", mk("844", "84", ""), "dead space"},
		{"a phone page that scrolls sideways fails", strings.Replace(mk("1176", "22", ""), `"scrollW":390`, `"scrollW":500`, 1), "sideways"},
		{"a phone page with a cut box fails even when it scrolls", mk("1176", "22", `,"overflow":[{"el":"panel orders","dh":40,"dw":0}]`), "overflows its box"},
		{"a phone panel pushed out to the right fails", mk("1176", "22", `,"outside":[{"el":"panel orders","right":500,"bottom":900}]`), "ends outside the window"},
		{"a glyph box 2 px taller than its line is not a cut", mk("844", "8", `,"overflow":[{"el":"char-who em","dh":2,"dw":0}]`), ""},
		{"3 px of height is", mk("844", "8", `,"overflow":[{"el":"char-who em","dh":3,"dw":0}]`), "overflows its box"},
		{"width overflow is never slack", mk("844", "8", `,"overflow":[{"el":"char-who em","dh":2,"dw":4}]`), "overflows its box"},
	}
	for _, c := range cases {
		dir := t.TempDir()
		n, body := fitDump("fit-cockpit-390x844", c.json)
		os.WriteFile(filepath.Join(dir, n+".html"), []byte(body), 0o644)
		lines, ok, err := fitDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if c.want == "" && !ok || c.want != "" && (ok || !strings.Contains(strings.Join(lines, "\n"), c.want)) {
			t.Errorf("%s: ok=%v %v, want %q", c.name, ok, lines, c.want)
		}
	}
	// the layout button is not offered on a phone: not hittable there is no defect; on a wide window it is one, and so is any other target
	for _, c := range []struct{ name, w, h, hits, want string }{
		{"fit-cockpit-menu-390x844", "390", "844", `{"#menu-more":true,"#btn-layout":false}`, ""},
		{"fit-cockpit-menu-390x844", "390", "844", `{"#menu-more":false,"#btn-layout":false}`, "#menu-more is not hittable"},
		{"fit-cockpit-menu-1366x768", "1366", "768", `{"#menu-more":true,"#btn-layout":false}`, "#btn-layout is not hittable"},
	} {
		j := `{"vw":` + c.w + `,"vh":` + c.h + `,"scrollH":` + c.h + `,"scrollW":` + c.w + `,"overflow":[],"slots":[{"slot":"band"}],"covered":96,"hole":8,"holeAt":[0,0],"clicked":"#btn-more","hits":` + c.hits + `}`
		dir := t.TempDir()
		n, body := fitDump(c.name, j)
		os.WriteFile(filepath.Join(dir, n+".html"), []byte(body), 0o644)
		lines, ok, _ := fitDir(dir)
		if c.want == "" && !ok || c.want != "" && (ok || !strings.Contains(strings.Join(lines, "\n"), c.want)) {
			t.Errorf("%s %s: ok=%v %v, want %q", c.name, c.hits, ok, lines, c.want)
		}
	}
	// a wide page that scrolls is still red
	dir := t.TempDir()
	wide := `{"vw":1366,"vh":768,"scrollH":1500,"scrollW":1366,"overflow":[],"slots":[{"slot":"band"},{"slot":"phoneTabs"}],"covered":96,"hole":8,"holeAt":[0,0]}`
	n, body := fitDump("fit-cockpit-1366x768", wide)
	os.WriteFile(filepath.Join(dir, n+".html"), []byte(body), 0o644)
	if lines, ok, _ := fitDir(dir); ok || !strings.Contains(strings.Join(lines, "\n"), "the page scrolls") {
		t.Errorf("a wide page that scrolls must fail: %v", lines)
	}
}
