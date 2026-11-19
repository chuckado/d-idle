package task_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"d2lod-bot/internal/task"
)

// The hand-in steps of slice S7b: the Cain report to Akara, the Horadric Staff to Deckard Cain and the Golden Bird
// through Meshif and Alkor. Every dialogue row and quest word rule is U until the live drills in docs/questboard.md.
// The steps are pure functions of the world (no class input), so the class dimension is covered where it exists, in
// internal/questboard and internal/bot.

type talkRec struct {
	task.Actions
	talks []string
	err   error
}

func (r *talkRec) Talk(_ context.Context, npc, pick string) error {
	r.talks = append(r.talks, npc+"/"+pick)
	return r.err
}

func step(t *testing.T, group, suffix string) task.Step {
	t.Helper()
	for _, n := range task.Handler(group) {
		if n.Step.ID == "quest."+group+"/"+suffix {
			return n.Step
		}
	}
	t.Fatalf("group %s has no step %s", group, suffix)
	return task.Step{}
}

func w0(act int, items map[string]int, quests map[string]task.QuestState) task.World {
	return task.World{Act: act, InTown: true, Level: 30, Items: items, Quests: quests, Killed: map[string]bool{}, Cleared: map[task.Area]bool{}, WPs: map[task.Area]bool{}}
}

func need(s task.Step, w task.World) bool { n, _ := s.Need(w); return n }

func TestCainReportOnlyOnReaderEvidence(t *testing.T) {
	s := step(t, task.QCainBoard, "report")
	if s.Phase != "turn-in" {
		t.Fatalf("report must be a turn-in: %+v", s)
	}
	// T3 only (the cage was clicked, nothing readable): the step does not claim the objective
	if need(s, w0(1, map[string]int{}, nil)) {
		t.Fatal("report needed without evidence that Cain is free")
	}
	cleared := map[string]task.QuestState{task.QCain: {Stage: task.StageCleared, Ev: task.EvMemory}}
	if !need(s, w0(1, map[string]int{}, cleared)) {
		t.Fatal("freed (reward pending) but the report is not needed")
	}
	done := map[string]task.QuestState{task.QCain: {Stage: task.StageDone, Ev: task.EvSave}}
	if need(s, w0(1, map[string]int{}, done)) {
		t.Fatal("reported but still needed")
	}
	// restart mid-step: the world is the only memory, so the same answer comes back
	if !need(s, w0(1, map[string]int{}, cleared)) {
		t.Fatal("re-derivation after a restart changed the answer")
	}
	r := &talkRec{}
	if err := s.Do(context.Background(), r); err != nil || len(r.talks) != 1 || r.talks[0] != "Akara/cain report" {
		t.Fatalf("talk = %v err %v", r.talks, err)
	}
	// NPC not found: the error reaches the executor (it classifies and skips an optional step)
	r = &talkRec{err: task.Errorf(task.KindNotFound, "Akara not found")}
	if err := s.Do(context.Background(), r); err == nil || task.Classify(err) != task.KindNotFound {
		t.Fatalf("not found = %v", err)
	}
	// in the field the step asks for a trip home first
	w := w0(1, map[string]int{}, cleared)
	w.InTown = false
	if f := s.Ready(w); f == nil || len(f.Insert) != 1 {
		t.Fatalf("field ready = %+v", f)
	}
}

func TestStaffCainHandIn(t *testing.T) {
	s := step(t, task.QStaffBoard, "cain")
	if s.Phase != "turn-in" {
		t.Fatalf("cain step: %+v", s)
	}
	if need(s, w0(2, map[string]int{}, nil)) {
		t.Fatal("needed without a Staff")
	}
	if need(s, w0(2, map[string]int{task.ItemShaft: 1, task.ItemStaffTop: 1}, nil)) {
		t.Fatal("the pieces are not the Staff")
	}
	staff := map[string]int{task.ItemStaff: 1}
	if !need(s, w0(2, staff, nil)) {
		t.Fatal("Staff in the pack, not placed: the hand-in is needed")
	}
	if need(s, w0(2, staff, map[string]task.QuestState{task.QStaff: {Stage: task.StageDone, Ev: task.EvSave}})) {
		t.Fatal("quest word done: not needed")
	}
	placed := w0(2, map[string]int{task.ItemStaff: 1, task.ItemStaffPlaced: 1}, nil)
	if need(s, placed) {
		t.Fatal("placed: not needed")
	}
	r := &talkRec{}
	if err := s.Do(context.Background(), r); err != nil || r.talks[0] != "Deckard Cain/horadric staff" {
		t.Fatalf("talk = %v err %v", r.talks, err)
	}
	r = &talkRec{err: errors.New("unreachable")}
	if err := s.Do(context.Background(), r); err == nil {
		t.Fatal("a failed talk must surface")
	}
	// the Cube step comes first in the board group; the campaign group has no Cain talk
	ns := task.Handler(task.QStaffBoard)
	if ns[0].Step.ID != "quest.horadric_staff_board/assemble" || ns[1].Step.ID != "quest.horadric_staff_board/cain" {
		t.Fatalf("order: %s, %s", ns[0].Step.ID, ns[1].Step.ID)
	}
	for _, n := range task.Handler(task.QStaff) {
		if strings.HasSuffix(n.Step.ID, "/cain") {
			t.Fatal("the campaign group talks to Cain")
		}
	}
	for _, n := range task.Handler(task.QCain) {
		if strings.HasSuffix(n.Step.ID, "/report") {
			t.Fatal("the campaign group reports to Akara")
		}
	}
}

func TestGoldenBoardNeedsPositiveEvidence(t *testing.T) {
	g := task.QGoldenBoard
	meshif, alkor := step(t, g, "meshif"), step(t, g, "alkor")
	// nothing in the pack, no reader: both are "needed" (never done by default) and both say what is missing
	empty := w0(3, map[string]int{}, nil)
	if !need(meshif, empty) || !need(alkor, empty) {
		t.Fatal("a hero that never had a figurine reads as done")
	}
	if f := meshif.Ready(empty); f == nil || !strings.Contains(f.Why, "Jade Figurine") || len(f.Insert) != 0 {
		t.Fatalf("meshif ready = %+v", f)
	}
	if f := alkor.Ready(empty); f == nil || !strings.Contains(f.Why, "Golden Bird") {
		t.Fatalf("alkor ready = %+v", f)
	}
	// the figurine in the pack: Meshif can run, in town; in the field a trip home is spliced in
	figurine := w0(3, map[string]int{task.ItemJade: 1}, nil)
	if f := meshif.Ready(figurine); f != nil {
		t.Fatalf("meshif with the figurine = %+v", f)
	}
	figurine.InTown = false
	if f := meshif.Ready(figurine); f == nil || len(f.Insert) != 1 {
		t.Fatalf("meshif in the field = %+v", f)
	}
	// the bird made: Meshif is done, Alkor can run
	bird := w0(3, map[string]int{task.ItemGolden: 1}, nil)
	if need(meshif, bird) || !need(alkor, bird) || alkor.Ready(bird) != nil {
		t.Fatal("bird in the pack: meshif done, alkor open")
	}
	// the quest word (T1/T2) ends both; a proxy never does
	word := map[string]task.QuestState{task.QGolden: {Stage: task.StageDone, Ev: task.EvMemory}}
	if need(meshif, w0(3, nil, word)) || need(alkor, w0(3, nil, word)) {
		t.Fatal("quest word done: nothing left")
	}
	// the legacy group is untouched
	if n := task.Handler(task.QGolden); len(n) != 2 || !n[0].Step.Optional {
		t.Fatalf("legacy golden group changed: %d nodes", len(n))
	}
	r := &talkRec{}
	if err := alkor.Do(context.Background(), r); err != nil || r.talks[0] != "Alkor/golden bird" {
		t.Fatalf("talk = %v err %v", r.talks, err)
	}
}
