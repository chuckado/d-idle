// Package drill is the scripted-scenario runner of the live primitive drills (card m1-liveproof, docs/primitives.md):
// a drill sets a state up through logged cheats, drives the primitive through the bot's real code path, reads the end
// state back and returns PASS, FAIL or BLOCKED with the intended and the observed state. The runner wraps every drill
// in the P35 UI gate (cursor empty, no panel, no dialog before and after), a hard time limit and a repeat count, and
// writes one result line per drill, a table and JSON.
//
// The package knows nothing about d2d: a World (UI read, recovery, logged cheats) is all it needs, so the runner is
// tested offline with a fake. cmd/d2drill supplies the live World and the drills.
package drill

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Verdict is the result of a drill.
type Verdict string

const (
	Pass    Verdict = "PASS"
	Fail    Verdict = "FAIL"
	Blocked Verdict = "BLOCKED"
)

// UI is what the P35 gate sees. The zero value is a clean UI.
type UI struct {
	Cursor string   // code or name of the item on the cursor, "" none
	Panels []string // open panels and dialogs by name
}

// Clean reports an empty cursor and no open panel or dialog.
func (u UI) Clean() bool { return u.Cursor == "" && len(u.Panels) == 0 }

func (u UI) String() string {
	if u.Clean() {
		return "clean"
	}
	var p []string
	if u.Cursor != "" {
		p = append(p, "cursor holds "+u.Cursor)
	}
	if len(u.Panels) > 0 {
		p = append(p, "open: "+strings.Join(u.Panels, ","))
	}
	return strings.Join(p, "; ")
}

// World is the instance under test, as far as the runner needs it.
type World interface {
	// UI reads the cursor and the open panels.
	UI(ctx context.Context) (UI, error)
	// Recover tries to close what the UI gate found (Escape, a click that drops the cursor).
	Recover(ctx context.Context) error
	// Cheat runs one setup cheat (a d2d "debug" verb) and records it with the reason in the cheat ledger. The result is
	// the verb's reply. A cheat without a reason is refused.
	Cheat(ctx context.Context, actor, verb, reason string) (string, error)
}

// Outcome is what a drill body returns.
type Outcome struct {
	Verdict  Verdict
	Reason   string // one line: why this verdict
	Intended string // the end state the pass condition asks for
	Observed string // the end state read from the StateReader
}

// Pass, Fail and Blocked build an Outcome.
func Passed(reason, intended, observed string) Outcome {
	return Outcome{Pass, reason, intended, observed}
}
func Failed(reason, intended, observed string) Outcome {
	return Outcome{Fail, reason, intended, observed}
}
func Blocks(reason string) Outcome { return Outcome{Verdict: Blocked, Reason: reason} }

// Run is the context a drill body gets: the drill id as the cheat actor and the record of the cheats used.
type Run struct {
	W      World
	Drill  string
	Cheats []string
}

// Cheat runs a logged setup cheat; reason is required. The verb and reason are kept for the result.
func (r *Run) Cheat(ctx context.Context, verb, reason string) (string, error) {
	if strings.TrimSpace(reason) == "" {
		return "", fmt.Errorf("drill %s: cheat %q without a reason", r.Drill, verb)
	}
	out, err := r.W.Cheat(ctx, "drills:"+r.Drill, verb, reason)
	r.Cheats = append(r.Cheats, verb)
	return out, err
}

// Drill is one scripted scenario.
type Drill struct {
	ID        string // "D1.P1"
	Primitive string // "P1"
	Title     string
	// Needs lists missing cheats or verbs. A drill with a missing piece is BLOCKED without running, naming the piece.
	Needs []string
	// Limit is the hard time limit of one run (0: the runner default).
	Limit time.Duration
	// Do runs the drill once. It may use the cheats through r.Cheat only.
	Do func(ctx context.Context, r *Run) Outcome
}

// RunResult is one repeat of a drill.
type RunResult struct {
	Outcome
	Cheats  []string      `json:"cheats,omitempty"`
	Elapsed time.Duration `json:"elapsed_ns"`
	UIPre   string        `json:"ui_pre"`
	UIPost  string        `json:"ui_post"`
}

// Result is the verdict of a drill over all its repeats.
type Result struct {
	ID            string      `json:"id"`
	Primitive     string      `json:"primitive"`
	Title         string      `json:"title,omitempty"`
	Verdict       Verdict     `json:"verdict"`
	Reason        string      `json:"reason"`
	Intended      string      `json:"intended"`
	Observed      string      `json:"observed"`
	Runs          []RunResult `json:"runs,omitempty"`
	Deterministic bool        `json:"deterministic"`
	Cheats        []string    `json:"cheats,omitempty"`
}

// Line is the one-line form: DRILL <id> <primitive> PASS|FAIL|BLOCKED <reason> <intended vs observed>.
func (r Result) Line() string {
	ev := ""
	if r.Intended != "" || r.Observed != "" {
		ev = fmt.Sprintf(" | intended: %s | observed: %s", r.Intended, r.Observed)
	}
	det := ""
	if len(r.Runs) > 1 {
		det = " | runs: " + fmt.Sprint(len(r.Runs)) + ", " + map[bool]string{true: "identical", false: "DIFFER"}[r.Deterministic]
	}
	return fmt.Sprintf("DRILL %s %s %s %s%s%s", r.ID, r.Primitive, r.Verdict, oneLine(r.Reason), ev, det)
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }
