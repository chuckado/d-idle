package bot

import (
	"context"
	"sync"
	"time"

	"d2lod-bot/internal/arbiter"
	"d2lod-bot/internal/d2"
)

// The intent arbiter in shadow mode (docs/arbiter.md, knob bot.arbiterShadow). Beside runGoal, the arbiter is asked what
// it WOULD choose from the same world facts; the answer is only logged. runGoal and every hook in it are unchanged.

type arbShadow struct {
	mu sync.Mutex
	sh *arbiter.Shadow
	t0 time.Time
}

// minuteKV is the "arb" attribute of the minute line: per hook family diverged/ticks since the last minute line.
func (a *arbShadow) minuteKV() []any {
	a.mu.Lock()
	sh := a.sh
	a.mu.Unlock()
	if sh == nil {
		return nil
	}
	return []any{"arb", sh.Drain()}
}

// Arbiter is the last shadow decision for the watch API and the debug page (read only; false while the knob is off).
func (r *Runner) Arbiter() (arbiter.Decision, bool) {
	r.arb.mu.Lock()
	sh := r.arb.sh
	r.arb.mu.Unlock()
	if sh == nil {
		return arbiter.Decision{}, false
	}
	return sh.Last()
}

// runGoalShadowed runs the chain and, with the knob on, asks the shadow. The shadow never blocks, never acts and its
// panics are recovered inside; the chain's result is returned untouched.
func (r *Runner) runGoalShadowed(ctx context.Context, snap *d2.Snapshot, c control) bool {
	t := r.e.Tune.Bot
	if t.ArbiterShadow == 0 {
		return r.runGoal(ctx, snap, c)
	}
	v, ok := r.shadowView(snap, c)
	acted := r.runGoal(ctx, snap, c)
	if ok {
		r.shadowTick(v, acted)
	}
	return acted
}

func (r *Runner) shadowView(snap *d2.Snapshot, c control) (v arbiter.View, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	p := snap.Player
	t := r.e.Tune.Bot
	l := &r.loop
	b := func(x bool) int {
		if x {
			return 1
		}
		return 0
	}
	inTown := p.InTown()
	area := "town"
	if !inTown {
		area = r.areaName(p.Area)
	}
	kit := r.kitOf(snap)
	r.obMu.Lock()
	real := r.ob.real
	r.obMu.Unlock()
	pg := r.d.PlayerGear
	f := map[string]int{
		arbiter.FactInTown:       b(inTown),
		arbiter.FactLifePct:      p.LifePct(),
		arbiter.FactHostile:      r.hostileNear(snap, t.ClearRadius),
		arbiter.FactPlayerOrder:  b(pg != nil && pg.Pending() > 0),
		arbiter.FactTrip:         b(l.trip != nil),
		arbiter.FactTripEmerg:    b(l.trip != nil && l.trip.emergency),
		arbiter.FactClaim:        b(!inTown && c.spec.camp == nil && len(l.questPending) > 0 && r.d.Quester != nil),
		arbiter.FactPotions:      kit.BeltHP,
		arbiter.FactPotionsMin:   t.ResupplyPotions,
		arbiter.FactPackGold:     p.Gold,
		arbiter.FactStashGold:    p.StashGold,
		arbiter.FactRecovery:     b(r.recoveryActive()),
		arbiter.FactGoldFarm:     b(r.goldFarmOn()),
		arbiter.FactObjective:    b(c.spec.camp != nil),
		arbiter.FactSurviveBelow: r.e.Tune.Combat.DrinkLifePct,
	}
	return arbiter.View{Now: r.shadowNow(), Area: area, Progress: real, Facts: f}, true
}

func (r *Runner) shadowNow() float64 {
	a := &r.arb
	a.mu.Lock()
	defer a.mu.Unlock()
	now := r.e.Clock.Now()
	if a.sh == nil {
		t := r.e.Tune.Bot
		cfg := arbiter.DefaultConfig()
		cfg.MinDwell, cfg.TravelWindow = float64(t.ArbiterDwellSec), float64(t.ArbiterTravelSec)
		a.sh = arbiter.NewShadow(cfg, arbiter.ShadowConfig{LogGapSec: float64(t.ArbiterLogGapSec)}, arbiter.StdSources()...)
		a.t0 = now
	}
	return now.Sub(a.t0).Seconds()
}

func (r *Runner) shadowTick(v arbiter.View, acted bool) {
	defer func() { _ = recover() }()
	a := &r.arb
	a.mu.Lock()
	sh := a.sh
	a.mu.Unlock()
	chain := ""
	if acted {
		chain = chainFamily(r.currentWhy())
	}
	if line := sh.Tick(v, chain); line != "" {
		r.e.Log.Info("arbiter shadow", "diverge", line)
	}
}

func (r *Runner) currentWhy() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.why.Code
}

// chainFamily maps the whyline code of the activity the chain just started or continued to the arbiter's hook family.
// U: the code can lag one activity behind when the chain acted without announcing a reason; unknown codes count as no
// family (the tick is not compared).
func chainFamily(code string) string {
	switch code {
	case WhyTownTrip, WhyTownChores, WhyHealTrip, WhyWaitHeal, WhyToTown, WhyBackToField, WhyRepairTrip, WhyRepairWait,
		WhyAfterDeath, WhyDeathCorpse, WhyPlayerReset:
		return "chores"
	case WhyClaimTrip:
		return "claim"
	case WhyQuestStep, WhyJourneyBlocked:
		return "quest"
	case WhyFarmUser:
		return "goldfarm"
	case WhySavingFor, WhyDeathGold, WhyFarmGold:
		return "recovery"
	case WhyPlayerItems:
		return "player"
	case WhyGuardExit:
		return "survive"
	case WhyWalk, WhyFight, WhyAreaDone, WhyFarmRuns, WhyFarmReset, WhyRepairFarm:
		return "level"
	}
	return ""
}
