package bot

import (
	"testing"

	"d2lod-bot/internal/questboard"
	"d2lod-bot/internal/task"
)

// Chunk evidence of the three hand-in chunks of slice S7b (Cain report, Staff assembly, Golden Bird). Pure functions of
// the task world, so a restart is a second call with the same world (nothing is persisted). All rows U until the drills
// of docs/questboard.md; the tests are V for the logic.

func s7bWorld(items map[string]int, quests map[string]task.QuestState) task.World {
	return task.World{Act: 1, InTown: true, Level: 30, Items: items, Quests: quests, Killed: map[string]bool{}, Cleared: map[task.Area]bool{}, WPs: map[task.Area]bool{}}
}

func s7bChunk(t *testing.T, quest, chunk string) (questboard.Quest, questboard.Chunk) {
	t.Helper()
	b, err := questboard.Load()
	if err != nil {
		t.Fatal(err)
	}
	q, err := b.Quest(quest)
	if err != nil {
		t.Fatal(err)
	}
	c, err := b.Chunk(quest, chunk)
	if err != nil {
		t.Fatal(err)
	}
	return q, c
}

func TestCainReportChunk(t *testing.T) {
	q, rep := s7bChunk(t, "search_for_cain", "report")
	if rep.Unsupported == "" {
		t.Fatal("report must stay unsupported until the Tristram part runs live (cainfix)")
	}
	rep.Unsupported = "" // the evidence rules below are the ones that ship when it is switched on
	if _, ok := handlerNodes(q, rep); !ok {
		t.Fatal("report has no handler")
	}
	if _, ok := handInNode(q, rep); !ok {
		t.Fatal("report has no hand-in step")
	}
	exact := func(s task.Stage, ev task.Evidence) task.World {
		return s7bWorld(map[string]int{}, map[string]task.QuestState{task.QCain: {Stage: s, Ev: ev}})
	}
	// T3 only (no reader): nothing of the chain counts, report is not done
	if nodesDone(q, rep, s7bWorld(map[string]int{}, nil), 0) {
		t.Fatal("report reads done on nothing")
	}
	// freed, reward open (T1): the cage part is done, the report is not (it waits for the tap)
	w := exact(task.StageCleared, task.EvMemory)
	w.Area, w.Cleared[task.Tristram] = task.Tristram, true
	w.Items[task.ItemPortal] = 1
	if nodesDone(q, rep, w, 0) {
		t.Fatal("report done before the talk")
	}
	// the cage part is unsupported for now, so it never reads done (even for a hero who freed Cain by hand); this
	// flips with the switch-on of the Tristram part
	if priorDone(q, rep, w) {
		t.Fatal("an unsupported cage part must not read done")
	}
	// reported (T2 save word Done): done; a restart (same world) gives the same answer
	w2 := exact(task.StageDone, task.EvSave)
	for i := 0; i < 2; i++ {
		if !nodesDone(q, rep, w2, 0) {
			t.Fatalf("report not done on the save word (pass %d)", i)
		}
	}
	// T3-only conflict: Cain's scroll in the pack says "started", never "reported"
	conflict := s7bWorld(map[string]int{task.ItemCainSc: 1}, nil)
	if nodesDone(q, rep, conflict, 0) {
		t.Fatal("a proxy claimed the report")
	}
}

func TestStaffAssembleChunk(t *testing.T) {
	q, asm := s7bChunk(t, "horadric_staff", "assemble")
	nodes, ok := handlerNodes(q, asm)
	if !ok || len(nodes) != 1 || nodes[0].Step.ID != "quest.horadric_staff_board/assemble" {
		t.Fatalf("field nodes = %v ok %v", nodes, ok)
	}
	hand, ok := handInNode(q, asm)
	if !ok || hand.Step.ID != "quest.horadric_staff_board/cain" {
		t.Fatalf("hand-in = %v ok %v", hand.Step.ID, ok)
	}
	pieces := map[string]int{task.ItemCube: 1, task.ItemShaft: 1, task.ItemStaffTop: 1}
	// both pieces in the pack: the Cube step is needed, so the chunk is not done and has no objective yet
	if nodesDone(q, asm, s7bWorld(pieces, nil), 0) {
		t.Fatal("assemble done with the pieces unassembled")
	}
	// the Staff exists, Cain not told (no quest word): not done, the hand-in is the next thing
	staff := s7bWorld(map[string]int{task.ItemCube: 1, task.ItemStaff: 1}, nil)
	if nodesDone(q, asm, staff, 0) {
		t.Fatal("assemble done before the hand-in")
	}
	if n, _ := hand.Step.Need(staff); !n {
		t.Fatal("hand-in not needed with the Staff in the pack")
	}
	if f := hand.Step.Ready(staff); f != nil {
		t.Fatalf("hand-in blocked in town: %+v", f)
	}
	// the quest word (T2) done: the chunk is done; restart gives the same
	word := s7bWorld(map[string]int{task.ItemCube: 1, task.ItemStaff: 1}, map[string]task.QuestState{task.QStaff: {Stage: task.StageDone, Ev: task.EvSave}})
	for i := 0; i < 2; i++ {
		if !nodesDone(q, asm, word, 0) {
			t.Fatalf("assemble not done on the quest word (pass %d)", i)
		}
	}
	// nothing at all: the predecessors (cube, shaft) are not done, so the chunk is not
	if nodesDone(q, asm, s7bWorld(map[string]int{}, nil), 0) {
		t.Fatal("assemble reads done on an empty pack")
	}
}

func TestGoldenBirdChunk(t *testing.T) {
	q, bird := s7bChunk(t, "golden_bird", "bird")
	nodes, ok := handlerNodes(q, bird)
	if !ok || len(nodes) != 1 || nodes[0].Step.ID != "quest.golden_bird_board/meshif" {
		t.Fatalf("field nodes = %v ok %v", nodes, ok)
	}
	if legacyChunk(q, bird) {
		t.Fatal("the bird chunk must read its facts from the steps")
	}
	// a hero that never found a figurine must not read as done (the hand-in step is not satisfied by absence)
	if nodesDone(q, bird, s7bWorld(map[string]int{}, nil), 0) {
		t.Fatal("bird done on an empty pack")
	}
	if nodesDone(q, bird, s7bWorld(map[string]int{task.ItemJade: 1}, nil), 0) {
		t.Fatal("bird done with the figurine still in the pack")
	}
	if nodesDone(q, bird, s7bWorld(map[string]int{task.ItemGolden: 1}, nil), 0) {
		t.Fatal("bird done before Alkor")
	}
	done := s7bWorld(map[string]int{}, map[string]task.QuestState{task.QGolden: {Stage: task.StageDone, Ev: task.EvMemory}})
	for i := 0; i < 2; i++ {
		if !nodesDone(q, bird, done, 0) {
			t.Fatalf("bird not done on the quest word (pass %d)", i)
		}
	}
	// readiness fact: the chunk says what is missing, with no trip spliced in front
	meshif := nodes[0].Step
	if f := meshif.Ready(s7bWorld(map[string]int{}, nil)); f == nil || len(f.Insert) != 0 || f.Why == "" {
		t.Fatalf("missing figurine fault = %+v", f)
	}
}

// The group of the board's chunk is its own: the quest word of the quest itself (T1/T2) still sets the tier.
func TestBoardGroupTierFollowsTheQuestWord(t *testing.T) {
	q, bird := s7bChunk(t, "golden_bird", "bird")
	if q.HandlerOf(bird).Group != "golden_bird_board" || taskID(q) != "golden_bird" {
		t.Fatalf("group %q task %q", q.HandlerOf(bird).Group, taskID(q))
	}
}
