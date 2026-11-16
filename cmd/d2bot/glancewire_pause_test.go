package main

import (
	"testing"
	"time"

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

// The frozen pause: the band says "paused since HH:MM", the glance carries paused and pausedSince, and an ended pause
// leaves an info row.
func TestGlanceFrozenPause(t *testing.T) {
	since := time.Date(2026, 11, 15, 14, 5, 0, 0, time.Local)
	v := bot.PlayerPauseView{State: bot.PausePaused, Frozen: true, Since: &since}
	var src watch.GlanceSources
	for _, f := range glanceExtra {
		f(&src, glanceDeps{PauseV: func() bot.PlayerPauseView { return v }})
	}
	on, text := src.Paused()
	if !on || text != "paused since 14:05" || len(text) > 31 {
		t.Fatalf("band %v %q", on, text)
	}
	if s, ok := src.Frozen(); !ok || !s.Equal(since) {
		t.Fatalf("frozen %v %v", s, ok)
	}
	v = bot.PlayerPauseView{State: bot.PauseRunning, Ended: &bot.PauseEnded{Minutes: 61, Auto: true, At: since}}
	if _, ok := src.Frozen(); ok {
		t.Fatal("running must not read as frozen")
	}
	needs := src.Needs()
	if len(needs) != 1 || needs[0].Kind != "info" || needs[0].Sev != "info" || needs[0].Text != "Paused for 61 min; the pause cap resumed the game" {
		t.Fatalf("needs %+v", needs)
	}
	v.Ended = nil
	if len(src.Needs()) != 0 {
		t.Fatal("no row without a note")
	}
}
