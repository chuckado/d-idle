package bot

import (
	"context"

	"d2lod-bot/internal/arbiter"
	"d2lod-bot/internal/d2"
)

// The arbiter's control path (docs/arbiter.md, knobs bot.arbiterControl and bot.arbiterFamilies, ledger row in
// data/darklaunch.json). Default off. With it on, the arbiter's decision is taken BEFORE the chain; when it chose a
// family that is under control, that family's hook runs first and the chain's lower hooks do not start. When the arbiter
// chose nothing, panicked, or the controlled hook found nothing to do, the old chain runs unchanged: the fallback.
// Wired so far: survive (a reflex: no chore, claim or quest step starts while the hero is under the drink line with
// hostiles near; it fights where it stands and the guard's reflex drink does the drinking) and player (item actions from
// the watch page run first). The other family bits are accepted and ignored. U: not run live.

// runGoalControlled is runGoalArb for a hero under control. It returns what the chain or the controlled hook did.
func (r *Runner) runGoalControlled(ctx context.Context, snap *d2.Snapshot, c control, v arbiter.View) bool {
	t := r.e.Tune.Bot
	r.shadowNow() // creates the shadow
	r.arb.mu.Lock()
	sh := r.arb.sh
	r.arb.mu.Unlock()
	d, ok := sh.Decide(v)
	acted, took := false, false
	if ok && d.Chosen != nil {
		acted, took = r.arbAct(ctx, snap, c, d, t.ArbiterFamilies)
	}
	if !took {
		acted = r.runGoal(ctx, snap, c)
	}
	r.arbCompare(sh, d, v.Now, acted, took)
	return acted
}

// arbAct runs the controlled hook for the family the arbiter chose. took is false when the family is not under control
// or its hook did nothing, so the chain must run.
func (r *Runner) arbAct(ctx context.Context, snap *d2.Snapshot, c control, d arbiter.Decision, mask int) (acted, took bool) {
	defer func() {
		if recover() != nil {
			acted, took = false, false
		}
	}()
	fam := arbiter.Family(d.Chosen.Source)
	if !arbiter.Controlled(mask, fam) {
		return false, false
	}
	switch fam {
	case arbiter.SrcSurvive:
		p := snap.Player
		if p == nil || p.InTown() {
			return false, false
		}
		r.setWhy(WhyFight, "area", r.areaName(p.Area), "life", itoa(p.LifePct()))
		acted = r.sweep(ctx, snap, c, p.Area)
	case arbiter.SrcPlayer:
		acted = r.playerActionStep(ctx, snap, c)
	default:
		return false, false
	}
	if acted {
		r.arb.ctl.Add(1)
	}
	return acted, acted
}

// arbCompare is the second half of shadowTick for a decision taken before the chain acted.
func (r *Runner) arbCompare(sh *arbiter.Shadow, d arbiter.Decision, now float64, acted, took bool) {
	defer func() { _ = recover() }()
	chain := ""
	switch {
	case took:
		chain = arbiter.Family(d.Chosen.Source) // the arbiter acted: agreement by construction
	case acted:
		chain = chainFamily(r.currentWhy())
	}
	if line := sh.Compare(d, now, chain); line != "" {
		r.e.Log.Info("arbiter shadow", "diverge", line)
	}
	for _, in := range d.Starved {
		r.inv.noteStarved(in.Key() + ": " + in.Reason)
	}
}
