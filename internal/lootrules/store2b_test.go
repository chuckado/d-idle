package lootrules

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func saveCustom(t *testing.T, s *Store, mut func(*RuleSet)) *RuleSet {
	t.Helper()
	in := *s.Active()
	in.Preset = PresetCustom
	in.Rules = append([]Rule(nil), in.Rules...)
	mut(&in)
	rs, err := s.Put(body(t, &in))
	if err != nil {
		t.Fatal(err)
	}
	return rs
}

func TestStoreUndoGoesBackAndForth(t *testing.T) {
	s, _ := openTmp(t)
	if s.HasUndo() {
		t.Error("undo offered before any save")
	}
	if _, err := s.Undo(1); !errors.Is(err, ErrNoUndo) {
		t.Fatalf("undo with no history: %v", err)
	}
	saveCustom(t, s, func(r *RuleSet) { r.Rules[0].Off = true }) // seq 2
	saveCustom(t, s, func(r *RuleSet) { r.Rules[1].Off = true }) // seq 3
	if !s.HasUndo() {
		t.Fatal("no undo after two saves")
	}
	var seen *RuleSet
	s.OnChange(func(rs *RuleSet) { seen = rs })
	if _, err := s.Undo(2); !errors.As(err, new(ErrConflict)) {
		t.Fatalf("stale seq must conflict: %v", err)
	}
	rs, err := s.Undo(3)
	if err != nil || rs.Seq != 4 || !rs.Rules[0].Off || rs.Rules[1].Off || seen != rs {
		t.Fatalf("undo: %v %+v", err, rs)
	}
	rs, err = s.Undo(4) // a second undo is the redo
	if err != nil || !rs.Rules[1].Off {
		t.Fatalf("redo: %v", err)
	}
}

func TestStoreImportExportRoundTripAndLimits(t *testing.T) {
	s, _ := openTmp(t)
	b, err := s.Export()
	if err != nil || !strings.Contains(string(b), `"schema": "lootrules/1"`) {
		t.Fatalf("export: %v", err)
	}
	for _, bad := range []string{"{", `{"schema":"lootrules/1","limits":{"maxActions":1},"rules":[],"junk":1}`, `{"schema":"lootrules/2","limits":{"maxActions":1},"rules":[]}`} {
		if _, err := s.Import([]byte(bad), 1); err == nil {
			t.Errorf("accepted %q", bad)
		}
	}
	if s.Active().Seq != 1 {
		t.Fatal("a refused import changed the store")
	}
	rs := Starter()
	rs.Name = "From a friend"
	rs.Seq, rs.Preset = 77, "starter"
	out, _ := rs.Marshal()
	got, err := s.Import(out, 1)
	if err != nil || got.Seq != 2 || got.Preset != PresetCustom || got.Name != "From a friend" {
		t.Fatalf("import: %v %+v", err, got)
	}
	if _, err := s.Import(out, 1); !errors.As(err, new(ErrConflict)) {
		t.Errorf("stale import: %v", err)
	}
	// too many rules, too long a text
	big := Starter()
	for i := 0; i < maxImportRules; i++ {
		big.Rules = append(big.Rules, Rule{ID: "x" + strings.Repeat("a", 3) + string(rune('a'+i%26)) + string(rune('a'+i/26%26)) + string(rune('a'+i/676)), Action: Ignore, Reason: "r"})
	}
	bb, _ := big.Marshal()
	if _, err := s.Import(bb, 2); err == nil || !strings.Contains(err.Error(), "more than") {
		t.Errorf("rule cap: %v", err)
	}
	long := Starter()
	long.Rules[0].Reason = strings.Repeat("w", maxImportString+1)
	lb, _ := long.Marshal()
	if _, err := s.Import(lb, 2); err == nil || !strings.Contains(err.Error(), "longer than") {
		t.Errorf("text cap: %v", err)
	}
}

func TestStoreClassFallbackForANewHero(t *testing.T) {
	s, p := openTmp(t)
	var seen *RuleSet
	s.OnChange(func(rs *RuleSet) { seen = rs })
	if s.UseClassFallback("Necromancer") {
		t.Error("unknown class changed the set")
	}
	if !s.UseClassFallback("Sorceress") || s.Active().Preset != "class:Sorceress" || s.Active().Seq != 1 || seen != s.Active() {
		t.Fatalf("fallback: %+v", s.Active())
	}
	for _, r := range s.Active().Rules {
		if len(r.Classes) > 0 {
			t.Errorf("class tag left on %s", r.ID)
		}
	}
	if s.UseClassFallback("Amazon") {
		t.Error("a second fallback changed a set the hero already has")
	}
	saveCustom(t, s, func(r *RuleSet) { r.Rules[0].Off = true })
	if s.UseClassFallback("Amazon") {
		t.Error("fallback over a saved set")
	}
	// a hero with a file never falls back
	s2 := OpenStore(p, testClasses)
	if s2.UseClassFallback("Amazon") || s2.Active().Seq != 2 {
		t.Error("a hero with a saved file fell back")
	}
	// a hero without one (another character's folder) does
	s3 := OpenStore(filepath.Join(t.TempDir(), "other.json"), testClasses)
	if !s3.UseClassFallback("Amazon") {
		t.Error("a new hero did not get its class preset")
	}
	_ = os.Remove(p)
}

func TestSellScopeRulesAreValidAndDescribed(t *testing.T) {
	for _, r := range Starter().Rules {
		if has(r.Scope, PlaceSell) {
			if d := Describe(r); !strings.HasPrefix(d, "At the vendor: ") {
				t.Errorf("%s: %q", r.ID, d)
			}
		}
	}
	rs := Starter()
	rs.Rules = append(rs.Rules, Rule{ID: "bad", Scope: []string{PlaceSell}, Action: Drop, Reason: "r"})
	if es := rs.Validate(); len(es) == 0 || !strings.Contains(es.Error(), "not valid in scope sell") {
		t.Errorf("drop at the vendor: %v", es)
	}
}
