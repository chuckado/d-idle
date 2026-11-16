package town

import (
	"testing"

	"d2lod-bot/internal/tidy"
)

// planDecided with nothing decided is tidy.Policy.Plan; engine decisions replace the policy's for their items and keep
// the order (stash best value first, drops cheapest first) and the cap. Two classes.
func TestPlanDecidedMatchesPolicyPlanAndAppliesOverrides(t *testing.T) {
	p, err := tidy.Default()
	if err != nil {
		t.Fatal(err)
	}
	for _, class := range []string{"Sorceress", "Amazon"} {
		c := tidy.Ctx{Class: class, PackFullPct: 90, StashFullPct: 10}
		items := []tidy.Item{
			{ID: 1, Facts: tidy.Facts{Place: "pack", Quality: "normal", Identified: true, Code: "r01", Type: "rune", SellValue: 5}},
			{ID: 2, Facts: tidy.Facts{Place: "pack", Quality: "normal", Identified: true, Code: "r02", Type: "rune", SellValue: 50}},
			{ID: 3, Facts: tidy.Facts{Place: "pack", Quality: "normal", Identified: true, Code: "cap", Type: "helm", SellValue: 3}},
		}
		if got, want := planDecided(items, nil, p, c, 12), p.Plan(items, c, 12); len(got) != len(want) {
			t.Fatalf("%s: undecided plan %v differs from policy %v", class, got, want)
		}
		decided := map[uint32]tidy.Decision{1: {Action: tidy.Keep, Rule: "r", Why: "engine keeps"}, 3: {Action: tidy.Stash, Rule: "r", Why: "engine stashes"}}
		got := planDecided(items, decided, p, c, 12)
		if len(got) != 2 || got[0].ID != 2 && got[0].ID != 3 {
			t.Fatalf("%s: plan %v", class, got)
		}
		for _, e := range got {
			if e.ID == 1 {
				t.Errorf("%s: item 1 was decided keep but is planned", class)
			}
		}
		if got := planDecided(items, decided, p, c, 1); len(got) != 1 {
			t.Errorf("%s: cap not applied: %v", class, got)
		}
	}
}
