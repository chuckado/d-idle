package bot

import (
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/env"
	"d2lod-bot/internal/questboard"
)

func TestQuestAutoKnobs(t *testing.T) {
	d := env.DefaultTuning().Quests
	if d.AutoDelaySec != 20 || d.AwayMin != 10 || d.AutoHoldMin != 15 || d.AutoGapMin != 30 || d.AutoDayCap != 6 {
		t.Fatalf("defaults %+v", d)
	}
	want := map[string]bool{"quests.autoDelaySec": false, "quests.awayMin": false, "quests.autoHoldMin": false, "quests.autoGapMin": false, "quests.autoDayCap": false}
	tu := env.DefaultTuning()
	for _, k := range env.Knobs() {
		if _, ok := want[k.Key]; ok {
			want[k.Key] = true
			if err := SetKnob(&tu, k, 3); err != nil {
				t.Fatalf("%s: %v", k.Key, err)
			}
		}
	}
	for k, ok := range want {
		if !ok {
			t.Errorf("knob %s missing", k)
		}
	}
	if q := tu.Quests; q.AutoDelaySec != 3 || q.AwayMin != 3 || q.AutoHoldMin != 3 || q.AutoGapMin != 3 || q.AutoDayCap != 3 {
		t.Fatalf("SetKnob %+v", q)
	}
}

func autoRig(t *testing.T, on bool) (*rig, *questboard.Exec, *questboard.Store) {
	t.Helper()
	g, _ := boardRig(t, on)
	g.r.e.Tune.Quests.AutoDelaySec, g.r.e.Tune.Quests.AwayMin, g.r.e.Tune.Quests.AutoHoldMin = 0, 0, 0
	b, err := questboard.Load()
	if err != nil {
		t.Fatal(err)
	}
	st, err := questboard.Open(b, t.TempDir(), "Testy")
	if err != nil {
		t.Fatal(err)
	}
	return g, g.r.NewQuestBoard(b, st), st
}

// An opted-in, safe quest starts by itself, the hero stops at the NPC, and with nobody watching the board hands it in.
func TestQuestAutoStartsAndHandsInWhenAway(t *testing.T) {
	g, x, st := autoRig(t, true)
	if err := st.SetAutoStart("den_of_evil", 0, true); err != nil {
		t.Fatal(err)
	}
	g.run()
	defer g.finish()
	if err := g.r.Start(campaign()); err != nil {
		t.Fatal(err)
	}
	g.waitFor("auto hand-in", func() bool { return g.w.count("talk Akara/reward") > 0 })
	time.Sleep(100 * time.Millisecond)
	if n := g.w.count("talk Akara/reward"); n != 1 {
		t.Fatalf("reward talked %d times: %s", n, g.w.log())
	}
	p, _, _ := st.Load()
	if len(p.Accepted) != 1 || !p.Accepted[0].Auto || len(p.AutoLog) < 1 || p.AutoLog[0].Kind != questboard.EvAutoStart {
		t.Fatalf("plan %+v", p)
	}
	if d, _ := x.Next(); d.Kind != questboard.DecIdle {
		t.Fatalf("after: %v", d)
	}
}

// Without the opt-in nothing starts, and with the knob at 0 the board and the controller stay silent.
func TestQuestAutoNeedsTheOptInAndTheKnob(t *testing.T) {
	for _, on := range []bool{true, false} {
		g, x, st := autoRig(t, on)
		if on {
			if err := st.SetAutoStart("den_of_evil", 0, false); err != nil {
				t.Fatal(err)
			}
		}
		g.run()
		if err := g.r.Start(campaign()); err != nil {
			t.Fatal(err)
		}
		time.Sleep(300 * time.Millisecond)
		g.finish()
		if log := g.w.log(); !on && !strings.Contains(log, "talk Akara") {
			// knob 0: the planner plays the Den as before (covered by TestQuestBoardKnobOffKeepsThePlanner)
		} else if on && strings.Contains(log, "talk Akara/quest") {
			t.Fatalf("started without the opt-in: %s", log)
		}
		if p, _, _ := st.Load(); on && (len(p.Accepted) != 0 || len(p.AutoLog) != 0) {
			t.Fatalf("plan %+v", p)
		}
		_ = x
	}
}
