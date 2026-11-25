package bot

import (
	"context"
	"sync"
	"time"

	"d2lod-bot/internal/d2"
)

// The corpse in town (user request 2026-11-25): a corpse that could not be fetched in the field is not given up. The game
// moves a corpse left behind to the town of the next game (V d2d: docs/d2d-spike.md, "Save and reload with a corpse": the
// corpse stands in town one cell from the arrival spot, also when it was made in the field), so after the last failed try the
// bot starts a fresh game and takes the corpse in town first thing. A corpse with nothing on it is not restored (V d2d), so
// a corpse worth nothing is not chased.

// corpseTownMaxAge: a pending corpse pickup older than this is dropped (the game was lost, the hero was moved on).
const corpseTownMaxAge = 15 * time.Minute

type corpseTownState struct {
	mu          sync.Mutex
	pending     bool
	since       time.Time
	equipBefore int
	tries       int
}

func (c *corpseTownState) set(now time.Time, equip int) {
	c.mu.Lock()
	c.pending, c.since, c.equipBefore, c.tries = true, now, equip, 0
	c.mu.Unlock()
}

func (c *corpseTownState) get() (pending bool, since time.Time, equip, tries int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.pending, c.since, c.equipBefore, c.tries
}

func (c *corpseTownState) clear() {
	c.mu.Lock()
	c.pending = false
	c.mu.Unlock()
}

func (c *corpseTownState) tried() {
	c.mu.Lock()
	c.tries++
	c.mu.Unlock()
}

// resetForCorpse abandons the corpse of the field and starts a fresh game, so the game moves it to town. It goes home first
// when the hero stands in the field (the fresh game saves and leaves from anywhere, the walk home only keeps the hero safe).
func (r *Runner) resetForCorpse(ctx context.Context, snap *d2.Snapshot, why string) {
	if r.d.Cycler == nil || r.d.CorpseTake == nil || snap == nil || snap.Player == nil {
		return
	}
	r.e.Log.Info("corpse: abandoned, resetting", "why", why)
	if !snap.Player.InTown() {
		if err := r.d.Town.ToTown(ctx); err != nil {
			r.e.Log.Warn("corpse: the way home failed, resetting from the field", "err", err)
		}
	}
	s := r.latest()
	if s == nil || s.Player == nil || isDead(s.Player) || ctx.Err() != nil {
		return
	}
	r.corpseTown.set(r.e.Clock.Now(), equipCount(s))
	if err := r.startFreshGame(ctx, s); err != nil {
		r.e.Log.Warn("corpse: the fresh game failed", "err", err)
		r.corpseTown.clear()
	}
}

// corpseTownStep takes the corpse in town in the new game, before anything else. It reports whether it acted.
func (r *Runner) corpseTownStep(ctx context.Context, snap *d2.Snapshot) bool {
	pending, since, before, tries := r.corpseTown.get()
	if !pending || snap == nil || snap.Player == nil || isDead(snap.Player) {
		return false
	}
	if r.e.Clock.Now().Sub(since) > corpseTownMaxAge || tries >= 3 {
		r.e.Log.Info("corpse: not found in town, giving up", "tries", tries)
		r.corpseTown.clear()
		return false
	}
	if !snap.Player.InTown() || snap.UI.BlockingPanelOpen() || r.hostileNear(snap, r.e.Tune.Bot.ClearRadius) > 0 {
		return false
	}
	if r.d.CorpseTake == nil {
		r.corpseTown.clear()
		return false
	}
	r.corpseTown.tried()
	found, err := r.d.CorpseTake.TakeCorpse(ctx)
	if err != nil {
		r.e.Log.Warn("corpse: the click failed", "err", err)
	}
	after := before
	if s := r.poll(); s != nil && s.Player != nil {
		after = equipCount(s)
	}
	if found && after > before {
		r.e.Log.Info("corpse: recovered in town", "worn", after, "before", before)
		r.corpseTown.clear()
		r.baselineAfterCorpse()
	}
	return true
}

// baselineAfterCorpse: the returned gear is not a drop.
func (r *Runner) baselineAfterCorpse() {
	if s := r.latest(); s != nil && s.Player != nil {
		r.obMu.Lock()
		r.baselineItems(s)
		r.obMu.Unlock()
	}
}
