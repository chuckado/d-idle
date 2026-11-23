package bot

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/questboard"
	"d2lod-bot/internal/talk"
	"d2lod-bot/internal/task"
)

func boardRig(t *testing.T, on bool) (*rig, *questboard.Exec) {
	t.Helper()
	q := &fakeQuests{}
	g := campaignRig(t, q, 3)
	g.mover.exploreFn = func(ctx context.Context, done func() bool) error { return nil }
	g.r.e.Tune.Quests.Board = 0
	if on {
		g.r.e.Tune.Quests.Board = 1
	}
	b, err := questboard.Load()
	if err != nil {
		t.Fatal(err)
	}
	st, err := questboard.Open(b, t.TempDir(), "Testy")
	if err != nil {
		t.Fatal(err)
	}
	return g, g.r.NewQuestBoard(b, st)
}

func TestQuestBoardKnobDefaultsOn(t *testing.T) {
	if env.DefaultTuning().Quests.Board != 1 {
		t.Fatal("quests.board must default to 1 (S9: the board is the only quest driver)")
	}
	found := false
	for _, k := range env.Knobs() {
		if k.Key == "quests.board" {
			found = k.Default == 1 && k.Min == 0 && k.Max == 1
		}
	}
	if !found {
		t.Fatal("knob quests.board missing or not 0..1 default 1")
	}
	tu := env.DefaultTuning()
	for _, k := range env.Knobs() {
		if k.Key == "quests.board" {
			if err := SetKnob(&tu, k, 1); err != nil || tu.Quests.Board != 1 {
				t.Fatalf("SetKnob: %v %d", err, tu.Quests.Board)
			}
		}
	}
}

// Knob 0 with a board wired: the planner plays the Den exactly as before (the board never acts, the reward is claimed by
// the campaign itself).
func TestQuestBoardKnobOffKeepsThePlanner(t *testing.T) {
	g, x := boardRig(t, false)
	g.run()
	defer g.finish()
	if _, err := x.Offers(); err != questboard.ErrOff {
		t.Fatalf("offers with the knob off: %v", err)
	}
	if err := g.r.Start(campaign()); err != nil {
		t.Fatal(err)
	}
	g.waitFor("reward talk", func() bool { return g.w.count("talk Akara/reward") > 0 })
	if g.r.questBoardOn() {
		t.Fatal("knob on")
	}
}

// Knob 1: nothing starts by itself; Start runs the chunk, the hero stops with the quest awaiting, and only the HandIn tap
// talks the reward.
func TestQuestBoardAcceptRunStopHandIn(t *testing.T) {
	g, x := boardRig(t, true)
	g.run()
	defer g.finish()
	if err := g.r.Start(campaign()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	if log := g.w.log(); strings.Contains(log, "talk Akara") {
		t.Fatalf("the planner started a quest by itself: %s", log)
	}
	off, err := x.Offers()
	if err != nil || len(off) == 0 {
		t.Fatalf("offers: %v %v", off, err)
	}
	if err := x.Accept("den_of_evil", "clear_den"); err != nil {
		t.Fatal(err)
	}
	g.waitFor("awaiting hand-in", func() bool {
		d, _ := x.Next()
		return d.Kind == questboard.DecAwait
	})
	if strings.Contains(g.w.log(), "talk Akara/reward") {
		t.Fatalf("the bot handed in by itself: %s", g.w.log())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	res, err := x.HandIn(ctx, "den_of_evil", "clear_den")
	if err != nil || res.Outcome == questboard.HandInBlocked {
		t.Fatalf("hand in: %+v %v", res, err)
	}
	if n := g.w.count("talk Akara/reward"); n != 1 {
		t.Fatalf("reward talked %d times: %s", n, g.w.log())
	}
	if res.Outcome != questboard.HandInDone {
		t.Fatalf("the save reads Done at once in this fake: %+v", res)
	}
	if d, _ := x.Next(); d.Kind != questboard.DecIdle {
		t.Fatalf("after the hand-in: %v", d)
	}
}

func TestBoardPlanHidesTheHandIn(t *testing.T) {
	b, err := questboard.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"den_of_evil", "sisters_burial_grounds"} {
		q, _ := b.Quest(id)
		c := q.Chunks[0]
		nodes, ok := handlerNodes(q, c)
		if !ok || len(nodes) == 0 {
			t.Fatalf("%s: no handler nodes", id)
		}
		hand, has := handInNode(q, c)
		if !has {
			t.Fatalf("%s: no hand-in node", id)
		}
		for _, n := range nodes {
			if n.Step.ID == hand.Step.ID {
				t.Fatalf("%s keeps the hand-in step %s", id, n.Step.ID)
			}
		}
	}
	// a chunk of a multi-chunk quest runs only its own steps (slice S7)
	q, _ := b.Quest("search_for_cain")
	scroll, _ := b.Chunk("search_for_cain", "scroll")
	nodes, ok := handlerNodes(q, scroll)
	if !ok || len(nodes) != 3 {
		t.Fatalf("scroll chunk nodes = %d %v", len(nodes), ok)
	}
	read, _ := b.Chunk("search_for_cain", "read")
	if nodes, ok := handlerNodes(q, read); !ok || len(nodes) != 0 {
		t.Fatalf("read chunk field nodes = %d %v", len(nodes), ok)
	}
	if n, has := handInNode(q, read); !has || stepSuffix("search_for_cain", n.Step) != "read" {
		t.Fatal("read chunk has no hand-in step")
	}
	te, _ := b.Quest("terrors_end")
	seals, _ := b.Chunk("terrors_end", "seals")
	if _, ok := handlerNodes(te, seals); ok {
		t.Fatal("an unsupported chunk must have no handler")
	}
	_ = task.QDen
}

// The margin query (Deps.QuestMargin, slice S4) replaces the legacy words: a boss the survival model calls too hard is
// never offered, whatever the old clvl floor says, and the API gets the reasons behind the word.
func TestQuestBoardAssessReplacesLegacyWords(t *testing.T) {
	g, x := boardRig(t, true)
	var calls atomic.Int32
	g.r.SetQuestMargin(func(s *d2.Snapshot, boss, diff string, potions int) (questboard.Hero, *questboard.Encounter, bool) {
		calls.Add(1)
		h := questboard.Hero{Class: "Paladin", Level: s.Player.Level, MercHired: true, MercAlive: true, Potions: 8, LowestResist: 0}
		if boss == "" {
			return h, nil, true
		}
		return h, &questboard.Encounter{Name: boss, Kind: "boss", Best: 0.2, Low: 0.1, Known: true}, true
	})
	g.run()
	defer g.finish()
	g.waitFor("snapshot", func() bool { return g.r.QuestClass() != "" })
	off, err := x.Offers()
	if err != nil {
		t.Fatal(err)
	}
	if offered(off, "sisters_burial_grounds") {
		t.Fatalf("Blood Raven offered although the margin says too hard: %+v", off)
	}
	b, _ := questboard.Load()
	q, _ := b.Quest("sisters_burial_grounds")
	rd, ok := g.r.QuestReadiness(q, q.Chunks[0])
	if !ok || rd.Word != questboard.ReadyTooHard || rd.Sentence == "" {
		t.Fatalf("readiness: %+v %v", rd, ok)
	}
	n := calls.Load()
	g.r.QuestReadiness(q, q.Chunks[0])
	if calls.Load() != n {
		t.Fatal("the margin query ran again inside the throttle window")
	}
}

func offered(off []questboard.Offer, quest string) bool {
	for _, o := range off {
		if o.Quest == quest {
			return true
		}
	}
	return false
}

// The runner builds the board of the loaded hero through the maker (the name is only known from the first snapshot) and
// follows the knob.
func TestQuestBoardMakerBuildsPerHero(t *testing.T) {
	g, _ := boardRig(t, true)
	var heroes []string
	var mu sync.Mutex
	got := func() []string { mu.Lock(); defer mu.Unlock(); return append([]string(nil), heroes...) }
	dir := t.TempDir()
	g.r.SetQuestBoardMaker(func(hero string) (*questboard.Board, *questboard.Store, error) {
		mu.Lock()
		heroes = append(heroes, hero)
		mu.Unlock()
		b, err := questboard.Load()
		if err != nil {
			return nil, nil, err
		}
		st, err := questboard.Open(b, dir, hero)
		return b, st, err
	})
	g.run()
	stopped := false
	defer func() {
		if !stopped {
			g.finish()
		}
	}()
	var x *questboard.Exec
	g.waitFor("board for the hero", func() bool {
		var st *questboard.Store
		x, st = g.r.QuestBoard()
		return x != nil && st != nil && len(got()) > 0
	})
	if h := got(); len(h) != 1 || h[0] == "" {
		t.Fatalf("maker calls: %v", h)
	}
	if x2, _ := g.r.QuestBoard(); x2 != x || len(got()) != 1 {
		t.Fatal("the board was rebuilt for the same hero")
	}
	g.finish() // the knob is plain data the loop reads: change it only with the loop stopped (-race)
	stopped = true
	g.r.e.Tune.Quests.Board = 0
	if x3, _ := g.r.QuestBoard(); x3.Enabled() {
		t.Fatal("the board did not follow the knob off")
	}
}

// Every chunk of every act file either names a handler that exists in the task layer or says why it cannot run
// (slice S7). The "evidence rule" of a chunk is its steps' own Done predicates plus the quest word.
func TestEveryChunkHasAHandlerOrIsUnsupported(t *testing.T) {
	b, err := questboard.Load()
	if err != nil {
		t.Fatal(err)
	}
	npcs := talk.ByAct
	for _, q := range b.Quests() {
		if q.Accept != nil {
			if _, ok := npcs[q.Accept.NPC]; !ok {
				t.Errorf("%s: accept NPC %q is not in the talk tables", q.ID, q.Accept.NPC)
			}
		}
		for _, c := range q.Chunks {
			at := q.ID + "/" + c.ID
			if c.Giver != nil {
				if _, ok := npcs[c.Giver.NPC]; !ok {
					t.Errorf("%s: hand-in NPC %q is not in the talk tables", at, c.Giver.NPC)
				}
			}
			nodes, ok := handlerNodes(q, c)
			if c.Unsupported != "" {
				if ok {
					t.Errorf("%s: unsupported but a handler runs it", at)
				}
				continue
			}
			if !ok {
				t.Errorf("%s: no handler (group %q) and no unsupported reason", at, q.HandlerOf(c).Group)
				continue
			}
			ref := q.HandlerOf(c)
			have := map[string]bool{}
			for _, n := range task.Handler(ref.Group) {
				have[stepSuffix(ref.Group, n.Step)] = true
			}
			for _, n := range c.Nodes {
				if !have[n] {
					t.Errorf("%s: node %q is not a step of group %q", at, n, ref.Group)
				}
			}
			for _, n := range nodes {
				if n.Step.Need == nil {
					t.Errorf("%s: step %s has no done check", at, n.Step.ID)
				}
			}
			if c.Ends == "handin" {
				if _, has := handInNode(q, c); !has {
					t.Errorf("%s: ends handin but group %q has no turn-in step %q", at, ref.Group, ref.HandIn)
				}
			}
		}
	}
}

// A hand-in step reads done before its item exists (the bark is not in the pack yet): a part never counts done
// before the parts it comes after.
func TestCainChunkEvidenceFollowsTheChain(t *testing.T) {
	b, _ := questboard.Load()
	q, _ := b.Quest("search_for_cain")
	chunk := func(id string) questboard.Chunk { c, _ := b.Chunk(q.ID, id); return c }
	w := task.World{Items: map[string]int{}, Quests: map[string]task.QuestState{}}
	if nodesDone(q, chunk("scroll"), w, 0) || nodesDone(q, chunk("read"), w, 0) {
		t.Fatal("nothing fetched: no part is done")
	}
	w.Items[task.ItemBark] = 1
	if !nodesDone(q, chunk("scroll"), w, 0) || nodesDone(q, chunk("read"), w, 0) {
		t.Fatal("bark in the pack: the scroll part is done, the reading is not")
	}
	delete(w.Items, task.ItemBark)
	w.Items[task.ItemCainSc] = 1
	if !nodesDone(q, chunk("scroll"), w, 0) || !nodesDone(q, chunk("read"), w, 0) {
		t.Fatal("Cain's scroll in the pack: both parts are done")
	}
}

// A failed board step must spend game time, or a lockstep clock never reaches the failure gap and the chunk spins forever.
func TestFailSpend(t *testing.T) {
	if failSpend(true, false) <= 0 || failSpend(true, false) >= boardFailGap {
		t.Fatalf("a failed non-idle step spends a part of the gap: %v", failSpend(true, false))
	}
	if failSpend(false, false) != 0 || failSpend(true, true) != 0 {
		t.Fatal("a step that did not fail, or an idle tick that already waits, spends nothing extra")
	}
}

// necro stage 1: a long goal travel yields once to a runnable chunk, never to the same chunk twice, and never without one.
func TestBoardYieldOncePerChunk(t *testing.T) {
	for _, class := range []d2.Class{d2.Necromancer, d2.Paladin} {
		t.Run(class.String(), func(t *testing.T) {
			g, x := boardRig(t, true)
			g.w.set(func(w *world) { w.class = class; w.area = 2 })
			snap, _ := g.w.snapshot()
			g.r.obMu.Lock()
			g.r.ob.lastSnap = snap
			g.r.obMu.Unlock()
			if g.r.boardYield() {
				t.Fatal("no accepted chunk: nothing to yield to")
			}
			if err := x.Accept("den_of_evil", "clear_den"); err != nil {
				t.Fatal(err)
			}
			if !g.r.boardYield() {
				t.Fatal("an accepted runnable chunk must stop the travel at the next hop")
			}
			if g.r.boardYield() {
				t.Fatal("the same chunk must not stop the travel twice (starvation)")
			}
		})
	}
}
