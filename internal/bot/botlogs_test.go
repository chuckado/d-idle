package bot

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
)

// Countable guard lines (docs/botlogs.md): each event writes exactly one line, nothing else does.
func logRig(t *testing.T) (*rig, *bytes.Buffer) {
	t.Helper()
	g := newRig(t)
	var buf bytes.Buffer
	g.e.Log = slog.New(slog.NewTextHandler(&buf, nil))
	return g, &buf
}

func count(buf *bytes.Buffer, msg string) int { return strings.Count(buf.String(), "msg="+msg+" ") }

func TestExploreGiveupLine(t *testing.T) {
	g, buf := logRig(t)
	ctx := context.Background()
	fail := errors.New("nav: timeout")
	g.r.exploreGaveUp(ctx, 2, fail)
	if n := count(buf, "explore-giveup"); n != 0 {
		t.Fatalf("one failure: %d lines", n)
	}
	g.r.exploreGaveUp(ctx, 2, fail)
	if n := count(buf, "explore-giveup"); n != 1 {
		t.Fatalf("two failures: %d lines", n)
	}
	for _, w := range []string{"area=2", "failures=2", "kills=0"} {
		if !strings.Contains(buf.String(), w) {
			t.Errorf("missing %s in %s", w, buf)
		}
	}
	g.r.exploreGaveUp(ctx, 2, fail) // the count restarted
	if n := count(buf, "explore-giveup"); n != 1 {
		t.Fatalf("third failure alone: %d lines", n)
	}
	g.r.loop.run = &LevelRun{Area: 3, Kind: RunQuest}
	g.r.exploreGaveUp(ctx, 3, fail)
	g.r.exploreGaveUp(ctx, 3, fail)
	if n := count(buf, "explore-giveup"); n != 1 {
		t.Fatalf("a quest run keeps roaming: %d lines", n)
	}
}

func TestTPDebounceLines(t *testing.T) {
	var d tpDebounce
	t0 := time.Unix(0, 0)
	look := func(raw bool, s int) tpEvent { _, ev := d.look(raw, t0.Add(time.Duration(s)*time.Second)); return ev }
	if ev := look(true, 0); ev.kind != "" {
		t.Fatalf("first look: %+v", ev)
	}
	for _, s := range []int{2, 4, 6} {
		if ev := look(false, s); ev.kind != "" {
			t.Fatalf("inside a run: %+v", ev)
		}
	}
	if ev := look(true, 8); ev != (tpEvent{"hold", 3, 6 * time.Second}) {
		t.Fatalf("hold: %+v", ev)
	}
	if ev := look(true, 10); ev.kind != "" {
		t.Fatalf("steady true: %+v", ev)
	}
	look(false, 20)
	if ev := look(false, 20+int(tpDebounceSpan/time.Second)); ev != (tpEvent{"flip", 2, tpDebounceSpan}) {
		t.Fatalf("flip: %+v", ev)
	}
	if ev := look(false, 400); ev.kind != "" {
		t.Fatalf("steady false: %+v", ev)
	}
}

// Through hasPortalSource on a mem.Fake world: a dropped-out tome that returns is one hold line, one that stays gone is one flip line.
func TestTPDebounceLinesLogged(t *testing.T) {
	g, buf := logRig(t)
	g.w.set(func(w *world) {
		w.area = 2
		w.items = append(w.items, d2.Unit{ID: 900, Type: d2.TypeItem, Class: 5, Item: &d2.Item{Where: d2.ContInventory}})
	})
	s, _ := g.w.snapshot()
	if !g.r.hasPortalSource(s) {
		t.Fatal("portal source expected")
	}
	g.w.set(func(w *world) { w.items = w.items[:len(w.items)-1] })
	s0, _ := g.w.snapshot()
	g.clock.Sleep(5 * time.Second)
	if !g.r.hasPortalSource(s0) {
		t.Fatal("held")
	}
	g.clock.Sleep(5 * time.Second)
	if !g.r.hasPortalSource(s) {
		t.Fatal("back")
	}
	if count(buf, "tp-debounce-hold") != 1 || count(buf, "tp-debounce-flip") != 0 {
		t.Fatalf("hold run: %s", buf)
	}
	for _, w := range []string{"verdict=true", "raw=false", "looks=1", "held_s=5"} {
		if !strings.Contains(buf.String(), w) {
			t.Errorf("missing %s in %s", w, buf)
		}
	}
	g.r.hasPortalSource(s0)
	g.clock.Sleep(tpDebounceSpan + time.Second)
	if g.r.hasPortalSource(s0) {
		t.Fatal("flip expected")
	}
	g.r.hasPortalSource(s0)
	if count(buf, "tp-debounce-hold") != 1 || count(buf, "tp-debounce-flip") != 1 {
		t.Fatalf("flip run: %s", buf)
	}
}

func TestKitgateTripLine(t *testing.T) {
	g, buf := logRig(t)
	g.r.logTrip(&trip{plan: ChorePlan{Reason: "[free heal]"}}, 400)
	if count(buf, "kitgate-trip") != 0 {
		t.Fatalf("non-kit trip logged: %s", buf)
	}
	g.r.obMu.Lock()
	g.r.ob.level = 11
	g.r.obMu.Unlock()
	g.r.logTrip(&trip{plan: ChorePlan{Reason: "[kit: healing potions 1/4]"}}, 400)
	g.r.logTrip(&trip{plan: ChorePlan{Reason: "[kit: town portal]"}}, 300)
	if n := count(buf, "kitgate-trip"); n != 2 {
		t.Fatalf("want 2 lines, got %d: %s", n, buf)
	}
	for _, w := range []string{"level=11", `missing="healing potions 1/4"`, "gold=400", "trip=1", "trip=2", `missing="town portal"`} {
		if !strings.Contains(buf.String(), w) {
			t.Errorf("missing %s in %s", w, buf)
		}
	}
}

// kitgate-skip fires once when the cap is reached from KitGateLevel up (the gate is bypassed there), and never below it.
func TestKitgateSkipCondition(t *testing.T) {
	for _, tc := range []struct {
		level, want int
	}{{7, 0}, {8, 1}, {14, 1}} {
		var buf bytes.Buffer
		g, s := kitField(t, tc.level, 1, true, 500)
		g.e.Log = slog.New(slog.NewTextHandler(&buf, nil))
		g.r.obMu.Lock()
		g.r.ob.kitTrips = g.e.Tune.Resource.KitTripCap
		g.r.obMu.Unlock()
		for i := 0; i < 3; i++ {
			g.r.kitTrip(s)
		}
		if n := count(&buf, "kitgate-skip"); n != tc.want {
			t.Errorf("clvl %d: %d skip lines, want %d", tc.level, n, tc.want)
		}
	}
}
