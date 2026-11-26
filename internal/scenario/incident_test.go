package scenario

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeIncident(t *testing.T, class string, bug bool) string {
	t.Helper()
	in := map[string]any{"version": IncidentVersion, "kind": "death", "time": time.Date(2026, 11, 25, 2, 51, 36, 0, time.UTC),
		"areaName": "Cold Plains", "class": class, "tags": []string{"SETUP"},
		"pre":   map[string]any{"start": "field", "area": 3, "level": 8, "gold": 40, "life_pct": 90},
		"facts": map[string]any{"beltHP": 3, "beltMP": 1, "bug": bug, "summary": "x"}}
	b, _ := json.Marshal(in)
	p := filepath.Join(t.TempDir(), "i.json")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// An incident becomes a valid nightly row for two classes, with the belt and the position of the record.
func TestIncidentScenario(t *testing.T) {
	for _, class := range []string{"Sorceress", "Paladin"} {
		in, err := LoadIncident(writeIncident(t, class, true))
		if err != nil {
			t.Fatal(err)
		}
		s, err := in.Scenario("", 7)
		if err != nil {
			t.Fatalf("%s: %v", class, err)
		}
		if s.Tier != TierNightly || s.Pre.Area != 3 || s.Pre.Start != StartField || len(s.Pre.Belt) != 4 || s.Matrix.Classes[0] != class {
			t.Fatalf("%s: %+v", class, s)
		}
		if len(s.Assert) != 3 { // died, the potion bug, no writes
			t.Fatalf("%s: assert %+v", class, s.Assert)
		}
		if _, err := BuildSetup(s, class, 7); err != nil {
			t.Fatalf("%s: setup: %v", class, err)
		}
	}
	if _, err := LoadIncident(filepath.Join(t.TempDir(), "none.json")); err == nil {
		t.Fatal("missing file accepted")
	}
}

// The checked-in Questy rows load and validate (the report-only directory is not part of the default scenario set).
func TestQuestyIncidentRows(t *testing.T) {
	scs, err := LoadDir("../../tools/provingground/incidents")
	if err != nil || len(scs) != 3 {
		t.Fatalf("%d rows, %v", len(scs), err)
	}
	for _, s := range scs {
		if s.Tier != TierNightly {
			t.Errorf("%s: tier %s", s.ID, s.Tier)
		}
	}
}
