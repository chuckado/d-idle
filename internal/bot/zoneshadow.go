package bot

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/survival"
)

// ZoneCand is one zone of the plan offered to the survival shadow controller.
type ZoneCand struct {
	Area     int
	Depleted bool // the planner marked it cleared this game
}

// Zone event kinds (ZoneObserve).
const (
	ZoneEvArea     = "area"       // an area summary: Kills, Seconds, XP
	ZoneEvChicken  = "chicken"    // a chicken (SafetyChicken carried out)
	ZoneEvDeath    = "death"      // the character died
	ZoneEvMercDied = "merc_death" // the merc went from alive to dead
)

// ZoneEvent is one observation for the per-zone ledger. Diff is the lowercase difficulty name.
type ZoneEvent struct {
	Kind     string
	Area     int
	Diff     string
	Kills    int
	Seconds  float64
	XP       float64
	LifeLost int // life lost over the stay (ZoneEvArea), net of max-life changes
	At       time.Time
}

// zoneMerc remembers the last merc reading so a death is reported once, at the edge.
type zoneMerc struct {
	seen, alive bool
	loop        survival.ReviveLoop // revive loop breaker; N, Within and Hold are set from the knobs on first use
	loopOnce    sync.Once
	blockNoted  atomic.Bool // the "left dead by the breaker" line was logged for this block
}

// zoneShadow logs the survival margin of the chosen zone next to the `planner` line (slice S-B) and the controller's
// shadow choice (slice S-C). Shadow only: nothing reads the result.
func (r *Runner) zoneShadow(s *d2.Snapshot, area int, diff string) {
	if (r.d.ZoneShadow == nil && r.d.ZoneChoice == nil) || s == nil || s.Player == nil {
		return
	}
	potions := -1
	if healing, _, known := r.beltAndInv(s); known {
		potions = healing
	}
	if r.d.ZoneShadow != nil {
		if line := r.d.ZoneShadow(s, area, diff, potions); line != "" {
			r.e.Log.Info(line)
		}
	}
	if r.d.ZoneChoice == nil || r.zoneModelOn() { // the model logs its own `zone choice:` line when it decides
		return
	}
	seen := map[int]bool{}
	var cands []ZoneCand
	for _, st := range r.d.Plan.Steps[diff] {
		if st.Area == 0 || seen[st.Area] || r.d.Plan.IsTown(st.Area) {
			continue
		}
		seen[st.Area] = true
		cands = append(cands, ZoneCand{Area: st.Area, Depleted: r.loop.cleared[st.Area]})
	}
	for _, line := range strings.Split(r.d.ZoneChoice(s, area, diff, potions, cands), "\n") {
		if line != "" {
			r.e.Log.Info(line)
		}
	}
}

// zoneObserve forwards one event to the ledger hook; the hook must be quick and lock nothing of the Runner.
func (r *Runner) zoneObserve(ev ZoneEvent) {
	if r.d.ZoneObserve != nil {
		r.d.ZoneObserve(ev)
	}
}

// zoneMercEdge reports a merc death once when the reading flips from alive to dead (caller holds obMu; the hook only
// takes its own lock).
func (r *Runner) zoneMercEdge(p *d2.Player, alive bool, now time.Time) {
	m := &r.zmerc
	diff := difficultyName(p.Difficulty)
	if m.seen && m.alive && !alive {
		r.zoneObserve(ZoneEvent{Kind: ZoneEvMercDied, Area: p.Area, Diff: diff, At: now})
	}
	if r.e == nil { // a bare runner (unit tests of the edge): no knobs, no log
		m.seen, m.alive = true, alive
		return
	}
	switch {
	case m.seen && m.alive && !alive:
		life, known, fast, tripped := r.mercLoop().Died(survival.ZoneKey{Area: p.Area, Diff: diff}, now)
		switch {
		case !known:
			r.e.Log.Info("merc revive: he died with no revive of this run before", "area", p.Area)
		case tripped:
			r.e.Log.Info("merc revive: died soon after the revive, loop breaker closes the zone", "area", p.Area, "afterSec", int(life.Seconds()),
				"limit", r.e.Tune.Bot.MercReviveLoopN, "windowMin", r.e.Tune.Bot.MercReviveLoopMin)
		default:
			r.e.Log.Info("merc revive: he died", "area", p.Area, "afterSec", int(life.Seconds()), "quick", fast)
		}
		m.blockNoted.Store(false)
	case m.seen && !m.alive && alive:
		r.mercLoop().Revived(now)
		r.e.Log.Info("merc revive: alive again", "area", p.Area)
	}
	m.seen, m.alive = true, alive
}

// mercLoop is the revive loop breaker, configured from the knobs on first use.
func (r *Runner) mercLoop() *survival.ReviveLoop {
	m := &r.zmerc
	m.loopOnce.Do(func() {
		t := r.e.Tune.Bot
		m.loop.N, m.loop.Within = t.MercReviveLoopN, time.Duration(t.MercReviveLoopMin)*time.Minute
		m.loop.Hold = time.Hour
	})
	return &m.loop
}
