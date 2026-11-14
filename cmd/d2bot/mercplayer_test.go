package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/merc"
)

// innerGear is the inventory desk's stand-in: nothing queued.
type innerGear struct{ ran int }

func (g *innerGear) Pending() int                        { return 0 }
func (g *innerGear) RunPending(context.Context) []string { g.ran++; return nil }

type playerRig struct {
	w    *mercWire
	rev  *fakeReviver
	rd   *fakeReader
	clk  *clock
	gear mercGear
	hero *d2.Snapshot
}

func newPlayerRig(t *testing.T, class d2.Class, gold int, st merc.MercState) *playerRig {
	t.Helper()
	rd := &fakeReader{st: st}
	rev := &fakeReviver{paid: 750}
	w, _, clk := testWire(t, rd, rev)
	hero := townSnap(gold, true)
	hero.Player.Class = class
	hero.Player.Area = 2 // Blood Moor: the field
	r := &playerRig{w: w, rev: rev, rd: rd, clk: clk, hero: hero}
	r.gear = mercGear{inner: &innerGear{}, snap: func() *d2.Snapshot { return r.hero }}
	latestMerc.Store(w)
	t.Cleanup(func() { latestMerc.Store(nil) })
	return r
}

var deadMerc = merc.MercState{Known: true, Owned: true, Level: 10} // price 750

func TestPlayerReviveDeadWithGoldGoesHomeAndRevives(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		r := newPlayerRig(t, class, 5000, deadMerc)
		api := mercAPIAdapter{}
		if v := api.View().(merc.PlayerView); v.State != "dead" || !v.Button.Enabled {
			t.Fatalf("%v: %+v", class, v)
		}
		if r.gear.ErrandsWaiting() != 0 || r.gear.Pending() != 0 {
			t.Fatalf("%v: busy without a request", class)
		}
		if _, err := api.Request(); err != nil {
			t.Fatal(err)
		}
		// in the field: the errand trip is wanted, nothing runs yet
		if r.gear.ErrandsWaiting() != 1 || r.gear.Pending() != 0 || r.gear.ErrandKind() != "revive_merc" {
			t.Fatalf("%v: errands %d pending %d", class, r.gear.ErrandsWaiting(), r.gear.Pending())
		}
		r.hero.Player.Area = 1 // the trip brought the hero home
		if r.gear.ErrandsWaiting() != 0 || r.gear.Pending() != 1 || r.gear.NextKind() != "revive_merc" {
			t.Fatalf("%v: in town errands %d pending %d", class, r.gear.ErrandsWaiting(), r.gear.Pending())
		}
		out := r.gear.RunPending(context.Background())
		if r.rev.calls != 1 || len(out) != 1 || !strings.Contains(out[0], "merc revived") {
			t.Fatalf("%v: calls %d out %v", class, r.rev.calls, out)
		}
		if v := api.View().(merc.PlayerView); v.Request == nil || v.Request.State != merc.ReqDone {
			t.Errorf("%v: %+v", class, v)
		}
		if r.gear.Pending() != 0 {
			t.Errorf("%v: still pending after the revive", class)
		}
		latestMerc.Store(nil)
	}
}

func TestPlayerReviveGoldShortAndAliveAreRefused(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		r := newPlayerRig(t, class, 700, deadMerc)
		api := mercAPIAdapter{}
		_, err := api.Request()
		var ref *merc.Refusal
		if !errors.As(err, &ref) || ref.Reason != "need 50 more gold" {
			t.Fatalf("%v: %v", class, err)
		}
		r.rd.set(merc.MercState{Known: true, Owned: true, Alive: true, Level: 10})
		r.clk.t = r.clk.t.Add(time.Minute) // the state read is cached for a few seconds
		r.hero.Player.Gold = 9000
		if _, err := api.Request(); err == nil || !strings.Contains(err.Error(), "alive") {
			t.Errorf("%v: %v", class, err)
		}
		latestMerc.Store(nil)
	}
}

func TestPlayerReviveGoldGoneBeforeTheTripMakesNoTrip(t *testing.T) {
	r := newPlayerRig(t, d2.Paladin, 5000, deadMerc)
	if _, err := (mercAPIAdapter{}).Request(); err != nil {
		t.Fatal(err)
	}
	if r.gear.ErrandsWaiting() != 1 {
		t.Fatal("no trip wanted")
	}
	r.hero.Player.Gold = 100 // spent on potions meanwhile
	if r.gear.ErrandsWaiting() != 0 {
		t.Error("a trip home for nothing")
	}
	if err := (mercAPIAdapter{}).Cancel(); err != nil {
		t.Errorf("cancel while waiting: %v", err)
	}
}

func TestPlayerReviveFailureKeepsTheReasonAndACooldown(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		r := newPlayerRig(t, class, 5000, deadMerc)
		r.rev.err = errors.New("hireling not found")
		api := mercAPIAdapter{}
		if _, err := api.Request(); err != nil {
			t.Fatal(err)
		}
		r.hero.Player.Area = 1
		r.gear.RunPending(context.Background())
		v := api.View().(merc.PlayerView)
		if v.Request == nil || v.Request.State != merc.ReqFailed || v.Request.Reason == "" || v.Button.Enabled || !strings.Contains(v.Button.Reason, "try again in") {
			t.Fatalf("%v: %+v", class, v)
		}
		if _, err := api.Request(); err == nil {
			t.Errorf("%v: request inside the cooldown", class)
		}
		r.clk.t = r.clk.t.Add(merc.PlayerReviveCooldown + time.Minute)
		r.rev.err = nil
		if v := api.View().(merc.PlayerView); !v.Button.Enabled {
			t.Errorf("%v: still off after the cooldown: %+v", class, v)
		}
		latestMerc.Store(nil)
	}
}

func TestPlayerReviveRoutesAreOffWithoutTheWiring(t *testing.T) {
	latestMerc.Store(nil)
	api := mercAPIAdapter{}
	v := api.View().(merc.PlayerView)
	if v.State != "none" || v.Button.Enabled || v.Button.Reason == "" {
		t.Errorf("%+v", v)
	}
	var ref *merc.Refusal
	if _, err := api.Request(); !errors.As(err, &ref) || ref.Code != "off" {
		t.Errorf("%v", err)
	}
	if mercGear{}.Pending() != 0 || mercGear{}.ErrandsWaiting() != 0 {
		t.Error("an empty gear is busy")
	}
}
