package bot

import (
	"d2lod-bot/internal/d2"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"
)

// Two heroes with different numbers: the rows take their lines from the facts, so nothing names a class.
type invHero struct {
	name                       string
	maxLife, drinkPct, chicken int
	kitMissing                 []string
}

var invHeroes = []invHero{
	{"sorc", 176, 55, 25, []string{"town portal", "mana potions 0/4"}},
	{"pala", 520, 45, 30, []string{"healing potions 0/3"}},
}

func (h invHero) facts(now time.Time) invFacts {
	return invFacts{now: now, area: 3, life: h.maxLife, maxLife: h.maxLife, lifePct: 100, xp: 1000, kills: 10, gold: 5000,
		beltHP: 4, tp: 2, state: string(StateClearing), step: "clear", drinkPct: h.drinkPct, chickenPct: h.chicken}
}

// run feeds facts once a second for n seconds, mutating them with edit, and returns the names that fired.
func runInv(m *invMon, f invFacts, n int, edit func(*invFacts, int)) map[string]int {
	got := map[string]int{}
	for i := 0; i < n; i++ {
		f.now = f.now.Add(time.Second)
		if edit != nil {
			edit(&f, i)
		}
		for _, h := range m.tick(f) {
			got[h.Name]++
		}
	}
	return got
}

var t0 = time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)

func TestInvariantTableComplete(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range invRows {
		if seen[r.Name] || r.Why == "" {
			t.Fatalf("row %q duplicated or without a why", r.Name)
		}
		seen[r.Name] = true
	}
	if len(invRows) < 12 {
		t.Fatalf("%d rows", len(invRows))
	}
}

func TestInvariantHealthyHeroFiresNothing(t *testing.T) {
	for _, h := range invHeroes {
		m := &invMon{}
		if got := runInv(m, h.facts(t0), 3000, func(f *invFacts, i int) { f.xp += uint32(i); f.step = fmt.Sprint("clear ", i/100) }); len(got) != 0 {
			t.Errorf("%s: healthy hero fired %v", h.name, got)
		}
	}
}

func TestInvariantMustFail(t *testing.T) {
	cases := []struct {
		name string
		n    int
		edit func(h invHero) func(*invFacts, int)
		want string
	}{
		{"low life, potions, none drunk", 20, func(h invHero) func(*invFacts, int) {
			return func(f *invFacts, _ int) { f.life, f.lifePct = h.maxLife/5, 20 }
		}, "potion-not-drunk"},
		{"chicken line, scrolls, hostiles", 20, func(h invHero) func(*invFacts, int) {
			return func(f *invFacts, _ int) { f.lifePct, f.hostiles, f.life = h.chicken-1, 3, 1 }
		}, "portal-not-cast"},
		{"stuck in a pack at half life", 20, func(h invHero) func(*invFacts, int) {
			return func(f *invFacts, _ int) { f.lifePct, f.life, f.hostiles = 50, h.maxLife/2, 7 }
		}, "stuck-in-pack"},
		{"resupplying in a field", 30, func(h invHero) func(*invFacts, int) {
			return func(f *invFacts, _ int) { f.state, f.step = string(StateResupplying), "chores" }
		}, "resupply-in-field"},
		{"life zero while alive", 10, func(h invHero) func(*invFacts, int) {
			return func(f *invFacts, _ int) { f.life, f.lifePct = 0, 0; f.inTown = true }
		}, "life-zero-alive"},
		{"xp frozen in the field", 2000, func(h invHero) func(*invFacts, int) { return nil }, "xp-frozen"},
		{"mode without exit", 1500, func(h invHero) func(*invFacts, int) {
			return func(f *invFacts, _ int) { f.xp = 1000; f.inTown = true }
		}, "mode-no-exit"},
		{"corpse never resolved", 5, func(h invHero) func(*invFacts, int) {
			return func(f *invFacts, _ int) { f.corpsePending, f.corpseSince = true, t0.Add(-time.Hour) }
		}, "corpse-unresolved"},
		{"poor for two hours while killing", 7300, func(h invHero) func(*invFacts, int) {
			return func(f *invFacts, i int) { f.gold, f.kills = 3, 10+i; f.xp += 1 }
		}, "economy-stall"},
	}
	for _, h := range invHeroes {
		for _, c := range cases {
			m := &invMon{}
			if got := runInv(m, h.facts(t0), c.n, c.edit(h)); got[c.want] == 0 {
				t.Errorf("%s/%s: %s did not fire (%v)", h.name, c.name, c.want, got)
			}
		}
	}
}

func TestInvariantDrinkClearsIt(t *testing.T) {
	h := invHeroes[0]
	m := &invMon{}
	got := runInv(m, h.facts(t0), 60, func(f *invFacts, i int) { f.life, f.lifePct = 40, 20; f.drank = int64(i / 3) })
	if got["potion-not-drunk"] != 0 {
		t.Fatalf("a hero that drinks every 3 s fired: %v", got)
	}
}

func TestInvariantFieldEntry(t *testing.T) {
	for _, h := range invHeroes {
		m := &invMon{}
		f := h.facts(t0)
		f.inTown = true
		runInv(m, f, 3, nil)
		f.inTown, f.lifePct, f.life = false, 30, h.maxLife*3/10
		f.kitKnown, f.kitMissing = true, h.kitMissing
		var names []string
		for _, x := range m.tick(f) {
			names = append(names, x.Name)
		}
		if strings.Join(names, ",") != "field-entry-hurt,field-entry-no-kit" {
			t.Errorf("%s: %v", h.name, names)
		}
	}
}

func TestInvariantTownThrash(t *testing.T) {
	m := &invMon{}
	f := invHeroes[1].facts(t0)
	got := runInv(m, f, 40*60, func(f *invFacts, i int) { f.inTown = i%120 < 60 })
	if got["town-thrash"] == 0 {
		t.Fatalf("20 trips in 40 min did not fire: %v", got)
	}
}

func TestInvariantRefusals(t *testing.T) {
	m := &invMon{}
	f := invHeroes[0].facts(t0)
	runInv(m, f, 2, nil)
	for i := 0; i < 10; i++ {
		m.noteRefusal()
	}
	m.noteRefusalFailure("quest board step", errors.New("not ready: early road"))
	var names []string
	f.now = f.now.Add(time.Second)
	for _, x := range m.tick(f) {
		names = append(names, x.Name)
	}
	if strings.Join(names, ",") != "refusal-as-failure,refusal-storm" {
		t.Fatalf("%v", names)
	}
	if !isRefusal(fmt.Errorf("x: %w", errAreaGated)) || !isRefusal(errors.New("not ready: kit")) || isRefusal(errors.New("walk failed")) {
		t.Fatal("isRefusal")
	}
}

func TestInvariantRateLimitAndMinute(t *testing.T) {
	m := &invMon{}
	f := invHeroes[0].facts(t0)
	got := runInv(m, f, 300, func(f *invFacts, _ int) { f.life, f.lifePct = 1, 1 })
	if n := got["potion-not-drunk"]; n < 3 || n > 6 {
		t.Fatalf("a 5 minute episode logged %d lines, want one per minute or so", n)
	}
	kv := m.minuteKV()
	if len(kv) != 2 || !strings.Contains(kv[1].(string), "potion-not-drunk=") {
		t.Fatalf("minute kv %v", kv)
	}
	if m.minuteKV() != nil || m.Totals()["potion-not-drunk"] != got["potion-not-drunk"] {
		t.Fatal("minute counts are drained, totals are not")
	}
}

func TestInvariantDeadFiresNothing(t *testing.T) {
	m := &invMon{}
	f := invHeroes[0].facts(t0)
	f.dead = true
	if got := runInv(m, f, 100, func(f *invFacts, _ int) { f.life, f.lifePct = 0, 0 }); len(got) != 0 {
		t.Fatalf("%v", got)
	}
}

// The hook: invTick builds the facts from a snapshot, logs one `invariant` line and calls OnInvariant; knob 0 is silent.
func TestInvTickLogsAndCallsOracle(t *testing.T) {
	for _, knob := range []int{1, 0} {
		g := newRig(t)
		buf := &lockedBuf{}
		g.e.Log = slog.New(slog.NewTextHandler(buf, nil))
		g.e.Tune.Bot.Invariants = knob
		var hits []string
		g.r.d.OnInvariant = func(n, _ string) { hits = append(hits, n) }
		snap := &d2.Snapshot{Player: &d2.Player{Area: 3, Life: 0, MaxLife: 100, Level: 5, Mode: 1}}
		for i := 0; i < 5; i++ {
			g.clock.Sleep(time.Second)
			g.r.invTick(snap, control{})
		}
		out := buf.String()
		if knob == 1 && (!strings.Contains(out, "invariant life-zero-alive") || len(hits) != 1 || g.r.InvariantHits()["life-zero-alive"] != 1) {
			t.Fatalf("knob 1: hits %v log %q", hits, out)
		}
		if knob == 0 && (strings.Contains(out, "invariant") || len(hits) != 0) {
			t.Fatalf("knob 0 logged: %v %q", hits, out)
		}
	}
}
