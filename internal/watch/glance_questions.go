package watch

import (
	"encoding/json"
	"strconv"

	"d2lod-bot/internal/decide"
)

// questionNeeds turns the hero's open questions and confirms into Needs you rows (S7). The old page answered them in its
// Thoughts card, which a cockpit does not have; so a question that waits for the player is a row with its first two choices
// as buttons (the answer routes are the old page's: /api/answer and /api/confirm). Every real choice is a button; "Not sure" is
// what the hero does on its own when the question fades, so it needs no button. Oldest question first: it fades first.
func (h *Hub) questionNeeds() []GlanceNeed {
	var out []GlanceNeed
	for _, t := range h.Thoughts().Thoughts {
		out = append(out, confirmNeedFor(t))
	}
	return out
}

// confirmNeedFor builds the row of one open thought (a question or a confirm).
func confirmNeedFor(t ThoughtItem) GlanceNeed {
	n := GlanceNeed{ID: "thought:" + strconv.Itoa(t.ID), Kind: "question", Sev: "warn", Actions: []GlanceVerb{}}
	n.Text = t.Headline
	if t.Kind == "confirm" {
		n.Text = t.Learn
	}
	if n.Text == "" {
		n.Text = t.Question
	}
	n.Sub = t.IfNothing
	if n.Sub == "" {
		n.Sub = "It fades in under a minute; then I decide."
	}
	route := "/api/answer"
	if t.Kind == "confirm" {
		route = "/api/confirm"
	}
	for _, c := range t.Choices {
		if c.ID == decide.NotSure {
			continue
		}
		body := map[string]any{"thought": t.ID, "answer": c.ID}
		if t.Kind == "confirm" {
			body = map[string]any{"thought": t.ID, "yes": c.ID == "yes"}
		}
		b, _ := json.Marshal(body)
		n.Actions = append(n.Actions, GlanceVerb{Label: c.Label, Post: route, Body: b})
	}
	return n
}
