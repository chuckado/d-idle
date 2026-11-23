package survival

import (
	"fmt"
	"time"
)

// ReviveInput compares the hero's best farm with and without the merc (design 5.5). With and Without are the same zones
// judged for a living and a dead merc (the caller builds both from the same hero).
type ReviveInput struct {
	Base          Input // Cands is ignored
	With, Without []Candidate
	Quote         int // gold the revive costs
	Reserve       int // gold the kit rule keeps back after the purchase
	Gold          int
}

// ReviveDecision is the answer; Why.Text is the sentence for the Status line when he stays dead.
type ReviveDecision struct {
	Revive bool
	Why    Why
	Gain   float64 // value units gained until he dies again
	Cost   float64 // the quote in the same units
	Zone   string  // the zone the gain is measured in
}

// bestEligible is the value and name of the best eligible candidate; ok false when none is eligible.
func (l *Ledger) bestEligible(in Input, cs []Candidate, lam float64) (float64, string, bool) {
	best, name, ok := 0.0, "", false
	for _, c := range cs {
		v := l.judge(in, c, lam, 1)
		if v.eligible && (!ok || v.value > best) {
			best, name, ok = v.value, c.Name, true
		}
	}
	return best, name, ok
}

// Revive decides whether paying the quote is worth it: gain = (best eligible value with him minus without) times the
// minutes he is expected to live in that zone (the ledger's merc deaths there, else the prior); worth it when the gain
// exceeds the quote scaled by ValuePerGold and lambda, and the purse keeps the kit reserve. A zone that killed him
// twice in the window is closed to the with-merc plan: the hero farms where he survives or without him.
func (l *Ledger) Revive(in ReviveInput) ReviveDecision {
	lam := lambda(l.P, Purse{Gold: in.Gold, Need: in.Quote + in.Reserve})
	var with []Candidate
	for _, c := range in.With {
		if !l.NoReviveNow(c.Key, in.Base.Now) {
			with = append(with, c)
		}
	}
	vWith, zone, okW := l.bestEligible(in.Base, with, lam)
	vWithout, _, okO := l.bestEligible(in.Base, in.Without, lam)
	if !okW && len(with) < len(in.With) {
		return ReviveDecision{Why: Why{WhyMercNoRevive, "the zone kills him; the hero farms where he is not needed"}}
	}
	if !okW {
		return ReviveDecision{Why: Why{WhyMercDead, "he would not make any zone safe enough"}}
	}
	if !okO {
		vWithout = 0 // no safe zone without him: any safe farm with him is all gain
	}
	var key ZoneKey
	for _, c := range with {
		if c.Name == zone {
			key = c.Key
		}
	}
	minutes := l.MercMinutes(key)
	d := ReviveDecision{Gain: (vWith - vWithout) * minutes, Cost: float64(in.Quote) * l.P.ValuePerGold * (1 + lam), Zone: zone}
	switch {
	case in.Gold-in.Quote < in.Reserve:
		d.Why = Why{WhyMercDead, "left dead: the kit comes first and the purse cannot cover both"}
	case d.Gain <= d.Cost:
		d.Why = Why{WhyMercDead, fmt.Sprintf("left dead: in %s he adds too little for what he costs", zone)}
	default:
		d.Revive = true
		d.Why = Why{WhyMercRevive, fmt.Sprintf("worth reviving: he lives long enough in %s to pay back", zone)}
	}
	return d
}

// MercDeathsIn is the number of merc deaths in k within the window.
func (l *Ledger) MercDeathsIn(k ZoneKey, now time.Time) int {
	s := l.peek(k)
	if s == nil {
		return 0
	}
	return l.count(s, now, EvMercDeath)
}
