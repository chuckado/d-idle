package main

import (
	"encoding/json"
	"testing"

	"d2lod-bot/internal/goals"
	"d2lod-bot/internal/watch"
)

func TestJourneyBlockedRowAndCommands(t *testing.T) {
	for _, class := range []string{"Sorceress", "Paladin"} {
		reason := "no mercenary yet"
		st := goals.Status{Character: goals.Character{Class: class}, JourneyBlocked: reason}
		status := func() goals.Status { return st }

		var src watch.GlanceSources
		for _, f := range glanceExtra {
			f(&src, glanceDeps{Status: status, Chapter: func() string { return "blood-raven" }})
		}
		if src.Needs == nil {
			t.Fatal("no Needs source registered")
		}
		rows := src.Needs()
		if len(rows) != 1 || rows[0].Kind != "journey_blocked" || rows[0].Text != "Journey blocked: "+reason || rows[0].Sev != "warn" {
			t.Fatalf("%s: rows %+v", class, rows)
		}
		if len(rows[0].Actions) != 2 || rows[0].Actions[0].Label != "Go now" || rows[0].Actions[1].Label != "Hold chapter" {
			t.Fatalf("%s: actions %+v", class, rows[0].Actions)
		}
		var body map[string]any
		if err := json.Unmarshal(rows[0].Actions[1].Body, &body); err != nil || body["chapter"] != "blood-raven" || body["intent"] != "hold" {
			t.Errorf("hold body %s", rows[0].Actions[1].Body)
		}

		cv := watch.BuildCommands(watch.CommandSources{Hooks: watch.CommandHooks{Blocked: journeyBlockedHook(status)}})
		if len(cv.Blocked) != 1 || cv.Blocked[0].Code != "journey_blocked" {
			t.Fatalf("%s: blocked %+v", class, cv.Blocked)
		}

		st.JourneyBlocked = "" // cleared: both go
		if rows := src.Needs(); len(rows) != 0 {
			t.Errorf("%s: row stays after clearing: %+v", class, rows)
		}
		if cv := watch.BuildCommands(watch.CommandSources{Hooks: watch.CommandHooks{Blocked: journeyBlockedHook(status)}}); len(cv.Blocked) != 0 {
			t.Errorf("%s: blocked[] stays: %+v", class, cv.Blocked)
		}
	}
}

func TestBlockedNeedWithoutChapterHasNoActions(t *testing.T) {
	n, ok := blockedNeed("x", func() string { return "" })
	if !ok || len(n.Actions) != 0 {
		t.Errorf("%+v %v", n, ok)
	}
	if _, ok := blockedNeed("", nil); ok {
		t.Error("empty reason must give no row")
	}
	if journeyBlockedHook(nil) != nil {
		t.Error("nil status must give no hook")
	}
}
