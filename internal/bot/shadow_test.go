package bot

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestShadowOffByDefault(t *testing.T) {
	g := newRig(t)
	if g.r.e.Tune.Bot.ArbiterShadow != 0 {
		t.Fatal("bot.arbiterShadow must default to 0")
	}
	g.r.poll()
	g.r.runGoalShadowed(context.Background(), g.r.latest(), levelGoalControl())
	if _, ok := g.r.Arbiter(); ok {
		t.Fatal("the shadow ran with the knob off")
	}
	if kv := g.r.arb.minuteKV(); kv != nil {
		t.Fatalf("minute line carries %v with the knob off", kv)
	}
}

func TestShadowOnLogsAndKeepsTheChain(t *testing.T) {
	g := newRig(t)
	g.r.e.Tune.Bot.ArbiterShadow = 1
	g.w.set(func(w *world) { w.area = 5 })
	g.r.poll()
	for i := 0; i < 5; i++ {
		g.clock.Sleep(time.Second)
		g.r.runGoalShadowed(context.Background(), g.r.latest(), levelGoalControl())
	}
	if _, ok := g.r.Arbiter(); !ok {
		t.Fatal("no shadow decision with the knob on")
	}
	if kv := g.r.arb.minuteKV(); len(kv) != 2 || !strings.Contains(kv[1].(string), "panics=0") {
		t.Fatalf("minute kv %v", kv)
	}
}

func TestChainFamily(t *testing.T) {
	for code, want := range map[string]string{WhyQuestStep: "quest", WhyTownTrip: "chores", WhyClaimTrip: "claim", "": "", "bogus": ""} {
		if got := chainFamily(code); got != want {
			t.Errorf("%q: %q want %q", code, got, want)
		}
	}
}

func TestRefusalBecomesArbiterTempIntent(t *testing.T) {
	for _, tc := range []struct{ need, want string }{{"needs_potions", "resupply/refill"}, {"too_weak", "level/level_up"}} {
		g := newRig(t)
		g.w.set(func(w *world) { w.area = 5 })
		g.r.poll()
		g.r.arbRefused(tc.need, "w") // knob off: nothing happens
		if _, ok := g.r.Arbiter(); ok {
			t.Fatal("the arbiter ran with the knob off")
		}
		g.r.e.Tune.Bot.ArbiterShadow, g.r.e.Tune.Bot.ResupplyPotions = 1, 100
		g.r.arbRefused(tc.need, "w")
		g.clock.Sleep(time.Second)
		g.r.runGoalShadowed(context.Background(), g.r.latest(), levelGoalControl())
		d, ok := g.r.Arbiter()
		if !ok || d.Chosen == nil && len(d.Rejected) == 0 {
			t.Fatalf("%s: no decision", tc.need)
		}
		found := d.Chosen != nil && d.Chosen.Key() == tc.want
		for _, r := range d.Rejected {
			found = found || r.Intent.Key() == tc.want
		}
		if !found {
			t.Fatalf("%s: temporary intent %s not proposed: %s", tc.need, tc.want, d)
		}
	}
}
