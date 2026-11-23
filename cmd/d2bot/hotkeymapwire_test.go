package main

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"d2lod-bot/internal/agency"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/hotkeymap"
)

type tableBinder struct {
	table [hotkeymap.Slots]int
	drop  int // slot index whose write is lost; -1 none
	binds int
}

func newTable(drop int) *tableBinder {
	t := &tableBinder{drop: drop}
	for i := range t.table {
		t.table[i] = hotkeymap.Unbound
	}
	return t
}
func (t *tableBinder) Writable() int { return popupSlots }
func (t *tableBinder) Bind(_ context.Context, slot, id int) error {
	t.binds++
	if slot != t.drop {
		t.table[slot] = id
	}
	return nil
}
func (t *tableBinder) Read(context.Context) ([hotkeymap.Slots]int, error) { return t.table, nil }

func hkEnv(on int) *env.Env {
	tu := env.DefaultTuning()
	tu.Input.HotkeyMap = on
	return &env.Env{Tune: &tu}
}

func twoClasses() []struct {
	name string
	plan agency.Plan
	ids  map[string]int
} {
	c := func(n string) *agency.Choice { return &agency.Choice{Skill: n} }
	return []struct {
		name string
		plan agency.Plan
		ids  map[string]int
	}{
		{"Paladin", agency.Plan{Main: c("Zeal"), Attack: c("Zeal"), Aura: c("Might"), Buffs: []agency.Choice{{Skill: "Holy Shield"}}},
			map[string]int{"Zeal": 101, "Might": 98, "Holy Shield": 117, townPortalSkill: 220}},
		{"Sorceress", agency.Plan{Main: c("Fire Ball"), Attack: c("Fire Ball"), Secondary: c("Blizzard"), Movement: []agency.Choice{{Skill: "Teleport"}}},
			map[string]int{"Fire Ball": 47, "Blizzard": 59, "Teleport": 54}}, // no town portal row: left out, not an error
	}
}

func TestApplyHotkeyMapKnobOffDoesNothing(t *testing.T) {
	for _, c := range twoClasses() {
		tb := newTable(-1)
		called := false
		ok := applyHotkeyMap(context.Background(), hkEnv(0), func() (agency.Plan, error) { called = true; return c.plan, nil },
			func(n string) (int, bool) { id, ok := c.ids[n]; return id, ok }, townPortalSkill, tb, slog.New(slog.DiscardHandler))
		if ok || called || tb.binds != 0 {
			t.Fatalf("%s: knob off ran: ok %v plan asked %v binds %d", c.name, ok, called, tb.binds)
		}
	}
}

func TestApplyHotkeyMapKnobOnWritesAndReadsBack(t *testing.T) {
	for _, c := range twoClasses() {
		tb := newTable(-1)
		ok := applyHotkeyMap(context.Background(), hkEnv(1), func() (agency.Plan, error) { return c.plan, nil },
			func(n string) (int, bool) { id, ok := c.ids[n]; return id, ok }, townPortalSkill, tb, slog.New(slog.DiscardHandler))
		if !ok || tb.table[0] != c.ids[c.plan.Main.Skill] || tb.binds == 0 {
			t.Fatalf("%s: ok %v slot0 %d binds %d", c.name, ok, tb.table[0], tb.binds)
		}
	}
}

// Must fail on both classes: a write the game ignores is reported (false) because the read-back differs.
func TestApplyHotkeyMapMustFailOnLostWrite(t *testing.T) {
	for _, c := range twoClasses() {
		tb := newTable(0)
		ok := applyHotkeyMap(context.Background(), hkEnv(1), func() (agency.Plan, error) { return c.plan, nil },
			func(n string) (int, bool) { id, ok := c.ids[n]; return id, ok }, townPortalSkill, tb, slog.New(slog.DiscardHandler))
		if ok {
			t.Fatalf("%s: a lost write was reported as applied", c.name)
		}
	}
	if applyHotkeyMap(context.Background(), hkEnv(1), func() (agency.Plan, error) { return agency.Plan{}, errors.New("no plan") },
		func(string) (int, bool) { return 0, false }, "", newTable(-1), slog.New(slog.DiscardHandler)) {
		t.Fatal("a missing plan was reported as applied")
	}
}

type stateStub struct{ line string }

func (s stateStub) Do(context.Context, string) (*d2dctl.Reply, error) {
	return &d2dctl.Reply{Cmd: "state", Lines: []string{s.line}}, nil
}

func TestPopupBinderReadsTheStateLine(t *testing.T) {
	b := popupBinder{state: stateStub{"screen=ingame hotkeys=101,65535,65535,65535,65535,65535,65535,65535,65535,65535,65535,65535,65535,65535,65535,65535 x=1"}}
	got, err := b.Read(context.Background())
	if err != nil || got[0] != 101 || got[1] != hotkeymap.Unbound {
		t.Fatalf("%v %v", got, err)
	}
	var keys []string
	b.bind = func(_ context.Context, id int, key string) error { keys = append(keys, key); return nil }
	b.Bind(context.Background(), 0, 7)
	b.Bind(context.Background(), 7, 8)
	if len(keys) != 2 || keys[0] != "f1" || keys[1] != "f8" {
		t.Fatalf("keys %v", keys)
	}
}
