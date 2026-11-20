package buildfile

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"d2lod-bot/internal/leveler"
	"d2lod-bot/internal/plandata"
)

// Every shipped plan table exports to a build file that validates under the permissive rules, imports back to a plan
// table the leveler parses, and has the same cumulative allocation at every row.
func TestPlansRoundTripThroughFiles(t *testing.T) {
	names := PlanNames()
	if len(names) < 20 {
		t.Fatalf("PlanNames found %d tables", len(names))
	}
	for _, name := range names {
		raw, _ := plandata.FS.ReadFile("build-" + name + ".json")
		lb, err := leveler.ParseBuild(raw)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		rules := looseRules{lb.Class}
		b, notes, err := ExportPlan(name, rules)
		if err != nil {
			t.Fatal(err)
		}
		if ps := Validate(b, rules); len(ps) > 0 {
			t.Errorf("%s: exported file does not validate: %v (notes %v)", name, ps, notes)
			continue
		}
		out, err := ImportPlan(b, rules)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		back, err := leveler.ParseBuild(out)
		if err != nil {
			t.Errorf("%s: imported table does not parse: %v", name, err)
			continue
		}
		bad := 0
		for _, r := range lb.Rows {
			ws, wa := Cumulative(lb.Rows, r.Clvl)
			gs, ga := Cumulative(back.Rows, r.Clvl)
			if !reflect.DeepEqual(ws, gs) || wa != ga {
				bad++
			}
		}
		if bad > 0 {
			t.Errorf("%s: %d of %d rows differ after export and import", name, bad, len(lb.Rows))
		}
	}
}

func onlyCode(ps []Problem, code string) bool {
	for _, p := range ps {
		if p.Code != code {
			return false
		}
	}
	return true
}

func TestImportPlanRefusesAnInvalidFile(t *testing.T) {
	b := &Build{Schema: Schema, Name: "bad", Class: "sorceress", Skills: []Goal{{Skill: "No Such Skill", To: 5}}}
	if _, err := ImportPlan(b, fixtureRules{}); err == nil || !strings.Contains(err.Error(), "skill_unknown") {
		t.Fatalf("want skill_unknown, got %v", err)
	}
}

func TestSaveLoadFile(t *testing.T) {
	b, _, err := ExportPlan("sorceress_normal30", looseRules{"sorceress"})
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "x.build.json")
	if err := SaveFile(p, b, looseRules{"sorceress"}); err != nil {
		t.Fatal(err)
	}
	got, err := LoadFile(p, looseRules{"sorceress"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != b.Name || len(got.Skills) != len(b.Skills) {
		t.Fatalf("loaded %q with %d goals, saved %q with %d", got.Name, len(got.Skills), b.Name, len(b.Skills))
	}
	if _, err := LoadFile(filepath.Join(t.TempDir(), "missing.json"), looseRules{"sorceress"}); err == nil {
		t.Error("a missing file must be an error")
	}
}
