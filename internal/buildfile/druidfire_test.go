package buildfile_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"d2lod-bot/internal/agency"
	"d2lod-bot/internal/buildfile"
	"d2lod-bot/internal/leveler"
	"d2lod-bot/internal/plandata"
)

const (
	fireGoals = "testdata/druid-fire-guide.json"
	fireTable = "druid_fire_guide"
)

// druid1: the Druid fire guide track is a build file; the shipped plan table is its expansion plus engine data (selector,
// combat line, tier, hotkeys), and nothing else. UPDATE=1 rewrites the rows of the two table copies after a deliberate change.
func TestDruidFireGoalFileMatchesTable(t *testing.T) {
	c, err := agency.LoadCaps()
	if err != nil {
		t.Fatal(err)
	}
	rules := c.BuildRules()
	b, err := buildfile.LoadFile(fireGoals, rules) // strict parse + Validate under the capability rules (V skills.txt)
	if err != nil {
		t.Fatal(err)
	}
	gen, err := buildfile.ImportPlan(b, rules)
	if err != nil {
		t.Fatal(err)
	}
	var want struct {
		Rows json.RawMessage `json:"rows"`
	}
	if err := json.Unmarshal(gen, &want); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"../plandata/build-" + fireTable + ".json", "../../docs/plan/data/build-" + fireTable + ".json"} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var have map[string]json.RawMessage
		if err := json.Unmarshal(raw, &have); err != nil {
			t.Fatal(err)
		}
		if os.Getenv("UPDATE") != "" {
			have["rows"] = want.Rows
			out, _ := json.MarshalIndent(have, "", " ")
			if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		var a, g bytes.Buffer
		_ = json.Compact(&a, have["rows"])
		_ = json.Compact(&g, want.Rows)
		if a.String() != g.String() {
			t.Errorf("%s: rows differ from the goal file's expansion (UPDATE=1 go test ./internal/buildfile -run DruidDruidFireGoalFile)", path)
		}
	}
	// the table the leveler reads is legal row by row (cap, required level, prerequisite)
	raw, _ := plandata.FS.ReadFile("build-" + fireTable + ".json")
	lb, err := leveler.ParseBuild(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range buildfile.CheckTable(lb, rules) {
		t.Errorf("%s", p)
	}
}
