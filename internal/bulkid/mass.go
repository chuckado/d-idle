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
}

// Mass is the preview of the "Reveal filtered" button: which vault records it would queue, and what that costs.
type Mass struct {
	Count   int      `json:"count"`
	Skipped int      `json:"skipped"` // matched the filter but the price cannot cover them now
	IDs     []string `json:"ids"`
	Mode    string   `json:"mode"` // cain, scroll or none
	Gold    int      `json:"gold,omitempty"`
	Line    string   `json:"line"`
}

// MassReveal picks the filtered, unidentified vault records the hero can pay for now: all of them at Cain, else as many
// as it holds charges, best first (priority, value, id). Nothing is bought and nothing is queued here.
func MassReveal(f *Filter, recs []vault.Rec, qname func(int) string, p Price) Mass {
	type cand struct {
		id    string
		prio  int
		value int
	}
	var cs []cand
	for _, r := range recs {
		if r.Identified || r.RevQ || (r.State != vault.Recorded && r.State != vault.KeepQ) {
			continue
		}
		it := Item{Code: r.Code, Types: r.Chain, Quality: qname(r.Quality), Ilvl: r.Ilvl, Value: r.Value}
		if _, prio, ok := f.Match(it); ok {
			cs = append(cs, cand{r.ID, prio, r.Value})
		}
	}
	sort.SliceStable(cs, func(i, j int) bool {
		a, b := cs[i], cs[j]
		if a.prio != b.prio {
			return a.prio > b.prio
		}
		if a.value != b.value {
			return a.value > b.value
		}
		return a.id < b.id
	})
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
		m.IDs = append(m.IDs, c.id)
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
	return m
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
