package bot

import (
	"errors"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/goals"
)

// The ring keeps decisions, not ticks; refusals stand with their need and age out; the answer is the same for two classes.
func TestExplainRing(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		g := newRig(t)
		g.w.set(func(w *world) { w.class = class })
		s, _ := g.w.snapshot()
		r := g.r
		put := func(code string, kv ...string) {
			r.setWhy(code, kv...)
			r.mu.Lock()
			r.why, r.pendingWhy = r.pendingWhy, goals.Why{}
			r.mu.Unlock()
			r.explainTick(s, true)
		}
		put(WhyQuestStep, "step", "a")
		put(WhyQuestStep, "step", "a")
		put(WhyTownTrip, "why", "belt 1 under 4")
		put(WhyQuestStep, "step", "a")
		v := r.Explain()
		if len(v.Decisions) != 3 || v.Decisions[0].Ticks != 2 || v.Decisions[1].Family != "chores" {
			t.Fatalf("%v: ring %+v", class, v.Decisions)
		}
		r.explainTick(s, false)
		for i := 0; i < 300; i++ {
			put(WhyWalk, "area", string(rune('a'+i%26)), "level", itoa(i))
		}
		if n := len(r.Explain().Decisions); n != explainRing {
			t.Fatalf("%v: ring size %d, want %d", class, n, explainRing)
		}
		r.noteStanding("walk", errors.New("too hard").Error())
		r.noteStanding("walk", "too hard")
		v = r.Explain()
		if len(v.Refusals) != 1 || v.Refusals[0].Count != 2 || v.Refusals[0].Need != "too hard" {
			t.Fatalf("%v: refusals %+v", class, v.Refusals)
		}
		g.clock.Sleep(explainRefLife + time.Second)
		if len(r.Explain().Refusals) != 0 {
			t.Fatalf("%v: refusal did not age out", class)
		}
	}
}
