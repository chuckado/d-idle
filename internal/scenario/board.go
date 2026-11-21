package scenario

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Assertion kinds that read the quest board of the watch API (card bsm1, docs/design/fable-review-2026-11-21 appendix B.3).
const (
	ABoard        = "board"           // quest, chunk, state (+ or), reason (+ reason_or), when end|always|never
	ABoardHeadsLE = "board_heads_le"  // max: in every sample at most max chunks are in_progress or awaiting_handin
	ABoardReasons = "board_reasons"   // no args: in every sample each accepted, awaiting_handin or blocked chunk has a reason code (I5)
	AAPIReply     = "api_reply"       // index, status, ok, already, code, changed_has: the recorded reply of trigger.api[index]
	BoardWhenEnd  = "end"             // the last sample (default)
	BoardAlways   = "always"          // every sample
	BoardNever    = "never"           // no sample
	boardLiveA    = "in_progress"     // wire state names (internal/questboard)
	boardLiveB    = "awaiting_handin" // the second state a head run can be in
)

// BoardChunk is one chunk row of GET /api/quests: Reason is the reason code of the view (empty until the board has them).
type BoardChunk struct {
	Quest  string `json:"quest"`
	Chunk  string `json:"chunk"`
	State  string `json:"state"`
	Reason string `json:"reason,omitempty"`
}

// BoardSample is the board at one moment of the run; T is virtual minutes by the same linear map as LogLine.T.
type BoardSample struct {
	T      float64      `json:"t"`
	Chunks []BoardChunk `json:"chunks"`
}

// APIReply is the recorded answer to one trigger.api call. Changed lists "quest/chunk" of the runs the reply says it touched.
type APIReply struct {
	Index   int      `json:"index"`
	Path    string   `json:"path"`
	Status  int      `json:"status"`
	OK      bool     `json:"ok"`
	Already bool     `json:"already,omitempty"`
	Code    string   `json:"code,omitempty"`
	Reason  string   `json:"reason,omitempty"`
	Changed []string `json:"changed,omitempty"`
}

// MapSampleTimes converts the wall times of samples into virtual minutes over the span [first,last] of the bot log.
func MapSampleTimes(at []time.Time, first, last time.Time, virtualMin float64) []float64 {
	out := make([]float64, len(at))
	span := last.Sub(first)
	for i, t := range at {
		if span > 0 {
			out[i] = float64(t.Sub(first)) / float64(span) * virtualMin
		}
	}
	return out
}

func (a Assert) validateBoard() error {
	switch a.Kind {
	case ABoard:
		if a.Quest == "" || a.Chunk == "" || a.State == "" && a.Reason == "" {
			return fmt.Errorf("board needs quest, chunk and state or reason")
		}
		switch a.When {
		case "", BoardWhenEnd, BoardAlways, BoardNever:
		default:
			return fmt.Errorf("board when %q", a.When)
		}
	case ABoardHeadsLE:
		if a.Max == nil {
			return fmt.Errorf("board_heads_le needs max")
		}
	case AAPIReply:
		if a.Index == nil || *a.Index < 0 {
			return fmt.Errorf("api_reply needs index")
		}
	}
	return nil
}

func boardMatch(a Assert, c BoardChunk) bool {
	if a.State != "" && c.State != a.State && !contains(a.Or, c.State) {
		return false
	}
	if a.Reason != "" && c.Reason != a.Reason && !contains(a.ReasonOr, c.Reason) {
		return false
	}
	return true
}

func evalBoard(a Assert, st State) AssertResult {
	when := a.When
	if when == "" {
		when = BoardWhenEnd
	}
	want := fmt.Sprintf("%s/%s state %s", a.Quest, a.Chunk, strings.Join(append([]string{a.State}, a.Or...), " or "))
	if a.Reason != "" {
		want += " reason " + strings.Join(append([]string{a.Reason}, a.ReasonOr...), " or ")
	}
	want += " (" + when + ")"
	if !st.has("board") || len(st.Board) == 0 {
		return unobs(want, "the quest board")
	}
	find := func(s BoardSample) BoardChunk {
		for _, c := range s.Chunks {
			if c.Quest == a.Quest && c.Chunk == a.Chunk {
				return c
			}
		}
		return BoardChunk{Quest: a.Quest, Chunk: a.Chunk}
	}
	hits := 0
	for _, s := range st.Board {
		if boardMatch(a, find(s)) {
			hits++
		}
	}
	last := find(st.Board[len(st.Board)-1])
	got := fmt.Sprintf("last state %q reason %q, %d of %d samples match", last.State, last.Reason, hits, len(st.Board))
	switch when {
	case BoardAlways:
		return AssertResult{OK: hits == len(st.Board), Want: want, Got: got}
	case BoardNever:
		return AssertResult{OK: hits == 0, Want: want, Got: got}
	}
	return AssertResult{OK: boardMatch(a, last), Want: want, Got: got}
}

func evalBoardHeads(a Assert, st State) AssertResult {
	want := fmt.Sprintf("at most %d head chunks in every sample", *a.Max)
	if !st.has("board") || len(st.Board) == 0 {
		return unobs(want, "the quest board")
	}
	worst := 0
	for _, s := range st.Board {
		n := 0
		for _, c := range s.Chunks {
			if c.State == boardLiveA || c.State == boardLiveB {
				n++
			}
		}
		if n > worst {
			worst = n
		}
	}
	return AssertResult{OK: worst <= *a.Max, Want: want, Got: fmt.Sprintf("at most %d", worst)}
}

// evalBoardReasons is the sampled form of I5: a live chunk that does not move says why. A running head (in_progress)
// needs none; the page and the checker treat a hold as a reason.
func evalBoardReasons(st State) AssertResult {
	want := "every accepted, awaiting or blocked chunk has a reason code in every sample"
	if !st.has("board") || len(st.Board) == 0 {
		return unobs(want, "the quest board")
	}
	for i, s := range st.Board {
		for _, c := range s.Chunks {
			switch c.State {
			case "accepted", boardLiveB, "blocked":
				if c.Reason == "" {
					return AssertResult{Want: want, Got: fmt.Sprintf("sample %d: %s/%s is %s with no reason", i, c.Quest, c.Chunk, c.State)}
				}
			}
		}
	}
	return AssertResult{OK: true, Want: want, Got: fmt.Sprintf("%d samples", len(st.Board))}
}

func evalAPIReply(a Assert, st State) AssertResult {
	want := fmt.Sprintf("reply %d", *a.Index)
	if a.Status != 0 {
		want += fmt.Sprintf(" status %d", a.Status)
	}
	if a.OKIs != nil {
		want += fmt.Sprintf(" ok %v", *a.OKIs)
	}
	if a.AlreadyIs != nil {
		want += fmt.Sprintf(" already %v", *a.AlreadyIs)
	}
	if a.Code != "" {
		want += " code " + a.Code
	}
	if len(a.ChangedHas) > 0 {
		want += " changed has " + strings.Join(a.ChangedHas, ", ")
	}
	var r *APIReply
	for i := range st.APIReplies {
		if st.APIReplies[i].Index == *a.Index {
			r = &st.APIReplies[i]
		}
	}
	if r == nil {
		return unobs(want, "the api reply")
	}
	ok := (a.Status == 0 || r.Status == a.Status) && (a.OKIs == nil || r.OK == *a.OKIs) &&
		(a.AlreadyIs == nil || r.Already == *a.AlreadyIs) && (a.Code == "" || r.Code == a.Code)
	for _, w := range a.ChangedHas {
		ok = ok && contains(r.Changed, w)
	}
	return AssertResult{OK: ok, Want: want, Got: fmt.Sprintf("status %d ok %v already %v code %q changed %v", r.Status, r.OK, r.Already, r.Code, r.Changed)}
}

// LogSpan returns the time of the first and last bot log line (the span the log's linear time map uses).
func LogSpan(log []byte) (first, last time.Time, ok bool) {
	for _, l := range bytes.Split(log, []byte("\n")) {
		if len(l) == 0 || l[0] != '{' {
			continue
		}
		var m struct {
			Time string `json:"time"`
		}
		if json.Unmarshal(l, &m) != nil {
			continue
		}
		t, err := time.Parse(time.RFC3339Nano, m.Time)
		if err != nil {
			continue
		}
		if !ok {
			first, ok = t, true
		}
		last = t
	}
	return first, last, ok && last.After(first)
}

// NeedsBoardSamples is true when an assertion reads the board samples, so the runner must start the watch server and
// the sampler even for a scenario with no api calls and no tapper.
func NeedsBoardSamples(s Scenario) bool {
	for _, a := range s.Assert {
		if a.Kind == ABoard || a.Kind == ABoardHeadsLE || a.Kind == ABoardReasons {
			return true
		}
	}
	return false
}
