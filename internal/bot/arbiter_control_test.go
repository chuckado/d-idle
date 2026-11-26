package bot

import (
	"context"
	"testing"
	"time"

	"d2lod-bot/internal/arbiter"
	"d2lod-bot/internal/d2"
)

func controlRig(t *testing.T, class d2.Class, ctl, mask int) *rig {
	g := newRig(t)
	g.w.set(func(w *world) {
		w.area, w.life, w.class = 5, 30, class
		w.mons = []d2.Unit{{ID: 9, Type: d2.TypeMonster, Class: 10, Mode: d2.MonModeStand, X: 103, Y: 100}}
	})
	g.r.e.Tune.Bot.ArbiterShadow, g.r.e.Tune.Bot.ArbiterControl, g.r.e.Tune.Bot.ArbiterFamilies = 1, ctl, mask
	g.r.poll()
	g.clock.Sleep(time.Second)
	return g
}

// Survive under control: at low life with hostiles near the arbiter's reflex family acts (fights where it stands) and
// the tick counts as controlled. Two classes. The off cases are the must-fail half: knob 0, or the family bit not set,
// leave the chain alone.
func TestArbiterControlSurvive(t *testing.T) {
	for _, class := range []d2.Class{0, 1} {
		g := controlRig(t, class, 1, 3)
		g.r.runGoalShadowed(context.Background(), g.r.latest(), levelGoalControl())
		d, ok := g.r.Arbiter()
		if !ok || d.Chosen == nil || d.Chosen.Source != arbiter.SrcSurvive {
			t.Fatalf("class %d: the arbiter should choose survive: %v", class, d)
		}
		if n := g.r.arb.ctl.Load(); n != 1 {
			t.Fatalf("class %d: controlled ticks %d, want 1", class, n)
		}
		for name, mask := range map[string]int{"knob off": 3, "bit not set": 2} {
			ctl := 1
			if name == "knob off" {
				ctl = 0
			}
			h := controlRig(t, class, ctl, mask)
			h.r.runGoalShadowed(context.Background(), h.r.latest(), levelGoalControl())
			if n := h.r.arb.ctl.Load(); n != 0 {
				t.Fatalf("class %d, %s: the arbiter acted (%d)", class, name, n)
			}
		}
	}
}

func TestArbiterControlOffByDefault(t *testing.T) {
	g := newRig(t)
	if g.r.e.Tune.Bot.ArbiterControl != 0 {
		t.Fatal("bot.arbiterControl must default to 0")
	}
}

func fieldRig(t *testing.T, class d2.Class, ctl, mask int) *rig {
	g := newRig(t)
	g.w.set(func(w *world) { w.area, w.class, w.level = 5, class, 6 })
	g.w.belt()
	g.r.e.Tune.Bot.ArbiterShadow, g.r.e.Tune.Bot.ArbiterControl, g.r.e.Tune.Bot.ArbiterFamilies = 1, ctl, mask
	g.r.poll()
	g.clock.Sleep(time.Second)
	return g
}

// Chores under control: an empty belt in a quiet field, the arbiter picks the refill and the chores hook starts the trip
// (family bit 4 = chores). Two classes; knob off and bit not set are the must-fail half.
func TestArbiterControlChores(t *testing.T) {
	for _, class := range []d2.Class{0, 1} {
		g := fieldRig(t, class, 1, 7)
		g.r.runGoalShadowed(context.Background(), g.r.latest(), levelGoalControl())
		d, ok := g.r.Arbiter()
		if !ok || d.Chosen == nil || arbiter.Family(d.Chosen.Source) != "chores" {
			t.Fatalf("class %d: the arbiter should choose chores: %v", class, d)
		}
		if n := g.r.arb.ctl.Load(); n != 1 || g.r.loop.trip == nil && g.w.count("totown") == 0 {
			t.Fatalf("class %d: controlled ticks %d, no trip started", class, n)
		}
		for name, c := range map[string][2]int{"knob off": {0, 7}, "bit not set": {1, 3}} {
			h := fieldRig(t, class, c[0], c[1])
			h.r.runGoalShadowed(context.Background(), h.r.latest(), levelGoalControl())
			if n := h.r.arb.ctl.Load(); n != 0 {
				t.Fatalf("class %d, %s: the arbiter acted (%d)", class, name, n)
			}
		}
	}
}

// Claim under control: a pending reward in the field, the arbiter picks the claim and the claim hook goes to town.
func TestArbiterControlClaim(t *testing.T) {
	for _, class := range []d2.Class{0, 1} {
		for _, mask := range []int{3 | 1<<3, 3} {
			g := fieldRig(t, class, 1, mask)
			g.w.belt(101, 102, 103, 104)
			g.r.poll()
			g.r.d.Quester = fakeQuester{w: g.w}
			g.r.loop.questPending = map[string]bool{"den_of_evil": true}
			g.r.runGoalShadowed(context.Background(), g.r.latest(), levelGoalControl())
			want := int64(0)
			if mask&(1<<3) != 0 {
				want = 1
			}
			d, _ := g.r.Arbiter()
			if n := g.r.arb.ctl.Load(); n != want {
				t.Logf("decision %+v", d.Chosen)
				t.Fatalf("class %d mask %d: controlled ticks %d, want %d", class, mask, n, want)
			}
		}
	}
}
