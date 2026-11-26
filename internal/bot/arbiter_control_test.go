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
