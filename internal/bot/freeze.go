package bot

import (
	"context"
	"fmt"
	"time"
)

// The frozen Pause (docs/pause.md): the player's Pause stops the game itself, not only the bot. The backend freezes the
// game clock (Freezer; d2d: the fork's runtime "lockstep on" with no "step" behind it, "lockstep off" resumes); the
// runner parks its loop, cancels the running action with errPause and takes the pause out of every clock: the
// env clock stands still (clockFreezer), and the stall, stuck, drought and farm clocks are re-armed on resume. While
// frozen the loop touches nothing: no poll, no safety, no chicken, no goal step. Reset (break glass) thaws first.

// Freezer is the backend's real freeze of the game. A nil Freezer, or CanFreeze false, means "no freeze here": the
// Pause then keeps its older meaning (walk home, then stop) and the watch page shows no frozen state.
type Freezer interface {
	CanFreeze(ctx context.Context) bool
	Freeze(ctx context.Context, on bool) error
}

// clockFreezer is implemented by an env clock that can stand still (d2denv.LateClock on a wall clock).
type clockFreezer interface{ SetFrozen(on bool) }

const (
	freezeProbeTTL  = time.Minute      // a failed capability probe is repeated after this
	freezeCallLimit = 10 * time.Second // one freeze or thaw call
	freezeNoteTTL   = 30 * time.Minute // how long the "paused for N min" note stays on the view
)

// PauseEnded is the note that stays on the view after a frozen pause ended: the Needs you info row reads it.
type PauseEnded struct {
	Minutes int       `json:"minutes"`
	Auto    bool      `json:"auto"` // the cap ended it, not the player
	At      time.Time `json:"at"`
}

type freezeState struct {
	on      bool
	since   time.Time // wall time of the freeze
	can     bool      // last capability answer
	probeAt time.Time // when it was taken (zero: never)
	ended   *PauseEnded
	near    bool // monsters were within the clear radius when the game was frozen (info only)
}

func (r *Runner) wallNow() time.Time {
	if r.wall != nil {
		return r.wall()
	}
	return time.Now()
}

// Frozen reports whether the game is frozen by the player's Pause (lock-free enough for health probes).
func (r *Runner) Frozen() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.frz.on
}

// canFreeze is the cached capability probe; it never holds r.mu across the backend call.
func (r *Runner) canFreeze() bool {
	if r.d.Freezer == nil {
		return false
	}
	r.mu.Lock()
	if r.frz.on {
		r.mu.Unlock()
		return true
	}
	if !r.frz.probeAt.IsZero() && (r.frz.can || r.wallNow().Sub(r.frz.probeAt) < freezeProbeTTL) {
		can := r.frz.can
		r.mu.Unlock()
		return can
	}
	r.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), freezeCallLimit)
	defer cancel()
	can := r.d.Freezer.CanFreeze(ctx)
	r.mu.Lock()
	r.frz.can, r.frz.probeAt = can, r.wallNow()
	r.mu.Unlock()
	return can
}

func (r *Runner) setClockFrozen(on bool) {
	if cf, ok := r.e.Clock.(clockFreezer); ok {
		cf.SetFrozen(on)
	}
}

// freezeOn freezes the game. On a backend error nothing changes and the error is returned.
func (r *Runner) freezeOn() error {
	near := false
	if s := r.latest(); s != nil && s.Player != nil && !s.Player.InTown() {
		near = r.hostileNear(s, r.e.Tune.Bot.ClearRadius) > 0 || r.hostileWithin(s, r.e.Tune.Bot.ClearRadius) > 0
	}
	r.mu.Lock()
	if r.frz.on {
		r.mu.Unlock()
		return nil
	}
	// a go-home pause in progress ends: the freeze replaces it (the hero stays where he stands)
	if r.plp.held {
		r.paused, r.autoResumeAt = false, time.Time{}
	}
	r.plp.want, r.plp.held, r.plp.reason, r.plp.since = false, false, "", time.Time{}
	r.plp.save()
	r.frz.on, r.frz.since, r.frz.ended, r.frz.near = true, r.wallNow(), nil, near
	r.cancelLocked(errPause) // the running action ends now; the loop parks at its next step
	r.mu.Unlock()
	r.setClockFrozen(true)
	ctx, cancel := context.WithTimeout(context.Background(), freezeCallLimit)
	defer cancel()
	if err := r.d.Freezer.Freeze(ctx, true); err != nil {
		r.setClockFrozen(false)
		r.mu.Lock()
		r.frz.on = false
		r.mu.Unlock()
		r.e.Log.Warn("pause: the game could not be frozen", "err", err)
		return err
	}
	r.e.Log.Info("paused: the game is frozen; the bot stands still until Resume") // once; no per-minute lines while frozen
	return nil
}

// freezeOff thaws the game and re-arms every progress clock. auto marks the cap. force (Reset) clears the state even
// when the game refuses to thaw, so the loop can run its own recovery.
func (r *Runner) freezeOff(auto, force bool) error {
	r.mu.Lock()
	if !r.frz.on {
		r.mu.Unlock()
		return nil
	}
	since := r.frz.since
	r.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), freezeCallLimit)
	err := r.d.Freezer.Freeze(ctx, false)
	cancel()
	if err != nil && !force {
		r.e.Log.Warn("pause: the game did not thaw, still frozen", "err", err)
		return err
	}
	r.setClockFrozen(false)
	now := r.e.Clock.Now()
	r.obMu.Lock()
	r.ob.campaignAt, r.ob.rearmAt = now, now // a pause is not a stall: nothing counts the frozen time
	if r.ob.lastProgress.Before(now) {
		r.ob.lastProgress = now
	}
	r.ob.errSince = time.Time{}
	r.obMu.Unlock()
	r.mu.Lock()
	r.loop.stallStage = 0
	mins := int(r.wallNow().Sub(since) / time.Minute)
	r.frz.on, r.frz.ended = false, &PauseEnded{Minutes: mins, Auto: auto, At: r.wallNow()}
	r.mu.Unlock()
	r.e.Log.Info("resumed: the game runs again", "pausedMin", mins, "auto", auto)
	return err
}

// ForceThaw is Reset's break glass: whatever the pause state, the game runs again before the reset starts.
func (r *Runner) ForceThaw() {
	if r.d.Freezer != nil {
		r.freezeOff(false, true)
	}
}

// frozenStep is the first thing step does. While frozen it does nothing but watch the cap; it reports true when the
// loop must skip the rest of the step (the caller then sleeps one tick).
func (r *Runner) frozenStep() bool {
	r.mu.Lock()
	on, since := r.frz.on, r.frz.since
	r.mu.Unlock()
	if !on {
		return false
	}
	if capMin := r.e.Tune.Bot.PauseCapMin; capMin > 0 && r.wallNow().Sub(since) >= time.Duration(capMin)*time.Minute {
		if r.freezeOff(true, false) == nil {
			return true // the next step runs normally
		}
	}
	r.mu.Lock()
	if r.st != StatePaused || r.activity != "Paused" {
		r.setLocked(StatePaused, "Paused")
	}
	r.mu.Unlock()
	return true
}

// frozenView adds the freeze fields to a pause view (caller holds r.mu).
func (r *Runner) frozenViewLocked(v *PlayerPauseView) {
	f := &r.frz
	if f.on {
		since := f.since
		v.State, v.Frozen, v.Since, v.Reason, v.MonstersNear = PausePaused, true, &since, "the game is frozen", f.near
		if c := r.e.Tune.Bot.PauseCapMin; c > 0 {
			v.CapMin = c
			end := since.Add(time.Duration(c) * time.Minute)
			v.EndsAt = &end
		}
	}
	if f.ended != nil && r.wallNow().Sub(f.ended.At) < freezeNoteTTL {
		e := *f.ended
		v.Ended = &e
	}
}

// PauseNeedText is the Needs you info row after a frozen pause ended.
func (e PauseEnded) PauseNeedText() string {
	if e.Auto {
		return fmt.Sprintf("Paused for %d min; the pause cap resumed the game", e.Minutes)
	}
	return fmt.Sprintf("Paused for %d min", e.Minutes)
}
