package questboard

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// The chunked multi-part quests (slice S7). Tags: V verified by the test, U unverified live.

func perms(s []string) [][]string {
	if len(s) <= 1 {
		return [][]string{append([]string(nil), s...)}
	}
	var out [][]string
	for i := range s {
		rest := append(append([]string(nil), s[:i]...), s[i+1:]...)
		for _, p := range perms(rest) {
			out = append(out, append([]string{s[i]}, p...))
		}
	}
	return out
}

const (
	will  = "khalims_will"
	strt  = "search_for_cain"
	hill  = "horadric_staff"
	third = "terrors_end"
)

// Every order of Eye, Brain and Heart works on two builds; the Will waits for all three and the Flail, the orb
// for the Will (V).
func TestKhalimPiecesInEveryOrder(t *testing.T) {
	for _, bd := range []string{"Paladin", "Sorceress"} {
		for _, order := range perms([]string{"eye", "brain", "heart"}) {
			t.Run(bd+"/"+strings.Join(order, ","), func(t *testing.T) {
				r := newRig(t, bd, 40)
				r.f.DoneQuest["act2_done"] = TierSave
				for i, p := range order {
					off, err := r.x.Offers()
					if err != nil {
						t.Fatal(err)
					}
					for _, rest := range order[i:] {
						if !offered(off, will, rest) {
							t.Fatalf("after %v: %s is not offered", order[:i], rest)
						}
					}
					if offered(off, will, "assemble") {
						t.Fatalf("assemble offered with only %v done", order[:i])
					}
					if err := r.x.Accept(will, p); err != nil {
						t.Fatalf("%s: %v", p, err)
					}
					r.tick(2)
					if got := r.chunk(will, p).State; got != Done {
						t.Fatalf("%s after its steps: %s", p, got)
					}
				}
				if c := r.chunk(will, "assemble"); c.State != Locked || !strings.Contains(c.Why, "Blackened Temple") {
					t.Fatalf("assemble without the Flail quest: %s %q", c.State, c.Why)
				}
				r.f.DoneQuest["blackened_temple"] = TierSave
				if err := r.x.Accept(will, "assemble"); err != nil {
					t.Fatal(err)
				}
				r.tick(2)
				if c := r.chunk(will, "orb"); c.State != Ready {
					t.Fatalf("orb after the Will: %s %q", c.State, c.Why)
				}
				if err := r.x.Accept(will, "orb"); err != nil {
					t.Fatal(err)
				}
				r.tick(3)
				if got := r.chunk(will, "orb").State; got != Done {
					t.Fatalf("orb: %s", got)
				}
			})
		}
	}
}

// A restart between any two pieces re-derives the same offers: nothing of the chain is stored (V).
func TestKhalimRestartBetweenPieces(t *testing.T) {
	for _, order := range perms([]string{"eye", "brain", "heart"}) {
		r := newRig(t, "Paladin", 40)
		r.f.DoneQuest["act2_done"] = TierSave
		for i, p := range order {
			if err := r.x.Accept(will, p); err != nil {
				t.Fatal(err)
			}
			r.tick(1)
			r.restart() // half way through the piece
			if got := r.chunk(will, p).State; got != InProgress && got != Accepted {
				t.Fatalf("%v: %s after a restart mid-piece: %s", order, p, got)
			}
			r.tick(1)
			r.restart() // after it
			if got := r.chunk(will, p).State; got != Done {
				t.Fatalf("%v: %s after finishing: %s", order, p, got)
			}
			if i < 2 {
				off, _ := r.x.Offers()
				for _, rest := range order[i+1:] {
					if !offered(off, will, rest) {
						t.Fatalf("%v: %s lost its offer after a restart", order, rest)
					}
				}
			}
		}
	}
}

// The Staff: the Cube and the Staff of Kings in any order, then the assembly with its Cain hand-in (S7b, two builds, V).
// The assembly needs the Tainted Sun quest too (the amulet).
func TestStaffPiecesAndAssemblyWithCain(t *testing.T) {
	for _, bd := range []string{"Sorceress", "Paladin"} {
		for _, order := range perms([]string{"cube", "shaft"}) {
			r := newRig(t, bd, 40)
			r.f.DoneQuest["act1_done"] = TierSave
			r.f.DoneQuest["radament"] = TierSave
			for _, p := range order {
				if err := r.x.Accept(hill, p); err != nil {
					t.Fatalf("%s %v %s: %v", bd, order, p, err)
				}
				r.tick(3)
				if got := r.chunk(hill, p).State; got != Done {
					t.Fatalf("%s %v %s: %s", bd, order, p, got)
				}
			}
			if cv := r.chunk(hill, "assemble"); cv.State != Locked {
				t.Fatalf("%s: assemble before the amulet quest = %s %q", bd, cv.State, cv.Why)
			}
			r.f.DoneQuest["tainted_sun"] = TierSave
			cv := r.chunk(hill, "assemble")
			if cv.State != Ready || strings.Contains(cv.Why, "not supported yet") {
				t.Fatalf("%s: assemble = %s %q", bd, cv.State, cv.Why)
			}
			if err := r.x.Accept(hill, "assemble"); err != nil {
				t.Fatalf("%s: Start on the assembly: %v", bd, err)
			}
		}
	}
}

// The Cain chain is strictly ordered and its last part, the report to Akara, is a hand-in that is offered once the cage
// part reads done (S7b, two builds, V); nothing is offered before.
func TestCainChainInOrder(t *testing.T) {
	b := mustBoard(t)
	for _, bd := range []string{"Paladin", "Sorceress"} {
		ready := readinessFor(t, b, bd, 40)
		in := Input{Facts: facts{0: {}}, Ready: ready, Chunks: map[string]ChunkFact{}}
		get := func() QuestView { return view(t, b, in)[strt] }
		if v := get(); !reflect.DeepEqual(v.Offered, []string{"scroll"}) {
			t.Fatalf("%s start: %v", bd, v.Offered)
		}
		in.Chunks[ChunkKey(strt, "scroll")] = ChunkFact{Done: true, Tier: TierProxy}
		if v := get(); !reflect.DeepEqual(v.Offered, []string{"read"}) {
			t.Fatalf("%s after the scroll: %v", bd, v.Offered)
		}
		in.Chunks[ChunkKey(strt, "read")] = ChunkFact{Done: true, Tier: TierProxy}
		if v := get(); !reflect.DeepEqual(v.Offered, []string{"free_cain"}) {
			t.Fatalf("%s after the reading: %v", bd, v.Offered)
		}
		in.Chunks[ChunkKey(strt, "free_cain")] = ChunkFact{Done: true, Tier: TierProxy}
		v := get()
		if !reflect.DeepEqual(v.Offered, []string{"report"}) {
			t.Fatalf("%s after the cage: %v", bd, v.Offered)
		}
		if c := chunk(v, "report"); strings.Contains(c.Why, "not supported yet") {
			t.Fatalf("%s report = %s %q", bd, c.State, c.Why)
		}
	}
}

// The Golden Bird is one hand-in chunk, supported; the board does not claim to fetch the figurine (S7b, V).
func TestGoldenBirdIsSupported(t *testing.T) {
	b := mustBoard(t)
	for _, bd := range []string{"Paladin", "Sorceress"} {
		in := Input{Facts: facts{0: {"act2_done": done(TierSave)}}, Ready: readinessFor(t, b, bd, 40)}
		v := view(t, b, in)["golden_bird"]
		if c := chunk(v, "bird"); strings.Contains(c.Why, "not supported yet") {
			t.Fatalf("%s bird = %s %q", bd, c.State, c.Why)
		}
	}
}

// Terror's End: the seals are unsupported on their own and never block Diablo (V); leaving mid-seals may reset (U).
func TestTerrorsEndSealsDoNotBlockDiablo(t *testing.T) {
	b := mustBoard(t)
	in := Input{Facts: facts{0: {"izual": done(TierSave)}}, Ready: readinessFor(t, b, "Sorceress", 40)}
	v := view(t, b, in)[third]
	if c := chunk(v, "seals"); c.State != Available || !strings.Contains(c.Why, "not supported yet") {
		t.Fatalf("seals = %s %q", c.State, c.Why)
	}
	if !reflect.DeepEqual(v.Offered, []string{"diablo"}) {
		t.Fatalf("offered = %v", v.Offered)
	}
}

// Evidence conflicts: the T3 proxy never unlocks a quest prerequisite, the quest word beats a part's proxy, and a
// difficulty change starts the chain again (V).
func TestChunkEvidenceConflictsAndDifficulty(t *testing.T) {
	b := mustBoard(t)
	pieces := map[string]ChunkFact{}
	for _, p := range []string{"eye", "brain", "heart"} {
		pieces[ChunkKey(will, p)] = ChunkFact{Done: true, Tier: TierProxy}
	}
	base := facts{0: {"act2_done": done(TierSave), "blackened_temple": done(TierProxy)}}
	in := Input{Facts: base, Chunks: pieces, Ready: readinessFor(t, b, "Paladin", 40)}
	if c := chunk(view(t, b, in)[will], "assemble"); c.State != Locked {
		t.Fatalf("a proxy-only Flail quest unlocked the Will: %s", c.State)
	}
	base[0]["blackened_temple"] = done(TierMemory)
	if c := chunk(view(t, b, in)[will], "assemble"); c.State != Ready {
		t.Fatalf("T2 Flail quest: %s %q", c.State, c.Why)
	}
	// the quest word (T1) marks every part done without any part evidence
	base[0][will] = done(TierSave)
	in.Chunks = nil
	v := view(t, b, in)[will]
	for _, c := range v.Chunks {
		if c.State != Done {
			t.Errorf("%s with the quest word T1: %s", c.ID, c.State)
		}
	}
	// Nightmare: the normal word and the normal record do not carry over
	in.Diff = 1
	in.Plan = Plan{Accepted: []Record{rec(will, "eye", 0)}}
	nm := view(t, b, in)[will]
	if c := chunk(nm, "eye"); c.State == Done || c.State == Accepted {
		t.Errorf("eye in nightmare: %s", c.State)
	}
}

// The prerequisite graph over quests, chunk requires and chunk order has no cycle, whole board (V).
func TestWholeBoardGraphIsAcyclic(t *testing.T) {
	b := mustBoard(t)
	edges := map[string][]string{}
	for _, q := range b.Quests() {
		for _, r := range q.Requires {
			edges["q:"+q.ID] = append(edges["q:"+q.ID], "q:"+r)
		}
		for _, c := range q.Chunks {
			n := "c:" + q.ID + "/" + c.ID
			edges[n] = append(edges[n], "q:"+q.ID)
			for _, r := range c.Requires {
				edges[n] = append(edges[n], "q:"+r)
			}
			for _, a := range c.After {
				edges[n] = append(edges[n], "c:"+q.ID+"/"+a)
			}
		}
		for _, s := range q.SoftAfter {
			edges["q:"+q.ID] = append(edges["q:"+q.ID], "q:"+s)
		}
	}
	state := map[string]int{}
	var visit func(n string, path []string)
	visit = func(n string, path []string) {
		switch state[n] {
		case 1:
			t.Fatalf("cycle: %s", strings.Join(append(path, n), " -> "))
		case 2:
			return
		}
		state[n] = 1
		for _, m := range edges[n] {
			visit(m, append(path, n))
		}
		state[n] = 2
	}
	for n := range edges {
		visit(n, nil)
	}
}

// Every chunk of every act: steps named, handler fields consistent, an unsupported reason that says why, and no
// supported chunk waiting behind an unsupported one inside its own quest (a dead chain) (V).
func TestEveryChunkIsWellFormed(t *testing.T) {
	b := mustBoard(t)
	for _, q := range b.Quests() {
		unsupported := map[string]bool{}
		for _, c := range q.Chunks {
			at := fmt.Sprintf("%s/%s", q.ID, c.ID)
			if len(c.Steps) == 0 {
				t.Errorf("%s: no steps", at)
			}
			if c.Unsupported != "" {
				unsupported[c.ID] = true
				if len(c.Unsupported) < 20 {
					t.Errorf("%s: unsupported reason %q says nothing", at, c.Unsupported)
				}
				continue
			}
			if len(q.Chunks) > 1 && c.Handler == "" && len(c.Nodes) == 0 && c.HandInStep == "" {
				t.Errorf("%s: a part of a multi-chunk quest over the shared handler lists no nodes or handInStep", at)
			}
		}
		for _, c := range q.Chunks {
			for _, a := range c.After {
				if unsupported[a] && c.Unsupported == "" {
					t.Errorf("%s/%s comes after the unsupported %s: a dead chain", q.ID, c.ID, a)
				}
			}
		}
	}
}

// Group and class tables: the same data answers two builds (no class literal in Go).
func TestHandlerOfDefaults(t *testing.T) {
	b := mustBoard(t)
	q, _ := b.Quest("sisters_burial_grounds")
	if ref := q.HandlerOf(q.Chunks[0]); ref.Group != "sisters_burial" || !ref.All || ref.HandIn != "reward" {
		t.Fatalf("burial = %+v", ref)
	}
	q, _ = b.Quest(strt)
	c, _ := b.Chunk(strt, "scroll")
	if ref := q.HandlerOf(c); ref.All || len(ref.Field) != 3 {
		t.Fatalf("cain scroll = %+v", ref)
	}
	q, _ = b.Quest(will)
	c, _ = b.Chunk(will, "eye")
	if ref := q.HandlerOf(c); ref.Group != "khalims_eye" || !ref.All {
		t.Fatalf("eye = %+v", ref)
	}
}
