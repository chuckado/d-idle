package watch

import (
	"encoding/json"
	"regexp"
	"testing"
)

// The quest board UI (slice S5, docs/questboard.md "Watch UI"): the state to words mapping, the buttons, the Hand in read-back and
// the refusal text against the goldens of GET /api/quests, in a stub DOM (testdata/questui/quest_ui.mjs).
func TestQuestUIJS(t *testing.T) {
	runNode(t, "testdata/questui/quest_ui.mjs", "web/panels")
}

var questPanels = []string{"quest-strip", "quest-board", "quest-list"}

// The three quest panels keep the page rules (no timers, requests or queries of their own, text through textContent, no style
// writes) and declare themselves as opt-in cockpit slot panels reading the one /api/quests endpoint.
func TestQuestPanelsFollowTheContract(t *testing.T) {
	bad := regexp.MustCompile(`setInterval|setTimeout|fetch\(|EventSource|document\.|innerHTML|\.style\b|window\.`)
	for _, id := range questPanels {
		b, err := webFS.ReadFile("web/panels/" + id + "/panel.js")
		if err != nil {
			t.Fatal(err)
		}
		if m := bad.Find(b); m != nil {
			t.Errorf("%s: %q breaks the panel contract", id, m)
		}
		mb, _ := webFS.ReadFile("web/panels/" + id + "/panel.json")
		var m PanelManifest
		if err := json.Unmarshal(mb, &m); err != nil {
			t.Fatal(err)
		}
		if m.ID != id || !m.OptIn || !m.Slot || !m.CockpitOnly || m.Endpoint != QuestsRoute {
			t.Errorf("%s manifest: %+v", id, m)
		}
	}
}

// With the quest board on, glance says so (the cockpit then hides the Journey tile); off, the field is absent.
func TestGlanceQuestBoardFlag(t *testing.T) {
	h := newSimHub(t, "sorceress")
	h.SetGlanceSources(GlanceSources{QuestOn: func() bool { return true }})
	if !h.Glance().QuestBoard {
		t.Error("questBoard missing while the board is on")
	}
	h.SetGlanceSources(GlanceSources{})
	b, _ := json.Marshal(h.Glance())
	if regexp.MustCompile(`questBoard`).Match(b) {
		t.Errorf("questBoard present while off: %s", b)
	}
}
