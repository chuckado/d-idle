package bot

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
)

// kitField is a Sorceress in the Stony Field (area 4) with the kit parts asked for. Mana potions have no catalog entry in
// the fake, so the tests set kitMinMP to 0 except where noted.
func kitField(t *testing.T, level, hp int, tp bool, gold int) (*rig, *d2.Snapshot) {
	t.Helper()
	g := newRig(t)
	g.e.Tune.Resource.KitMinMP = 0
	g.w.set(func(w *world) { w.level, w.area, w.gold0 = level, 4, gold })
	ids := make([]uint32, hp)
	for i := range ids {
		ids[i] = uint32(101 + i)
	}
	g.w.belt(ids...)
	if tp {
		g.w.set(func(w *world) {
			w.items = append(w.items, d2.Unit{ID: 900, Type: d2.TypeItem, Class: 5, Item: &d2.Item{Where: d2.ContInventory}})
		})
	}
	g.r.loop.broke.on = true
	s, _ := g.w.snapshot()
	return g, s
}

// KITGATELEVEL rule 1 and 4: below clvl 8 the old gate holds on an incomplete kit; from clvl 8 only a character with no
// portal and no healing potion is held, every other incomplete kit goes on.
func TestKitGateLevelHold(t *testing.T) {
	for _, tc := range []struct {
		name     string
		level    int
		hp       int
		tp       bool
		wantHold bool
	}{
		{"clvl 7 no belt, portal", 7, 0, true, true},
		{"clvl 7 two potions, no portal", 7, 2, false, true},
		{"clvl 7 complete", 7, 4, true, false},
		{"clvl 8 no belt, portal", 8, 0, true, false},
		{"clvl 8 two potions, no portal", 8, 2, false, false},
		{"clvl 8 no portal, no potions", 8, 0, false, true},
		{"clvl 14 no portal, no potions", 14, 0, false, true},
		{"clvl 14 complete", 14, 4, true, false},
	} {
		g, s := kitField(t, tc.level, tc.hp, tc.tp, 500)
		if got := g.r.kitGateActive(s); got != tc.wantHold {
			t.Errorf("%s: gate %v, want %v", tc.name, got, tc.wantHold)
		}
	}
}

// Rule 2: from clvl 8 an incomplete kit with gold makes a town trip and the area choice is not the safe tier.
func TestKitGateLevelTripThenProgress(t *testing.T) {
	for _, level := range []int{8, 11, 14} {
		g, s := kitField(t, level, 1, true, 500)
		plan, _, ok := g.r.choresDue(s)
		if !ok || !plan.Potions || !strings.Contains(plan.Reason, "kit: healing potions 1/4") {
			t.Fatalf("clvl %d: want a kit trip, got ok=%v %+v", level, ok, plan)
		}
		if g.r.kitGateActive(s) {
			t.Errorf("clvl %d: the gate must not hold", level)
		}
		if a, ok := g.r.recoveryArea(s); ok && kitSafeTier[a] {
			t.Errorf("clvl %d: picked %d, the Blood Moor / Cold Plains tier", level, a)
		}
	}
}

// Rule 3 and 4: gold short with a portal source keeps going (no trip, no hold); gold short without portal and potions holds.
func TestKitGateLevelGoldShort(t *testing.T) {
	g, s := kitField(t, 11, 1, true, 10)
	if plan, _, ok := g.r.choresDue(s); ok && strings.Contains(plan.Reason, "kit:") {
		t.Errorf("no gold: no kit trip, got %+v", plan)
	}
	if g.r.kitGateActive(s) {
		t.Error("a portal source and a potion: no hold")
	}
	g, s = kitField(t, 11, 0, false, 10)
	if !g.r.kitGateActive(s) {
		t.Error("no portal and no potions: hold")
	}
}

// Rule 5: after KitTripCap kit trips with the kit still incomplete the trips stop and kitgate-skip is logged once.
func TestKitGateLevelCapLogsSkip(t *testing.T) {
	var buf bytes.Buffer
	g, s := kitField(t, 11, 1, true, 500)
	g.e.Log = slog.New(slog.NewTextHandler(&buf, nil))
	if _, ok := g.r.kitTrip(s); !ok {
		t.Fatal("first trip wanted")
	}
	for i := 0; i < 2; i++ {
		g.r.logTrip(&trip{plan: ChorePlan{Reason: "[kit: healing potions 1/4]"}}, 400)
	}
	for i := 0; i < 3; i++ {
		if _, ok := g.r.kitTrip(s); ok {
			t.Fatal("cap reached: proceed")
		}
	}
	if n := strings.Count(buf.String(), "kitgate-skip"); n != 1 {
		t.Fatalf("want one kitgate-skip line, got %d in %q", n, buf.String())
	}
	for _, want := range []string{"trips=2", "gold=500", "healing potions 1/4"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("log line lacks %s: %s", want, buf.String())
		}
	}
	// a complete kit resets the count
	g.w.belt(101, 102, 103, 104)
	s2, _ := g.w.snapshot()
	g.r.kitTrip(s2)
	if g.r.ob.kitTrips != 0 || g.r.ob.kitSkipped {
		t.Error("complete kit must reset the trip count")
	}
}

// A complete kit never makes a trip or holds, at any level.
func TestKitGateLevelCompleteKit(t *testing.T) {
	for _, level := range []int{1, 7, 8, 11, 14} {
		g, s := kitField(t, level, 4, true, 500)
		if r, ok := g.r.kitTrip(s); ok {
			t.Errorf("clvl %d: complete kit trip %q", level, r)
		}
		if g.r.kitGateActive(s) {
			t.Errorf("clvl %d: complete kit held", level)
		}
	}
}

// The false empty-tome pattern (stat 70 drops out of the dump): the debounced verdict still says "has a portal", so no
// trip is made; a tome that stays empty for the whole span does make one.
func TestKitGateLevelTomeDropout(t *testing.T) {
	g, s := kitField(t, 11, 4, true, 500)
	if _, ok := g.r.kitTrip(s); ok {
		t.Fatal("complete kit")
	}
	g.w.set(func(w *world) { w.items = w.items[:len(w.items)-1]; w.area = 2 }) // the tome vanishes from the read; Blood Moor, walkable home
	s2, _ := g.w.snapshot()
	if r, ok := g.r.kitTrip(s2); ok {
		t.Fatalf("drop-out inside the debounce span made a trip: %q", r)
	}
	g.clock.Sleep(tpDebounceSpan + time.Second)
	if r, ok := g.r.kitTrip(s2); !ok || !strings.Contains(r, "town portal") {
		t.Fatalf("a really empty tome must make a trip, got %q %v", r, ok)
	}
}

func TestKitMissingList(t *testing.T) {
	k := kit{TP: true, HP: 2, MP: 1, minHP: 4, minMP: 3}
	if m := strings.Join(k.missing(), ";"); m != "healing potions 2/4;mana potions 1/3" {
		t.Fatalf("got %q", m)
	}
	if len((kit{TP: true, HP: 4, MP: 0, minHP: 4}).missing()) != 0 {
		t.Fatal("a build without mana need is complete")
	}
}
