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
	if rep.Unsupported != "" {
		t.Fatalf("report is supported since CAIN-5b, got %q", rep.Unsupported)
	}
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
	if !priorDone(q, rep, w) {
		t.Fatal("the cage part does not read done on the reader's freed word")
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

// V lab drill CAIN-L3 (2026-11-19, sandbox-like run on a d2d lab slot): the scroll chunk reads done only while the
// Key to the Cairn Stones is in the pack, the stones chunk only while the hero stands in Tristram or holds a portal.
// The first town trip from Tristram (the potion trip comes at once) made both read open again: the board offered
// the Scroll of Inifuss again and the hero walked back to the Dark Wood. A chunk seen done stays done for the process.
func TestCainAutoChunksStayDoneAfterTheEvidenceVanishes(t *testing.T) {
	b, err := questboard.Load()
	if err != nil {
		t.Fatal(err)
	}
	{ // class-free rule: no class, skill or build data is read, so one run covers every build
		rt := &boardRT{handed: map[string]bool{}, reached: map[string]bool{}}
		q, err := b.Quest("search_for_cain")
		if err != nil {
			t.Fatal(err)
		}
		tristram := s7bWorld(map[string]int{task.ItemCainSc: 1, task.ItemPortal: 1}, map[string]task.QuestState{task.QCain: {Stage: task.StageStarted, Ev: task.EvMemory}})
		tristram.Area, tristram.InTown = task.Tristram, false
		town := s7bWorld(map[string]int{}, map[string]task.QuestState{task.QCain: {Stage: task.StageStarted, Ev: task.EvMemory}})
		for _, id := range []string{"scroll", "stones"} {
			c, err := b.Chunk("search_for_cain", id)
			if err != nil {
				t.Fatal(err)
			}
			if rt.chunkFact(q, c, town).Done {
				t.Fatalf("%s: done on a hero that never did it", id)
			}
			if !rt.chunkFact(q, c, tristram).Done {
				t.Fatalf("%s: not done in Tristram", id)
			}
			if !rt.chunkFact(q, c, town).Done {
				t.Fatalf("%s: read open again in town after it was done", id)
			}
		}
	}
}

// V lab drill CAIN-L3, second live run: the stones chunk comes after the scroll chunk, and the Key to the Cairn Stones
// is used up when the portal opens, so in Tristram the scroll chunk read open again, the stones chunk with it ("all
// steps done but the quest does not read as finished", the board stood in Tristram for good). The earlier chunk seen
// done counts for the later one as well.
func TestCainStonesChunkDoneAfterTheKeyIsUsedUp(t *testing.T) {
	b, err := questboard.Load()
	if err != nil {
		t.Fatal(err)
	}
	q, err := b.Quest("search_for_cain")
	if err != nil {
		t.Fatal(err)
	}
	rt := &boardRT{handed: map[string]bool{}, reached: map[string]bool{}}
	started := map[string]task.QuestState{task.QCain: {Stage: task.StageStarted, Ev: task.EvMemory}}
	withKey := s7bWorld(map[string]int{task.ItemCainSc: 1}, started)
	withKey.Area, withKey.InTown = task.StonyField, false
	scroll, _ := b.Chunk("search_for_cain", "scroll")
	stones, _ := b.Chunk("search_for_cain", "stones")
	if !rt.chunkFact(q, scroll, withKey).Done {
		t.Fatal("scroll not done with the key in the pack")
	}
	tristram := s7bWorld(map[string]int{}, started) // the key is gone
	tristram.Area, tristram.InTown = task.Tristram, false
	if !rt.chunkFact(q, stones, tristram).Done {
		t.Fatal("stones not done in Tristram after the key was used up")
	}
}

// V lab drill CAIN-L3, third live run: the free-Cain chunk (cage, home) read DONE whenever the hero stood outside
// Tristram, because a field step whose area precondition is false counts as "not needed". After the first chicken from
// Tristram the board dropped the chunk (and the report to Akara could never be offered). The chunk is done on the quest
// word (reward pending, record 4 bit 1) and a hero in town, nothing else.
func TestCainFreeChunkIsDoneOnTheQuestWordOnly(t *testing.T) {
	b, err := questboard.Load()
	if err != nil {
		t.Fatal(err)
	}
	q, _ := b.Quest("search_for_cain")
	c, _ := b.Chunk("search_for_cain", "free_cain")
	started := map[string]task.QuestState{task.QCain: {Stage: task.StageStarted, Ev: task.EvMemory}}
	cleared := map[string]task.QuestState{task.QCain: {Stage: task.StageCleared, Ev: task.EvMemory}}
	town := func(qs map[string]task.QuestState) task.World { return s7bWorld(map[string]int{}, qs) }
	field := func(qs map[string]task.QuestState, area task.Area) task.World {
		w := s7bWorld(map[string]int{}, qs)
		w.Area, w.InTown = area, false
		return w
	}
	// the chunks before it are done (the board remembered them), so only its own evidence is in question
	rt := &boardRT{handed: map[string]bool{"search_for_cain/read": true}, reached: map[string]bool{"search_for_cain/scroll/": true, "search_for_cain/stones/": true}}
	for name, w := range map[string]task.World{"town": town(started), "stony field": field(started, task.StonyField), "tristram": field(started, task.Tristram)} {
		if rt.chunkFact(q, c, w).Done {
			t.Errorf("free_cain reads done in %s with Cain still caged", name)
		}
	}
	if rt.chunkFact(q, c, field(cleared, task.Tristram)).Done {
		t.Error("free_cain reads done before the hero is back in camp")
	}
	if !rt.chunkFact(q, c, town(cleared)).Done {
		t.Error("free_cain not done with Cain free and the hero in camp")
	}
}
