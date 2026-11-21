package bulkid

import (
	"fmt"
	"sort"

	"d2lod-bot/internal/vault"
)

// MassMax is the most ids one mass reveal names (the vault routes take at most this many per press).
const MassMax = 50

// Price is what a reveal costs the hero right now (the vault's RevealInfo plus the scroll price).
type Price struct {
	Free       bool // Cain stands here
	Charges    int  // scrolls plus tome charges held, -1 unknown
	ScrollGold int  // gold of one scroll (information only; the mass button never buys)
	EtherPer   int  // Ether one reveal costs (identify.etherPerReveal), 0: Ether is not offered
	EtherBal   int  // Ether in the wallet, -1 none
}

// Mass is the preview of the "Reveal filtered" button: which vault records it would queue, and what that costs.
type Mass struct {
	Count   int      `json:"count"`
	Skipped int      `json:"skipped"` // matched the filter but the price cannot cover them now
	IDs     []string `json:"ids"`
	Mode    string   `json:"mode"` // cain, scroll or none
	Gold    int      `json:"gold,omitempty"`
	Line    string   `json:"line"`
	// Ether is the second option of the button, present while Ether is a payer and Cain is not here (Cain's reveal is free).
	Ether *MassEther `json:"ether,omitempty"`
}

// MassEther previews the mass reveal paid with Ether: no scroll or tome charge is used, the Ether is charged per item
// after each unlock. Only records whose affixes the vault holds can be unlocked this way.
type MassEther struct {
	Count   int      `json:"count"`
	Skipped int      `json:"skipped"` // matched but the wallet cannot cover them (or the vault holds no affixes)
	IDs     []string `json:"ids"`
	Cost    int      `json:"cost"` // Ether for Count items
	Per     int      `json:"per"`
	Balance int      `json:"balance"`
	Line    string   `json:"line"`
}

// Cand is one unidentified vault record the identify filter selects, in reveal order (priority, value, id).
type Cand struct {
	ID     string
	Prio   int
	Value  int
	Sealed bool // the vault holds its affixes: Ether can unlock it without a scroll
}

// Candidates lists the filtered, unidentified, not yet queued vault records best first.
func Candidates(f *Filter, recs []vault.Rec, qname func(int) string) []Cand {
	var cs []Cand
	for _, r := range recs {
		if r.Identified || r.RevQ || (r.State != vault.Recorded && r.State != vault.KeepQ) {
			continue
		}
		it := Item{Code: r.Code, Types: r.Chain, Quality: qname(r.Quality), Ilvl: r.Ilvl, Value: r.Value}
		if _, prio, ok := f.Match(it); ok {
			cs = append(cs, Cand{r.ID, prio, r.Value, r.Sealed != ""})
		}
	}
	sort.SliceStable(cs, func(i, j int) bool {
		a, b := cs[i], cs[j]
		if a.Prio != b.Prio {
			return a.Prio > b.Prio
		}
		if a.Value != b.Value {
			return a.Value > b.Value
		}
		return a.ID < b.ID
	})
	return cs
}

// MassReveal picks the filtered, unidentified vault records the hero can pay for now: all of them at Cain, else as many
// as it holds charges, best first (priority, value, id). Nothing is bought and nothing is queued here.
func MassReveal(f *Filter, recs []vault.Rec, qname func(int) string, p Price) Mass {
	cs := Candidates(f, recs, qname)
	m := Mass{IDs: []string{}}
	room := MassMax
	switch {
	case p.Free:
		m.Mode = ViaCain
	case p.Charges > 0:
		m.Mode = ViaScroll
		room = min(room, p.Charges)
	default:
		m.Mode = "none"
		room = 0
	}
	for i, c := range cs {
		if i >= room {
			m.Skipped = len(cs) - i
			break
		}
		m.IDs = append(m.IDs, c.ID)
	}
	m.Count = len(m.IDs)
	switch {
	case len(cs) == 0:
		m.Line = "No unidentified items match the identify filter"
	case m.Mode == ViaCain:
		m.Line = fmt.Sprintf("Reveal %d filtered item%s: free at Cain", m.Count, plural(m.Count))
	case m.Mode == ViaScroll:
		m.Gold = m.Count * p.ScrollGold
		m.Line = fmt.Sprintf("Reveal %d filtered item%s: %d charge%s (you have %d)", m.Count, plural(m.Count), m.Count, plural(m.Count), p.Charges)
		if m.Skipped > 0 {
			m.Line += fmt.Sprintf("; %d more need scrolls", m.Skipped)
		}
	default:
		m.Skipped = len(cs)
		m.Line = fmt.Sprintf("%d filtered item%s: no scroll or tome charge and Cain is not here; buy scrolls or visit Cain", len(cs), plural(len(cs)))
	}
	if p.EtherPer > 0 && !p.Free {
		m.Ether = massEther(cs2ids(cs, func(c Cand) (string, bool) { return c.ID, c.Sealed }), len(cs), p)
	}
	return m
}

// cs2ids lists the ids (in order) of the candidates the pick accepts.
func cs2ids[T any](cs []T, f func(T) (string, bool)) []string {
	var ids []string
	for _, c := range cs {
		if id, ok := f(c); ok {
			ids = append(ids, id)
		}
	}
	return ids
}

func massEther(sealed []string, matched int, p Price) *MassEther {
	e := &MassEther{IDs: []string{}, Per: p.EtherPer, Balance: p.EtherBal}
	room := min(MassMax, max(p.EtherBal, 0)/p.EtherPer)
	e.IDs = append(e.IDs, sealed[:min(room, len(sealed))]...)
	e.Count = len(e.IDs)
	e.Cost = e.Count * p.EtherPer
	e.Skipped = matched - e.Count
	switch {
	case matched == 0:
		e.Line = "No unidentified items match the identify filter"
	case e.Count == 0 && len(sealed) == 0:
		e.Line = fmt.Sprintf("%d filtered item%s: the vault holds no affixes for them; use scrolls or Cain", matched, plural(matched))
	case e.Count == 0:
		e.Line = fmt.Sprintf("Reveal with Ether: %d each, you have %d: not enough", p.EtherPer, p.EtherBal)
	default:
		e.Line = fmt.Sprintf("Reveal %d filtered item%s with Ether: %d Ether (you have %d), no scroll used", e.Count, plural(e.Count), e.Cost, p.EtherBal)
		if e.Skipped > 0 {
			e.Line += fmt.Sprintf("; %d more cannot be paid or have no held affixes", e.Skipped)
		}
	}
	return e
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
