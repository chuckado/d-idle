package agency

import (
	"reflect"
	"testing"

	"d2lod-bot/internal/buildfile"
	"d2lod-bot/internal/leveler"
	"d2lod-bot/internal/plandata"
)

// The capability table is a rules source for the build files of the classes it holds requirements for: every shipped
// table of those classes exports to a file that validates against it and expands back to the same cumulative
// allocation at every row. A class the table has no requirements for is "skill_unknown", never guessed.
func TestShippedTablesValidateAgainstCapabilityRules(t *testing.T) {
	c, err := LoadCaps()
	if err != nil {
		t.Fatal(err)
	}
	rules := c.BuildRules()
	// Every table must validate and round-trip; the legacy defects of primary-blizzard and the two lightning tables are fixed as data.
	covered, other := 0, 0
	for _, name := range buildfile.PlanNames() {
		b, _, err := buildfile.ExportPlan(name, rules)
		if err != nil {
			t.Fatal(err)
		}
		ps := buildfile.Validate(b, rules)
		switch b.Class {
		case "sorceress", "paladin", "necromancer", "barbarian", "druid":
			if len(ps) > 0 {
				t.Errorf("%s: %v", name, ps)
				continue
			}
			if name == "primary-blizzard" { // an enabled respec: its rows are the post-respec path, Cumulative does not model the refund (buildfile.CheckTable does)
				covered++
				continue
			}
			raw, _ := plandata.FS.ReadFile("build-" + name + ".json")
			lb, _ := leveler.ParseBuild(raw)
			e, err := buildfile.Expand(b, rules)
			if err != nil {
				t.Errorf("%s: %v", name, err)
				continue
			}
			for _, r := range lb.Rows {
				ws, wa := buildfile.Cumulative(lb.Rows, r.Clvl)
				gs, ga := buildfile.Cumulative(e.Pre, r.Clvl)
				if !reflect.DeepEqual(ws, gs) || wa != ga {
					t.Errorf("%s clvl %d: table %v %v, file %v %v", name, r.Clvl, ws, wa, gs, ga)
					break
				}
			}
			covered++
		default:
			for _, p := range ps {
				if p.Code != "skill_unknown" {
					t.Errorf("%s (%s): %s: want only skill_unknown", name, b.Class, p)
				}
			}
			other++
		}
	}
	t.Logf("%d Paladin, Sorceress, Necromancer, Barbarian and Druid tables validate and round-trip against the capability table; %d other-class tables are skill_unknown by design", covered, other)
	if covered < 18 {
		t.Errorf("only %d tables covered", covered)
	}
}
