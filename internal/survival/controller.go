package survival

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"d2lod-bot/internal/power"
)

// Candidate is one zone the hero could farm, with the numbers the caller computed for the hero as it is now.
type Candidate struct {
	Key       ZoneKey
	Name      string
	ZM        power.ZoneMargin
	Eco       Eco
	TravelMin float64
	Depleted  bool // the zone is cleared (the levelsim supply term)
}

// Purse is the gold side of lambda: what the kit and the next purchase need against what the hero holds.
type Purse struct {
	Gold int
	Need int // kit, revive quote, repair; 0 when nothing is due
}

// Input is one decision.
type Input struct {
	Now          time.Time
	Dial         power.Dial
	Current      ZoneKey // zero Area = not in a farm zone
	Power        float64 // scalar hero power (PowerOf)
	Purse        Purse
	Cands        []Candidate
	FreshAllowed bool
}

// Alt is one zone that was not chosen, with the reason in words (API) and with numbers (log only).
type Alt struct {
	Area   int    `json:"id"`
	Name   string `json:"name"`
	Margin string `json:"margin"` // "safe", "risky" or "too hard"
	Why    Why    `json:"why"`
	detail string
}

// Decision is the shadow choice: what the controller would farm, and why.
type Decision struct {
	Area     int
	Name     string
	Diff     string
	Margin   string // word
	Why      Why
	Stay     bool // the choice is the current zone
	Fresh    bool // restart the game and farm the same zone again
	Hole     bool // nothing is safe: first field, the hero needs help
	Alts     []Alt
	Lambda   float64
	NeedsYou bool
	line     string
}

// Line is the one log line per decision (design 8.1).
func (d Decision) Line() string { return d.line }

// PowerOf is the scalar the 10 percent power-change rule compares: max life, with a living merc counted as a quarter
// more (I, U). It deliberately ignores the belt, which moves by the minute.
func PowerOf(h power.Hero) float64 {
	v := float64(h.MaxLife)
	if h.Merc.Hired && h.Merc.Alive {
		v *= 1.25
	}
	return v
}

// verdict is a candidate judged against the dial and the ledger.
type verdict struct {
	c        Candidate
	best     float64 // margin of the bad pack after the ledger correction
	low      float64
	k        float64 // k_zone x k_hero
	eligible bool
	word     string
	why      Why
	value    float64
}

func word(best, low, enter, stay float64) string {
	switch {
	case low >= enter:
		return "safe"
	case best >= stay:
		return "risky"
	}
	return "too hard"
}

func lambda(p *Params, pu Purse) float64 {
	if pu.Need <= 0 {
		return 0
	}
	short := float64(pu.Need-pu.Gold) / float64(pu.Need)
	return math.Min(1, math.Max(0, short)) * p.LambdaMax
}

func (l *Ledger) judge(in Input, c Candidate, lam float64) verdict {
	l.Refresh(c.Key, in.Power)
	v := verdict{c: c}
	v.k = l.KZone(c.Key, in.Now) * l.KHero()
	v.best, v.low = c.ZM.Bad.Best/v.k, c.ZM.Bad.Low/v.k
	enter, stay := in.Dial.Enter, in.Dial.Stay
	cur := c.Key == in.Current
	v.word = word(v.best, v.low, enter, stay)
	switch {
	case l.Locked(c.Key, in.Power) && !cur:
		v.why = Why{WhyLeftForCost, "left because it cost too much; closed until the hero is stronger"}
	default:
		if out, kind := l.ForcedOut(c.Key, in.Dial.ID, in.Now); out {
			if kind == "died" {
				v.why = Why{WhyDied, "the hero died here recently"}
			} else {
				v.why = Why{WhyCost, "chickens or a merc death here crossed the limit of this setting"}
			}
			break
		}
		if cur && v.best >= stay {
			v.eligible = true
		} else if !cur && v.low >= enter {
			v.eligible = true
		} else if cur {
			v.why = Why{WhyStayLine, "no longer safe enough to stay"}
		} else if v.best >= stay {
			v.why = Why{WhyMarginLow, "close: one more level or a full belt would do it"}
		} else {
			v.why = Why{WhyMarginLow, "too hard for the hero as it is"}
		}
	}
	stayMin := l.P.StayMinutes
	v.value = (c.Eco.XPPerMin + lam*c.Eco.Net()) * stayMin / (stayMin + c.TravelMin)
	return v
}

// Choose decides the zone (design 5.3). It marks the zone it leaves for cost in the ledger. Pure otherwise.
func (l *Ledger) Choose(in Input) (Decision, error) {
	if len(in.Cands) == 0 {
		return Decision{}, fmt.Errorf("survival: no candidate zones")
	}
	lam := lambda(l.P, in.Purse)
	vs := make([]verdict, 0, len(in.Cands))
	for _, c := range in.Cands {
		vs = append(vs, l.judge(in, c, lam))
	}
	var elig []*verdict
	var cur *verdict
	for i := range vs {
		if vs[i].c.Key == in.Current {
			cur = &vs[i]
		}
		if vs[i].eligible {
			elig = append(elig, &vs[i])
		}
	}
	sort.SliceStable(elig, func(i, j int) bool { return elig[i].value > elig[j].value })
	d := Decision{Lambda: lam}
	var pick *verdict
	switch {
	case len(elig) == 0:
		// a real hole: the zone with the highest corrected margin (the first field), with a "needs you" line
		best := &vs[0]
		for i := range vs {
			if vs[i].best > best.best {
				best = &vs[i]
			}
		}
		pick = best
		d.Hole, d.NeedsYou = true, true
		d.Why = Why{WhyHole, "nothing is safe enough: the easiest field, and the hero needs help"}
	default:
		pick = elig[0]
		d.Why = Why{WhyBestPay, "best pay among the safe zones"}
		if pick.c.Depleted {
			var next *verdict
			for _, e := range elig {
				if !e.c.Depleted {
					next = e
					break
				}
			}
			fresh := in.FreshAllowed && l.P.FreshRatio > 0 &&
				(next == nil || next.value < l.P.FreshRatio*pick.value*l.P.StayMinutes/(l.P.StayMinutes+l.P.FreshMinutes))
			switch {
			case fresh:
				d.Fresh = true
				d.Why = Why{WhyFresh, "the best zone is cleared and a new game beats walking to a poor one"}
			case next != nil:
				pick = next
				d.Why = Why{WhyDepleted, "the best zone is cleared; the next safe one"}
			}
		}
		if cur != nil && cur.eligible && pick != cur && !cur.c.Depleted && pick.value < cur.value*(1+l.P.SwitchMargin) {
			pick = cur
			d.Why = Why{WhyStay, "another zone pays a little more, not enough to leave"}
		}
	}
	d.Area, d.Name, d.Diff, d.Margin = pick.c.Key.Area, pick.c.Name, pick.c.Key.Diff, pick.word
	d.Stay = pick.c.Key == in.Current
	if cur != nil && !cur.eligible && !d.Stay {
		l.MarkLeft(cur.c.Key, in.Now, in.Power)
	}
	for i := range vs {
		v := &vs[i]
		if v == pick {
			continue
		}
		a := Alt{Area: v.c.Key.Area, Name: v.c.Name, Margin: v.word, Why: v.why,
			detail: fmt.Sprintf("margin %.1f (low %.1f, need %.1f, k %.2f)", v.best, v.low, in.Dial.Enter, v.k)}
		if v.eligible {
			a.Why = Why{WhyBestPay, "safe, but pays less"}
			a.detail = fmt.Sprintf("margin %.1f, value %.0f < %.0f", v.best, v.value, pick.value)
		}
		d.Alts = append(d.Alts, a)
	}
	d.line = decisionLine(d, pick, in)
	return d, nil
}

func decisionLine(d Decision, p *verdict, in Input) string {
	var b strings.Builder
	fmt.Fprintf(&b, "zone choice: %s  margin %.1f (need %.1f; low %.1f)  xp/min %.0f  gold/min %+.0f  dial %s  why: %s (%s)",
		d.Name, p.best, in.Dial.Enter, p.low, p.c.Eco.XPPerMin, p.c.Eco.Net(), in.Dial.ID, d.Why.Code, d.Why.Text)
	if d.Fresh {
		b.WriteString("  fresh game")
	}
	for _, a := range d.Alts {
		fmt.Fprintf(&b, "  | not %s: %s, %s", a.Name, a.Why.Code, a.detail)
	}
	return b.String()
}
