package arbiter

// Explained is the arbiter's decision in the form of the explain record (Fable review 2026-11-25, P7): the intent it
// chose, its hook family and reason, and what it displaced (the proposals it turned down, with the reason each).
type Explained struct {
	Intent    string      `json:"intent"` // source/kind
	Family    string      `json:"family"`
	Reason    string      `json:"reason,omitempty"`
	Held      bool        `json:"held,omitempty"`    // chosen only because its lease holds
	Resumed   bool        `json:"resumed,omitempty"` // continued after a higher class preempted it
	Displaced []Displaced `json:"displaced,omitempty"`
}

// Displaced is one proposal the choice put aside.
type Displaced struct {
	Intent string `json:"intent"`
	Why    string `json:"why"`
}

const explainedMaxDisplaced = 6

// Explained returns the decision as an explain record; false when the arbiter chose nothing.
func (d Decision) Explained() (Explained, bool) {
	if d.Chosen == nil {
		return Explained{}, false
	}
	e := Explained{Intent: d.Chosen.Key(), Family: Family(d.Chosen.Source), Reason: d.Chosen.Reason, Held: d.Held, Resumed: d.Resumed}
	for _, r := range d.Rejected {
		if len(e.Displaced) == explainedMaxDisplaced {
			break
		}
		e.Displaced = append(e.Displaced, Displaced{Intent: r.Intent.Key(), Why: r.Why})
	}
	return e, true
}
