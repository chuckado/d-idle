package bot

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/goals"
)

// The player's Reset command (docs/reset.md, user request 2026-11-15): leave the game and rejoin the same character; the
// bot then resumes its plan. The runner keeps the goal chain, the orders, the Loop, the pause and the journey in its own
// memory and state files, so a leave-and-rejoin loses none of them. This file is the soft level (the loop asks the
// Cycler to start a new game at the next safe moment); the hard level (the process exits like the watchdog) is
// cmd/d2bot/resetwire.go, which only records it here (BeginHardReset) so the next process can finish the story.
// Safety (chicken, potions, portal) is not touched, and nothing here is an equipment or combat decision.

// Reset states, the values of ResetView.State.
const (
	ResetIdle      = "idle"
	ResetResetting = "resetting" // asked; waiting for the moment, or leaving the game
	ResetRejoining = "rejoining" // the game is gone, the character is going back in (soft), or the process restarted (hard)
	ResetFailed    = "failed"
)

// ResetSchema is the schema field of the watch route's reply.
const ResetSchema = "watch.reset/1"

// resetPatience is how long a soft reset waits for nothing hostile near before it leaves anyway: a logout is itself an
// escape, so danger never blocks it for long, and `force` skips the wait.
const resetPatience = 20 * time.Second

// ResetResult is the last finished reset.
type ResetResult struct {
	Result string    `json:"result"` // done or failed
	At     time.Time `json:"at"`
	Hard   bool      `json:"hard,omitempty"`
	Reason string    `json:"reason,omitempty"`
}

// ResetView is what GET /api/reset shows.
type ResetView struct {
	Schema  string       `json:"schema"`
	State   string       `json:"state"`
	Since   *time.Time   `json:"since,omitempty"`
	Reason  string       `json:"reason,omitempty"`
	Hard    bool         `json:"hard,omitempty"`
	Already bool         `json:"already,omitempty"` // POST only: a reset was already running, this press changed nothing
	Last    *ResetResult `json:"last,omitempty"`
}

type playerReset struct {
	state     string
	since     time.Time
	reason    string
	hard      bool
	force     bool
	restarted bool // a hard reset: this is the process that came back
	last      *ResetResult
	path      string
}

type resetFile struct {
	State string       `json:"state"`
	Since time.Time    `json:"since,omitempty"`
	Hard  bool         `json:"hard,omitempty"`
	Last  *ResetResult `json:"last,omitempty"`
}

// SetResetPath loads the persisted reset and keeps it up to date. A hard reset that was under way when the process
// exited shows as rejoining until the first in-game snapshot of this process. Never call it in a measured run.
func (r *Runner) SetResetPath(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p := &r.rst
	p.path = path
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var f resetFile
	if json.Unmarshal(b, &f) != nil {
		return
	}
	p.last = f.Last
	if f.Hard && (f.State == ResetResetting || f.State == ResetRejoining) {
		p.state, p.since, p.hard, p.restarted, p.reason = ResetRejoining, f.Since, true, true, "the bot restarted, rejoining the game"
	}
}

// CanSoftReset says the runner can leave and rejoin the game itself (a Cycler is wired).
func (r *Runner) CanSoftReset() bool { return r.d.Cycler != nil }

// PlayerReset starts a soft reset. A press while one runs changes nothing (Already). force skips the wait for a safe moment.
func (r *Runner) PlayerReset(force bool) ResetView {
	r.mu.Lock()
	p := &r.rst
	if p.state == ResetResetting || p.state == ResetRejoining {
		v := r.resetViewLocked()
		r.mu.Unlock()
		v.Already = true
		return v
	}
	now := r.e.Clock.Now()
	p.state, p.since, p.hard, p.force, p.restarted = ResetResetting, now, false, force, false
	p.reason = "waiting for a safe moment to leave the game"
	if force {
		p.reason = "leaving the game"
	}
	p.save()
	v := r.resetViewLocked()
	r.mu.Unlock()
	r.e.Log.Info("reset requested", "hard", false, "force", force)
	return v
}

// BeginHardReset records a hard reset the caller is about to carry out (the process exits next). It is idempotent.
func (r *Runner) BeginHardReset() ResetView {
	r.mu.Lock()
	p := &r.rst
	if p.state == ResetResetting || p.state == ResetRejoining {
		v := r.resetViewLocked()
		r.mu.Unlock()
		v.Already = true
		return v
	}
	p.state, p.since, p.hard, p.reason = ResetResetting, r.e.Clock.Now(), true, "restarting the bot"
	p.save()
	v := r.resetViewLocked()
	r.mu.Unlock()
	r.e.Log.Info("reset requested", "hard", true)
	return v
}

// ResetView reports the state.
func (r *Runner) ResetView() ResetView {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.resetViewLocked()
}

func (r *Runner) resetViewLocked() ResetView {
	p := &r.rst
	v := ResetView{Schema: ResetSchema, State: ResetIdle, Last: p.last}
	if p.state != "" {
		v.State = p.state
	}
	if v.State != ResetIdle {
		since := p.since
		v.Since, v.Reason, v.Hard = &since, p.reason, p.hard
	}
	return v
}

func (p *playerReset) save() {
	if p.path == "" {
		return
	}
	st := p.state
	if st == ResetFailed || st == ResetIdle {
		st = ResetIdle
	}
	b, _ := json.Marshal(resetFile{State: st, Since: p.since, Hard: p.hard, Last: p.last})
	if os.MkdirAll(filepath.Dir(p.path), 0o755) == nil {
		tmp := p.path + ".tmp"
		if os.WriteFile(tmp, b, 0o644) == nil {
			os.Rename(tmp, p.path)
		}
	}
}

// finishReset ends the running reset with a result. The failed state stays visible until the next press.
func (r *Runner) finishReset(result, why string) {
	r.mu.Lock()
	p := &r.rst
	now := r.e.Clock.Now()
	p.last = &ResetResult{Result: result, At: now, Hard: p.hard, Reason: why}
	if result == "done" {
		p.state, p.reason, p.hard, p.restarted = ResetIdle, "", false, false
	} else {
		p.state, p.reason = ResetFailed, why
	}
	p.save()
	r.mu.Unlock()
	if result == "done" {
		r.e.Log.Info("reset done", "why", why)
	} else {
		r.e.Log.Warn("reset failed", "why", why)
	}
}

func (r *Runner) setResetState(state, reason string) {
	r.mu.Lock()
	r.rst.state, r.rst.reason = state, reason
	r.rst.save()
	r.mu.Unlock()
}

// resetStep runs in the loop right after the death and recovery handling, whether or not the goal is paused (a paused hero is
// reset too and stays paused). It reports whether it acted.
func (r *Runner) resetStep(ctx context.Context, snap *d2.Snapshot, c control) bool {
	r.mu.Lock()
	p := r.rst
	r.mu.Unlock()
	if p.state != ResetResetting && p.state != ResetRejoining {
		return false
	}
	if snap == nil || snap.Player == nil {
		return false
	}
	if p.restarted { // the hard reset's new process is in a game again: the story ends here
		r.finishReset("done", "the bot restarted and rejoined")
		return false
	}
	if p.state != ResetResetting {
		return false
	}
	if r.d.Cycler == nil {
		r.finishReset("failed", "this backend cannot leave the game by itself")
		return false
	}
	now := r.e.Clock.Now()
	if !p.force && now.Sub(p.since) < resetPatience && (r.hostileNear(snap, r.e.Tune.Bot.ClearRadius) > 0 || r.hostileWithin(snap, r.e.Tune.Bot.ClearRadius) > 0) {
		r.setResetState(ResetResetting, "waiting for a safe moment to leave the game")
		return false
	}
	r.setResetState(ResetRejoining, "leaving the game and rejoining")
	diff := difficultyName(snap.Player.Difficulty)
	r.setWhy(WhyPlayerReset)
	ok := r.do(ctx, c, StateRecovering, "Reset: leaving the game and rejoining", 5*time.Minute, "reset", r.newGameRun(func(ctx context.Context) error {
		return r.d.Cycler.NewGame(ctx, diff)
	}))
	switch {
	case ok:
		r.finishReset("done", "left the game and rejoined")
	case ctx.Err() == nil:
		r.finishReset("failed", "the game did not come back")
	}
	return true
}

// StuckFor is how long the bot has made no progress while it has a goal and is not paused or idle (the stall detector's own
// clock, observe.go lastProgress), and 0 when it is not stuck. The loop's heartbeat age (LastStep) covers a hung loop.
func (r *Runner) StuckFor() time.Duration {
	r.mu.Lock()
	st, goal, paused := r.st, r.goal, r.paused
	r.mu.Unlock()
	if goal == nil || paused || skippedState(st, goal.Type == goals.Idle) {
		return 0
	}
	r.obMu.Lock()
	defer r.obMu.Unlock()
	if r.ob.lastProgress.IsZero() {
		return 0
	}
	return max(0, r.e.Clock.Now().Sub(r.ob.lastProgress))
}
