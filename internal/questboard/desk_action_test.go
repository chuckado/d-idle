package questboard

import "testing"

// Every board state either offers one button or the row says why not (Reason/Why); the awaiting chunk without a record
// (case C5, no `awaiting` object in the view) offers Start, the one with a run offers Hand in (user bug 2026-11-21).
func TestChunkActionEveryState(t *testing.T) {
	startable := &OfferView{Startable: true}
	hand := &ReasonView{Code: CodeHandInReady}
	run := &Run{Status: RunAwaiting}
	for _, c := range []struct {
		name   string
		st     State
		off    *OfferView
		run    *Run
		reason *ReasonView
		paused bool
		verb   string // "": no action
		off0   bool
	}{
		{name: "locked", st: Locked},
		{name: "available", st: Available},
		{name: "ready", st: Ready, off: startable, verb: "start"},
		{name: "ready not startable", st: Ready, off: &OfferView{}},
		{name: "accepted", st: Accepted, run: &Run{Status: RunRunning}, verb: "cancel"},
		{name: "in progress", st: InProgress, verb: "cancel"},
		{name: "awaiting with run", st: Awaiting, run: run, verb: "handin"},
		{name: "awaiting no run, no object", st: Awaiting, verb: "handin"},
		{name: "awaiting handin_ready", st: Awaiting, reason: hand, verb: "start"},
		{name: "awaiting handin_ready paused", st: Awaiting, reason: hand, paused: true, verb: "start", off0: true},
		{name: "handing in", st: Awaiting, run: &Run{Status: RunHandingIn}, verb: "handin", off0: true},
		{name: "blocked", st: Blocked, verb: "retry"},
		{name: "done", st: Done},
		{name: "skipped", st: Skipped},
	} {
		a := chunkAction(c.st, c.off, c.run, c.reason, c.paused)
		if c.verb == "" {
			if a != nil {
				t.Errorf("%s: unexpected action %+v", c.name, a)
			}
			continue
		}
		if a == nil || a.Verb != c.verb || a.Label == "" {
			t.Errorf("%s: action %+v, want verb %s", c.name, a, c.verb)
			continue
		}
		if a.Disabled != c.off0 || a.Disabled && a.Why == "" {
			t.Errorf("%s: disabled %v why %q", c.name, a.Disabled, a.Why)
		}
	}
	for _, s := range AllStates { // a new state must be added to the table above
		if s != Locked && s != Available && s != Ready && s != Accepted && s != InProgress && s != Awaiting && s != Done && s != Skipped && s != Blocked {
			t.Errorf("state %s has no row in the table", s)
		}
	}
}
