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
	// OffenceUnread is true when the hero's weapon terms were not read (power quality default): k_hero is then capped.
	OffenceUnread bool
	// Active is true when the controller decides (bot.zoneModel 1), not when it only shadows: only then do the realized
	// yield, the mana floor and the fresh-game guard act (card survfix2). The log lines are the same either way.
	Active    bool
	Archetype string // power profile id: selects the archetype row of params (data)
	MercDead  bool   // the hired merc is dead
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
	Area   int
	Name   string
	Diff   string
	Margin string // word
	Why    Why
	Stay   bool // the choice is the current zone
	Fresh  bool // restart the game and farm the same zone again
	// FreshRefused is the reason a fresh game was not taken although the numbers asked for it ("" = not refused).
	FreshRefused string
	Hole         bool // nothing is safe: first field, the hero needs help
	Alts         []Alt
	Lambda       float64
	NeedsYou     bool
	// CurK is k_zone x k_hero of the current zone (the picked zone when the hero is in none), for the bot's re-ask line.
	CurK float64
	line string
}

// Line is the one log line per decision (design 8.1).
func (d Decision) Line() string { return d.line }

// PowerOf is the scalar the 10 percent power-change rule compares: max life. It ignores the belt, which moves by the
// minute, and the merc: his death or revive moved this scalar by 20 percent, which reset every learned row (sums and the
// hard multiplier) right after the event the row had just recorded (card survledger).
func PowerOf(h power.Hero) float64 {
	return float64(h.MaxLife)
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
	xp       float64 // xp/min the value rests on (model prior, or the blend with the realized one)
	realized float64 // the realized xp/min behind the blend, 0 = the prior stands
	starving bool
	riskPaid bool // eligible only through the dial's pay-for-risk route
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
	v.k = l.KZoneFor(c.Key, in.Now, in.OffenceUnread) * l.KHeroFor(in.OffenceUnread)
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
		row, _ := l.P.stepDown(in.Dial.ID) // an unknown dial has no risk route; ForcedOut above already named it
		typ := c.ZM.Typical.Low / v.k
		riskOK := row.TypicalEnter > 0 && typ >= row.TypicalEnter && v.best >= row.BadFloor
		poor := riskOK && !cur && in.Purse.Gold < row.RiskMinGold
		if poor {
			riskOK = false
		}
		if cur && v.best >= stay {
			v.eligible = true
		} else if !cur && v.low >= enter {
			v.eligible = true
		} else if riskOK {
			v.eligible, v.riskPaid = true, true
		} else if poor {
			v.why = Why{WhyPoor, "too poor to pay for a risk: the potions would not be rebought"}
		} else if cur {
			v.why = Why{WhyStayLine, "no longer safe enough to stay"}
		} else if v.best >= stay {
			v.why = Why{WhyMarginLow, "close: one more level or a full belt would do it"}
		} else {
			v.why = Why{WhyMarginLow, "too hard for the hero as it is"}
		}
	}
	stayMin := l.P.StayMinutes
	v.xp = c.Eco.XPPerMin
	if in.Active {
		v.xp, v.realized = l.BlendYield(c.Key, c.Eco.XPPerMin, in.Archetype)
		v.starving = l.Starving(c.Key, in.Archetype)
	}
	v.value = (v.xp + lam*c.Eco.Net()) * stayMin / (stayMin + c.TravelMin)
	if v.riskPaid {
		row, _ := l.P.stepDown(in.Dial.ID)
		v.value *= row.RiskValueMul
	}
	if v.starving {
		v.value *= l.P.StarveValueMul // the hero cannot cast here for want of mana: the model's kill speed is a fiction
	}
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
		if pick.riskPaid {
			d.Why = Why{WhyRiskPaid, "a little riskier than the safe zones, and it pays clearly more"}
		}
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
			if fresh && in.Active && (in.OffenceUnread || in.MercDead && l.P.FreshNeedsMerc) {
				fresh = false // a fresh game is a bet on the model's kill speed: not with an unread weapon (or without the merc when freshNeedsMerc)
				d.FreshRefused = "offence unread"
				if in.MercDead {
					d.FreshRefused = "merc dead"
				}
			}
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
	d.CurK = pick.k
	if cur != nil {
		d.CurK = cur.k
	}
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
		if v.realized > 0 || v.starving {
			a.detail += fmt.Sprintf(" xp/min %.0f", v.xp)
			if v.starving {
				a.detail += " mana-starved"
			}
		}
		if v.eligible {
			a.Why = Why{WhyBestPay, "safe, but pays less"}
			if v.riskPaid {
				a.Why = Why{WhyRiskPaid, "a little risky, and pays less"}
			}
			a.detail = fmt.Sprintf("margin %.1f, value %.0f < %.0f", v.best, v.value, pick.value)
		}
		d.Alts = append(d.Alts, a)
	}
	d.line = decisionLine(d, pick, in)
	return d, nil
}

func decisionLine(d Decision, p *verdict, in Input) string {
	var b strings.Builder
	fmt.Fprintf(&b, "zone choice: %s  margin %.1f (need %.1f; low %.1f)  k %.2f  xp/min %.0f  gold/min %+.0f  dial %s  why: %s (%s)",
		d.Name, p.best, in.Dial.Enter, p.low, p.k, p.xp, p.c.Eco.Net(), in.Dial.ID, d.Why.Code, d.Why.Text)
	if p.realized > 0 {
		fmt.Fprintf(&b, "  realized %.0f (model %.0f)", p.realized, p.c.Eco.XPPerMin)
	}
	if p.starving {
		b.WriteString("  mana-starved")
	}
	if d.Fresh {
		b.WriteString("  fresh game")
	}
	if d.FreshRefused != "" {
		b.WriteString("  fresh game refused: " + d.FreshRefused)
	}
	for _, a := range d.Alts {
		fmt.Fprintf(&b, "  | not %s: %s, %s", a.Name, a.Why.Code, a.detail)
	}
	return b.String()
}
