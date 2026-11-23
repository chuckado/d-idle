package questboard

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"
)

var smT0 = time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)

func smBoard(t *testing.T) *Board {
	t.Helper()
	b, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// chains returns every quest of the shipped board with at least one `after` link: the fixtures stand in for builds,
// the machine knows no class (build-flexibility rule), so the tests run over every multi-part quest of Acts 1 to 5.
func chains(b *Board) []Quest {
	var out []Quest
	for _, q := range b.Quests() {
		for _, c := range q.Chunks {
			if len(c.After) > 0 {
				out = append(out, q)
				break
			}
		}
	}
	return out
}

func codeFor(ev Event) string {
	switch ev {
	case EvRefuse:
		return "needs_potions"
	case EvCap:
		return "gave_up"
	}
	return ""
}

// I8: the table is total, and a live target always carries a code of the closed list (I5).
func TestStepTotal(t *testing.T) {
	for _, from := range RunStatuses {
		for _, ev := range Events {
			tr, err := Step(from, ev, codeFor(ev))
			if err != nil {
				t.Fatalf("%s %s: %v", from, ev, err)
			}
			if !from.Live() && !tr.Noop {
				t.Fatalf("%s moved on %s", from, ev)
			}
			if !tr.Noop && tr.To.Live() && tr.To != RunRunning {
				if _, ok := ReasonOf(tr.Code); !ok {
					t.Fatalf("%s %s -> %s without a code (%q)", from, ev, tr.To, tr.Code)
				}
			}
			if tr.Noop && tr.To != from {
				t.Fatalf("no-op changed status %s %s", from, ev)
			}
		}
	}
}

func TestStepNeedsRightCode(t *testing.T) {
	if _, err := Step(RunRunning, EvRefuse, "gave_up"); !errors.Is(err, ErrReason) {
		t.Fatalf("a park code is not a hold: %v", err)
	}
	if _, err := Step(RunRunning, EvRefuse, ""); !errors.Is(err, ErrReason) {
		t.Fatalf("no code: %v", err)
	}
	if _, err := Step(RunRunning, EvCap, "needs_potions"); !errors.Is(err, ErrReason) {
		t.Fatalf("a hold is not a park: %v", err)
	}
	// bug 4: a refusal for a fixable reason is a hold, never a counted failure or a park
	tr, _ := Step(RunRunning, EvRefuse, "needs_potions")
	if tr.To != RunHeld {
		t.Fatalf("refuse -> %s", tr.To)
	}
}

func TestStepPaths(t *testing.T) {
	walk := func(from RunStatus, evs ...Event) RunStatus {
		for _, ev := range evs {
			tr, err := Step(from, ev, codeFor(ev))
			if err != nil || tr.Noop {
				t.Fatalf("%s on %s: %+v %v", ev, from, tr, err)
			}
			from = tr.To
		}
		return from
	}
	if got := walk(RunQueued, EvPromote, EvObjective, EvHandInTap, EvHandInDone); got != RunFinished {
		t.Fatalf("hand-in path ends %s", got)
	}
	if got := walk(RunQueued, EvPromote, EvRefuse, EvClear); got != RunRunning {
		t.Fatalf("hold path ends %s", got)
	}
	if got := walk(RunQueued, EvPromote, EvCap, EvRetry); got != RunQueued {
		t.Fatalf("park path ends %s", got)
	}
	if got := walk(RunHandingIn, EvRestart); got != RunAwaiting {
		t.Fatalf("restart of a dialogue ends %s", got)
	}
}

func TestWire(t *testing.T) {
	want := map[RunStatus]State{RunQueued: Accepted, RunRunning: InProgress, RunHeld: InProgress, RunAwaiting: Awaiting, RunHandingIn: Awaiting, RunParked: Blocked, RunFinished: Done, RunStopped: ""}
	for s, w := range want {
		if s.Wire() != w {
			t.Fatalf("%s wires to %q", s, s.Wire())
		}
	}
}

func TestReasonTable(t *testing.T) {
	for _, c := range ReasonCodes() {
		r, _ := ReasonOf(c)
		if r.Text == "" || r.Remedy == "" {
			t.Fatalf("%s: incomplete row", c)
		}
		if r.Kind == KindHold && r.Remedy == RemedyPlayer {
			t.Fatalf("%s: a hold needs no tap", c)
		}
		if r.Kind == KindNeeds && r.Remedy != RemedyPlayer {
			t.Fatalf("%s: a Needs-you code is a player remedy", c)
		}
	}
	if got := Sentence("needs_potions", map[string]string{"have": "3", "need": "4", "kind": "healing"}); got != "Going to town for potions: 3 of 4 healing potions." {
		t.Fatalf("sentence %q", got)
	}
	if got := Sentence("nope", nil); got != "No reason was given (report this)" {
		t.Fatalf("unknown code: %q", got)
	}
}

func cainBoard(t *testing.T) (*Board, *SM) {
	return smBoard(t), &SM{}
}

func TestStartAcceptsChain(t *testing.T) {
	b, m := cainBoard(t)
	changed, already, err := m.Start(b, "normal", "search_for_cain", "stones", "player", smT0, nil)
	if err != nil || already || len(changed) != 3 || changed[0].Chunk != "scroll" || changed[2].Chunk != "stones" {
		t.Fatalf("start: %v %v %v", changed, already, err)
	}
	live := m.Live(b, "normal")
	if live[0].Status != RunRunning || live[1].Status != RunQueued || live[1].By != "chain" || live[2].By != "player" {
		t.Fatalf("runs %+v", live)
	}
	if v := m.Check(b, "normal"); len(v) != 0 {
		t.Fatalf("violations %v", v)
	}
	// I7
	before, _ := json.Marshal(m)
	if _, already, _ := m.Start(b, "normal", "search_for_cain", "stones", "player", smT0, nil); !already {
		t.Fatal("second start must be already")
	}
	after, _ := json.Marshal(m)
	if string(before) != string(after) {
		t.Fatal("second start changed the store")
	}
}

func TestStartRefusals(t *testing.T) {
	b, m := cainBoard(t)
	boom := errors.New("too hard")
	_, _, err := m.Start(b, "normal", "search_for_cain", "stones", "player", smT0, func(r Ref) error {
		if r.Chunk == "read" {
			return boom
		}
		return nil
	})
	var ns ErrNotStartable
	if !errors.As(err, &ns) || ns.Part.Chunk != "read" || len(m.Runs) != 0 {
		t.Fatalf("refusal %v, runs %d (a refusal writes nothing)", err, len(m.Runs))
	}
	if _, _, err := m.Start(b, "normal", "search_for_cain", "scroll", "player", smT0, nil); err != nil {
		t.Fatal(err)
	}
	_, _, err = m.Start(b, "normal", "den_of_evil", "clear_den", "player", smT0, nil)
	var busy ErrBusy
	if !errors.As(err, &busy) || busy.Quest != "search_for_cain" {
		t.Fatalf("D5: %v", err)
	}
	m.NoteDone(b, "normal", "search_for_cain", "scroll", TierProxy, smT0, "pack")
	if _, _, err := m.Start(b, "normal", "search_for_cain", "scroll", "player", smT0, nil); !errors.Is(err, ErrDone) {
		t.Fatalf("done chunk: %v", err)
	}
}

func TestStopCascadesAndIdempotent(t *testing.T) {
	b, m := cainBoard(t)
	m.Start(b, "normal", "search_for_cain", "stones", "player", smT0, nil)
	changed, already := m.Stop(b, "normal", "search_for_cain", "scroll")
	if already || len(changed) != 3 {
		t.Fatalf("stop: %v %v", changed, already)
	}
	if len(m.Live(b, "normal")) != 0 {
		t.Fatal("a run survived the cascade")
	}
	before, _ := json.Marshal(m)
	if _, already := m.Stop(b, "normal", "search_for_cain", "scroll"); !already {
		t.Fatal("second stop must be already")
	}
	after, _ := json.Marshal(m)
	if string(before) != string(after) {
		t.Fatal("second stop changed the store")
	}
	// stopping a later part leaves the earlier ones
	m = &SM{}
	m.Start(b, "normal", "search_for_cain", "stones", "player", smT0, nil)
	if ch, _ := m.Stop(b, "normal", "search_for_cain", "stones"); len(ch) != 1 || len(m.Live(b, "normal")) != 2 {
		t.Fatalf("stop of the last part: %v", ch)
	}
}

func TestParkedHeadStopsTheRest(t *testing.T) {
	b, m := cainBoard(t)
	m.Start(b, "normal", "search_for_cain", "stones", "player", smT0, nil)
	if _, err := m.Apply(b, "normal", "search_for_cain", "scroll", EvCap, "died_too_often"); err != nil {
		t.Fatal(err)
	}
	live := m.Live(b, "normal")
	if live[0].Status != RunParked || live[1].Status != RunQueued || live[2].Status != RunQueued {
		t.Fatalf("D3: %+v", live)
	}
	if v := m.Check(b, "normal"); len(v) != 0 {
		t.Fatalf("violations %v", v)
	}
	if _, already, err := m.Retry(b, "normal", "search_for_cain", "scroll", false, smT0); err != nil || already {
		t.Fatalf("retry %v %v", already, err)
	}
	if m.Live(b, "normal")[0].Status != RunRunning {
		t.Fatal("retry must make the head run again")
	}
	if _, already, _ := m.Retry(b, "normal", "search_for_cain", "scroll", false, smT0); !already {
		t.Fatal("retry of a running run must be already")
	}
}

// I9: save, load, decide equals decide.
func TestRoundTripKeepsDecision(t *testing.T) {
	b, m := cainBoard(t)
	m.Start(b, "normal", "search_for_cain", "stones", "player", smT0, nil)
	m.Apply(b, "normal", "search_for_cain", "scroll", EvCap, "gave_up")
	raw, _ := json.Marshal(m)
	var m2 SM
	if err := json.Unmarshal(raw, &m2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m.Live(b, "normal"), m2.Live(b, "normal")) {
		t.Fatal("load changed the live runs")
	}
	if len(m2.Promote(b, "normal")) != 0 {
		t.Fatal("a parked head was promoted after load (Q5: a park survives)")
	}
}

func TestLatchOnlyRises(t *testing.T) {
	b, m := cainBoard(t)
	m.Start(b, "normal", "search_for_cain", "scroll", "player", smT0, nil)
	if !m.NoteDone(b, "normal", "search_for_cain", "scroll", TierProxy, smT0, "pack") {
		t.Fatal("first read must write")
	}
	if len(m.Live(b, "normal")) != 0 {
		t.Fatal("I4: the run closes with the latch")
	}
	if m.NoteDone(b, "normal", "search_for_cain", "scroll", TierNone, smT0, "x") || m.Latches[0].Tier != TierProxy {
		t.Fatal("a lower tier must change nothing")
	}
	if !m.NoteDone(b, "normal", "search_for_cain", "scroll", TierSave, smT0, "save") || m.Latches[0].Tier != TierSave {
		t.Fatal("a higher tier raises the latch")
	}
	// R1: silence of a proxy never clears it
	if _, cleared := m.Regress(b, "normal", "search_for_cain", "scroll", TierProxy); cleared || !m.Latched("search_for_cain", "scroll", "normal") {
		t.Fatal("R1")
	}
}

func TestRegressParksAndAsks(t *testing.T) {
	b, m := cainBoard(t)
	m.Start(b, "normal", "search_for_cain", "scroll", "player", smT0, nil)
	m.NoteDone(b, "normal", "search_for_cain", "scroll", TierSave, smT0, "save")
	m.Start(b, "normal", "search_for_cain", "read", "player", smT0, nil)
	parked, cleared := m.Regress(b, "normal", "search_for_cain", "scroll", TierSave)
	if !cleared || parked == nil || parked.Chunk != "read" {
		t.Fatalf("regress: %v %v", parked, cleared)
	}
	live := m.Live(b, "normal")
	if len(live) != 1 || live[0].Status != RunParked || live[0].Code != "prereq_regressed" {
		t.Fatalf("R4: %+v", live)
	}
	if _, _, err := m.Retry(b, "normal", "search_for_cain", "read", true, smT0); err != nil {
		t.Fatal(err)
	}
	live = m.Live(b, "normal")
	if len(live) != 2 || live[0].Chunk != "scroll" || live[0].By != "chain" || live[0].Status != RunRunning {
		t.Fatalf("run it again: %+v", live)
	}
	if v := m.Check(b, "normal"); len(v) != 0 {
		t.Fatalf("violations %v", v)
	}
}

func TestRepairClosure(t *testing.T) {
	b, m := cainBoard(t)
	m.Runs = []Run{{Quest: "search_for_cain", Chunk: "stones", Diff: "normal", Status: RunQueued, Code: "waits_for", At: smT0}}
	if v := m.Check(b, "normal"); len(v) == 0 {
		t.Fatal("I2 not detected")
	}
	added, parked := m.Repair(b, "normal", smT0, nil)
	if len(added) != 2 || len(parked) != 0 || len(m.Check(b, "normal")) != 0 {
		t.Fatalf("repair %v %v %v", added, parked, m.Check(b, "normal"))
	}
	m = &SM{Runs: []Run{{Quest: "search_for_cain", Chunk: "stones", Diff: "normal", Status: RunQueued, Code: "waits_for", At: smT0}}}
	added, parked = m.Repair(b, "normal", smT0, func(Ref) error { return errors.New("no") })
	if len(added) != 0 || len(parked) != 1 || m.Live(b, "normal")[0].Code != "prereq_missing" {
		t.Fatalf("unstartable repair: %v %v", added, parked)
	}
}

// I1, I2, I7 over every multi-part quest of the shipped board and every subset of its chunks (the generated set).
func TestInvariantsOverEveryChain(t *testing.T) {
	b := smBoard(t)
	cs := chains(b)
	if len(cs) < 2 {
		t.Fatalf("want at least two multi-part quests as fixtures, have %d", len(cs))
	}
	for _, q := range cs {
		n := len(q.Chunks)
		if n > 6 {
			n = 6
		}
		for mask := 1; mask < 1<<n; mask++ {
			m := &SM{}
			for i := 0; i < n; i++ {
				if mask&(1<<i) != 0 {
					if _, _, err := m.Start(b, "normal", q.ID, q.Chunks[i].ID, "player", smT0, nil); err != nil {
						t.Fatalf("%s mask %d: %v", q.ID, mask, err)
					}
				}
			}
			if v := m.Check(b, "normal"); len(v) != 0 {
				t.Fatalf("%s mask %d: %v", q.ID, mask, v)
			}
			snap, _ := json.Marshal(m)
			for i := 0; i < n; i++ {
				m.Stop(b, "normal", q.ID, q.Chunks[i].ID)
				once, _ := json.Marshal(m)
				m.Stop(b, "normal", q.ID, q.Chunks[i].ID)
				twice, _ := json.Marshal(m)
				if string(once) != string(twice) {
					t.Fatalf("%s: stop is not idempotent", q.ID)
				}
				if v := m.Check(b, "normal"); len(v) != 0 {
					t.Fatalf("%s mask %d stop %d: %v", q.ID, mask, i, v)
				}
				_ = json.Unmarshal(snap, m)
			}
		}
	}
}
