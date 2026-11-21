package bot

import (
	"context"
	"time"

	"d2lod-bot/internal/d2"
)

// Recall now (docs/process/management-fast-path.md): a MANAGEMENT command, not a goal. The player presses it when an item
// action needs the town (a Vault item to equip, the stash, the vendor) and the hero is in the field. The runner walks home
// at the next safe moment (the errand trip's rule, safeMoment), ahead of the other player steps, and the queued town work
// (vault keep and equip orders, stash moves) runs on arrival in the same visit. It never walks through a fight and never
// undoes anything: after the walk the hero is simply in town and the goal loop resumes its goal. Gated by
// ui.managementFast; a measured run never installs the desk.

// Recall states, the values of RecallView.State.
const (
	RecallIdle    = "idle"    // nothing asked
	RecallWaiting = "waiting" // asked; the moment is not safe yet (Reason says why)
	RecallGoing   = "going"   // the hero is walking home
	RecallArrived = "arrived" // the hero reached town because of the command (shown until the next command)
	RecallFailed  = "failed"  // the walk failed; it is tried again after recallBackoff
)

// RecallSchema is the schema field of the watch route's reply.
const RecallSchema = "watch.recall/1"

// recallBackoff is the wait after a failed walk home before the next try.
const recallBackoff = 20 * time.Second

// RecallView is what GET /api/recall shows.
type RecallView struct {
	Schema  string     `json:"schema"`
	Enabled bool       `json:"enabled"` // ui.managementFast is on: the page offers the button
	State   string     `json:"state"`
	InTown  bool       `json:"in_town"`
	Since   *time.Time `json:"since,omitempty"`
	Reason  string     `json:"reason,omitempty"` // what the command waits for, or why the last try failed
	Error   string     `json:"error,omitempty"`  // why the press did nothing (knob off, no way home)
}

type recallState struct {
	state  string
	since  time.Time
	reason string
	nextAt time.Time
}

func (r *Runner) recallOn() bool { return r.e.Tune != nil && r.e.Tune.UI.ManagementFast != 0 }

// RecallNow asks the hero to walk home at the next safe moment and returns the new view. In town it is answered at once
// ("arrived": nothing to walk).
func (r *Runner) RecallNow() RecallView {
	if !r.recallOn() {
		v := r.RecallView()
		v.Error = "immediate management is off (knob ui.managementFast)"
		return v
	}
	if r.d.Town == nil {
		v := r.RecallView()
		v.Error = "this bot has no way home"
		return v
	}
	r.mu.Lock()
	rc := &r.rcl
	if rc.state != RecallWaiting && rc.state != RecallGoing && rc.state != RecallFailed {
		rc.state, rc.since, rc.reason, rc.nextAt = RecallWaiting, r.e.Clock.Now(), "waiting for the next look at the hero", time.Time{}
	}
	r.mu.Unlock()
	return r.RecallView()
}

// RecallCancel withdraws the command (a walk in progress is not interrupted; it ends in town).
func (r *Runner) RecallCancel() RecallView {
	r.mu.Lock()
	if r.rcl.state == RecallWaiting || r.rcl.state == RecallFailed {
		r.rcl = recallState{}
	}
	r.mu.Unlock()
	return r.RecallView()
}

// RecallView reports the state.
func (r *Runner) RecallView() RecallView {
	r.mu.Lock()
	rc := r.rcl
	r.mu.Unlock()
	v := RecallView{Schema: RecallSchema, Enabled: r.recallOn(), State: rc.state, Reason: rc.reason}
	if v.State == "" {
		v.State = RecallIdle
	}
	if r.e.Read != nil {
		if sn, err := r.e.Snapshot(); err == nil && sn != nil && sn.Player != nil {
			v.InTown = sn.Player.InTown()
		}
	}
	if v.State == RecallArrived && !v.InTown { // the hero went back out: the command is history
		v.State, v.Reason = RecallIdle, ""
	}
	if v.State != RecallIdle {
		since := rc.since
		v.Since = &since
	}
	return v
}

func (r *Runner) setRecall(state, reason string) {
	r.mu.Lock()
	r.rcl.state, r.rcl.reason = state, reason
	r.mu.Unlock()
}

// recallDue is the guard's check (see playerDue): the command is waiting and the moment is safe, so a walk in progress is
// cancelled and the loop runs the recall next.
func (r *Runner) recallDue(s *d2.Snapshot, now time.Time) bool {
	r.mu.Lock()
	st, next := r.rcl.state, r.rcl.nextAt
	r.mu.Unlock()
	if (st != RecallWaiting && st != RecallFailed) || s.Player.InTown() || now.Before(next) {
		return false
	}
	return r.hostileNear(s, r.e.Tune.Bot.ClearRadius) == 0 && r.hostileWithin(s, r.e.Tune.Bot.ClearRadius) == 0
}

// recallStep runs ahead of the other player steps in runGoal. It reports whether it acted.
func (r *Runner) recallStep(ctx context.Context, snap *d2.Snapshot, c control) bool {
	r.mu.Lock()
	st, next := r.rcl.state, r.rcl.nextAt
	r.mu.Unlock()
	if (st != RecallWaiting && st != RecallFailed && st != RecallGoing) || snap == nil || snap.Player == nil {
		return false
	}
	if snap.Player.InTown() {
		if r.loop.trip != nil {
			r.setRecall(st, "finishing the town trip first")
			return false
		}
		r.setRecall(RecallArrived, "in town")
		return false // the queued town work runs in this same tick (playerActionStep, vaultKeepStep)
	}
	if r.d.Town == nil {
		r.setRecall(RecallFailed, "no way home is wired")
		return false
	}
	if r.e.Clock.Now().Before(next) {
		return false
	}
	if _, ok, why := r.safeMoment(snap); !ok {
		r.setRecall(RecallWaiting, "waiting for a safe moment: "+why)
		return false
	}
	r.setRecall(RecallGoing, "going home")
	if r.do(ctx, c, StateResupplying, "Going to town because you asked", sec(r.cfg.TownTimeoutSec), "recall trip", r.d.Town.ToTown) {
		return true
	}
	r.mu.Lock()
	r.rcl.state, r.rcl.reason, r.rcl.nextAt = RecallFailed, "the walk home failed; trying again shortly", r.e.Clock.Now().Add(recallBackoff)
	r.mu.Unlock()
	r.e.Log.Warn("recall: the walk home failed; trying again shortly")
	return true
}
