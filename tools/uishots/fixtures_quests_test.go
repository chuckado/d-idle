package main

import (
	"os"
	"strings"
	"testing"
)

// Every quest fixture variant builds, in both builds, and shows the state its name says.
func TestQuestFixtureVariants(t *testing.T) {
	for _, class := range []string{"Paladin", "Sorceress"} {
		for variant, want := range map[string]string{"offer": "offers", "running": "running", "awaiting": "awaiting", "handing": "handIn", "confirming": "handIn", "failed": "handIn", "blocked": "blocked", "training": "offers", "done": "offers", "auto": "autoOn", "countdown": "countdown", "away": "away", "mid": "offers"} {
			v := questView(variant, class)
			if !v.Enabled {
				t.Fatalf("%s %s: the board is off", class, variant)
			}
			ok := false
			switch want {
			case "offers":
				ok = len(v.Offers) > 0
			case "autoOn":
				for _, r := range v.Quests {
					ok = ok || (r.ID == "den_of_evil" && r.Auto)
				}
			case "countdown":
				ok = v.AutoStart.Countdown != nil && v.AutoStart.Countdown.SecondsLeft == 10 && v.AutoStart.Countdown.Quest == "den_of_evil"
			case "away":
				ok = v.WhileAway != nil && len(v.WhileAway.Events) == 2
				has := false
				for _, n := range v.Needs {
					has = has || n.Kind == "quest_away"
				}
				ok = ok && has
			case "running":
				ok = v.Running != nil && v.Running.StepCount > 1 && v.Running.StepNo >= 1
			case "awaiting":
				ok = v.Awaiting != nil && v.Awaiting.NPC != ""
			case "handIn":
				ok = v.HandIn != nil && v.Awaiting != nil
			case "blocked":
				ok = v.Blocked != nil && v.Blocked.Why != ""
			}
			if !ok {
				t.Errorf("%s %s: view has no %s (needs %+v)", class, variant, want, v.Needs)
			}
		}
		if v := questView("off", class); v.Enabled || v.Reason == "" {
			t.Errorf("off view: %+v", v)
		}
	}
}

// Every quest defect, style or script, has a fit page and a verdict.
func TestQuestDefectsHavePages(t *testing.T) {
	sh, _ := os.ReadFile("../uishots.sh")
	names := map[string]bool{}
	for n := range defectCSS {
		if strings.HasPrefix(n, "q-") {
			names[n] = true
		}
	}
	for n := range questJS {
		names[n] = true
	}
	if len(names) != 4 {
		t.Errorf("want 4 quest defects, have %d", len(names))
	}
	for n := range names {
		rest := strings.TrimPrefix(n, "q-")
		if !strings.Contains(string(sh), "\nquestdefect-"+rest+" ") {
			t.Errorf("no fit page questdefect-%s", rest)
		}
		if _, ok := mustFail["fit-questdefect-"+rest+"-"]; !ok {
			t.Errorf("mustFail has no entry for questdefect-%s", rest)
		}
	}
}
