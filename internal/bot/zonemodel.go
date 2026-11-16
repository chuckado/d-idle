package bot

import (
	"sync"
	"time"

	"d2lod-bot/internal/d2"
)

// Survival slice S-D (docs/survival/zonemodel.md): with bot.zoneModel 1 the survival controller chooses the leveling
// zone among the plan's zones. The plan tables stay the candidate list and the fallback.

// zoneDialIDs maps bot.zoneDial to the dial ids of internal/power/profiles.json.
var zoneDialIDs = []string{"cautious", "balanced", "bold"}

// zoneRecheck is the dwell: a chosen zone is kept this long before the model is asked again (design 11: minimum dwell).
const zoneRecheck = 3 * time.Minute

// ZoneVerdict is the controller's decision as the bot needs it. Margin is a word (safe, risky, too hard).
type ZoneVerdict struct {
	Area       int
	Fresh      bool   // the best zone is cleared and a fresh game beats walking to a poorer one
	Hole       bool   // nothing is safe: the easiest field
	Margin     string // word
	Line       string // the `zone choice:` log line
	Why        string // sentence for the Status line
	MercDead   bool   // the merc is dead and the controller judged the revive
	MercRevive bool   // reviving pays; false with MercDead: leave him dead
	MercWhy    string
}

// zoneState is the last verdict, for the merc policy and the Status word.
type zoneState struct {
	mu     sync.Mutex
	v      ZoneVerdict
	at     time.Time
	have   bool
	noted  string // merc note already logged
	failed string // last error logged
}

// zoneDial is the dial id of the knob; an out-of-range value is the default (balanced).
func (r *Runner) zoneDial() string {
	i := r.e.Tune.Bot.ZoneDial
	if i < 0 || i >= len(zoneDialIDs) {
		i = 1
	}
	return zoneDialIDs[i]
}

func (r *Runner) zoneModelOn() bool { return r.e.Tune.Bot.ZoneModel == 1 && r.d.ZoneDecide != nil }

// zoneRecheckDue is true when the cached pick is old enough to ask the model again.
func (r *Runner) zoneRecheckDue() bool {
	if !r.zoneModelOn() {
		return false
	}
	z := &r.zst
	z.mu.Lock()
	defer z.mu.Unlock()
	return z.have && r.e.Clock.Now().Sub(z.at) >= zoneRecheck
}

// zoneCands lists the plan's zones of this difficulty: the eligible steps, and the cleared zones as depleted candidates.
func (r *Runner) zoneCands(eligible []Step, diff string, cleared map[int]bool) []ZoneCand {
	seen := map[int]bool{}
	var out []ZoneCand
	for _, st := range eligible {
		if !seen[st.Area] {
			seen[st.Area] = true
			out = append(out, ZoneCand{Area: st.Area})
		}
	}
	for _, st := range r.d.Plan.Steps[diff] {
		if st.Area != 0 && cleared[st.Area] && !seen[st.Area] && !r.d.Plan.IsTown(st.Area) {
			seen[st.Area] = true
			out = append(out, ZoneCand{Area: st.Area, Depleted: true})
		}
	}
	return out
}

// zoneModelPick is PickOptions.Model for the goal's leveling pick; nil while the model is off.
func (r *Runner) zoneModelPick(s *d2.Snapshot, diff string, cleared map[int]bool) func([]Step) (int, bool, bool) {
	if !r.zoneModelOn() || s == nil || s.Player == nil {
		return nil
	}
	return func(eligible []Step) (int, bool, bool) {
		potions := -1
		if healing, _, known := r.beltAndInv(s); known {
			potions = healing
		}
		cands := r.zoneCands(eligible, diff, cleared)
		v, err := r.d.ZoneDecide(s, s.Player.Area, diff, potions, r.zoneDial(), cands)
		z := &r.zst
		if err != nil {
			z.mu.Lock()
			first := z.failed != err.Error()
			z.failed, z.have = err.Error(), false
			z.mu.Unlock()
			if first { // one line per distinct error, never a line per tick
				r.e.Log.Warn("zone model: error, the plan tables decide", "err", err)
			}
			return 0, false, false
		}
		z.mu.Lock()
		z.v, z.at, z.have, z.failed = v, r.e.Clock.Now(), true, ""
		z.mu.Unlock()
		for _, line := range splitLines(v.Line) {
			r.e.Log.Info(line)
		}
		if v.Fresh {
			if r.d.Cycler == nil || len(cleared) == 0 {
				return 0, false, false // no way to take a fresh game: the plan rule decides
			}
			return 0, true, false
		}
		return v.Area, false, v.Area != 0
	}
}

// zoneLeaveMercDead is true while the model says the dead merc stays dead (the zone kills him, or he adds too little).
// It only ever suppresses the revive chore; hiring and the legacy gold rules are untouched.
func (r *Runner) zoneLeaveMercDead() bool {
	if !r.zoneModelOn() {
		return false
	}
	z := &r.zst
	z.mu.Lock()
	defer z.mu.Unlock()
	if !z.have || !z.v.MercDead || z.v.MercRevive || r.e.Clock.Now().Sub(z.at) > 4*zoneRecheck {
		return false
	}
	if z.noted != z.v.MercWhy {
		z.noted = z.v.MercWhy
		r.e.Log.Info("merc: left dead by the zone model", "why", z.v.MercWhy)
	}
	return true
}

// ZoneMargin is the word of the margin of the zone the model last chose ("" while the model is off or silent), for the
// Status sentence: words only, never a number (user decision 2026-11-14).
func (r *Runner) ZoneMargin() (word, why string) {
	z := &r.zst
	z.mu.Lock()
	defer z.mu.Unlock()
	if !z.have {
		return "", ""
	}
	return z.v.Margin, z.v.Why
}

func splitLines(s string) []string {
	var out []string
	for len(s) > 0 {
		i := 0
		for i < len(s) && s[i] != '\n' {
			i++
		}
		if i > 0 {
			out = append(out, s[:i])
		}
		if i == len(s) {
			break
		}
		s = s[i+1:]
	}
	return out
}
