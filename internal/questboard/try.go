package questboard

import "fmt"

// The owner's "try anyway" start (docs/questboard.md, ADR 0030 amended 2026-11-24): a deliberate, audited override of the
// READINESS refusal of one chunk for one attempt. It skips only the readiness words and floors (the board word "too hard
// for now", level and life floors, the Andariel hold); kit, safety and every combat reaction stay. It is not sticky: the
// try ends when the run parks, dies, finishes or is stopped, or when the process restarts.

// ErrNotOverridable is the refusal of a try-anyway start for a chunk whose refusal is not a readiness one.
type ErrNotOverridable struct{ State State }

func (e ErrNotOverridable) Error() string {
	return fmt.Sprintf("questboard: try anyway only overrides a readiness refusal; this part is %s", e.State)
}

// canStart is CanStart with the one-chunk readiness override of AcceptTry.
func (e *Exec) canStart(vs []QuestView, quest, chunk string) error {
	if e.tryNow {
		if _, cur, ok := chunkView(vs, quest, chunk); ok && cur.State == Available {
			return nil
		}
	}
	return CanStart(vs, quest, chunk)
}

// AcceptTry starts an Available chunk past its readiness refusal. It returns the refused reason (for the audit). The chunk
// must read Available (a readiness word or a floor); a locked, done, parked or running chunk is not overridable.
func (e *Exec) AcceptTry(quest, chunk string) (refused string, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.o.Enabled {
		return "", ErrOff
	}
	vs, plan, diff, derr := e.derive()
	if derr != nil {
		return "", derr
	}
	_, cur, ok := chunkView(vs, quest, chunk)
	if !ok {
		return "", fmt.Errorf("questboard: quest %q has no chunk %q", quest, chunk)
	}
	if cur.State != Available {
		return "", ErrNotOverridable{cur.State}
	}
	refused = cur.Why
	if refused == "" {
		refused = "readiness refused"
	}
	e.lastChanged = nil
	e.tryNow = true
	defer func() { e.tryNow = false }()
	if err := e.acceptLocked(vs, plan, diff, quest, chunk, nil); err != nil {
		return refused, err
	}
	e.try = ChunkKey(quest, chunk)
	e.o.Log("override: try anyway", "chunk", e.try, "refused", refused)
	return refused, nil
}

// TryKey is the chunk key of the try-anyway run in force ("" = none).
func (e *Exec) TryKey() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.try
}

// tryCheck ends the try when the head decision is no longer a live run of the chunk (parked, done, stopped, replaced).
func (e *Exec) tryCheck(d Decision, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.try == "" || err != nil {
		return // a board that cannot be read now says nothing about the run
	}
	live := ChunkKey(d.Quest, d.Chunk) == e.try && (d.Kind == DecRun || d.Kind == DecAwait || d.Kind == DecBusy)
	if (d.Kind == DecBusy || d.Kind == DecPaused) && d.Quest == "" {
		live = true // a hand-in dialogue in flight or a pause: the decision carries no chunk
	}
	if !live {
		e.o.Log("override: try anyway ended", "chunk", e.try, "decision", string(d.Kind))
		e.try = ""
	}
}
