package bot

import "d2lod-bot/internal/d2"

// hpStat feeds the dmgTaken and hpMinPct fields of the minute line (gate v2 proxies, internal/d2dsoakstat).
type hpStat struct {
	prev    int  // life at the previous sample
	seen    bool // prev is valid
	dmg     int  // running counter since the bot started: raw life points lost, sum of negative life deltas
	minPct  int  // lowest life percent seen with a hostile in fight range since the last minute line
	fighted bool // minPct is valid

	// Intake (docs/survival/ledger.md "Observed intake"). lost is the life points lost since the bot started, summed over
	// the drops between consecutive polls; a poll where MaxLife changed (level-up, gear swap, aura) is only rebased, so a
	// max-life change never counts as damage. Gains are ignored. hits counts the polls that saw a drop (a lower bound of the
	// hits: two hits between two polls count as one and their drops merge into one delta). polls counts the samples.
	prevMax   int
	lost      int
	hits      int
	polls     int
	minLost   int // since the last minute line
	minHits   int
	minPolls  int
	recent    []int // life lost in the last intakeWindow completed minute lines, oldest first
	stayStart int   // lost at entry of the current area (area summary)
}

// intakeWindow is the number of minute lines the glance rate averages (the "window").
const intakeWindow = 5

// noteHP is called on every alive sample (and once with the zero-life sample of a death). Life gains from potions
// and regeneration are ignored; only drops add to dmg. Caller holds obMu.
func (o *observer) noteHP(p *d2.Player, hostiles int) {
	if o.hp.seen && p.Life < o.hp.prev {
		o.hp.dmg += o.hp.prev - p.Life
		if p.MaxLife == o.hp.prevMax {
			d := o.hp.prev - p.Life
			o.hp.lost, o.hp.minLost = o.hp.lost+d, o.hp.minLost+d
			o.hp.hits, o.hp.minHits = o.hp.hits+1, o.hp.minHits+1
		}
	}
	o.hp.polls, o.hp.minPolls = o.hp.polls+1, o.hp.minPolls+1
	o.hp.prev, o.hp.prevMax, o.hp.seen = p.Life, p.MaxLife, true
	if hostiles > 0 {
		if pct := p.LifePct(); !o.hp.fighted || pct < o.hp.minPct {
			o.hp.minPct, o.hp.fighted = pct, true
		}
	}
}

// takeHPMin returns the minute's lowest fight life percent (100 when no fight) and starts the next minute.
func (o *observer) takeHPMin() int {
	v := 100
	if o.hp.fighted {
		v = o.hp.minPct
	}
	o.hp.fighted = false
	return v
}

// takeIntake returns the minute's life lost, hits and polls, closes the minute into the rolling window and starts the
// next one. Caller holds obMu.
func (o *observer) takeIntake() (lost, hits, polls int) {
	h := &o.hp
	lost, hits, polls = h.minLost, h.minHits, h.minPolls
	h.minLost, h.minHits, h.minPolls = 0, 0, 0
	h.recent = append(h.recent, lost)
	if len(h.recent) > intakeWindow {
		h.recent = h.recent[len(h.recent)-intakeWindow:]
	}
	return lost, hits, polls
}

// Intake is the observed damage intake over the last completed minute lines (docs/survival/ledger.md).
type Intake struct {
	LifeLostPerMin float64 // mean life lost per minute line in the window
	WindowMinutes  int     // minute lines averaged (0 = nothing yet)
}

// Intake reads the rolling intake; the bool is false until the first minute line.
func (r *Runner) Intake() (Intake, bool) {
	r.obMu.Lock()
	defer r.obMu.Unlock()
	rec := r.ob.hp.recent
	if len(rec) == 0 {
		return Intake{}, false
	}
	sum := 0
	for _, v := range rec {
		sum += v
	}
	return Intake{LifeLostPerMin: float64(sum) / float64(len(rec)), WindowMinutes: len(rec)}, true
}
