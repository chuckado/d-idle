package cheat

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// RunEnv names the measured run an entry belongs to. The memory rule (docs/adr/0014, docs/cheats.md):
// writes after a measured run starts make it CHEATED. MarkRun writes the start marker; WritesAfterStart
// counts what came after it, for the "writes after start: N" line of a soak or gate record.
const RunEnv = "D2BOT_RUN"

// KindRunStart is a ledger marker, not a write (it touches no game state).
const KindRunStart Kind = "runstart"

// MarkRun appends the run-start marker for run. Setup writes belong BEFORE it.
func (l *Ledger) MarkRun(run, actor, instance, reason string, now time.Time) error {
	if strings.TrimSpace(run) == "" {
		return errors.New("cheat: run id required")
	}
	return l.Append(Entry{Time: now.UTC(), Actor: actor, Instance: instance, Kind: KindRunStart,
		Reason: reason, Success: true, Run: run})
}

// WritesAfterStart counts successful writes after the latest start marker of run (file order), on the
// marker's instance when it names one. Refused and failed attempts are not writes. An unknown run is an
// error: "no marker" must never read as "0 writes".
func (l *Ledger) WritesAfterStart(run string) (int, []Entry, error) {
	es, err := l.List(Filter{})
	if err != nil {
		return 0, nil, err
	}
	start := -1
	for i, e := range es {
		if e.Kind == KindRunStart && e.Run == run {
			start = i
		}
	}
	if start < 0 {
		return 0, nil, fmt.Errorf("cheat: no run-start marker for run %q", run)
	}
	var out []Entry
	for _, e := range es[start+1:] {
		if e.Kind == KindRunStart || !e.Success {
			continue
		}
		if inst := es[start].Instance; inst != "" && e.Instance != "" && e.Instance != inst {
			continue
		}
		out = append(out, e)
	}
	return len(out), out, nil
}
