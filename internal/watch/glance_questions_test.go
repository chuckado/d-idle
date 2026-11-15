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
		if found.Text == "" || len(found.Actions) == 0 || len(found.Actions) > 2 {
			t.Fatalf("%s: row %+v", build, found)
		}
		a := found.Actions[0]
		var body map[string]any
		if err := json.Unmarshal(a.Body, &body); err != nil || a.Post != "/api/answer" || int(body["thought"].(float64)) != th.ID || body["answer"] == "notsure" || strings.TrimSpace(a.Label) == "" {
			t.Errorf("%s: first action %+v body %v", build, a, body)
		}
	}
}
