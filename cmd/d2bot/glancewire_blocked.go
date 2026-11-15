package main

import (
	"encoding/json"

	"d2lod-bot/internal/goals"
	"d2lod-bot/internal/watch"
)

// B6: the journey's "blocked" reason (goals.Status.JourneyBlocked, set by internal/bot/staletarget.go while the quest
// step is refused and the hero farms the level plan meanwhile) reaches the watch screen twice: as a Needs you row of
// /api/glance and as blocked[] of /api/commands. An empty reason clears both. No new state: both read the runner's
// status on every poll.

func init() {
	glanceExtra = append(glanceExtra, func(src *watch.GlanceSources, d glanceDeps) {
		if d.Status == nil {
			return
		}
		prev := src.Needs
		src.Needs = func() []watch.GlanceNeed {
			var out []watch.GlanceNeed
			if prev != nil {
				out = prev()
			}
			if n, ok := blockedNeed(d.Status().JourneyBlocked, d.Chapter); ok {
				out = append(out, n)
			}
			return out
		}
	})
}

// blockedNeed is the Needs you row for a reason; ok is false for "". The actions go to the journey's own intent route
// (Go now sends the hero, Hold parks the chapter) and exist only while the chapter is known.
func blockedNeed(reason string, chapter func() string) (watch.GlanceNeed, bool) {
	if reason == "" {
		return watch.GlanceNeed{}, false
	}
	n := watch.GlanceNeed{ID: "journey:blocked", Kind: "journey_blocked", Sev: "warn", Text: "Journey blocked: " + reason, Actions: []watch.GlanceVerb{}}
	if chapter != nil {
		if ch := chapter(); ch != "" {
			go_, _ := json.Marshal(map[string]any{"chapter": ch, "goNow": true})
			hold, _ := json.Marshal(map[string]any{"chapter": ch, "intent": "hold"})
			n.Actions = []watch.GlanceVerb{
				{Label: "Go now", Post: watch.CampaignIntentRoute, Body: go_},
				{Label: "Hold chapter", Post: watch.CampaignIntentRoute, Body: hold},
			}
		}
	}
	return n, true
}

// journeyBlockedHook is CommandHooks.Blocked: the same reason for /api/commands.
func journeyBlockedHook(status func() goals.Status) func() string {
	if status == nil {
		return nil
	}
	return func() string { return status().JourneyBlocked }
}

// journeyChapter is the chapter of the journey's last decision, "" without a journey or a decision.
func journeyChapter(ref *runnerRef) string {
	if ref == nil || ref.jc == nil {
		return ""
	}
	if d := ref.jc.View().Decision; d != nil {
		return d.Chapter
	}
	return ""
}
