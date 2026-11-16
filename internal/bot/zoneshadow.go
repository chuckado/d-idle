package bot

import (
	"strings"
	"time"

	"d2lod-bot/internal/d2"
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
type zoneMerc struct{ seen, alive bool }

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
	if r.d.ZoneChoice == nil {
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
	if m.seen && m.alive && !alive {
		r.zoneObserve(ZoneEvent{Kind: ZoneEvMercDied, Area: p.Area, Diff: difficultyName(p.Difficulty), At: now})
	}
	m.seen, m.alive = true, alive
}
