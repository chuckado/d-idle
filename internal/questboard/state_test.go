package questboard

import (
	"reflect"
	"strings"
	"testing"
)

type facts map[int]map[string]Fact

func (f facts) Quest(d int, id string) Fact { return f[d][id] }

func done(t Tier) Fact    { return Fact{Known: true, Done: true, Tier: t} }
func notDone(t Tier) Fact { return Fact{Known: true, Tier: t} }

func rec(quest, chunk string, diff int) Record {
	return Record{Quest: quest, Chunk: chunk, Diff: Difficulties[diff]}
}

func mustBoard(t *testing.T) *Board {
	t.Helper()
	b, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func view(t *testing.T, b *Board, in Input) map[string]QuestView {
	t.Helper()
	vs, err := Derive(b, in)
	if err != nil {
		t.Fatal(err)
	}
	again, _ := Derive(b, in)
	if !reflect.DeepEqual(vs, again) {
		t.Fatal("derive twice gave different results")
	}
	m := map[string]QuestView{}
	for _, v := range vs {
		m[v.ID] = v
	}
	return m
}

func chunk(v QuestView, id string) ChunkView {
	for _, c := range v.Chunks {
		if c.ID == id {
			return c
		}
	}
	return ChunkView{}
}

// readinessFor stands in for slice S2: the transitional legacy floor per class family (design 1.1), safe at the
// floor, risky within 2 levels, else too hard. It is what lets two builds see different offers.
func readinessFor(t *testing.T, b *Board, class string, clvl int) map[string]Word {
	t.Helper()
	out := map[string]Word{}
	for _, q := range b.Quests() {
		floor, ok, err := b.Floor(q, class)
		if err != nil {
			t.Fatal(err)
		}
		r := ReadySafe
		if ok {
			switch {
			case clvl >= floor:
			case clvl >= floor-2:
				r = ReadyRisky
			default:
				r = ReadyTooHard
			}
		}
		for _, c := range q.Chunks {
			out[ChunkKey(q.ID, c.ID)] = r
		}
	}
	return out
}

func TestStatesPerEvidence(t *testing.T) {
	b := mustBoard(t)
	allSafe := func(quest string) map[string]Word {
		return map[string]Word{ChunkKey(quest, "clear_den"): ReadySafe}
	}
	act1Open := facts{0: {"act1_done": done(TierSave)}}
	cases := []struct {
		name       string
		in         Input
		quest      string
		chunk      string
		state      State
		tier       Tier
		confirmed  bool
		whyHas     string
		awards     int
		afterCheck func(t *testing.T, v QuestView)
	}{
		{name: "locked: prerequisite word only on proxy never unlocks", quest: "radament", chunk: "radament", state: Locked, whyHas: "locked until",
			in: Input{Facts: facts{0: {"act1_done": done(TierProxy)}}}},
		{name: "locked: prerequisite not known", quest: "radament", chunk: "radament", state: Locked, whyHas: "Sail east",
			in: Input{Facts: facts{}}},
		{name: "locked to available: prerequisite on T1, readiness unknown", quest: "radament", chunk: "radament", state: Available, whyHas: "readiness not known",
			in: Input{Facts: act1Open}},
		{name: "locked to ready: prerequisite on T2 and safe", quest: "radament", chunk: "radament", state: Ready,
			in: Input{Facts: facts{0: {"act1_done": done(TierMemory)}}, Ready: map[string]Word{"radament/radament": ReadySafe}}},
		{name: "available: too hard", quest: "den_of_evil", chunk: "clear_den", state: Available, whyHas: "too hard",
			in: Input{Facts: facts{}, Ready: map[string]Word{"den_of_evil/clear_den": ReadyTooHard}}},
		{name: "available: missing kit keeps it a readiness fact, not a failure", quest: "den_of_evil", chunk: "clear_den", state: Available, whyHas: "3 antidotes",
			in: Input{Facts: facts{}, Ready: allSafe("den_of_evil"), Chunks: map[string]ChunkFact{"den_of_evil/clear_den": {Missing: "needs 3 antidotes"}}}},
		{name: "ready: safe", quest: "den_of_evil", chunk: "clear_den", state: Ready,
			in: Input{Facts: facts{}, Ready: allSafe("den_of_evil")}},
		{name: "ready: risky still offers", quest: "den_of_evil", chunk: "clear_den", state: Ready,
			in: Input{Facts: facts{}, Ready: map[string]Word{"den_of_evil/clear_den": ReadyRisky}}},
		{name: "accepted: the tap, handler not running", quest: "den_of_evil", chunk: "clear_den", state: Accepted,
			in: Input{Facts: facts{}, Plan: Plan{Accepted: []Record{rec("den_of_evil", "clear_den", 0)}}}},
		{name: "in progress: step is the first unsatisfied check", quest: "den_of_evil", chunk: "clear_den", state: InProgress,
			in: Input{Facts: facts{}, Plan: Plan{Accepted: []Record{rec("den_of_evil", "clear_den", 0)}},
				Chunks: map[string]ChunkFact{"den_of_evil/clear_den": {Running: true, Steps: []bool{true, true, false, false}}}},
			afterCheck: func(t *testing.T, v QuestView) {
				if c := chunk(v, "clear_den"); c.Step != 2 || c.StepName != "clear_den" {
					t.Errorf("step = %d %q", c.Step, c.StepName)
				}
			}},
		{name: "in progress: steps checks all true but not done yet leaves no open step", quest: "den_of_evil", chunk: "clear_den", state: InProgress,
			in: Input{Facts: facts{}, Plan: Plan{Accepted: []Record{rec("den_of_evil", "clear_den", 0)}},
				Chunks: map[string]ChunkFact{"den_of_evil/clear_den": {Running: true, Steps: []bool{true, true, true, true}}}}},
		{name: "awaiting: objective met on a chunk the player started (T3 enough to stop)", quest: "den_of_evil", chunk: "clear_den", state: Awaiting, tier: TierProxy,
			in: Input{Facts: facts{}, Plan: Plan{Accepted: []Record{rec("den_of_evil", "clear_den", 0)}},
				Chunks: map[string]ChunkFact{"den_of_evil/clear_den": {Running: true, Objective: true}}}},
		{name: "awaiting: restart with the objective already met", quest: "den_of_evil", chunk: "clear_den", state: Awaiting, tier: TierProxy,
			in: Input{Facts: facts{}, Plan: Plan{Accepted: []Record{rec("den_of_evil", "clear_den", 0)}},
				Chunks: map[string]ChunkFact{"den_of_evil/clear_den": {Objective: true}}}},
		{name: "awaiting: pending bit on T1 needs no record", quest: "den_of_evil", chunk: "clear_den", state: Awaiting, tier: TierSave,
			in: Input{Facts: facts{0: {"den_of_evil": {Known: true, Pending: true, Started: true, Tier: TierSave}}}}},
		{name: "not awaiting: pending bit on a proxy cannot stop or ask", quest: "den_of_evil", chunk: "clear_den", state: Available,
			in: Input{Facts: facts{0: {"den_of_evil": {Known: true, Pending: true, Tier: TierProxy}}}}},
		{name: "not awaiting: objective met but never accepted", quest: "den_of_evil", chunk: "clear_den", state: Available,
			in: Input{Facts: facts{}, Chunks: map[string]ChunkFact{"den_of_evil/clear_den": {Objective: true}}}},
		{name: "done: T1 confirmed, board ran it, pays once", quest: "den_of_evil", chunk: "clear_den", state: Done, tier: TierSave, confirmed: true, awards: 1,
			in: Input{Facts: facts{0: {"den_of_evil": done(TierSave)}}, Plan: Plan{Accepted: []Record{rec("den_of_evil", "clear_den", 0)}}}},
		{name: "done: T2 confirmed", quest: "den_of_evil", chunk: "clear_den", state: Done, tier: TierMemory, confirmed: true, awards: 1,
			in: Input{Facts: facts{0: {"den_of_evil": done(TierMemory)}}, Plan: Plan{Accepted: []Record{rec("den_of_evil", "clear_den", 0)}}}},
		{name: "done: T3 only shows confirming and pays nothing yet", quest: "den_of_evil", chunk: "clear_den", state: Done, tier: TierProxy, confirmed: false, awards: 0,
			in: Input{Facts: facts{0: {"den_of_evil": done(TierProxy)}}, Plan: Plan{Accepted: []Record{rec("den_of_evil", "clear_den", 0)}}}},
		{name: "done by hand: quiet, no award (no backfill)", quest: "den_of_evil", chunk: "clear_den", state: Done, tier: TierSave, confirmed: true, awards: 0, whyHas: "already done",
			in: Input{Facts: facts{0: {"den_of_evil": done(TierSave)}}}},
		{name: "conflict: T2 says not done, a proxy says done, single chunk: T2 wins", quest: "den_of_evil", chunk: "clear_den", state: Available,
			in: Input{Facts: facts{0: {"den_of_evil": notDone(TierMemory)}}, Chunks: map[string]ChunkFact{"den_of_evil/clear_den": {Done: true, Tier: TierProxy}}}},
		{name: "conflict: T1 stale says not done, a proxy says done: shown as confirming", quest: "den_of_evil", chunk: "clear_den", state: Done, tier: TierProxy,
			in: Input{Facts: facts{0: {"den_of_evil": notDone(TierSave)}}, Chunks: map[string]ChunkFact{"den_of_evil/clear_den": {Done: true, Tier: TierProxy}}}},
		{name: "skipped: the player's Skip", quest: "den_of_evil", chunk: "clear_den", state: Skipped,
			in: Input{Facts: facts{}, Plan: Plan{Skipped: []SkipRecord{{Quest: "den_of_evil", Diff: "normal"}}}}},
		{name: "skipped never beats done", quest: "den_of_evil", chunk: "clear_den", state: Done, tier: TierSave, confirmed: true,
			in: Input{Facts: facts{0: {"den_of_evil": done(TierSave)}}, Plan: Plan{Skipped: []SkipRecord{{Quest: "den_of_evil", Diff: "normal"}}}}},
		{name: "skipped in another difficulty does not apply", quest: "den_of_evil", chunk: "clear_den", state: Available,
			in: Input{Facts: facts{}, Plan: Plan{Skipped: []SkipRecord{{Quest: "den_of_evil", Diff: "nm"}}}}},
		{name: "blocked: the executor gave up, with a record", quest: "den_of_evil", chunk: "clear_den", state: Blocked, whyHas: "twice",
			in: Input{Facts: facts{}, Plan: Plan{Accepted: []Record{rec("den_of_evil", "clear_den", 0)}},
				Chunks: map[string]ChunkFact{"den_of_evil/clear_den": {Blocked: "killed me twice"}}}},
		{name: "blocked reason is ignored without a record", quest: "den_of_evil", chunk: "clear_den", state: Available,
			in: Input{Facts: facts{}, Chunks: map[string]ChunkFact{"den_of_evil/clear_den": {Blocked: "stale"}}}},
		{name: "blocked: the game failed the quest, no retry", quest: "den_of_evil", chunk: "clear_den", state: Blocked, whyHas: "failed in this game",
			in: Input{Facts: facts{0: {"den_of_evil": {Known: true, Failed: true, Tier: TierMemory}}}}},
		{name: "difficulty change: Normal done is not Nightmare done", quest: "den_of_evil", chunk: "clear_den", state: Ready,
			in: Input{Diff: 1, Facts: facts{0: {"den_of_evil": done(TierSave)}}, Ready: map[string]Word{"den_of_evil/clear_den": ReadySafe}}},
		{name: "difficulty change: Nightmare record does not accept Normal", quest: "den_of_evil", chunk: "clear_den", state: Available,
			in: Input{Facts: facts{}, Plan: Plan{Accepted: []Record{rec("den_of_evil", "clear_den", 1)}}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vs := view(t, b, c.in)
			v := vs[c.quest]
			cv := chunk(v, c.chunk)
			if cv.State != c.state {
				t.Fatalf("chunk state = %s (%s), want %s", cv.State, cv.Why, c.state)
			}
			if v.State != c.state {
				t.Errorf("quest state = %s, want %s", v.State, c.state)
			}
			if c.tier != TierNone && cv.Tier != c.tier {
				t.Errorf("tier = %s, want %s", cv.Tier, c.tier)
			}
			if c.state == Done && cv.Confirmed != c.confirmed {
				t.Errorf("confirmed = %v, want %v", cv.Confirmed, c.confirmed)
			}
			if c.whyHas != "" && !strings.Contains(cv.Why+v.Why, c.whyHas) {
				t.Errorf("why = %q, want it to mention %q", cv.Why, c.whyHas)
			}
			if len(v.Awards) != c.awards {
				t.Errorf("awards = %v, want %d", v.Awards, c.awards)
			}
			if c.afterCheck != nil {
				c.afterCheck(t, v)
			}
		})
	}
}

func TestAwardKeyIsAccountWideAndPerDifficulty(t *testing.T) {
	b := mustBoard(t)
	in := Input{Facts: facts{0: {"den_of_evil": done(TierSave)}}, Plan: Plan{Accepted: []Record{rec("den_of_evil", "clear_den", 0)}}}
	a := view(t, b, in)["den_of_evil"].Awards
	if len(a) != 1 || a[0].Key != "quest/den_of_evil/-/normal" || a[0].Amount != 2 {
		t.Fatalf("awards = %+v", a)
	}
	// Nightmare: a different key and amount; the hero is never in the key.
	in2 := Input{Diff: 1, Facts: facts{1: {"den_of_evil": done(TierSave)}}, Plan: Plan{Accepted: []Record{rec("den_of_evil", "clear_den", 1)}}}
	a2 := view(t, b, in2)["den_of_evil"].Awards
	if len(a2) != 1 || a2[0].Key != "quest/den_of_evil/-/nm" || a2[0].Amount != 3 {
		t.Fatalf("nm awards = %+v", a2)
	}
}

// Cain's chain: A then B then C then D, a hard chain; out-of-order Start taps are refused.
func TestChunkOrderChain(t *testing.T) {
	b := mustBoard(t)
	safe := map[string]Word{}
	for _, c := range []string{"scroll", "read", "free_cain", "report"} {
		safe[ChunkKey("search_for_cain", c)] = ReadySafe
	}
	in := Input{Facts: facts{}, Ready: safe}
	vs, _ := Derive(b, in)
	v := view(t, b, in)["search_for_cain"]
	if v.State != Ready || !reflect.DeepEqual(v.Offered, []string{"scroll"}) {
		t.Fatalf("start: %s offered %v", v.State, v.Offered)
	}
	if c := chunk(v, "read"); c.State != Locked || !strings.Contains(c.Why, "Get the Scroll") {
		t.Errorf("read = %s %q", c.State, c.Why)
	}
	// Out of order: refuse a locked chunk, accept the ready one.
	if err := CanStart(vs, "search_for_cain", "free_cain"); err == nil || !strings.Contains(err.Error(), "locked") {
		t.Errorf("out of order start = %v", err)
	}
	if err := CanStart(vs, "search_for_cain", "scroll"); err != nil {
		t.Errorf("ready start refused: %v", err)
	}
	// The scroll is in the pack (T3): the next chunk's offer appears, it skips no work.
	in.Chunks = map[string]ChunkFact{"search_for_cain/scroll": {Done: true, Tier: TierProxy}}
	v = view(t, b, in)["search_for_cain"]
	if !reflect.DeepEqual(v.Offered, []string{"read"}) || chunk(v, "scroll").Confirmed {
		t.Errorf("after scroll: offered %v, scroll %+v", v.Offered, chunk(v, "scroll"))
	}
	// Whole quest Done on T1 marks every chunk done.
	in.Facts = facts{0: {"search_for_cain": done(TierSave)}}
	v = view(t, b, in)["search_for_cain"]
	for _, c := range v.Chunks {
		if c.State != Done {
			t.Errorf("chunk %s = %s after the quest word", c.ID, c.State)
		}
	}
	if err := CanStart(vs, "ghost", "x"); err == nil || !strings.Contains(err.Error(), `"ghost"`) {
		t.Errorf("unknown quest start = %v", err)
	}
	if err := CanStart(vs, "search_for_cain", "zz"); err == nil || !strings.Contains(err.Error(), `"zz"`) {
		t.Errorf("unknown chunk start = %v", err)
	}
}

// The Staff pieces are offered in parallel in any order; assembling waits for both and the amulet (T1/T2).
func TestChunkOrderAnyThenAssemble(t *testing.T) {
	b := mustBoard(t)
	safe := map[string]Word{}
	for _, c := range []string{"cube", "shaft", "assemble"} {
		safe[ChunkKey("horadric_staff", c)] = ReadySafe
	}
	base := facts{0: {"act1_done": done(TierSave), "radament": done(TierSave)}}
	in := Input{Facts: base, Ready: safe}
	v := view(t, b, in)["horadric_staff"]
	if !reflect.DeepEqual(v.Offered, []string{"cube", "shaft"}) {
		t.Fatalf("parallel offers = %v", v.Offered)
	}
	// Shaft first, then cube: any order.
	in.Chunks = map[string]ChunkFact{"horadric_staff/shaft": {Done: true, Tier: TierProxy}}
	v = view(t, b, in)["horadric_staff"]
	if !reflect.DeepEqual(v.Offered, []string{"cube"}) || chunk(v, "assemble").State != Locked {
		t.Fatalf("shaft first: offered %v assemble %s", v.Offered, chunk(v, "assemble").State)
	}
	in.Chunks["horadric_staff/cube"] = ChunkFact{Done: true, Tier: TierProxy}
	v = view(t, b, in)["horadric_staff"]
	if c := chunk(v, "assemble"); c.State != Locked || !strings.Contains(c.Why, "Tainted Sun") {
		t.Errorf("assemble without the amulet = %s %q", c.State, c.Why)
	}
	// Amulet word on a proxy: still locked (T3 never unlocks a quest prerequisite). On T1: ready.
	base[0]["tainted_sun"] = done(TierProxy)
	if c := chunk(view(t, b, in)["horadric_staff"], "assemble"); c.State != Locked {
		t.Errorf("assemble on a proxy amulet = %s", c.State)
	}
	base[0]["tainted_sun"] = done(TierSave)
	// The order and the prerequisites are met: the assembly (with its Cain hand-in, S7b) is offered.
	v = view(t, b, in)["horadric_staff"]
	if c := chunk(v, "assemble"); !reflect.DeepEqual(v.Offered, []string{"assemble"}) || c.State != Ready {
		t.Errorf("assemble = %s %q offers %v", c.State, c.Why, v.Offered)
	}
}

// Restart mid-quest: the same inputs after a fresh process give the same views, with and without the executor.
func TestRestartMidQuestDerivesTheSame(t *testing.T) {
	b := mustBoard(t)
	plan := Plan{Accepted: []Record{rec("sisters_burial_grounds", "blood_raven", 0)}}
	steps := map[string]ChunkFact{"sisters_burial_grounds/blood_raven": {Running: true, Steps: []bool{true, true, true, false, false}}}
	f := facts{0: {"sisters_burial_grounds": {Known: true, Started: true, Tier: TierSave}}}
	running := view(t, b, Input{Facts: f, Plan: plan, Chunks: steps})["sisters_burial_grounds"]
	if running.State != InProgress || chunk(running, "blood_raven").StepName != "kill_blood_raven" {
		t.Fatalf("running = %s step %q", running.State, chunk(running, "blood_raven").StepName)
	}
	// The process restarts: nothing is running yet, the record survives, the hero resumes without a new tap.
	after := view(t, b, Input{Facts: f, Plan: plan})["sisters_burial_grounds"]
	if after.State != Accepted {
		t.Fatalf("after restart = %s", after.State)
	}
	// And a restart does not change a Done quest.
	f[0]["sisters_burial_grounds"] = done(TierSave)
	a := view(t, b, Input{Facts: f, Plan: plan})["sisters_burial_grounds"]
	c := view(t, b, Input{Facts: f, Plan: plan, Chunks: steps})["sisters_burial_grounds"]
	if a.State != Done || !reflect.DeepEqual(a, c) {
		t.Errorf("done differs with executor state: %+v vs %+v", a, c)
	}
}

// Two builds see different offers at the same level, from data (the legacy floor per class family).
func TestOffersDifferByClassFamily(t *testing.T) {
	b := mustBoard(t)
	for _, c := range []struct {
		class string
		clvl  int
		want  State
	}{
		{"paladin", 7, Ready},   // melee floor 7: safe
		{"sorceress", 7, Ready}, // caster floor 8: one level short is risky, still offered
		{"paladin", 4, Available},
		{"sorceress", 4, Available},
		{"sorceress", 6, Ready},
		{"sorceress", 5, Available},
		{"paladin", 5, Ready},
	} {
		in := Input{Facts: facts{}, Ready: readinessFor(t, b, c.class, c.clvl)}
		got := view(t, b, in)["sisters_burial_grounds"].State
		if got != c.want {
			t.Errorf("%s clvl %d: %s, want %s", c.class, c.clvl, got, c.want)
		}
	}
	if _, _, err := b.Floor(Quest{}, "bard"); err == nil {
		t.Error("unknown class accepted")
	}
}

func TestDeriveRejectsBadInput(t *testing.T) {
	b := mustBoard(t)
	cases := []struct {
		name string
		in   Input
		want string
	}{
		{"difficulty", Input{Diff: 3, Facts: facts{}}, "out of range"},
		{"no facts", Input{}, "no facts"},
		{"unknown quest in a record", Input{Facts: facts{}, Plan: Plan{Accepted: []Record{{Quest: "ghost", Chunk: "c", Diff: "normal"}}}}, `"ghost"`},
		{"unknown chunk in a record", Input{Facts: facts{}, Plan: Plan{Accepted: []Record{{Quest: "den_of_evil", Chunk: "zz", Diff: "normal"}}}}, `"zz"`},
		{"unknown difficulty in a record", Input{Facts: facts{}, Plan: Plan{Accepted: []Record{{Quest: "den_of_evil", Chunk: "clear_den", Diff: "easy"}}}}, `"easy"`},
		{"unknown skip", Input{Facts: facts{}, Plan: Plan{Skipped: []SkipRecord{{Quest: "ghost", Diff: "normal"}}}}, `"ghost"`},
		{"unknown auto", Input{Facts: facts{}, Plan: Plan{Auto: []AutoRecord{{Quest: "ghost", Diff: "normal"}}}}, `"ghost"`},
		{"bad chunk fact key", Input{Facts: facts{}, Chunks: map[string]ChunkFact{"nope": {}}}, "quest/chunk"},
		{"unknown chunk fact", Input{Facts: facts{}, Chunks: map[string]ChunkFact{"den_of_evil/zz": {}}}, `"zz"`},
	}
	for _, c := range cases {
		if _, err := Derive(b, c.in); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: error = %v, want it to name %q", c.name, err, c.want)
		}
	}
}

// Every row of the Transitions table is reachable and needs the evidence it says, checked on den_of_evil.
func TestTransitionTableEvidence(t *testing.T) {
	b := mustBoard(t)
	has := map[State]bool{}
	for _, tr := range Transitions {
		has[tr.To] = true
		if tr.Source == "" {
			t.Errorf("%s -> %s has no source", tr.From, tr.To)
		}
	}
	for _, s := range AllStates {
		if !has[s] {
			t.Errorf("state %s is never entered in the table", s)
		}
	}
	if len(AllStates) != 9 {
		t.Fatalf("%d states, want 9", len(AllStates))
	}
	// Done needs T1 to be QUIET-confirmed and T3 shows only "confirming": the table's Need for "" -> Done is T1.
	for _, tr := range Transitions {
		if tr.To == Done && tr.From == "" && tr.Need != TierSave {
			t.Errorf("quiet done needs %s, want T1", tr.Need)
		}
		if tr.From == Locked && !tr.Need.Exact() {
			t.Errorf("unlocking needs %s, T3 must never unlock", tr.Need)
		}
	}
	// A T3-only prerequisite never leaves locked, T1 and T2 do.
	for tier, want := range map[Tier]State{TierProxy: Locked, TierSave: Available, TierMemory: Available} {
		v := view(t, b, Input{Facts: facts{0: {"act1_done": done(tier)}}})["radament"]
		if v.State != want {
			t.Errorf("prerequisite on %s: %s, want %s", tier, v.State, want)
		}
	}
}

// Act 2 is closed until Act 1's hop is done on exact evidence, in each difficulty on its own.
func TestActGateIsPerDifficulty(t *testing.T) {
	b := mustBoard(t)
	f := facts{0: {"act1_done": done(TierSave)}}
	if got := view(t, b, Input{Facts: f})["radament"].State; got == Locked {
		t.Errorf("normal radament = %s", got)
	}
	if got := view(t, b, Input{Diff: 1, Facts: f})["radament"].State; got != Locked {
		t.Errorf("nm radament = %s, want locked until the nm hop", got)
	}
}
