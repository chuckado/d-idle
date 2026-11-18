package main

import (
	"fmt"
	"time"

	"d2lod-bot/internal/bot"
	"d2lod-bot/internal/watch"
)

// B1 on the band: a pause the player asked for (docs/pause.md) shows as the word Paused with its reason, for every state
// of the pause (frozen, going home first, held in town). A frozen pause also fills `paused` and `pausedSince` of the
// glance, and a pause that just ended gets a Needs you info row. In a measured run the pause is never installed and stays
// running.

func init() {
	glanceExtra = append(glanceExtra, func(src *watch.GlanceSources, d glanceDeps) {
		if d.PauseV != nil {
			src.Paused = func() (bool, string) { return pauseBandView(d.PauseV()) }
			src.Frozen = func() (time.Time, bool) {
				v := d.PauseV()
				if !v.Frozen || v.Since == nil {
					return time.Time{}, false
				}
				return *v.Since, true
			}
			prev := src.Needs
			src.Needs = func() []watch.GlanceNeed {
				var out []watch.GlanceNeed
				if prev != nil {
					out = prev()
				}
				if n, ok := pauseEndedNeed(d.PauseV()); ok {
					out = append(out, n)
				}
				return out
			}
			return
		}
		if d.Pause != nil {
			src.Paused = func() (bool, string) { return pauseBandText(d.Pause()) }
		}
	})
}

// pauseBandView is pauseBandText with the frozen pause's own sentence, "paused since 14:05" (31 characters at most).
func pauseBandView(v bot.PlayerPauseView) (bool, string) {
	if v.Frozen && v.Since != nil {
		return true, "paused since " + v.Since.Local().Format("15:04")
	}
	return pauseBandText(v.State)
}

// pauseBandText is the band's reason for a pause state (31 characters at most, so it fits at 1280 px).
func pauseBandText(state string) (bool, string) {
	switch state {
	case bot.PauseGoingHome:
		return true, "paused by you, going home first"
	case bot.PausePaused:
		return true, "paused by you, Resume continues"
	}
	return false, ""
}

// pauseEndedNeed is the info row that says how long a frozen pause lasted (and that the cap ended it, when it did).
func pauseEndedNeed(v bot.PlayerPauseView) (watch.GlanceNeed, bool) {
	if v.Ended == nil {
		return watch.GlanceNeed{}, false
	}
	n := watch.GlanceNeed{ID: "paused-for", Kind: "info", Sev: "info", Text: v.Ended.PauseNeedText(),
		Sub: fmt.Sprintf("The game runs again at %s", v.Ended.At.Local().Format("15:04")), Actions: []watch.GlanceVerb{}}
	return n, true
}
