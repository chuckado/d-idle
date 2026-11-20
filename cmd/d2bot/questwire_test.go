package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"d2lod-bot/internal/charsel"
	"d2lod-bot/internal/questboard"
)

// The plan files carry the hero name, so a Character Select switch must leave them where they are (parking them would hide
// the other hero's plan from it).
func TestQuestPlanFilesAreNotParkedBySwitch(t *testing.T) {
	inst := t.TempDir()
	dir := questStoreDir(filepath.Join(inst, "state"))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	b, err := questboard.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, hero := range []string{"Alpha", "Bravo"} {
		st, err := questboard.Open(b, dir, hero)
		if err != nil {
			t.Fatal(err)
		}
		if err := st.Accept("den_of_evil", "clear_den", 0, "train"); err != nil {
			t.Fatal(err)
		}
	}
	if err := charsel.SwapState(inst, "Alpha", "Bravo"); err != nil {
		t.Fatal(err)
	}
	for _, hero := range []string{"Alpha", "Bravo"} {
		st, _ := questboard.Open(b, dir, hero)
		p, _, err := st.Load()
		if err != nil || len(p.Accepted) != 1 {
			t.Fatalf("%s lost its plan after the switch: %+v %v", hero, p, err)
		}
	}
}

func TestQuestWireNeedsAStateDirAndSurvivesNoModel(t *testing.T) {
	if d := wireQuestBoard("", &runnerRef{}, nil, func(string, ...any) {}); d != nil {
		t.Fatal("a desk without a runner or state dir")
	}
	if _, _, ok := questMargin(nil, "", "normal", 0); ok {
		t.Fatal("margin without a model")
	}
}

func TestQuestNeedsBecomeGlanceRows(t *testing.T) {
	rows := []questboard.NeedRow{{ID: "quest:handin:a/b", Kind: "quest_handin", Sev: "event", Text: "Akara is waiting: hand in Den of Evil",
		Actions: []questboard.Verb{{Label: "Hand in", Post: "/api/quests/a/handin", Body: map[string]any{"chunk": "b"}}}}}
	got := questNeeds(rows)
	if len(got) != 1 || got[0].Actions[0].Post != "/api/quests/a/handin" || !strings.Contains(string(got[0].Actions[0].Body), `"chunk":"b"`) {
		t.Fatalf("%+v", got)
	}
}
