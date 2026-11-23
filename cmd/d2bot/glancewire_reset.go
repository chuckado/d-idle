package main

import (
	"fmt"
	"time"

	"d2lod-bot/internal/bot"
	"d2lod-bot/internal/watch"
)

// The Reset button on the band and in Needs you (docs/reset.md): a running reset shows as the word Resetting; a bot that
// made no progress for resetOfferAfter, or whose loop heartbeat is older than resetOfferBeat, gets a Needs you row with
// one action chip, Reset. The chip opens the cockpit's confirm sheet; it never resets by itself.
const (
	resetOfferAfter = 5 * time.Minute // the stall detector's clock (bot.Runner.StuckFor)
	resetOfferBeat  = 2 * time.Minute // the loop made no iteration at all (the watchdog fires at ten)
)

func init() {
	glanceExtra = append(glanceExtra, func(src *watch.GlanceSources, d glanceDeps) {
		if d.Reset == nil {
			return
		}
		src.Resetting = func() (bool, string) { return resetBandText(d.Reset()) }
		prev := src.Needs
		src.Needs = func() []watch.GlanceNeed {
			var out []watch.GlanceNeed
			if prev != nil {
				out = prev()
			}
			var stuck time.Duration
			if d.Stuck != nil {
				stuck = d.Stuck()
			}
			var beat time.Duration
			if d.Beat != nil {
				if t := d.Beat(); !t.IsZero() {
					beat = time.Since(t)
				}
			}
			if n, ok := stuckNeed(d.Reset().State, stuck, beat); ok {
				out = append(out, n)
			}
			return out
		}
	})
}

func resetBandText(v bot.ResetView) (bool, string) {
	switch v.State {
	case bot.ResetResetting, bot.ResetRejoining:
		if v.Reason != "" {
			return true, "reset: " + v.Reason
		}
		return true, "reset: leaving the game and rejoining"
	}
	return false, ""
}

// stuckNeed is the Needs you row for a stuck bot; ok is false while the bot moves or a reset already runs.
func stuckNeed(state string, stuck, beat time.Duration) (watch.GlanceNeed, bool) {
	if state == bot.ResetResetting || state == bot.ResetRejoining {
		return watch.GlanceNeed{}, false
	}
	var text string
	switch {
	case beat >= resetOfferBeat:
		text = fmt.Sprintf("The bot has not moved for %d min", int(beat/time.Minute))
	case stuck >= resetOfferAfter:
		text = fmt.Sprintf("The bot is stuck: no progress for %d min", int(stuck/time.Minute))
	default:
		return watch.GlanceNeed{}, false
	}
	return watch.GlanceNeed{ID: "stuck", Kind: "stuck", Sev: "warn", Text: text, Sub: "Reset leaves the game and rejoins; the plan and orders are kept",
		Actions: []watch.GlanceVerb{{Label: "Reset", Post: watch.ResetRoute, Body: []byte(`{}`)}}}, true
}
