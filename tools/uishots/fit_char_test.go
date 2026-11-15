package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The character screen's fit rules (card uic4, review c2c3 R14): each defect the first fit check let through is a verdict now. The
// numbers are what fit.js measures; a page that shows the defect again fails with the words below.
func TestFitCharacterRules(t *testing.T) {
	ok := `{"vw":1440,"vh":900,"scrollH":900,"scrollW":1440,"overflow":[],"slots":[{"slot":"band"},{"slot":"who"}],"covered":96,"hole":8,"holeAt":[0,44]%s}`
	cv := func(zero, cut, tiles string, size int) string {
		return `,"cv":{"zero":[` + zero + `],"cut":[` + cut + `],"tiles":[` + tiles + `],"inner":{"size":` + itoa(size) + `,"at":[300,200],"panel":"char-who"}}`
	}
	cases := []struct{ name, extra, want string }{
		{"a clean character page", cv("", "", "", 60), ""},
		{"no mercenary: 330 px empty beside the hero (R3)", cv("", "", "", 330), "dead space inside char-who"},
		{"the selected box with its empty band (R5)", cv("", "", "", 145), "dead space inside char-who"},
		{"the phone's storage side column, 0 px wide (R1)", cv(`"char-store div.cv-store"`, "", "", 60), "content with no width"},
		{"the phone's hero sheet: 'L...' (R2)", cv("", `"char-who label \"L...\" [40/20]"`, "", 60), "a values tile is cut"},
		{"the mercenary's stretched tile (R4)", cv("", "", `"char-who div 148x157 holds 31 px"`, 60), "a values tile is stretched"},
		{"the pending outline is gone after a repaint (R6)", `,"expect":{"[data-key=item-11][data-pending]":false}`, "expected [data-key=item-11][data-pending] on the page"},
		{"the phone's storage tab lost its belt (R1)", `,"clicked":"#cv-tabbar [data-tab=pack]","hits":{"#belt":false}`, "#belt is not hittable"},
		{"a long slot name clipped mid-line (R10)", `,"overflow":[{"el":"char-who span.nm","dh":0,"dw":55}]`, "overflows its box"},
		{"the dead mercenary's header overflows (R13)", `,"overflow":[{"el":"char-who div.cv-mhead","dh":0,"dw":66}]`, "overflows its box"},
		{"the band's headline cut (R8)", `,"overflow":[{"el":"band div.d-panel","dh":14,"dw":0}]`, "overflows its box"},
	}
	for _, c := range cases {
		j := strings.Replace(ok, "%s", c.extra, 1)
		if strings.Contains(c.extra, `"overflow"`) {
			j = strings.Replace(strings.Replace(ok, `"overflow":[],`, "", 1), "%s", c.extra, 1)
		}
		dir := t.TempDir()
		n, body := fitDump("fit-character-1440x900", j)
		if err := os.WriteFile(filepath.Join(dir, n+".html"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		lines, good, err := fitDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if c.want == "" && (!good || !strings.HasPrefix(lines[0], "ok")) || c.want != "" && (good || !strings.Contains(strings.Join(lines, "\n"), c.want)) {
			t.Errorf("%s: ok=%v %v, want %q", c.name, good, lines, c.want)
		}
	}
}

// Every fixture variant the fit pages name exists, and each one differs from the plain view in the way its name says.
func TestFitFixtureVariants(t *testing.T) {
	plain := variantView("")
	if plain == nil {
		t.Fatal("no plain view")
	}
	if _, ok := variantView("nomerc").(map[string]any)["merc"].(map[string]any)["owned"]; !ok {
		t.Error("nomerc: the merc block says nobody is hired")
	}
	if _, has := variantView("absentmerc").(map[string]any)["merc"]; has {
		t.Error("absentmerc: no merc block")
	}
	m := variantView("deadmerc").(map[string]any)["merc"].(map[string]any)
	if m["state"] != "dead" || m["revive"] == nil {
		t.Errorf("deadmerc: dead with a revive verb: %v", m)
	}
	eq := variantView("longnames").(map[string]any)["equipment"].(map[string]any)
	if name := eq["belt"].(map[string]any)["name"].(string); len(name) < 30 {
		t.Errorf("longnames: the belt name is not long: %q", name)
	}
	if q := variantView("pending").(map[string]any)["queue"].([]any); len(q) < 3 || q[0].(map[string]any)["wait"] != "town" {
		t.Errorf("pending: requests, one waiting for town: %v", q)
	}
	for name := range fxExpect {
		if fxVariants[name] == nil {
			t.Errorf("fxExpect names %q, which is no variant", name)
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for ; n > 0; n /= 10 {
		s = string(rune('0'+n%10)) + s
	}
	return s
}

// uishots.sh lists one fit page per variant, and the pending page expects exactly what fxExpect says it must show.
func TestFitPagesListVariants(t *testing.T) {
	b, err := os.ReadFile("../uishots.sh")
	if err != nil {
		t.Fatal(err)
	}
	sh := string(b)
	for name := range fxVariants {
		if !strings.Contains(sh, "fx="+name) {
			t.Errorf("tools/uishots.sh has no fit page for the variant %q", name)
		}
	}
	for name, sels := range fxExpect {
		for _, sel := range sels {
			if !strings.Contains(sh, "fx="+name) || !strings.Contains(sh, sel) {
				t.Errorf("the %s fit page does not expect %s", name, sel)
			}
		}
	}
	if !strings.Contains(sh, "character-phonepack") || !strings.Contains(sh, "#belt,#cv-sel,#errands") {
		t.Error("the phone's storage tab has no fit page that must hit the belt, the selected box and the errands")
	}
}
