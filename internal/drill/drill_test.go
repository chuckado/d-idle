package drill

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeWorld struct {
	ui        UI
	recovers  int
	recovery  bool // Recover clears the UI
	uiErr     error
	cheats    []string
	noReasons int
}

func (f *fakeWorld) UI(context.Context) (UI, error) { return f.ui, f.uiErr }
func (f *fakeWorld) Recover(context.Context) error {
	f.recovers++
	if f.recovery {
		f.ui = UI{}
	}
	return nil
}
func (f *fakeWorld) Cheat(_ context.Context, actor, verb, reason string) (string, error) {
	f.cheats = append(f.cheats, actor+" "+verb+" "+reason)
	return "ok", nil
}

func pass(string) func(context.Context, *Run) Outcome {
	return func(context.Context, *Run) Outcome { return Passed("fine", "a", "b") }
}

func TestLineAndTable(t *testing.T) {
	w := &fakeWorld{}
	r := &Runner{W: w, Repeat: 2}
	g := r.RunGroup(context.Background(), []Drill{{ID: "D1.P1", Primitive: "P1", Do: func(ctx context.Context, run *Run) Outcome {
		if _, err := run.Cheat(ctx, "debug stat 12 5", "level for the drill"); err != nil {
			return Failed(err.Error(), "", "")
		}
		return Passed("belted", "3 hp in columns 0,1,3", "hp@0 hp@1 hp@3")
	}}})
	if len(g.Results) != 1 || g.Results[0].Verdict != Pass || !g.Results[0].Deterministic {
		t.Fatalf("%+v", g.Results)
	}
	l := g.Results[0].Line()
	for _, want := range []string{"DRILL D1.P1 P1 PASS belted", "intended: 3 hp in columns 0,1,3", "observed: hp@0 hp@1 hp@3", "2, identical"} {
		if !strings.Contains(l, want) {
			t.Errorf("line %q lacks %q", l, want)
		}
	}
	if len(w.cheats) != 2 || !strings.HasPrefix(w.cheats[0], "drills:D1.P1 debug stat 12 5") {
		t.Errorf("cheats %v", w.cheats)
	}
	if !strings.Contains(g.Table(), "cheat verbs used") {
		t.Error(g.Table())
	}
}

func TestCheatNeedsReason(t *testing.T) {
	run := &Run{W: &fakeWorld{}, Drill: "x"}
	if _, err := run.Cheat(context.Background(), "debug kill", "  "); err == nil {
		t.Fatal("cheat without a reason must be refused")
	}
}

func TestBlockedNamesThePiece(t *testing.T) {
	r := &Runner{W: &fakeWorld{}}
	g := r.RunGroup(context.Background(), []Drill{{ID: "D1.P3m", Primitive: "P3", Needs: []string{"kill merc cheat"}, Do: pass("")}})
	if g.Results[0].Verdict != Blocked || !strings.Contains(g.Results[0].Reason, "kill merc cheat") {
		t.Fatalf("%+v", g.Results[0])
	}
}

func TestNondeterminismFails(t *testing.T) {
	n := 0
	r := &Runner{W: &fakeWorld{}, Repeat: 2}
	g := r.RunGroup(context.Background(), []Drill{{ID: "D", Primitive: "P", Do: func(context.Context, *Run) Outcome {
		n++
		return Passed("ok", "x", map[int]string{1: "A", 2: "B"}[n])
	}}})
	res := g.Results[0]
	if res.Deterministic || res.Verdict != Fail || !strings.Contains(res.Observed, "run1: A") {
		t.Fatalf("%+v", res)
	}
}

func TestOneRepeatFailsTheDrill(t *testing.T) {
	n := 0
	r := &Runner{W: &fakeWorld{}, Repeat: 2}
	g := r.RunGroup(context.Background(), []Drill{{ID: "D", Primitive: "P", Do: func(context.Context, *Run) Outcome {
		n++
		if n == 2 {
			return Failed("second run failed", "x", "y")
		}
		return Passed("ok", "x", "y")
	}}})
	if g.Results[0].Verdict != Fail {
		t.Fatalf("%+v", g.Results[0])
	}
}

func TestDirtyUIBeforeIsRecoveredOnce(t *testing.T) {
	w := &fakeWorld{ui: UI{Panels: []string{"inventory"}}, recovery: true}
	r := &Runner{W: w}
	g := r.RunGroup(context.Background(), []Drill{{ID: "D1", Primitive: "P1", Do: pass("")}})
	if w.recovers != 1 {
		t.Fatalf("recover calls %d", w.recovers)
	}
	if len(g.Results) != 2 || g.Results[0].Primitive != "P35" || g.Results[0].Verdict != Fail || g.Results[1].Verdict != Pass {
		t.Fatalf("%+v", g.Results)
	}
	if !strings.Contains(g.Results[0].Observed, "recovered") {
		t.Error(g.Results[0].Observed)
	}
}

func TestStuckCursorStopsTheGroup(t *testing.T) {
	w := &fakeWorld{}
	r := &Runner{W: w}
	ds := []Drill{
		{ID: "D1", Primitive: "P1", Do: func(context.Context, *Run) Outcome {
			w.ui = UI{Cursor: "hp1"} // the drill leaves an item on the cursor and recovery does not help
			return Passed("ok", "", "")
		}},
		{ID: "D2", Primitive: "P2", Do: pass("")},
	}
	g := r.RunGroup(context.Background(), ds)
	if g.Stopped == "" {
		t.Fatal("group must stop")
	}
	var d1 Result
	for _, x := range g.Results {
		if x.ID == "D1" {
			d1 = x
		}
	}
	if d1.Verdict != Fail || !strings.Contains(d1.Reason, "dirty UI") {
		t.Fatalf("D1 %+v", d1)
	}
	last := g.Results[len(g.Results)-1]
	if last.ID != "D2" || last.Verdict != Blocked || !strings.Contains(last.Reason, "not run") {
		t.Fatalf("D2 %+v", last)
	}
	if g.Results[0].Primitive != "P35" {
		t.Fatalf("P35 line missing: %+v", g.Results)
	}
}

func TestTimeLimit(t *testing.T) {
	r := &Runner{W: &fakeWorld{}, Grace: 50 * time.Millisecond}
	g := r.RunGroup(context.Background(), []Drill{{ID: "D", Primitive: "P", Limit: 30 * time.Millisecond, Do: func(ctx context.Context, _ *Run) Outcome {
		time.Sleep(time.Second) // ignores its context
		return Passed("late", "", "")
	}}})
	if g.Results[0].Verdict != Fail || !strings.Contains(g.Results[0].Reason, "time limit") {
		t.Fatalf("%+v", g.Results[0])
	}
}

func TestTimeLimitHonouredContext(t *testing.T) {
	r := &Runner{W: &fakeWorld{}}
	g := r.RunGroup(context.Background(), []Drill{{ID: "D", Primitive: "P", Limit: 20 * time.Millisecond, Do: func(ctx context.Context, _ *Run) Outcome {
		<-ctx.Done()
		return Passed("late", "", "")
	}}})
	if g.Results[0].Verdict != Fail {
		t.Fatalf("a pass after the deadline must fail: %+v", g.Results[0])
	}
}

func TestPanicIsFail(t *testing.T) {
	r := &Runner{W: &fakeWorld{}}
	g := r.RunGroup(context.Background(), []Drill{{ID: "D", Primitive: "P", Do: func(context.Context, *Run) Outcome { panic("boom") }}})
	if g.Results[0].Verdict != Fail || !strings.Contains(g.Results[0].Reason, "boom") {
		t.Fatalf("%+v", g.Results[0])
	}
}

func TestUnreadableUIBlocks(t *testing.T) {
	r := &Runner{W: &fakeWorld{uiErr: errors.New("no socket")}}
	g := r.RunGroup(context.Background(), []Drill{{ID: "D", Primitive: "P", Do: pass("")}})
	if g.Results[0].Verdict != Blocked || g.Stopped == "" {
		t.Fatalf("%+v", g)
	}
}

func TestWriteFiles(t *testing.T) {
	dir := t.TempDir()
	g := Group{Results: []Result{{ID: "D", Primitive: "P1", Verdict: Pass, Reason: "r", Deterministic: true}}}
	if err := g.Write(dir, "2026-10-31-d1"); err != nil {
		t.Fatal(err)
	}
}
