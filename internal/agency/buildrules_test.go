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
	// Findings (V, this test against docs/research/sorc-skill-matrix.md, which read the rules from skills.txt): the two legacy
	// lightning tables take Lightning Mastery at clvl 26 (required level 30) and Chain Lightning at 27 without a point in
	// Lightning, its prerequisite. The game's rule (d2d can_learn) refuses both, so the expanded file differs from clvl 26
	// on (the mastery waits for 30, the prerequisite costs a point); the rows before are compared, the tables are untouched.
	legacyUntil := map[string]int{"fallback-lightning": 25, "fallback-lightning-v2": 25}
	covered, other := 0, 0
	for _, name := range buildfile.PlanNames() {
		b, _, err := buildfile.ExportPlan(name, rules)
		if err != nil {
			t.Fatal(err)
		}
		ps := buildfile.Validate(b, rules)
		switch b.Class {
		case "sorceress", "paladin":
			if name == "primary-blizzard" { // V: the legacy table puts 43 points into Static Field (cap 20)
				for _, p := range ps {
					if p.Code != "target_range" {
						t.Errorf("%s: unexpected %s", name, p)
					}
				}
				if len(ps) == 0 {
					t.Errorf("%s: the known over-cap defect is gone, drop the exception", name)
				}
				continue
			}
			if len(ps) > 0 {
				t.Errorf("%s: %v", name, ps)
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
				if u, ok := legacyUntil[name]; ok && r.Clvl > u {
					break
				}
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
	t.Logf("%d Paladin and Sorceress tables validate and round-trip against the capability table; %d other-class tables are skill_unknown by design", covered, other)
	if covered < 10 {
		t.Errorf("only %d tables covered", covered)
	}
}
