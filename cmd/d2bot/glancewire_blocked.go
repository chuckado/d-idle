package main

import (
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
			if n, ok := blockedNeed(d.Status().JourneyBlocked); ok {
				out = append(out, n)
			}
			return out
		}
	})
}

// blockedNeed is the Needs you row for a reason; ok is false for "". It has no actions: the Journey and its Go now and
// Hold verbs are retired (quest board, S9), the reason only says why the floor planner's quest step is refused.
func blockedNeed(reason string) (watch.GlanceNeed, bool) {
	if reason == "" {
		return watch.GlanceNeed{}, false
	}
	return watch.GlanceNeed{ID: "journey:blocked", Kind: "journey_blocked", Sev: "warn", Text: "Quest step blocked: " + reason, Actions: []watch.GlanceVerb{}}, true
}

// journeyBlockedHook is CommandHooks.Blocked: the same reason for /api/commands.
func journeyBlockedHook(status func() goals.Status) func() string {
	if status == nil {
		return nil
	}
	return func() string { return status().JourneyBlocked }
}
