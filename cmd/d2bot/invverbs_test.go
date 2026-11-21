package main

import (
	"errors"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/invdesk"
)

// The instant lane stays closed while something else owns the hands: a cursor item, an open panel (vendor, stash, NPC
// menu, Esc menu) or a safety action; open otherwise.
func TestInputHeldClosesTheLane(t *testing.T) {
	var panel int
	for i := 0; i < len(d2.UIState{}); i++ {
		var u d2.UIState
		u[i] = 1
		if u.BlockingPanelOpen() {
			panel = i
			break
		}
	}
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		snap := func() *d2.Snapshot {
			return &d2.Snapshot{Player: &d2.Player{Class: class}, Inventory: &d2.Inventory{}}
		}
		busy := false
		hold := inputHeld(func() bool { return busy })
		if why := hold(snap()); why != "" {
			t.Errorf("%v: closed by %q", class, why)
		}
		s := snap()
		s.Inventory.Cursor = &d2.Unit{}
		if hold(s) == "" {
			t.Errorf("%v: cursor item did not close the lane", class)
		}
		s = snap()
		s.UI[panel] = 1
		if hold(s) == "" {
			t.Errorf("%v: open panel did not close the lane", class)
		}
		busy = true
		if hold(snap()) == "" {
			t.Errorf("%v: safety action did not close the lane", class)
		}
	}
}

// The client adapter carries the merc verbs and keeps the fork's refusal codes for the desk (B9).
func TestCoderMercVerbRefusalCodes(t *testing.T) {
	err := refusal(&d2dctl.ItemRefused{Code: d2dctl.ItemErrMercType, Why: "a Desert Guard does not carry a bow"})
	var vr *invdesk.VerbRefusal
	if !errors.As(err, &vr) || vr.Code != "merctype" {
		t.Fatalf("%v", err)
	}
	if d2dctl.CapItemMercEquip != invdesk.CapItemMercEquip || d2dctl.CapItemMercUnequip != invdesk.CapItemMercUnequip {
		t.Error("capability names of the client and the desk differ")
	}
}

// ui.managementFast opens the verbs, the input names and the instant lane together; with it off the three old knobs decide.
func TestManagementFastStandsForTheThreeKnobs(t *testing.T) {
	tu := env.DefaultTuning()
	e := &env.Env{Tune: &tu}
	v, in, ins := verbKnobs(e)
	if v() || in() || ins() {
		t.Fatal("default must be all off")
	}
	tu.UI.ManagementFast = 1
	if !v() || !in() || !ins() {
		t.Fatal("ui.managementFast must open all three")
	}
	tu.UI.ManagementFast = 0
	tu.Invdesk.Instant = 1
	if v() || in() || !ins() {
		t.Fatal("invdesk.instant alone must open only the lane's own switch")
	}
	if v2, in2, ins2 := verbKnobs(&env.Env{}); v2() || in2() || ins2() {
		t.Fatal("no tuning: off")
	}
}
