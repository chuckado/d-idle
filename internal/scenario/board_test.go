package scenario

import "testing"

func boardState(samples ...BoardSample) State {
	return State{Has: []string{"board"}, Board: samples}
}

func ch(q, c, s, r string) BoardChunk { return BoardChunk{Quest: q, Chunk: c, State: s, Reason: r} }

func TestBoardAssertWhen(t *testing.T) {
	st := boardState(
		BoardSample{T: 1, Chunks: []BoardChunk{ch("q", "a", "accepted", "waits_for")}},
		BoardSample{T: 2, Chunks: []BoardChunk{ch("q", "a", "in_progress", "")}},
	)
	run := func(a Assert) AssertResult { a.Kind, a.Quest, a.Chunk = ABoard, "q", "a"; return evalOne(a, st) }
	if r := run(Assert{State: "in_progress"}); !r.OK {
		t.Fatalf("end: %+v", r)
	}
	if r := run(Assert{State: "accepted", Or: []string{"in_progress"}, When: BoardAlways}); !r.OK {
		t.Fatalf("always with or: %+v", r)
	}
	if r := run(Assert{State: "blocked", When: BoardNever}); !r.OK {
		t.Fatalf("never: %+v", r)
	}
	if r := run(Assert{State: "accepted", When: BoardNever}); r.OK {
		t.Fatalf("never must fail when a sample matches: %+v", r)
	}
	if r := run(Assert{State: "accepted", Reason: "waits_for", When: BoardNever}); r.OK {
		t.Fatalf("reason match: %+v", r)
	}
	// a chunk that is not in the sample reads as the empty state
	if r := evalOne(Assert{Kind: ABoard, Quest: "q", Chunk: "zz", State: "in_progress"}, st); r.OK {
		t.Fatalf("absent chunk matched: %+v", r)
	}
	if r := evalOne(Assert{Kind: ABoard, Quest: "q", Chunk: "a", State: "done"}, State{}); !r.Unobserved {
		t.Fatalf("no board must be unobserved: %+v", r)
	}
}

func TestBoardHeadsLE(t *testing.T) {
	two := boardState(BoardSample{Chunks: []BoardChunk{ch("q", "a", "in_progress", ""), ch("q", "b", "awaiting_handin", "")}})
	one := 1
	if r := evalOne(Assert{Kind: ABoardHeadsLE, Max: &one}, two); r.OK {
		t.Fatalf("two heads passed: %+v", r)
	}
	two.Board[0].Chunks[1].State = "accepted"
	if r := evalOne(Assert{Kind: ABoardHeadsLE, Max: &one}, two); !r.OK {
		t.Fatalf("one head failed: %+v", r)
	}
}

func TestAPIReply(t *testing.T) {
	yes, no := true, false
	zero, one := 0, 1
	st := State{APIReplies: []APIReply{{Index: 0, Status: 200, OK: true, Code: "", Changed: []string{"q/a", "q/b"}}, {Index: 1, Status: 409, Code: "busy"}}}
	cases := []struct {
		a  Assert
		ok bool
	}{
		{Assert{Index: &zero, Status: 200, OKIs: &yes, AlreadyIs: &no, ChangedHas: []string{"q/b"}}, true},
		{Assert{Index: &zero, ChangedHas: []string{"q/c"}}, false},
		{Assert{Index: &one, OKIs: &no, Code: "busy"}, true},
		{Assert{Index: &one, Code: "done"}, false},
	}
	for i, c := range cases {
		c.a.Kind = AAPIReply
		if r := evalOne(c.a, st); r.OK != c.ok {
			t.Fatalf("case %d: %+v", i, r)
		}
	}
	two := 2
	if r := evalOne(Assert{Kind: AAPIReply, Index: &two}, st); !r.Unobserved {
		t.Fatalf("missing reply: %+v", r)
	}
}

func TestBoardAssertValidate(t *testing.T) {
	bad := []Assert{{Kind: ABoard, Quest: "q", Chunk: "a"}, {Kind: ABoard, Quest: "q", Chunk: "a", State: "x", When: "sometimes"}, {Kind: ABoardHeadsLE}, {Kind: AAPIReply}}
	for i, a := range bad {
		if a.Validate() == nil {
			t.Fatalf("case %d validated", i)
		}
	}
}
