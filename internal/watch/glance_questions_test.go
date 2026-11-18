package watch

import (
	"encoding/json"
	"strings"
	"testing"
)

// S7: an open question is a Needs you row with its first two choices as buttons that post to the old answer route, for both
// classes; a rig with nothing open has no such row.
func TestOpenQuestionIsANeedsRow(t *testing.T) {
	for _, build := range []string{"sorceress", "paladin"} {
		r := newRig(t, build)
		if n := len(r.sim.Hub.Glance().Needs); n != 0 {
			t.Fatalf("%s: needs before a question: %d", build, n)
		}
		r.fight()
		th := r.nextThought()
		var found *GlanceNeed
		for _, n := range r.sim.Hub.Glance().Needs {
			if n.Kind == "question" {
				n := n
				found = &n
			}
		}
		if found == nil {
			t.Fatalf("%s: no question row for thought %d", build, th.ID)
		}
		if found.Text == "" || len(found.Actions) == 0 || len(found.Actions) > 8 {
			t.Fatalf("%s: row %+v", build, found)
		}
		a := found.Actions[0]
		var body map[string]any
		if err := json.Unmarshal(a.Body, &body); err != nil || a.Post != "/api/answer" || int(body["thought"].(float64)) != th.ID || body["answer"] == "notsure" || strings.TrimSpace(a.Label) == "" {
			t.Errorf("%s: first action %+v body %v", build, a, body)
		}
	}
}

// Every answer of a question survives the cleaning (some points have five); other needs keep three buttons.
func TestQuestionKeepsAllAnswers(t *testing.T) {
	var acts []GlanceVerb
	for _, l := range []string{"a", "b", "c", "d", "e"} {
		acts = append(acts, GlanceVerb{Label: l, Post: "/api/answer"})
	}
	if n := cleanNeed(GlanceNeed{Kind: "question", Actions: acts}); len(n.Actions) != 5 {
		t.Errorf("question answers: %d", len(n.Actions))
	}
	if n := cleanNeed(GlanceNeed{Kind: "gear", Actions: acts}); len(n.Actions) != 3 {
		t.Errorf("other need buttons: %d", len(n.Actions))
	}
}

// A confirm is a row whose buttons post yes or no to /api/confirm.
func TestConfirmNeedPostsToConfirm(t *testing.T) {
	r := newRig(t, "sorceress")
	h := r.sim.Hub
	_ = h
	// the row builder is the same code path: build the need from a confirm item by hand
	n := confirmNeedFor(ThoughtItem{ID: 7, Kind: "confirm", Learn: "Always drink here. OK?", Choices: []ChoiceView{{"yes", "Yes, always", ""}, {"no", "No", ""}}})
	if len(n.Actions) != 2 || n.Actions[0].Post != "/api/confirm" || !strings.Contains(string(n.Actions[0].Body), `"yes":true`) || !strings.Contains(string(n.Actions[1].Body), `"yes":false`) {
		t.Errorf("confirm row: %+v", n)
	}
}
