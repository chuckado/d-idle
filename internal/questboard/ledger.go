package questboard

import "strings"

// The Ether hooks (slice S10, design section 7, docs/ether.md "Earn call sites"). The board never knows the wallet: it
// asks a Ledger, and the key `quest/<quest>/<chunk or ->/<difficulty>` is account-wide, so a quest pays once per account
// per difficulty. The award is derived from T1/T2 evidence (Award in state.go) every time, so a restart between the
// evidence read and the ledger write only means the next settle pays; the ledger refuses the duplicate.

// KindEther is the currency word in `pays` and `awarded`.
const KindEther = "ether"

// LedgerResult is one Earn call.
type LedgerResult struct {
	Key      string
	Credited int
	Already  bool // the key was booked before (or the day cap left nothing)
}

// Ledger is the account's wallet as the board sees it. Quest and chunk ("" for the whole quest; the key spells it "-") and the difficulty
// word (normal, nm, hell) make the key; amount is the data column.
type Ledger interface {
	Earn(quest, chunk, diff string, amount int) (LedgerResult, error)
	// Paid reports what the key paid, when it is booked.
	Paid(quest, chunk, diff string) (int, bool)
}

// Pay is one line of `pays` (an offer) and of `awarded` (a payoff).
type Pay struct {
	Kind   string `json:"kind"`
	Amount int    `json:"amount"`
	Key    string `json:"key,omitempty"`
}

// ledger returns the wallet now, nil while there is none (no vault, knob loot.vault 0): the whole hook is inert then.
func (e *Exec) ledger() Ledger {
	e.mu.Lock()
	f := e.o.Ledger
	e.mu.Unlock()
	if f == nil {
		return nil
	}
	return f()
}

// Ledger is the wallet the executor settles into (nil: none).
func (e *Exec) Ledger() Ledger { return e.ledger() }

// settle books the awards the evidence earned. It returns the credits made now; a booked key is not an error and not
// listed. Callers do not hold e.mu.
func (e *Exec) settle(awards []Award, diff int) []Pay {
	l := e.ledger()
	if l == nil {
		return nil
	}
	var out []Pay
	for _, a := range awards {
		quest, chunk, ok := splitAwardKey(a.Key)
		if !ok {
			continue
		}
		r, err := l.Earn(quest, chunk, Difficulties[diff], a.Amount)
		if err != nil {
			e.o.Log("quest award not booked", "key", a.Key, "err", err)
			continue
		}
		if r.Credited > 0 && !r.Already {
			out = append(out, Pay{Kind: KindEther, Amount: r.Credited, Key: r.Key})
		}
	}
	return out
}

func splitAwardKey(k string) (quest, chunk string, ok bool) {
	p := strings.Split(k, "/")
	if len(p) != 4 || p[0] != "quest" {
		return "", "", false
	}
	if p[2] == "-" {
		return p[1], "", true
	}
	return p[1], p[2], true
}

// Settle books the awards of every chunk and quest that evidence has confirmed now (a late T1/T2 confirmation of a
// hand-in the board saw only as T3). Returns what was credited.
func (e *Exec) Settle() []Pay {
	e.mu.Lock()
	if !e.o.Enabled {
		e.mu.Unlock()
		return nil
	}
	vs, _, diff, err := e.derive()
	e.mu.Unlock()
	if err != nil {
		return nil
	}
	var all []Award
	for _, v := range vs {
		all = append(all, v.Awards...)
	}
	return e.settle(all, diff)
}

// payFor lists what a chunk of a quest would pay in the difficulty and has not paid yet: the chunk's own column, plus
// the quest's column on its last chunk (U: the quest award lands when the quest completes).
func payFor(q Quest, c Chunk, diff int, l Ledger) []Pay {
	d := Difficulties[diff]
	var out []Pay
	add := func(chunk string, n int) {
		if n <= 0 {
			return
		}
		if _, paid := l.Paid(q.ID, chunk, d); paid {
			return
		}
		out = append(out, Pay{Kind: KindEther, Amount: n, Key: AwardKey(q.ID, chunk, diff)})
	}
	add(c.ID, c.Meta[d])
	if len(q.Chunks) > 0 && q.Chunks[len(q.Chunks)-1].ID == c.ID {
		add("", q.Meta[d])
	}
	return out
}

// paidFor lists what the chunk (and the quest on its last chunk) has paid, for the history row.
func paidFor(q Quest, c Chunk, diff int, l Ledger) []Pay {
	d := Difficulties[diff]
	var out []Pay
	get := func(chunk string) {
		if n, ok := l.Paid(q.ID, chunk, d); ok && n > 0 {
			out = append(out, Pay{Kind: KindEther, Amount: n, Key: AwardKey(q.ID, chunk, diff)})
		}
	}
	get(c.ID)
	if len(q.Chunks) > 0 && q.Chunks[len(q.Chunks)-1].ID == c.ID {
		get("")
	}
	return out
}

// settleChunk settles the awards of one chunk (and its quest) from fresh evidence.
func (e *Exec) settleChunk(quest, chunk string) []Pay {
	e.mu.Lock()
	vs, _, diff, err := e.derive()
	e.mu.Unlock()
	if err != nil {
		return nil
	}
	qv, _, _ := chunkView(vs, quest, chunk)
	return e.settle(chunkAwards(qv, quest, chunk), diff)
}
