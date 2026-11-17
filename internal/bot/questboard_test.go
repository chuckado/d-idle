package bot

import (
	"context"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/env"
	"d2lod-bot/internal/questboard"
	"d2lod-bot/internal/task"
)

func boardRig(t *testing.T, on bool) (*rig, *questboard.Exec) {
	t.Helper()
	q := &fakeQuests{}
	g := campaignRig(t, q, 3)
	g.mover.exploreFn = func(ctx context.Context, done func() bool) error { return nil }
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

func TestQuestBoardKnobDefaultsOff(t *testing.T) {
	if env.DefaultTuning().Quests.Board != 0 {
		t.Fatal("quests.board must default to 0")
	}
	found := false
	for _, k := range env.Knobs() {
		if k.Key == "quests.board" {
			found = k.Default == 0 && k.Min == 0 && k.Max == 1
		}
	}
	if !found {
		t.Fatal("knob quests.board missing or not 0..1 default 0")
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
		nodes, ok := handlerNodes(q)
		if !ok || len(nodes) == 0 {
			t.Fatalf("%s: no handler nodes", id)
		}
		for _, n := range nodes {
			if isHandIn(n.Step) {
				t.Fatalf("%s keeps the hand-in step %s", id, n.Step.ID)
			}
		}
		if _, ok := handInNode(q); !ok {
			t.Fatalf("%s: no hand-in node", id)
		}
	}
	// a quest with several chunks has no handler of its own before S7
	q, _ := b.Quest("search_for_cain")
	if _, ok := handlerNodes(q); ok {
		t.Fatal("a multi-chunk quest must not be runnable in S3")
	}
	_ = task.QDen
}
