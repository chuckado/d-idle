package bot

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"d2lod-bot/internal/d2"
)

// The player's Pause command (docs/pause.md, user answer 6): the hero starts nothing new, walks home at the next safe
// moment (the errand trip's rule, safeMoment) and only then enters the paused state, so a pause never leaves the hero in
// a field. Safety (chicken, potions, portal) keeps running while it goes home because it ranks above every player
// command and is not touched here. Resume clears the request. The request is persisted so a restart stays paused.

// Player pause states, the values of PlayerPauseView.State.
const (
	PauseRunning   = "running"
	PauseGoingHome = "going_home"
	PausePaused    = "paused"
)

// PlayerPauseSchema is the schema field of the watch route's reply.
const PlayerPauseSchema = "watch.pause/1"

// pauseHomeBackoff is the wait after a failed walk home before the next try.
const pauseHomeBackoff = 30 * time.Second

// PlayerPauseView is what GET /api/pause shows.
type PlayerPauseView struct {
	Schema       string      `json:"schema"`
	State        string      `json:"state"`
	Since        *time.Time  `json:"since,omitempty"`
	Reason       string      `json:"reason,omitempty"`
	Frozen       bool        `json:"frozen,omitempty"`     // the game itself is stopped (freeze.go); State is then "paused"
	CanFreeze    bool        `json:"can_freeze,omitempty"` // this backend can freeze the game: the page shows Pause
	CapMin       int         `json:"cap_min,omitempty"`    // minutes until a frozen pause ends by itself
	EndsAt       *time.Time  `json:"ends_at,omitempty"`
	Ended        *PauseEnded `json:"ended,omitempty"`         // a frozen pause that just ended (Needs you info row)
	Error        string      `json:"error,omitempty"`         // why the last press did nothing
	MonstersNear bool        `json:"monsters_near,omitempty"` // the hero stood near monsters when the game froze (info only)
}

type playerPause struct {
	want   bool      // the player asked for a pause
	held   bool      // the runner is paused because of it (in town)
	since  time.Time // when the request was made
	reason string    // what the pause waits for, or why it stopped
	path   string
	nextAt time.Time // no walk home before this (after a failed one)
}

type pauseFile struct {
	On    bool      `json:"on"`
	Since time.Time `json:"since,omitempty"`
}

// SetPausePath loads the persisted request and keeps it up to date. A persisted pause makes the hero go home and stop
// as soon as the loop runs, so a restart stays paused. Never call it in a measured run (D2BOT_RUN).
func (r *Runner) SetPausePath(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.plp.path = path
	if b, err := os.ReadFile(path); err == nil {
		var f pauseFile
		if json.Unmarshal(b, &f) == nil && f.On {
			r.plp.want, r.plp.since, r.plp.reason = true, f.Since, "waiting for a safe moment to go home"
		}
	}
}

// PlayerPause sets or clears the player's pause request and returns the new view. With a backend that can freeze the
// game the pause is the frozen one (freeze.go); otherwise it goes home first.
func (r *Runner) PlayerPause(on bool) PlayerPauseView { return r.PlayerPauseMode(on, nil) }

// PlayerPauseMode is PlayerPause with the mode chosen: freeze nil means "freeze when the backend can", false the older
// walk-home pause, true a freeze that is refused (view.Error) when the backend cannot.
func (r *Runner) PlayerPauseMode(on bool, freeze *bool) PlayerPauseView {
	if !on {
		if r.Frozen() {
			if err := r.freezeOff(false, false); err != nil {
				v := r.PlayerPauseView()
				v.Error = "could not resume: " + err.Error()
				return v
			}
		}
	} else if freeze == nil || *freeze {
		if r.canFreeze() {
			if err := r.freezeOn(); err != nil {
				v := r.PlayerPauseView()
				v.Error = "could not freeze the game: " + err.Error()
				return v
			}
			return r.PlayerPauseView()
		}
		if freeze != nil {
			v := r.PlayerPauseView()
			v.Error = "this game cannot be frozen here"
			return v
		}
	}
	return r.playerPauseWalk(on)
}

func (r *Runner) playerPauseWalk(on bool) PlayerPauseView {
	r.mu.Lock()
	defer r.mu.Unlock()
	p := &r.plp
	switch {
	case on && !p.want:
		p.want, p.since, p.reason, p.nextAt = true, r.e.Clock.Now(), "waiting for a safe moment to go home", time.Time{}
	case !on && p.want:
		held := p.held
		p.want, p.held, p.reason, p.since = false, false, "", time.Time{}
		if held {
			r.paused = false
			r.autoResumeAt = time.Time{}
		}
	}
	p.save()
	return r.pauseViewLocked()
}

// PlayerPauseView reports the state.
func (r *Runner) PlayerPauseView() PlayerPauseView {
	can := r.canFreeze()
	r.mu.Lock()
	defer r.mu.Unlock()
	v := r.pauseViewLocked()
	v.CanFreeze = can
	return v
}

func (r *Runner) pauseViewLocked() PlayerPauseView {
	p := &r.plp
	v := PlayerPauseView{Schema: PlayerPauseSchema, State: PauseRunning}
	switch {
	case p.held:
		v.State = PausePaused
	case p.want:
		v.State = PauseGoingHome
	}
	if p.want {
		since := p.since
		v.Since, v.Reason = &since, p.reason
	}
	r.frozenViewLocked(&v)
	return v
}

func (p *playerPause) save() {
	if p.path == "" {
		return
	}
	b, _ := json.Marshal(pauseFile{On: p.want, Since: p.since})
	if os.MkdirAll(filepath.Dir(p.path), 0o755) == nil {
		tmp := p.path + ".tmp"
		if os.WriteFile(tmp, b, 0o644) == nil {
			os.Rename(tmp, p.path)
		}
	}
}

func (r *Runner) setPauseReason(s string) {
	r.mu.Lock()
	r.plp.reason = s
	r.mu.Unlock()
}

// pauseDue is the guard's check: a pause request waits, the hero is in the field and the moment is safe, so a walk in
// progress is cancelled and the loop takes the hero home.
func (r *Runner) pauseDue(s *d2.Snapshot, now time.Time) bool {
	r.mu.Lock()
	want, held, next := r.plp.want, r.plp.held, r.plp.nextAt
	r.mu.Unlock()
	if !want || held || s.Player.InTown() || now.Before(next) {
		return false
	}
	return r.hostileNear(s, r.e.Tune.Bot.ClearRadius) == 0 && r.hostileWithin(s, r.e.Tune.Bot.ClearRadius) == 0
}

// playerPauseStep runs first among the player's steps of runGoal. In town (and no town trip running) the runner pauses
// at once; in the field it walks home at the safe moment and pauses on arrival. It reports whether it acted.
func (r *Runner) playerPauseStep(ctx context.Context, snap *d2.Snapshot, c control) bool {
	r.mu.Lock()
	want, held, next := r.plp.want, r.plp.held, r.plp.nextAt
	r.mu.Unlock()
	if !want || held || snap == nil || snap.Player == nil || c.paused {
		return false
	}
	if snap.Player.InTown() {
		if r.loop.trip != nil {
			r.setPauseReason("finishing the town trip first")
			return false
		}
		r.mu.Lock()
		r.plp.held, r.plp.reason = true, "paused in town"
		r.paused = true
		r.activity = "Pausing"
		r.mu.Unlock()
		return true
	}
	if r.d.Town == nil {
		r.setPauseReason("no way home is wired")
		return false
	}
	if r.e.Clock.Now().Before(next) {
		return false
	}
	if _, ok, why := r.safeMoment(snap); !ok {
		r.setPauseReason("waiting for a safe moment: " + why)
		return false
	}
	r.setPauseReason("going home")
	if r.do(ctx, c, StateResupplying, "Going to town to pause", sec(r.cfg.TownTimeoutSec), "pause trip", r.d.Town.ToTown) {
		return true
	}
	r.mu.Lock()
	r.plp.nextAt = r.e.Clock.Now().Add(pauseHomeBackoff)
	r.mu.Unlock()
	r.e.Log.Warn("pause: the walk home failed; trying again shortly")
	return true
}
