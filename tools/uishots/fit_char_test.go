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
	if !strings.Contains(sh, "character-phonepack") || !strings.Contains(sh, "#belt,#cv-sel") {
		t.Error("the phone's storage tab has no fit page that must hit the belt and the selected box (the errands have their own phone page)")
	}
	for _, p := range []string{"character-phone ", "character-phone-sheet ", "character-phone-errands ", "character-phone-menu ", "character-phonehero ", "phonedefect-label ", "phonedefect-acts ", "phonedefect-cutchips ", "phonedefect-lock ", "phonedefect-hover ", "phonedefect-narrow "} {
		if !strings.Contains(sh, "\n"+p) {
			t.Errorf("tools/uishots.sh has no fit page %q (the phone sizes of card uiphone)", p)
		}
	}
	for _, sz := range []string{"412x780", "393x700", "360x640", "412x890"} {
		if !strings.Contains(sh, sz) {
			t.Errorf("tools/uishots.sh does not run the phone size %s", sz)
		}
	}
}

// The phone's storage tab (card uiphone): each defect of the 2026-11-15 screenshot is a verdict. The numbers are what fit.js measures in
// phoneChecks; the live pages phonedefect-* put the same defects back and must fail with these words.
func TestFitPhoneRules(t *testing.T) {
	ok := `{"vw":412,"vh":780,"scrollH":780,"scrollW":412,"overflow":[],"slots":[{"slot":"band"}],"covered":96,"hole":8,"holeAt":[0,44],"cv":{"zero":[],"cut":[],"tiles":[],"inner":{"size":60,"at":[0,0],"panel":"char-store"},"phone":{%s}}}`
	cases := []struct{ name, phone, want string }{
		{"a clean phone page", "", ""},
		{"a name painted over the art (the screenshot's 'Chipp ed')", `"artLabel":["button.item \"Chipped\" font 8.5px"]`, "label over art"},
		{"the Sell chip cut at the right edge", `"bar":["button#cv-sell \"Sell\" ends outside the bar [380..440 of 0..412]"]`, "toolbar cut"},
		{"the selected card's buttons clipped to a sliver", `"acts":["the button row is cut [60/10]"]`, "action buttons"},
		{"a button under 44 px", `"acts":["button \"Equip\" is 22 px high (least 44)"]`, "action buttons"},
		{"the lock line ends in an ellipsis", `"lock":["the lock line is cut [320/212 17/17]"]`, "lock line"},
		{"the canvas scrolls sideways", `"hover":["the canvas scrolls sideways [470/396]"]`, "horizontal overflow"},
		{"cells of 20 px", `"grids":["#pack cells are 20 px (least 32)"]`, "grid"},
		{"the stash at half the width", `"grids":["#stash is 190 px wide in a box of 380"]`, "grid"},
	}
	for _, c := range cases {
		dir := t.TempDir()
		n, body := fitDir0(strings.Replace(ok, "%s", c.phone, 1))
		if err := os.WriteFile(filepath.Join(dir, n+".html"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		lines, good, err := fitDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if c.want == "" && !good || c.want != "" && (good || !strings.Contains(strings.Join(lines, "\n"), c.want)) {
			t.Errorf("%s: ok=%v %v, want %q", c.name, good, lines, c.want)
		}
	}
	sh, _ := os.ReadFile("../uishots.sh")
	for name := range defectCSS {
		if strings.HasPrefix(name, "ov-") {
			continue
		}
		if rest, ok := strings.CutPrefix(name, "chars-"); ok { // the Characters page: charsdefect-NAME (fixtures_chars.go)
			if !strings.Contains(string(sh), "\ncharsdefect-"+rest+" ") {
				t.Errorf("no fit page charsdefect-%s", rest)
			}
			if _, ok := mustFail["fit-charsdefect-"+rest+"-"]; !ok {
				t.Errorf("mustFail has no entry for charsdefect-%s", rest)
			}
			continue
		}
		if rest, ok := strings.CutPrefix(name, "mf-"); ok { // the Orders More list: moredefect-NAME
			if !strings.Contains(string(sh), "\nmoredefect-"+rest+" ") {
				t.Errorf("no fit page moredefect-%s", rest)
			}
			if _, ok := mustFail["fit-moredefect-"+rest+"-"]; !ok {
				t.Errorf("mustFail has no entry for moredefect-%s", rest)
			}
			continue
		}
		if rest, ok := strings.CutPrefix(name, "mn-"); ok { // the "..." menu: menudefect-NAME (fixtures_menu.go); mn-tall is the healthy tall menu
			if rest == "tall" {
				if !strings.Contains(string(sh), "\nmenu-tall ") {
					t.Error("no fit page menu-tall")
				}
				continue
			}
			if !strings.Contains(string(sh), "\nmenudefect-"+rest+" ") {
				t.Errorf("no fit page menudefect-%s", rest)
			}
			if _, ok := mustFail["fit-menudefect-"+rest+"-"]; !ok {
				t.Errorf("mustFail has no entry for menudefect-%s", rest)
			}
			continue
		}
		if rest, ok := strings.CutPrefix(name, "q-"); ok { // the quest board: questdefect-NAME (fixtures_quests.go)
			if !strings.Contains(string(sh), "\nquestdefect-"+rest+" ") {
				t.Errorf("no fit page questdefect-%s", rest)
			}
			if _, ok := mustFail["fit-questdefect-"+rest+"-"]; !ok {
				t.Errorf("mustFail has no entry for questdefect-%s", rest)
			}
			continue
		}
		if rest, ok := strings.CutPrefix(name, "vault-"); ok { // the Vault page: vaultdefect-NAME (fixtures_vault.go)
			if !strings.Contains(string(sh), "\nvaultdefect-"+rest+" ") {
				t.Errorf("no fit page vaultdefect-%s", rest)
			}
			if _, ok := mustFail["fit-vaultdefect-"+rest+"-"]; !ok {
				t.Errorf("mustFail has no entry for vaultdefect-%s", rest)
			}
			continue
		}
		if !strings.Contains(string(sh), "\nphonedefect-"+name+" ") {
			t.Errorf("no fit page phonedefect-%s", name)
		}
		found := false
		for k := range mustFail {
			found = found || k == "fit-phonedefect-"+name+"-"
		}
		if !found {
			t.Errorf("mustFail has no entry for phonedefect-%s", name)
		}
	}
}

func fitDir0(j string) (string, string) { return fitDump("fit-character-phone-412x780", j) }

// The phone fixture is a real-looking hero (card uiphone): its grids hold what they say (no item leaves the pack or the stash, none sits on
// another), the belt carries slot numbers as x, and every cell has art.
func TestFitPhoneFixture(t *testing.T) {
	v := variantView("phone").(map[string]any)
	grid := func(name string, cols, rows int) {
		used := map[[2]int]bool{}
		for _, x := range v[name].([]any) {
			it := x.(map[string]any)
			if (it["art"] == nil || it["art"] == "") && it["code"] != "zzz" { // zzz: the deliberate no-icon case
				t.Errorf("%s: %v has no art", name, it["name"])
			}
			x0, y0, w, h := it["x"].(int), it["y"].(int), it["w"].(int), it["h"].(int)
			if x0+w > cols || y0+h > rows {
				t.Errorf("%s: %v leaves the %dx%d grid", name, it["name"], cols, rows)
			}
			for dx := 0; dx < w; dx++ {
				for dy := 0; dy < h; dy++ {
					if used[[2]int{x0 + dx, y0 + dy}] {
						t.Errorf("%s: %v overlaps another item at %d,%d", name, it["name"], x0+dx, y0+dy)
					}
					used[[2]int{x0 + dx, y0 + dy}] = true
				}
			}
		}
	}
	grid("backpack", 10, 4)
	grid("stash", 6, 8)
	var big23, big24 bool
	for _, x := range v["backpack"].([]any) {
		it := x.(map[string]any)
		big23 = big23 || it["w"] == 2 && it["h"] == 3
		big24 = big24 || it["w"] == 2 && it["h"] == 4
	}
	if !big23 || !big24 {
		t.Error("the pack holds no 2x3 and 2x4 item")
	}
	if b := v["belt"].([]any); len(b) < 5 || b[len(b)-1].(map[string]any)["x"].(int) < 4 {
		t.Errorf("the belt has more than four potions with slot numbers as x: %v", b)
	}
	if png, err := (fixtureArt{}).PNG("item-8ss"); err != nil || len(png) < 50 {
		t.Errorf("fixture art: %v", err)
	}
}

// Every overlay defect has a fit page and a mustFail verdict, and the defect fires only for its own name.
func TestOverlayDefectsHavePages(t *testing.T) {
	sh, _ := os.ReadFile("../uishots.sh")
	names := map[string]bool{}
	for n := range defectCSS {
		if strings.HasPrefix(n, "ov-") {
			names[n] = true
		}
	}
	for n := range overlayJS {
		names[n] = true
	}
	if len(names) != 4 {
		t.Errorf("want 4 overlay defects, have %d", len(names))
	}
	for n := range names {
		short := strings.TrimPrefix(n, "ov-")
		if !strings.Contains(string(sh), "\noverlaydefect-"+short+" ") {
			t.Errorf("no fit page overlaydefect-%s", short)
		}
		if _, ok := mustFail["fit-overlaydefect-"+short+"-"]; !ok {
			t.Errorf("mustFail has no entry for overlaydefect-%s", short)
		}
	}
}

// The script defects of the Vault page have a fit page and a verdict too.
func TestVaultJSDefectsHavePages(t *testing.T) {
	sh, _ := os.ReadFile("../uishots.sh")
	if len(vaultJS) != 3 {
		t.Errorf("want 3 script defects, have %d", len(vaultJS))
	}
	for n := range vaultJS {
		rest := strings.TrimPrefix(n, "vault-")
		if !strings.Contains(string(sh), "\nvaultdefect-"+rest+" ") {
			t.Errorf("no fit page vaultdefect-%s", rest)
		}
		if _, ok := mustFail["fit-vaultdefect-"+rest+"-"]; !ok {
			t.Errorf("mustFail has no entry for vaultdefect-%s", rest)
		}
	}
}
