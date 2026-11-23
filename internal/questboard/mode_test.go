package questboard

import (
	"strings"
	"testing"
)

// The training target is the lowest-order open story quest and the meter words the gap, for two builds.
func TestNextTargetAndMeterTwoBuilds(t *testing.T) {
	for _, bd := range builds {
		t.Run(bd.class, func(t *testing.T) {
			r := newRig(t, bd.class, 1)
			st, err := r.x.State()
			if err != nil {
				t.Fatal(err)
			}
			tg, ok := NextTarget(r.b, st.Views, bd.class)
			if !ok || tg.Quest == "" || tg.Chunk == "" {
				t.Fatalf("no target: %+v", tg)
			}
			q, _ := r.b.Quest(tg.Quest)
			if q.Kind != "story" {
				t.Fatalf("the first target is %s (%s), want a story quest", tg.Quest, q.Kind)
			}
			m := BuildMeter("train", "train", tg, true, 1, "")
			if m.Target != tg.Quest || !strings.Contains(m.Line, "Training") || m.Line == "" {
				t.Fatalf("meter: %+v", m)
			}
			// a floor above the level reads as levels to go, in words
			tg.Offered, tg.Floor = false, 9
			m = BuildMeter("train", "train", tg, true, 7, "")
			if m.LevelsToGo != 2 || !strings.Contains(m.Line, "2 levels to go") {
				t.Fatalf("meter: %+v", m)
			}
			tg.Floor = 8
			if m = BuildMeter("train", "", tg, true, 7, ""); !strings.Contains(m.Line, "1 level to go") {
				t.Fatalf("meter: %+v", m)
			}
			// the floor reached but still too hard: not a level gap
			tg.Floor = 5
			if m = BuildMeter("train", "", tg, true, 7, ""); m.LevelsToGo != 0 || !strings.Contains(m.Line, "too hard") {
				t.Fatalf("meter: %+v", m)
			}
		})
	}
}

func TestMeterWordsEveryMode(t *testing.T) {
	for _, tc := range []struct{ mode, running, want string }{
		{"idle", "", "Waiting in town"},
		{"farm", "", "Farming loot and gold"},
		{"train", "", "no quest is left"},
		{"idle", "Den of Evil", "On a quest: Den of Evil. After it the hero returns to idle mode."},
	} {
		m := BuildMeter(tc.mode, tc.mode, Target{}, false, 5, tc.running)
		if !strings.Contains(m.Line, tc.want) {
			t.Errorf("%s/%s: %q", tc.mode, tc.running, m.Line)
		}
	}
}

// The Offered target says the player decides; Train never starts the quest.
func TestMeterOfferedTarget(t *testing.T) {
	m := BuildMeter("train", "train", Target{Quest: "x", Title: "Blood Raven", Offered: true}, true, 9, "")
	if !m.Offered || !strings.Contains(m.Line, "tap Start") {
		t.Fatalf("%+v", m)
	}
}
