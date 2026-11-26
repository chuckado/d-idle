package bot

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/task"
)

func refusedRes(code, why string) task.Result {
	return task.Result{Outcome: task.OutRetry, Kind: task.KindDanger, Code: code, Err: errors.New("danger: " + why)}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		name string
		res  task.Result
		kind OutcomeKind
		need string
	}{
		{"done", task.Result{Outcome: task.OutDone}, OutDoneKind, ""},
		{"ran", task.Result{Outcome: task.OutRan}, OutProgressKind, ""},
		{"coded refusal", refusedRes("needs_gold_farm", "not ready: x"), OutRefusedKind, "needs_gold_farm"},
		{"early road, no code", refusedRes("", "not ready: early road (belt 0)"), OutRefusedKind, "route_closed"},
		{"too weak, no code", refusedRes("", "not ready: clvl 5 < 9"), OutRefusedKind, "too_weak"},
		{"blocked not ready, other kind", task.Result{Outcome: task.OutBlocked, Kind: task.KindResource, Err: errors.New("not ready: kit")}, OutRefusedKind, "needs_potions"},
		{"nav fault", task.Result{Outcome: task.OutRetry, Kind: task.KindNotFound, Err: errors.New("no path")}, OutFailedKind, ""},
		{"stuck pause", task.Result{Outcome: task.OutBlocked, Kind: task.KindDanger, Stuck: true, Err: errors.New("not ready")}, OutFailedKind, ""},
		{"interrupted", task.Result{Outcome: task.OutInterrupted}, OutInterruptedKind, ""},
	}
	for _, c := range cases {
		o := classify(c.res)
		if o.Kind != c.kind || o.Need != c.need {
			t.Errorf("%s: got %s/%q want %s/%q", c.name, o.Kind, o.Need, c.kind, c.need)
		}
		if o.Kind == OutRefusedKind && (o.Wake == "" || o.asError() == nil) {
			t.Errorf("%s: a refusal names its wake and travels as an error", c.name)
		}
	}
	for need := range needTable {
		if needTable[need].Wake == "" || needTable[need].PauseSec <= 0 {
			t.Errorf("need %s has no wake or pause", need)
		}
	}
}

// 226 refusals in two hours became one line and no failure count: the same refusal 50 times, then a change of the facts.
func TestRefusedHandledOnceNoStrike(t *testing.T) {
	for _, level := range []int{5, 40} { // two heroes: a clvl 5 starter and a clvl 40 hero with other kit numbers
		g := newRig(t)
		buf := &lockedBuf{}
		g.r.e.Log = slog.New(slog.NewTextHandler(buf, nil))
		g.r.obMu.Lock()
		g.r.ob.lastSnap = &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Area: 2, Level: level, MaxLife: 100, Life: 100, Mode: 1, Gold: 10}}
		g.r.obMu.Unlock()
		ref := classify(refusedRes("", "not ready: early road (belt 0)")).asError()
		c := control{}
		for i := 0; i < 50; i++ {
			ok := g.r.do(context.Background(), c, StateTraveling, "x", 0, "quest board step", func(context.Context) error { return ref })
			if ok {
				t.Fatal("a refusal is not success")
			}
			if g.r.loop.failures != 0 {
				t.Fatalf("level %d: refusal counted as a failure (%d)", level, g.r.loop.failures)
			}
		}
		out := buf.String()
		if n := strings.Count(out, "step refused"); n != 1 || strings.Contains(out, "action failed") {
			t.Fatalf("level %d: %d refusal lines, log:\n%s", level, n, out)
		}
		if !strings.Contains(out, "need=route_closed") || !strings.Contains(out, "wake=") {
			t.Fatalf("line lacks need or wake: %s", out)
		}
		// a change of the facts (the hero levels up) logs again
		g.r.obMu.Lock()
		g.r.ob.lastSnap.Player.Level++
		g.r.obMu.Unlock()
		g.r.do(context.Background(), c, StateTraveling, "x", 0, "quest board step", func(context.Context) error { return ref })
		if n := strings.Count(buf.String(), "step refused"); n != 2 {
			t.Fatalf("level %d: want a second line after the wake fact changed, got %d", level, n)
		}
	}
}

// The legacy error path stays: the same words as a plain error still count (knob bot.typedOutcome 0 returns them).
func TestLegacyRefusalStillCounts(t *testing.T) {
	g := newRig(t)
	g.r.do(context.Background(), control{}, StateTraveling, "x", 0, "quest board step", func(context.Context) error {
		return errors.New("danger: not ready: early road (belt 0)")
	})
	if g.r.loop.failures != 1 {
		t.Fatalf("legacy path failures = %d", g.r.loop.failures)
	}
}
