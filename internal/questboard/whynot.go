package questboard

import "time"

// The "Why not?" row of a chunk (Fable review 2026-11-25, P7): for a chunk that is not startable or does not move, the
// readiness word, the floor, what it needs, what wakes it and when it was last refused. Data only: every field comes from
// the readiness answer, the typed refusal the bot holds and the desk's own refusal log, never from a per-quest branch.

// RefusalFacts is the hero's freshest typed refusal (need code, the fact that ends it, the admission words).
type RefusalFacts struct {
	Need      string    // closed need code (bot needTable), "" when none
	Wake      string    // the fact that ends the refusal
	Admission string    // the admission or guard line in words
	Seen      time.Time // when the refusal was last seen
}

// WhyNotView is the row.
type WhyNotView struct {
	Word        Word       `json:"word,omitempty"`        // readiness word
	Floor       int        `json:"floor,omitempty"`       // legacy clvl floor, 0 = none
	Need        string     `json:"need,omitempty"`        // need code (typed refusal) or the first readiness reason code
	Wake        string     `json:"wake,omitempty"`        // the fact that ends it
	Admission   string     `json:"admission,omitempty"`   // the guard words
	LastRefused *time.Time `json:"lastRefused,omitempty"` // last time a start or the bot refused this chunk
}

// noteRefusal remembers when a start was last refused for the chunk.
func (d *Desk) noteRefusal(quest, chunk string, r Reply) Reply {
	if r.OK || r.Code == "" || chunk == "" && r.Chunk == "" {
		return r
	}
	if chunk == "" {
		chunk = r.Chunk
	}
	d.mu.Lock()
	if d.refused == nil {
		d.refused = map[string]time.Time{}
	}
	d.refused[ChunkKey(quest, chunk)] = d.cfg.Now()
	d.mu.Unlock()
	return r
}

// whyNot builds the row for a chunk that carries a reason or a why; nil for one that moves or is done.
func (d *Desk) whyNot(q Quest, c Chunk, cv ChunkView, cr ChunkRow, rf *RefusalFacts) *WhyNotView {
	if cv.State == Done || cv.State == Skipped || cr.Why == "" && cr.Reason == nil && cr.Readiness == "" {
		return nil
	}
	w := &WhyNotView{Word: cr.Readiness}
	if d.cfg.Explain != nil {
		if rd, ok := d.cfg.Explain(q, c); ok {
			w.Floor = rd.LegacyFloor
			if len(rd.Reasons) > 0 {
				w.Need = rd.Reasons[0].Code
			}
		}
	}
	live := cv.State == Accepted || cv.State == InProgress || cv.State == Awaiting || cv.State == Blocked
	if live && rf != nil {
		if rf.Need != "" {
			w.Need = rf.Need
		}
		w.Wake, w.Admission = rf.Wake, rf.Admission
		if !rf.Seen.IsZero() {
			t := rf.Seen
			w.LastRefused = &t
		}
	}
	d.mu.Lock()
	if t, ok := d.refused[ChunkKey(q.ID, c.ID)]; ok && (w.LastRefused == nil || t.After(*w.LastRefused)) {
		w.LastRefused = &t
	}
	d.mu.Unlock()
	return w
}
