package bot

import (
	"context"
	"time"

	"d2lod-bot/internal/d2"
)

// vaultKeepWait is the pause between two attempts while the queue cannot move (stash full, game busy, hero not ready): the
// executor reports why, the bot does not hammer the verb. A change of the queue length (a new press) tries at once.
const vaultKeepWait = 30 * time.Second

// vaultKeepStep creates the items the player kept in the Ethereal Vault (docs/vault.md, Keep). In town at a safe moment it
// runs the executor, which is bounded per call and per visit; in the field it walks home only when the player pressed "Keep
// now" (the errand trip of errandTripStep, same backoff). It never runs in a fight and never inside a trip. It reports
// whether it acted.
func (r *Runner) vaultKeepStep(ctx context.Context, snap *d2.Snapshot, c control) bool {
	vk := r.d.VaultKeep
	if vk == nil || snap == nil || snap.Player == nil {
		return false
	}
	n := vk.Pending()
	if n == 0 {
		r.loop.vaultKeepPend, r.loop.vaultKeepNext = 0, time.Time{}
		return false
	}
	l := &r.loop
	now := r.e.Clock.Now()
	if !snap.Player.InTown() {
		if r.d.Town == nil || !vk.TripWanted() || now.Before(l.errandSkipUntil) {
			return false
		}
		if _, ok, _ := r.safeMoment(snap); !ok {
			return false
		}
		r.setWhy(WhyPlayerItems, "what", "vault keep")
		if r.do(ctx, c, StateResupplying, "Going to town to keep your vault items", sec(r.cfg.TownTimeoutSec), "vault keep trip", r.d.Town.ToTown) {
			return true
		}
		l.errandSkipUntil = now.Add(errandBackoff)
		r.e.Log.Warn("vault keep trip failed; no new errand trip for a while", "minutes", int(errandBackoff/time.Minute))
		return true
	}
	if n == l.vaultKeepPend && now.Before(l.vaultKeepNext) {
		return false
	}
	if _, ok, why := r.safeMoment(snap); !ok {
		r.noteDelayed(now, why)
		return false
	}
	l.vaultKeepNext = now.Add(vaultKeepWait)
	r.setWhy(WhyPlayerItems, "what", "vault keep")
	r.do(ctx, c, StateInTown, "Keeping vault items in the stash", sec(r.cfg.TownTimeoutSec), "vault keep", func(ctx context.Context) error {
		for _, line := range vk.Run(ctx) {
			r.e.Log.Info(line)
		}
		return ctx.Err()
	})
	// The per-visit cap leaves the rest for the next attempt, after the chores had their turn.
	if l.vaultKeepPend = vk.Pending(); l.vaultKeepPend == 0 {
		l.vaultKeepNext = time.Time{}
	}
	return true
}
