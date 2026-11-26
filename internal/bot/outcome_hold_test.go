package bot

import (
	"log/slog"
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
)

// A kit trip held back for gold, a spent trip cap and the town chores' skipped and blocked lines are typed holds: one
// refusal line per change of words or facts, never one per tick; knob off: none. Two classes.
func TestKitGateAndChoreHoldsAreTyped(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Barbarian} {
		g, s := kitField(t, 14, 0, true, 5) // clvl 14, no healing potion, a scroll, 5 gold: held for gold
		g.w.set(func(w *world) { w.class = class })
		s, _ = g.w.snapshot()
		buf := &lockedBuf{}
		g.r.e.Log = slog.New(slog.NewTextHandler(buf, nil))
		for i := 0; i < 30; i++ {
			if _, ok := g.r.kitTrip(s); ok {
				t.Fatalf("%v: a trip with no gold", class)
			}
		}
		out := buf.String()
		if strings.Count(out, "need=needs_gold_farm") != 1 || !strings.Contains(out, "wake=") {
			t.Fatalf("%v: log:\n%s", class, out)
		}
		g.r.choreHolds(ChoreReport{Skipped: []string{"stash: no room"}, Blocked: []string{"weapon"}})
		g.r.choreHolds(ChoreReport{Skipped: []string{"stash: no room"}, Blocked: []string{"weapon"}})
		out = buf.String()
		if strings.Count(out, "need=too_weak") != 1 || strings.Count(out, "step refused") != 3 {
			t.Fatalf("%v: chore holds:\n%s", class, out)
		}
		g.r.e.Tune.Bot.TypedOutcome = 0
		before := strings.Count(buf.String(), "step refused")
		g.r.holdRefused(holdNeed("too_weak", "other words"))
		if strings.Count(buf.String(), "step refused") != before {
			t.Fatal("knob off must stay silent")
		}
	}
}
