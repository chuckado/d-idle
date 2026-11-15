package main

import (
	"d2lod-bot/internal/bot"
	"d2lod-bot/internal/watch"
)

// B1 on the band: a pause the player asked for (docs/pause.md) shows as the word Paused with its reason, for both
// states of the pause (going home first, held in town). In a measured run the pause is never installed and stays running.

func init() {
	glanceExtra = append(glanceExtra, func(src *watch.GlanceSources, d glanceDeps) {
		if d.Pause != nil {
			src.Paused = func() (bool, string) { return pauseBandText(d.Pause()) }
		}
	})
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
