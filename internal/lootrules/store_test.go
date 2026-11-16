package lootrules

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var testClasses = []string{"Amazon", "Sorceress"}

func openTmp(t *testing.T) (*Store, string) {
	p := filepath.Join(t.TempDir(), "lootrules.json")
	return OpenStore(p, testClasses), p
}

func body(t *testing.T, rs *RuleSet) []byte {
	b, err := json.Marshal(rs)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestStoreDefaultsToStarterAndSavesWithBackup(t *testing.T) {
	s, p := openTmp(t)
	if s.Active().Preset != PresetStarter || s.Active().Seq != 1 || len(s.Active().Rules) != len(Starter().Rules) {
		t.Fatalf("default %+v", s.Active())
	}
	var seen *RuleSet
	s.OnChange(func(rs *RuleSet) { seen = rs })
	in := *s.Active()
	in.Preset = PresetCustom
	in.Rules = append([]Rule(nil), in.Rules...)
	in.Rules[0].Off = true
	rs, err := s.Put(body(t, &in))
	if err != nil || rs.Seq != 2 || seen != rs {
		t.Fatalf("put: %v %+v", err, rs)
	}
	if b, err := os.ReadFile(p + ".bak"); err != nil {
		t.Error("the first save keeps the preset it replaced, so Undo can go back to it")
	} else if old, err := Parse(b); err != nil || old.Seq != 1 || old.Rules[0].Off {
		t.Errorf("first backup is not the starting set: %v", err)
	}
	in2 := *rs
	in2.Rules = append([]Rule(nil), rs.Rules...)
	in2.Rules[1].Off = true
	if rs, err = s.Put(body(t, &in2)); err != nil || rs.Seq != 3 {
		t.Fatalf("second put: %v", err)
	}
	bak, err := os.ReadFile(p + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	old, _ := Parse(bak)
	if old.Seq != 2 || !old.Rules[0].Off || old.Rules[1].Off {
		t.Errorf("backup is not the previous version: seq %d", old.Seq)
	}
	// a reopen reads what was saved
	again := OpenStore(p, testClasses)
	if again.Active().Seq != 3 || !again.Active().Rules[1].Off {
		t.Error("reopen lost the set")
	}
	left, _ := filepath.Glob(filepath.Join(filepath.Dir(p), ".lootrules-*.tmp"))
	if len(left) != 0 {
		t.Errorf("temp files left: %v", left)
	}
}

func TestStoreRefusesWithLineAndPath(t *testing.T) {
	s, _ := openTmp(t)
	_, err := s.Put([]byte("{\n \"schema\": \"lootrules/1\",\n \"rules\": [ ,\n}"))
	var es Errors
	if !errors.As(err, &es) || es[0].Line != 3 {
		t.Fatalf("syntax error %v", err)
	}
	in := *s.Active()
	in.Preset = PresetCustom
	in.Rules = append([]Rule(nil), in.Rules...)
	in.Rules[2].Action = "explode"
	_, err = s.Put(body(t, &in))
	if !errors.As(err, &es) || es[0].Path != "rules[2].action" {
		t.Fatalf("field error %v", err)
	}
	if s.Active().Seq != 1 {
		t.Error("a refused set must not change the store")
	}
	if _, err = s.Put([]byte(`{"schema":"lootrules/1","seq":1,"typo":3}`)); err == nil {
		t.Error("unknown key accepted")
	}
}

func TestStoreConflict(t *testing.T) {
	s, _ := openTmp(t)
	in := *s.Active()
	in.Preset = PresetCustom
	if _, err := s.Put(body(t, &in)); err != nil {
		t.Fatal(err)
	}
	_, err := s.Put(body(t, &in)) // still based on seq 1
	var c ErrConflict
	if !errors.As(err, &c) || c.Seq != 2 {
		t.Fatalf("want conflict at 2, got %v", err)
	}
}

func TestStoreDamagedFileFallsBackToStarter(t *testing.T) {
	p := filepath.Join(t.TempDir(), "lootrules.json")
	if err := os.WriteFile(p, []byte("{nope"), 0o600); err != nil {
		t.Fatal(err)
	}
	s := OpenStore(p, testClasses)
	if s.LoadErr == "" || s.Active().Name != "Starter" {
		t.Errorf("damaged file: %q %q", s.LoadErr, s.Active().Name)
	}
	if _, err := os.Stat(p + ".bad"); err != nil {
		t.Error("the damaged file must be kept as .bad")
	}
}

func TestPresetsPerClass(t *testing.T) {
	s, _ := openTmp(t)
	if len(s.Presets()) != 3 {
		t.Fatalf("presets %+v", s.Presets())
	}
	amz, err := s.Preset("class:Amazon")
	if err != nil {
		t.Fatal(err)
	}
	sorc, err := s.Preset("class:Sorceress")
	if err != nil {
		t.Fatal(err)
	}
	if len(amz.Rules) != len(sorc.Rules)+1 || !hasRule(amz, "amazon-javelins") || hasRule(sorc, "amazon-javelins") {
		t.Errorf("javelin rule: amazon %d sorceress %d rules", len(amz.Rules), len(sorc.Rules))
	}
	for _, r := range amz.Rules {
		if len(r.Classes) != 0 {
			t.Errorf("rule %s kept its class tag", r.ID)
		}
	}
	if _, err := s.Preset("class:Nobody"); err == nil {
		t.Error("unknown class accepted")
	}
	// choosing a class preset through Put
	rs, err := s.Put([]byte(`{"schema":"lootrules/1","seq":1,"name":"x","limits":{"maxActions":1},"preset":"class:Sorceress","rules":[]}`))
	if err != nil || hasRule(rs, "amazon-javelins") || rs.Preset != "class:Sorceress" || len(rs.Rules) != len(sorc.Rules) {
		t.Fatalf("put preset: %v", err)
	}
	for _, cls := range testClasses {
		pre, _ := s.Preset("class:" + cls)
		if _, err := Compile(pre, cls, ""); err != nil {
			t.Errorf("%s preset does not compile: %v", cls, err)
		}
	}
}

func hasRule(rs *RuleSet, id string) bool {
	for _, r := range rs.Rules {
		if r.ID == id {
			return true
		}
	}
	return false
}

func TestCustomizeOptionsEditNamedRules(t *testing.T) {
	no := false
	rs, err := Customize(Starter(), Options{SellGear: "never", MinRune: 5, KeepSockets: &no, MinGold: 40})
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]Rule{}
	for _, r := range rs.Rules {
		byID[r.ID] = r
	}
	if !byID["g-magic-rare"].Off || !byID["g-rw3"].Off || !byID["t-sockets"].Off || byID["g-runes"].When.All == nil || byID["g-spare-gold"].When.All == nil {
		t.Errorf("options not applied: %+v", byID["g-runes"].When)
	}
	if es := rs.Validate(); len(es) > 0 {
		t.Fatalf("customised set invalid: %v", es)
	}
	// behaviour: rune 3 (below the line) is no longer a pickup, rune 9 is; two classes
	for _, cls := range testClasses {
		c, err := Compile(rs, cls, "")
		if err != nil {
			t.Fatal(err)
		}
		low := Item{Place: PlaceGround, Code: "r03", Types: []string{"rune"}, RuneNo: 3, Quality: "normal", Identified: true}
		high := Item{Place: PlaceGround, Code: "r09", Types: []string{"rune"}, RuneNo: 9, Quality: "normal", Identified: true}
		cx := &Ctx{Class: cls, CLvl: 20, Gold: 5000}
		if d := c.Decide(&low, cx); d.Action == Pickup {
			t.Errorf("%s: rune 3 still picked up (%s)", cls, d.Rule)
		}
		if d := c.Decide(&high, cx); d.Action != Pickup {
			t.Errorf("%s: rune 9 not picked up (%s)", cls, d.Rule)
		}
	}
	if _, err := Customize(Starter(), Options{SellGear: "bogus", MinRune: 99}); err == nil {
		t.Error("bad options accepted")
	}
	// the base is not changed
	if byID0 := Starter(); byID0.Rules[3].When.All != nil {
		t.Error("Starter mutated")
	}
}

func TestStorePutOptionsRebuildsFromPreset(t *testing.T) {
	s, _ := openTmp(t)
	rs, err := s.Put([]byte(`{"schema":"lootrules/1","seq":1,"name":"x","limits":{"maxActions":1},"preset":"starter","options":{"minRune":8},"rules":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	if rs.Options == nil || rs.Options.MinRune != 8 || rs.Limits.MaxActions != Starter().Limits.MaxActions {
		t.Errorf("not rebuilt from the preset: %+v", rs.Limits)
	}
	if _, err := s.Put([]byte(`{"schema":"lootrules/1","seq":2,"name":"x","limits":{"maxActions":1},"options":{"minRune":8},"rules":[]}`)); err == nil || !strings.Contains(err.Error(), "preset") {
		t.Errorf("options without a preset: %v", err)
	}
}

func TestDescribePlainLanguage(t *testing.T) {
	for _, r := range Starter().Rules {
		d := Describe(r)
		if d == "" || strings.Contains(d, "{") {
			t.Errorf("rule %s: %q", r.ID, d)
		}
	}
	rs := Starter()
	for _, r := range rs.Rules {
		if r.ID == "g-poor-gear" {
			got := Describe(r)
			if !strings.HasPrefix(got, "On the ground: sell when ") || !strings.Contains(got, "gold on hand less than 300") {
				t.Errorf("poor gear: %q", got)
			}
		}
	}
}
