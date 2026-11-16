package main

import "d2lod-bot/internal/watch"

// The farm loop switched itself off for this session after too many fresh games (internal/bot/farmreset.go, the thrash
// cap): one info row on the Needs you list says so, read from the runner's status (Status.FarmResetOff), gone when the
// knob is flipped off and on again or the bot restarts.

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
			if n, ok := farmOffNeed(d.Status().FarmResetOff); ok {
				out = append(out, n)
			}
			return out
		}
	})
}

// farmOffNeed is the info row for a reason; ok is false for "".
func farmOffNeed(why string) (watch.GlanceNeed, bool) {
	if why == "" {
		return watch.GlanceNeed{}, false
	}
	return watch.GlanceNeed{ID: "farmreset:off", Kind: "farm_reset_off", Sev: "info",
		Text: "Farm loop switched off for this session: " + why + ". Turn it off and on in Zone clear to try again.", Actions: []watch.GlanceVerb{}}, true
}
