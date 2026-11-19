package main

import (
	"d2lod-bot/internal/data"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/agency"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
)

// Without tables the wire is nil and every method is safe: the bot runs as before.
func TestAgencyWireNilIsSafe(t *testing.T) {
	var w *agencyWire
	w.OnSnapshot(&d2.Snapshot{Player: &d2.Player{}})
	b, err := json.Marshal(w.View())
	if err != nil {
		t.Fatal(err)
	}
	var v agency.View
	if err := json.Unmarshal(b, &v); err != nil || v.Mode != "off" || !v.ShadowOnly || v.Enabled {
		t.Fatalf("nil wire view %+v err %v", v, err)
	}
	if wireAgency(nil, &env.Env{}, nil, nil, nil) != nil {
		t.Error("no game tables, no wire")
	}
}

func TestChainSnapRunsEveryHookAndSkipsNil(t *testing.T) {
	var got []int
	f := chainSnap(func(*d2.Snapshot) { got = append(got, 1) }, nil, func(*d2.Snapshot) { got = append(got, 2) })
	f(&d2.Snapshot{})
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("order %v", got)
	}
}

// The knob defaults to off: shipping this card changes nothing until someone sets agency.planner.
func TestAgencyPlannerDefaultsOff(t *testing.T) {
	if env.DefaultTuning().Agency.Planner != 0 {
		t.Fatal("agency.planner must default to 0")
	}
}

// With the knob at 1 the wire logs one `planner shadow:` line a minute and serves the view; at 0 it does neither.
func TestAgencyWireShadowLineAndView(t *testing.T) {
	var logs strings.Builder
	log := slog.New(slog.NewTextHandler(&logs, nil))
	e := &env.Env{Tune: &env.Tuning{}}
	e.Tune.Agency.Planner = 1
	now := time.Unix(5000, 0)
	db := &data.DB{Skills: []*data.Skill{{Key: "Sacrifice"}}, Classes: []*data.CharClass{{}, {}, {}, {Int: 15}}}
	w := wireAgency(db, e, func() string { return "Testy" }, func() time.Time { return now }, log)
	if w == nil {
		t.Fatal("no wire")
	}
	w.owned = func(*d2.Player) map[string]int { return map[string]int{"Sacrifice": 5, "Might": 1} }
	snap := &d2.Snapshot{Player: &d2.Player{Class: d2.Paladin, Level: 8, Energy: 20, MaxLife: 200, Area: 2}}
	w.OnSnapshot(snap)
	w.OnSnapshot(snap) // inside the minute: no second line
	if n := strings.Count(logs.String(), "planner shadow:"); n != 1 {
		t.Fatalf("%d lines in the first minute:\n%s", n, logs.String())
	}
	for _, want := range []string{"class=paladin", "plan=Sacrifice/5", "profile=Sacrifice"} {
		if !strings.Contains(logs.String(), want) {
			t.Errorf("line lacks %q:\n%s", want, logs.String())
		}
	}
	v, ok := w.View().(agency.View)
	if !ok || !v.Enabled || v.Mode != "shadow" || !v.ShadowOnly || len(v.Rotation) == 0 || v.Rotation[0].Skill != "Sacrifice" {
		t.Fatalf("view %+v", v)
	}
	now = now.Add(time.Minute)
	w.OnSnapshot(snap)
	if n := strings.Count(logs.String(), "planner shadow:"); n != 2 {
		t.Errorf("a minute later: %d lines", n)
	}
	e.Tune.Agency.Planner = 0
	now = now.Add(time.Hour)
	w.OnSnapshot(snap)
	if n := strings.Count(logs.String(), "planner shadow:"); n != 2 {
		t.Errorf("knob 0 must log nothing: %d lines", n)
	}
	if v := w.View().(agency.View); v.Mode != "off" || v.Enabled {
		t.Errorf("knob 0 view %+v", v)
	}
}
