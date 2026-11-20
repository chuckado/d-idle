package questspec

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// act1BossQuests are the boss chunks that have a Proving Ground scenario (tools/provingground/scenarios/act1more.json).
var act1BossQuests = map[string]struct{ chunk, boss string }{
	"sisters_burial_grounds":   {"blood_raven", "blood_raven"},
	"tools_of_the_trade":       {"malus", "the_smith"},
	"sisters_to_the_slaughter": {"andariel", "andariel"},
}

// TestAct1BossChains: each boss chain kills its boss in an area it entered first, writes a kill-or-cleared fact the
// scenario can read as a quest record, and nothing after the kill needs a fact the kill does not give.
func TestAct1BossChains(t *testing.T) {
	chains, _, _, _ := shipped(t)
	byID := map[string]*Chain{}
	for _, c := range chains {
		byID[c.ID] = c
	}
	for q, want := range act1BossQuests {
		c := byID[q]
		if c == nil {
			t.Errorf("no chain %s", q)
			continue
		}
		var kill *Step
		for i := range c.Steps {
			if c.Steps[i].Act == "kill" {
				kill = &c.Steps[i]
			}
		}
		if kill == nil {
			t.Errorf("%s: no kill step", q)
			continue
		}
		if kill.Target.Name != want.boss || kill.Chunk != want.chunk {
			t.Errorf("%s: kill step targets %q in chunk %q, want %q in %q", q, kill.Target.Name, kill.Chunk, want.boss, want.chunk)
		}
		if len(kill.Post) != 1 || !strings.Contains(kill.Post[0], "killed("+want.boss+")|quest_state("+q+",cleared)") {
			t.Errorf("%s: kill post %v must offer the kill or the cleared record", q, kill.Post)
		}
		if kill.OnFail != "park" {
			t.Errorf("%s: a boss kill parks on failure (never loops), got %q", q, kill.OnFail)
		}
		entered := false
		for _, p := range kill.Pre {
			if strings.HasPrefix(p, "in_area(") {
				entered = true
				if c.Areas[strings.TrimSuffix(strings.TrimPrefix(p, "in_area("), ")")].Size == "" {
					t.Errorf("%s: kill pre %s names an area the chain does not declare", q, p)
				}
			}
		}
		if !entered {
			t.Errorf("%s: kill step has no in_area precondition", q)
		}
	}
}

// TestToolsMalusIsPickedBeforeCharsi: the Malus (hdm) is picked up after the Smith dies and before the walk back, so
// the hand-in step cannot run without it.
func TestToolsMalusIsPickedBeforeCharsi(t *testing.T) {
	chains, _, _, _ := shipped(t)
	for _, c := range chains {
		if c.ID != "tools_of_the_trade" {
			continue
		}
		idx := map[string]int{}
		for i, s := range c.Steps {
			idx[s.ID] = i
		}
		k, p, r := idx["kill_smith"], idx["pick_malus"], idx["return_charsi"]
		if !(k < p && p < r) {
			t.Fatalf("order kill_smith %d, pick_malus %d, return_charsi %d", k, p, r)
		}
		if c.Steps[p].Target.Code != "hdm" || !strings.Contains(c.Steps[r].Pre[0], "has_item(hdm)") {
			t.Errorf("the Malus code or the return precondition changed: %+v pre %v", c.Steps[p].Target, c.Steps[r].Pre)
		}
		return
	}
	t.Fatal("no tools_of_the_trade chain")
}

// TestBloodRavenTitleExists: the hero's title for the quest is a kill-count row of the hall table naming the monster
// the board's boss and the chain use (data check; the run has no Hall reader).
func TestBloodRavenTitleExists(t *testing.T) {
	b, err := os.ReadFile("../watch/hall_titles.json")
	if err != nil {
		t.Fatal(err)
	}
	var tab struct {
		Titles []struct {
			Text string `json:"text"`
			When struct {
				Monster string `json:"monster"`
				Min     int    `json:"min"`
			} `json:"when"`
		} `json:"titles"`
	}
	if err := json.Unmarshal(b, &tab); err != nil {
		t.Fatal(err)
	}
	for _, r := range tab.Titles {
		if r.Text == "Bane of Blood Raven" {
			if r.When.Monster != "Blood Raven" || r.When.Min != 1 {
				t.Errorf("title rule %+v, want Blood Raven x1", r.When)
			}
			return
		}
	}
	t.Error("no title 'Bane of Blood Raven' in the hall table")
}

// TestAct1BossScenarioFile: the scenario file names the board's quests and chunks of this table, one each.
func TestAct1BossScenarioFile(t *testing.T) {
	b, err := os.ReadFile("../../tools/provingground/scenarios/act1more.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Scenarios []struct {
			ID    string `json:"id"`
			Quest struct{ Quest, Chunk string }
		} `json:"scenarios"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, s := range f.Scenarios {
		w, ok := act1BossQuests[s.Quest.Quest]
		if !ok || w.chunk != s.Quest.Chunk {
			t.Errorf("%s: quest %s/%s is not in the boss table", s.ID, s.Quest.Quest, s.Quest.Chunk)
		}
		seen[s.Quest.Quest] = true
	}
	for q := range act1BossQuests {
		if !seen[q] {
			t.Errorf("no scenario for %s", q)
		}
	}
}
