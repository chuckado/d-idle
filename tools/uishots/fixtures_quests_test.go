package main

import "testing"

// Every quest fixture variant builds, in both builds, and shows the state its name says.
func TestQuestFixtureVariants(t *testing.T) {
	for _, class := range []string{"Paladin", "Sorceress"} {
		for variant, want := range map[string]string{"offer": "offers", "running": "running", "awaiting": "awaiting", "handing": "handIn", "confirming": "handIn", "failed": "handIn", "blocked": "blocked", "training": "offers", "done": "offers", "auto": "offers", "mid": "offers"} {
			v := questView(variant, class)
			if !v.Enabled {
				t.Fatalf("%s %s: the board is off", class, variant)
			}
			ok := false
			switch want {
			case "offers":
				ok = len(v.Offers) > 0
			case "running":
				ok = v.Running != nil
			case "awaiting":
				ok = v.Awaiting != nil && v.Awaiting.NPC != ""
			case "handIn":
				ok = v.HandIn != nil && v.Awaiting != nil
			case "blocked":
				ok = v.Blocked != nil && v.Blocked.Why != ""
			}
			if !ok {
				t.Errorf("%s %s: view has no %s: %+v", class, variant, want, v)
			}
		}
		if v := questView("off", class); v.Enabled || v.Reason == "" {
			t.Errorf("off view: %+v", v)
		}
	}
}
