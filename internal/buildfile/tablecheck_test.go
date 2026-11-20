package buildfile_test

import (
	"encoding/json"
	"testing"

	"d2lod-bot/internal/agency"
	"d2lod-bot/internal/buildfile"
	"d2lod-bot/internal/leveler"
	"d2lod-bot/internal/plandata"
)

// Must-fail: every shipped Sorceress and Paladin plan table is legal under the capability table's rules (V from
// skills.txt for the Sorceress, S for the Paladin). Two classes, so the rule is not Sorceress-only.
func TestShippedTablesAreLegal(t *testing.T) {
	c, err := agency.LoadCaps()
	if err != nil {
		t.Fatal(err)
	}
	rules := c.BuildRules()
	seen := map[string]int{}
	for _, name := range buildfile.PlanNames() {
		raw, _ := plandata.FS.ReadFile("build-" + name + ".json")
		lb, err := leveler.ParseBuild(raw)
		if err != nil {
			t.Fatal(err)
		}
		if lb.Class != "sorceress" && lb.Class != "paladin" {
			continue
		}
		seen[lb.Class]++
		// A respec marked "enabled": false is documentation (rows do not depend on it): the rows are checked as one path.
		var meta struct {
			Respec struct {
				Enabled *bool `json:"enabled"`
			} `json:"respec"`
		}
		_ = json.Unmarshal(raw, &meta)
		if e := meta.Respec.Enabled; e != nil && !*e {
			lb.Respec, lb.RespecPool = 0, nil
		}
		for _, p := range buildfile.CheckTable(lb, rules) {
			t.Errorf("%s: %s", name, p)
		}
	}
	if seen["sorceress"] < 5 || seen["paladin"] < 3 {
		t.Errorf("tables checked per class: %v", seen)
	}
}

// The checker itself rejects the three defect kinds, for both classes.
func TestCheckTableRejects(t *testing.T) {
	c, err := agency.LoadCaps()
	if err != nil {
		t.Fatal(err)
	}
	rules := c.BuildRules()
	row := func(clvl int, skill string, pts int) leveler.Row {
		return leveler.Row{Clvl: clvl, Skills: []leveler.SkillPts{{Skill: skill, Points: pts}}}
	}
	cases := []struct {
		name string
		lb   *leveler.Build
		code string
	}{
		{"sorceress over cap", &leveler.Build{Class: "sorceress", Rows: []leveler.Row{row(6, "Static Field", 21)}}, buildfile.PTableCap},
		{"sorceress early level", &leveler.Build{Class: "sorceress", Rows: []leveler.Row{row(26, "Lightning Mastery", 1)}}, buildfile.PTableLevel},
		{"sorceress prereq", &leveler.Build{Class: "sorceress", Rows: []leveler.Row{row(27, "Chain Lightning", 1)}}, buildfile.PTablePrereq},
		{"paladin over cap", &leveler.Build{Class: "paladin", Rows: []leveler.Row{row(2, "Might", 21)}}, buildfile.PTableCap},
		{"paladin early level", &leveler.Build{Class: "paladin", Rows: []leveler.Row{row(5, "Holy Shield", 1)}}, buildfile.PTableLevel},
		{"paladin prereq", &leveler.Build{Class: "paladin", Rows: []leveler.Row{row(30, "Fanaticism", 1)}}, buildfile.PTablePrereq},
	}
	for _, tc := range cases {
		ps := buildfile.CheckTable(tc.lb, rules)
		ok := false
		for _, p := range ps {
			ok = ok || p.Code == tc.code
		}
		if !ok {
			t.Errorf("%s: want %s, got %v", tc.name, tc.code, ps)
		}
	}
	if ps := buildfile.CheckTable(&leveler.Build{Class: "paladin", Rows: []leveler.Row{row(2, "Might", 1), row(3, "Might", 1)}}, rules); len(ps) != 0 {
		t.Errorf("legal table rejected: %v", ps)
	}
}
