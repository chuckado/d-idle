package questspec

import (
	"embed"
	"fmt"
	"strings"
	"testing"

	"d2lod-bot/internal/quest"
)

//go:embed testdata
var fixtures embed.FS

func fixChain(t *testing.T, name string) *Chain {
	t.Helper()
	b, err := fixtures.ReadFile("testdata/" + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := Load(b)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func fixCaps(t *testing.T, name string) *Caps {
	t.Helper()
	b, err := fixtures.ReadFile("testdata/" + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := LoadCaps(b)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func chainByID(t *testing.T, id string) *Chain {
	t.Helper()
	cs, _, _, _ := shipped(t)
	for _, c := range cs {
		if c.ID == id {
			return c
		}
	}
	t.Fatalf("no chain %s", id)
	return nil
}

func TestStrictLoad(t *testing.T) {
	for name, in := range map[string]string{
		"unknown key": `{"schema":"questspec/2","id":"x","steps":[],"bogus":1}`,
		"bad schema":  `{"schema":"questspec/1","id":"x"}`,
		"trailing":    `{"schema":"questspec/2","id":"x"} {}`,
	} {
		if _, err := Load([]byte(in)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if _, err := LoadFile([]byte(`{"schema":"questspec/2","act":1,"chains":[{"id":"a"},{"id":"a"}]}`)); err == nil {
		t.Error("duplicate chain id accepted")
	}
}

// The old Cain data (tree by display name, five stones read by lit state, scroll by name, unleased retries) on the
// backend as it was on 2026-11-17 reproduces the documented root causes RC1 to RC3.
func TestOldCainReproducesRootCauses(t *testing.T) {
	fs := Check(fixChain(t, "cain_old"), fixCaps(t, "old"))
	got := map[string]string{}
	rules := map[int]bool{}
	for _, f := range fs {
		got[f.RC] += fmt.Sprintf(" r%d/%s", f.Rule, f.Step)
		rules[f.Rule] = true
		t.Log(f)
	}
	for rc, w := range map[string]string{"RC1": "r4/find_tree", "RC2": "r1/stone_1", "RC3": "r5/*"} {
		if !strings.Contains(got[rc], w) {
			t.Errorf("%s: want %s, got %q", rc, w, got[rc])
		}
	}
	if !strings.Contains(got["RC1"], "r3/find_tree") {
		t.Errorf("RC1 hidden row: want r3/find_tree, got %q", got["RC1"])
	}
	for _, r := range []int{1, 3, 4, 5, 7, 8, 10} {
		if !rules[r] {
			t.Errorf("rule %d never fired on the old data", r)
		}
	}
}

// The tree named by display name and the stones read by a lit state that needs a lab verb are the two known bugs; the
// checker must name each, with the verb for the second.
func TestKnownBugsAreNamedPrecisely(t *testing.T) {
	var name7, verb8, lit1 bool
	shippedCaps := fixCaps(t, "old")
	for _, f := range Check(fixChain(t, "cain_old"), shippedCaps) {
		switch {
		case f.Rule == 7 && f.Step == "find_tree" && strings.Contains(f.Msg, "no class id"):
			name7 = true
		case f.Rule == 8 && f.Step == "stone_1" && strings.Contains(f.Msg, "debug objects"):
			verb8 = true
		case f.Rule == 1 && f.Step == "stone_3" && strings.Contains(f.Msg, "stones_lit(3)"):
			lit1 = true
		}
	}
	if !name7 || !verb8 || !lit1 {
		t.Errorf("tree-by-name %v, lab verb %v, unreadable lit state %v", name7, verb8, lit1)
	}
}

// On the lab verbs a stone's lit state reads (V probe L1), so the verb finding is about the shipped build only.
func TestLitStateNeedsLabVerb(t *testing.T) {
	c := fixChain(t, "cain_old")
	lab := fixCaps(t, "fixed") // lab verbs, knowledge, code matching
	for _, f := range Check(c, lab) {
		if f.Rule == 8 || (f.Rule == 1 && strings.Contains(f.Msg, "stones_lit")) {
			t.Errorf("lab backend: %s", f)
		}
	}
}

// The converted chain is clean on a backend with the fixes (room rows, a progress rule, lab verbs).
func TestShippedCainCleanOnFixedBackend(t *testing.T) {
	if fs := Check(chainByID(t, "search_for_cain"), fixCaps(t, "fixed")); len(fs) != 0 {
		t.Fatalf("findings on the fixed backend: %v", fs)
	}
}

// Resumption from facts: after a restart at any point the same non-movement steps run, none twice and none skipped.
func TestSimResumesAtEveryStep(t *testing.T) {
	cs, _, _, _ := shipped(t)
	for _, c := range cs {
		ref, err := Run(c, NewWorld(c, "rogue_encampment"), -1, "rogue_encampment")
		if err != nil || len(ref) == 0 {
			t.Fatalf("%s reference run: %v %v", c.ID, ref, err)
		}
		for k := 0; k <= len(c.Steps); k++ {
			got, err := Run(c, NewWorld(c, "rogue_encampment"), k, "rogue_encampment")
			if err != nil {
				t.Fatalf("%s restart after %d steps: %v", c.ID, k, err)
			}
			if strings.Join(got, ",") != strings.Join(ref, ",") {
				t.Errorf("%s restart after %d: ran %v, want %v", c.ID, k, got, ref)
			}
		}
	}
}

func TestSimCatchesWorldDisagreement(t *testing.T) {
	c := chainByID(t, "search_for_cain")
	w := NewWorld(c, "rogue_encampment")
	w.Effect["find_tree"] = func(*World) {} // the world never shows the tree
	if _, err := Run(c, w, -1, "rogue_encampment"); err == nil || !strings.Contains(err.Error(), "not observed") {
		t.Fatalf("want a 'not observed' error, got %v", err)
	}
}

// The classes the data names are the ones the bot uses (internal/quest), so a renumbered constant cannot drift.
func TestClassesMatchBotConstants(t *testing.T) {
	want := map[string]int{"akara": quest.Akara, "kashya": quest.Kashya, "charsi": quest.Charsi, "warriv": quest.Warriv,
		"tree_of_inifuss": quest.ObjTreeInifuss, "cairn_stones": quest.ObjCairnStone, "cains_gibbet": quest.ObjCainGibbet}
	cs, _, _, _ := shipped(t)
	n := 0
	for _, c := range cs {
		for _, s := range c.Steps {
			if w, ok := want[s.Target.Name]; ok {
				n++
				if s.Target.Class != w {
					t.Errorf("%s/%s: class %d, internal/quest has %d", c.ID, s.ID, s.Target.Class, w)
				}
			}
		}
	}
	if n == 0 {
		t.Fatal("no class checked")
	}
}

func TestBoardDriftIsReported(t *testing.T) {
	cs, _, _, board := shipped(t)
	var cain *Chain
	for _, c := range cs {
		if c.ID == "search_for_cain" {
			cp := *c
			cp.Steps = append([]Step(nil), c.Steps...)
			cain = &cp
		}
	}
	cain.Steps[1].Nodes = []string{"treee"}                // a renamed node
	cain.Steps[0].Board = []string{"travel_dark_woods"}    // a renamed board step
	cain.Steps[len(cain.Steps)-1].Chunk = "report_missing" // an unknown chunk
	fs := CheckBoard(cain, board)
	var msg []string
	for _, f := range fs {
		msg = append(msg, f.Msg)
	}
	all := strings.Join(msg, "\n")
	for _, w := range []string{`node "treee"`, `board step "travel_dark_woods"`, `chunk "report_missing"`, "realised by no chain step"} {
		if !strings.Contains(all, w) {
			t.Errorf("missing finding %q in:\n%s", w, all)
		}
	}
}

func TestAllowListStale(t *testing.T) {
	a, err := LoadAllow([]byte(`{"schema":"questallow/1","allowed":[{"chain":"c","caps":"b","step":"s","rule":4,"owner":"o","why":"w"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	used := map[int]bool{}
	if open := a.Filter("c", "b", []Finding{{Rule: 4, Step: "s"}, {Rule: 4, Step: "t"}}, used); len(open) != 1 || open[0].Step != "t" || !used[0] {
		t.Fatalf("filter: %v %v", open, used)
	}
	if _, err := LoadAllow([]byte(`{"schema":"questallow/1","allowed":[{"chain":"c"}]}`)); err == nil {
		t.Error("an entry without owner and reason was accepted")
	}
}

func load(t *testing.T, js string) *Chain {
	t.Helper()
	c, err := Load([]byte(js))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

const head = `{"schema":"questspec/2","id":"bad","initial":["a(x)"],"areas":{"town":{"size":"small"},"wild":{"size":"large"}},"steps":[`
const tail = `]}`

func step(id, pre, post, extra string) string {
	return `{"id":"` + id + `","act":"wait","target":{"kind":"area","name":"town"},"pre":[` + pre + `],"post":[` + post + `],"discover":"preset_always","budget_s":10,"on_fail":"retry","tries":2` + extra + `}`
}

func TestBrokenChains(t *testing.T) {
	cp := fixCaps(t, "fixed")
	cases := []struct {
		name string
		js   string
		rule int
	}{
		{"cycle", head + step("p", `"b(x)"`, `"a(y)"`, "") + "," + step("q", `"a(y)"`, `"b(x)"`, "") + tail, 2},
		{"unobservable", head + step("p", `"a(x)"`, ``, "") + tail, 3},
		{"post unreadable", head + step("p", `"a(x)"`, `"mystery(x)"`, "") + tail, 1},
		{"big area no explore", head + `{"id":"f","act":"find_object","target":{"kind":"object","name":"o","class":3},"pre":["in_area(wild)"],"post":["object_known(o)"],"discover":"preset_when_near","budget_s":10,"on_fail":"park"}` + tail, 4},
		{"item by display name", head + `{"id":"i","act":"pickup","target":{"kind":"item","name":"Scroll of Inifuss"},"pre":["a(x)"],"post":["has_item(bks)"],"discover":"preset_always","budget_s":10,"on_fail":"park"}` + tail, 5},
		{"hidden memory: same post twice", head + step("p", `"a(x)"`, `"visited(town)"`, "") + "," + step("q", `"a(x)"`, `"visited(town)"`, "") + tail, 6},
		{"only transient post", head + step("p", `"a(x)"`, `"in_area(town)"`, "") + tail, 6},
		{"only a per-run post", head + step("p", `"a(x)"`, `"visited(town)"`, "") + tail, 6},
		{"consumed post redone", head + step("p", `"a(x)"`, `"has_item(bks)"`, "") + "," + step("q", `"has_item(bks)"`, `"has_item(bkd)"`, `,"consumes":["has_item(bks)"]`) + tail, 6},
		{"npc by display name", head + `{"id":"n","act":"talk","target":{"kind":"npc","name":"Akara"},"pre":["a(x)"],"post":["quest_state(q,done)"],"discover":"preset_always","budget_s":10,"on_fail":"park"}` + tail, 7},
		{"verb the build lacks", head + step("p", `"a(x)"`, `"quest_state(q,done)"`, `,"verbs":["debug warps2"]`) + tail, 8},
		{"unknown knowledge source", head + `{"id":"f","act":"interact","target":{"kind":"object","name":"o","class":3},"pre":["a(x)"],"post":["quest_state(q,done)"],"discover":"room_then_explore","budget_s":10,"on_fail":"park"}` + tail, 9},
		{"loop without lease", head + strings.Replace(step("p", `"a(x)"`, `"quest_state(q,done)"`, ""), `,"tries":2`, "", 1) + tail, 10},
	}
	// "fixed" knows room rows; the knowledge case needs a backend without them
	noRooms := *cp
	noRooms.Knowledge = []string{"preset_list", "preset_near"}
	for _, tc := range cases {
		backend := cp
		if tc.rule == 9 {
			backend = &noRooms
		}
		fs := Check(load(t, tc.js), backend)
		hit := false
		for _, f := range fs {
			hit = hit || f.Rule == tc.rule
		}
		if !hit {
			t.Errorf("%s: rule %d not reported, got %v", tc.name, tc.rule, fs)
		}
	}
}
