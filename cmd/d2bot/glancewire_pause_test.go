package main

import (
	"testing"

	"d2lod-bot/internal/bot"
	"d2lod-bot/internal/watch"
)

func TestGlanceExtraBindsPause(t *testing.T) {
	state := bot.PauseRunning
	var src watch.GlanceSources
	for _, f := range glanceExtra {
		f(&src, glanceDeps{Pause: func() string { return state }})
	}
	if src.Paused == nil {
		t.Fatal("GlanceSources.Paused is not bound")
	}
	for _, c := range []struct {
		state string
		on    bool
		text  string
	}{
		{bot.PauseRunning, false, ""},
		{bot.PauseGoingHome, true, "paused by you, going home first"},
		{bot.PausePaused, true, "paused by you, Resume continues"},
	} {
		state = c.state
		on, text := src.Paused()
		if on != c.on || text != c.text {
			t.Errorf("%s: %v %q", c.state, on, text)
		}
		if len(text) > 31 {
			t.Errorf("%q is %d characters, band limit 31", text, len(text))
		}
	}
}
