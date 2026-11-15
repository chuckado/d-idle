package bot

import (
	"context"
	"fmt"
	"os"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/goals"
)

// The farm loop (docs/farmloop.md): monsters respawn only in a new game, so a leveling zone that is mostly cleared pays
// less per minute than a fresh one. At bot.farmResetPct coverage, and only when the value rule says a fresh game pays,
// the hero goes home, takes a fresh game and returns to the same zone.
const (
	farmWindow          = 3 * time.Minute  // the span of the observed XP rate
	farmSampleEvery     = 10 * time.Second // XP samples are kept this far apart
	farmMinVisit        = 4 * time.Minute  // no reset before the zone was farmed this long (the rate needs a window and a cycle)
	farmFreshSec        = 15.0             // Save and Exit to the next town arrival (V live, docs/levelsim.md section 2), until one was measured
	farmWalkFallbackSec = 60.0             // the walk back when this visit's own walk was not seen
	farmDeathHold       = 180 * time.Second
	farmPendingMax      = 5 * time.Minute // a reset that was not taken in this long is dropped
)

type farmSample struct {
	at  time.Time
	exp uint32
}

type farmLoop struct {
	area     int       // zone of the visit (0: none)
	since    time.Time // the visit began (first sweep in the zone this game)
	sinceExp uint32
	walkSec  float64
	smp      []farmSample

	pending    int // zone to reset
	pendingAt  time.Time
	pendingPct int

	returnTo    int // zone to go back to after the reset (one use)
	returnLevel int
	lastAt      time.Time // last reset
	freshSec    float64   // measured time of the last reset's fresh game
}

// forgetVisit starts the next game with no visit; the reset history stays.
func (f *farmLoop) forgetVisit() {
	f.area, f.since, f.sinceExp, f.walkSec, f.smp = 0, time.Time{}, 0, 0, nil
}

// FarmResetPays is the value rule. gained is the XP of the visit so far, fieldMin its minutes, nowRate the observed
// XP per minute of the last window, overheadMin the cost of a reset (fresh game plus the walk back). A reset now yields
// the visit's average over the cycle it would repeat, gained / (fieldMin + overheadMin): it pays when the rate the
// zone gives now is below that. Both numbers are observed in this run; no table constant enters (the clear curve of
// the AREAMON table is only the fallback of the simulation, docs/farmloop.md section 2).
func FarmResetPays(gained, fieldMin, nowRate, overheadMin float64) (pays bool, fresh float64) {
	if fieldMin <= 0 || gained <= 0 {
		return false, 0
	}
	fresh = gained / (fieldMin + overheadMin)
	return nowRate < fresh, fresh
}

// farmNote keeps the visit of the zone being swept: its start, the walk that led there and an XP sample every
// farmSampleEvery. Called from every sweep pass of a leveling slice.
func (r *Runner) farmNote(area int) {
	s := r.latest()
	if s == nil || s.Player == nil {
		return
	}
	f, now := &r.loop.farm, r.e.Clock.Now()
	if f.area != area {
		f.forgetVisit()
		f.area, f.since, f.sinceExp, f.walkSec = area, now, s.Player.Exp, farmWalkFallbackSec
		r.obMu.Lock()
		if left := r.ob.townLeftAt; !left.IsZero() && now.Sub(left) < 15*time.Minute && !left.After(now) {
			f.walkSec = now.Sub(left).Seconds()
		}
		r.obMu.Unlock()
	}
	if n := len(f.smp); n == 0 || now.Sub(f.smp[n-1].at) >= farmSampleEvery {
		f.smp = append(f.smp, farmSample{now, s.Player.Exp})
		if len(f.smp) > 64 {
			f.smp = f.smp[len(f.smp)-64:]
		}
	}
}

// farmResetDue is asked inside the roam: true ends the roam and marks the reset pending. Every guard of the brief is here:
// leveling slices only (no quest, boss, named run or gold/farm order), no fight, no recovery, 180 s after a recovery or
// death, cooldown, coverage, and the value rule.
func (r *Runner) farmResetDue(area int) bool {
	l, t := &r.loop, r.e.Tune.Bot
	f := &l.farm
	if f.pending != 0 {
		return f.pending == area
	}
	if t.FarmReset != 1 || !l.levelSlice || l.keepRoaming(area) || r.d.Cycler == nil || f.area != area || os.Getenv("D2BOT_RUN") != "" {
		return false
	}
	_, _, pct, _, ok := r.Explored()
	if !ok || pct < max(t.FarmResetPct, 50) {
		return false
	}
	s := r.latest()
	if s == nil || s.Player == nil || s.Player.InTown() || isDead(s.Player) {
		return false
	}
	now := r.e.Clock.Now()
	if now.Sub(f.since) < farmMinVisit || (!f.lastAt.IsZero() && now.Sub(f.lastAt) < time.Duration(t.FarmResetCooldownMin)*time.Minute) {
		return false
	}
	r.obMu.Lock()
	diedAt := r.ob.lastDeathAt
	r.obMu.Unlock()
	if (!diedAt.IsZero() && now.Sub(diedAt) < farmDeathHold) || r.freshHold(s.Player) > 0 || r.recoveryActive() || r.bossRun.Load() ||
		r.noTP.Load() || len(r.aliveHostiles(s)) > 0 || r.goldFarmOn() {
		return false
	}
	// the observed rate of the last window
	var old *farmSample
	for i := range f.smp {
		if now.Sub(f.smp[i].at) >= farmWindow {
			old = &f.smp[i]
		}
	}
	if old == nil {
		return false
	}
	nowRate := float64(s.Player.Exp-old.exp) / now.Sub(old.at).Minutes()
	fresh := f.freshSec
	if fresh <= 0 {
		fresh = farmFreshSec
	}
	overhead := fresh + f.walkSec
	pays, freshRate := FarmResetPays(float64(s.Player.Exp-f.sinceExp), now.Sub(f.since).Minutes(), nowRate, overhead/60)
	if !pays {
		return false
	}
	f.pending, f.pendingAt, f.pendingPct = area, now, pct
	r.e.Log.Info(fmt.Sprintf("farm reset: zone %s cov %d rate %.0f vs fresh %.0f overhead %.0f", r.areaName(area), pct, nowRate, freshRate, overhead),
		"area", area, "pct", pct, "nowXPMin", nowRate, "freshXPMin", freshRate, "overheadSec", overhead)
	return true
}

// farmResetStep is the first thing of a leveling slice while a reset is pending: home, then the fresh game. done is
// true when the slice's work for this call is spent.
func (r *Runner) farmResetStep(ctx context.Context, s *d2.Snapshot) (done bool, err error) {
	f := &r.loop.farm
	if f.pending == 0 {
		return false, nil
	}
	now := r.e.Clock.Now()
	if now.Sub(f.pendingAt) > farmPendingMax || r.d.Cycler == nil || r.freshHold(s.Player) > 0 {
		f.pending = 0
		return false, nil
	}
	pct := f.pendingPct
	r.setWhy(WhyFarmReset, "area", r.areaName(f.pending), "pct", itoa(pct))
	if !s.Player.InTown() {
		r.setState(StateResupplying, fmt.Sprintf("Resetting the zone: %d%% cleared, going home for a fresh game", pct))
		return true, r.d.Town.ToTown(ctx)
	}
	if !r.freshAllowed(ctx, s) {
		f.pending = 0
		return true, ctx.Err()
	}
	r.setState(StateInTown, fmt.Sprintf("Resetting the zone: %d%% cleared, a fresh game refills it", pct))
	zone := f.pending
	f.pending = 0
	t0 := r.e.Clock.Now()
	if err := r.startFreshGame(ctx, s); err != nil {
		return true, err
	}
	f.returnTo, f.returnLevel, f.lastAt = zone, s.Player.Level, now
	f.freshSec = max(r.e.Clock.Now().Sub(t0).Seconds(), 1)
	return true, nil
}

// farmReturn sends the planner back to the zone it left for the reset: after the fresh game the plan's route order
// would start again at the first area of the slice. One use, only at the same level, never over a ban, a recovery, the
// gear gate or an area that is cleared again.
func (r *Runner) farmReturn(s *d2.Snapshot, c control, a int) int {
	f := &r.loop.farm
	rt := f.returnTo
	if rt == 0 {
		return a
	}
	if s.Player.Level != f.returnLevel || s.Player.Area == rt || c.spec.kind != goals.ReachLevel {
		f.returnTo = 0
		return a
	}
	if a == rt || r.banned(rt) || r.loop.cleared[rt] || r.d.Plan.IsTown(rt) || r.recoveryActive() || r.gearVerdict(s).Gated {
		return a
	}
	return rt
}

// FarmResetShare is the coverage percent at which the farm loop resets a zone, 0 while the loop is off (the watch's
// Hero now line shows it next to the explored share).
func (r *Runner) FarmResetShare() int {
	if t := r.e.Tune.Bot; t.FarmReset == 1 {
		return max(t.FarmResetPct, 50)
	}
	return 0
}
