package merc

import (
	"fmt"
	"sync"
	"time"
)

// Revive mercenary as a player command (card botmerc, docs/merc-revive.md). The player asks; the bot goes home at the
// next safe moment and the EXISTING revive step runs there, so the gold bar, the hourly cap and the reserve are the
// policy's own (ShouldRevive, V read: the game wants gold only). This file holds the request, its button and its
// cooldown; it never decides anything the policy does not.

// ViewSchema is the schema word of GET /api/merc.
const ViewSchema = "watch.merc/1"

// PlayerReviveCooldown is how long a failed player revive keeps the button off (the repair trip's failed-request pause
// is the model; U live which value suits the NPC walk).
const PlayerReviveCooldown = 5 * time.Minute

// Request states.
const (
	ReqQueued    = "queued"
	ReqDoing     = "doing"
	ReqDone      = "done"
	ReqFailed    = "failed"
	ReqCancelled = "cancelled"
)

// Button actions.
const (
	ActionRequest = "request"
	ActionCancel  = "cancel"
	ActionNone    = "none"
)

// Button is what the page may press, the same shape as the repair button.
type Button struct {
	Action  string `json:"action"`
	Label   string `json:"label"`
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason,omitempty"`
}

// PlayerRequest is the open or the last request.
type PlayerRequest struct {
	State  string `json:"state"`
	Reason string `json:"reason,omitempty"` // why it failed
	At     int64  `json:"at"`
}

// PlayerView is the reply of GET /api/merc.
type PlayerView struct {
	Schema  string         `json:"schema"`
	State   string         `json:"state"` // alive, dead or none (no merc hired, or not read yet)
	Price   int            `json:"price"`
	Gold    int            `json:"gold"`
	Need    int            `json:"need"` // the policy's bar: price x the revive multiple plus the revive reserve
	Button  Button         `json:"button"`
	Request *PlayerRequest `json:"request,omitempty"`
}

// Refusal is returned by Request and Cancel; Kind is the code of the watch reply.
type Refusal struct{ Code, Reason string }

func (r *Refusal) Error() string { return r.Reason }
func (r *Refusal) Kind() string  { return r.Code }

// PlayerRevive holds the one request. The zero value is not usable: use NewPlayerRevive.
type PlayerRevive struct {
	now func() time.Time

	mu      sync.Mutex
	req     *PlayerRequest
	coolEnd time.Time
}

// NewPlayerRevive: now is the env clock (virtual under lockstep).
func NewPlayerRevive(now func() time.Time) *PlayerRevive {
	if now == nil {
		now = time.Now
	}
	return &PlayerRevive{now: now}
}

// Inputs are the facts a decision needs, read by the wiring from the same sources as the revive step.
type Inputs struct {
	State MercState
	Ctx   Ctx
	Quote int // the real revive quote when one was read, else 0 (the formula estimate)
}

// verdict says whether a revive could start now and, when not, the reason for the button.
func verdict(in Inputs) (ok bool, price, need int, reason string) {
	dec, price := ShouldRevive(in.State, in.Ctx, in.Quote)
	if price > 0 {
		need = ReviveNeed(in.Ctx, price)
	}
	if dec.Do {
		return true, price, need, ""
	}
	switch dec.Reason {
	case ReasonUnknown:
		reason = "the mercenary has not been read yet"
	case ReasonNotOwned:
		reason = "no mercenary is hired"
	case ReasonAlive:
		reason = "the mercenary is alive"
	case ReasonNoLevel:
		reason = "the mercenary's level is not known yet"
	case ReasonCap:
		reason = "the hourly revive limit is reached"
	default: // gold, stopped
		reason = fmt.Sprintf("need %d more gold", max(need-in.Ctx.Gold, 1))
	}
	return false, price, need, reason
}

func stateWord(s MercState) string {
	switch {
	case !s.Known || !s.Owned:
		return "none"
	case s.Alive:
		return "alive"
	}
	return "dead"
}

// View is the answer of GET /api/merc.
func (p *PlayerRevive) View(in Inputs) PlayerView {
	ok, price, need, why := verdict(in)
	v := PlayerView{Schema: ViewSchema, State: stateWord(in.State), Price: price, Gold: in.Ctx.Gold, Need: need}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.req != nil {
		r := *p.req
		v.Request = &r
	}
	label := "Revive mercenary"
	switch {
	case p.req != nil && p.req.State == ReqQueued:
		v.Button = Button{Action: ActionCancel, Label: "Cancel revive", Enabled: true}
	case p.req != nil && p.req.State == ReqDoing:
		v.Button = Button{Action: ActionNone, Label: label, Reason: "the bot is reviving him now"}
	case !ok:
		v.Button = Button{Action: ActionNone, Label: label, Reason: why}
	case p.now().Before(p.coolEnd):
		v.Button = Button{Action: ActionNone, Label: label, Reason: fmt.Sprintf("the last revive failed: try again in %d min", mins(p.coolEnd.Sub(p.now())))}
	default:
		v.Button = Button{Action: ActionRequest, Label: label, Enabled: true}
	}
	return v
}

func mins(d time.Duration) int { return int((d + time.Minute - 1) / time.Minute) }

// Request opens the request when the button allows it.
func (p *PlayerRevive) Request(in Inputs) (PlayerRequest, error) {
	v := p.View(in)
	if v.Button.Action != ActionRequest {
		code := "state"
		if v.Request != nil && (v.Request.State == ReqQueued || v.Request.State == ReqDoing) {
			code = "queued"
		}
		return PlayerRequest{}, &Refusal{code, v.Button.Reason}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.req = &PlayerRequest{State: ReqQueued, At: p.now().Unix()}
	return *p.req, nil
}

// Cancel withdraws a queued request.
func (p *PlayerRevive) Cancel() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.req == nil || p.req.State != ReqQueued {
		return &Refusal{"none", "there is no waiting revive request"}
	}
	p.req.State = ReqCancelled
	return nil
}

// Ready says whether the bot should go home for the request now: queued and the policy would pay.
func (p *PlayerRevive) Ready(in Inputs) bool {
	p.mu.Lock()
	queued := p.req != nil && p.req.State == ReqQueued
	p.mu.Unlock()
	if !queued {
		return false
	}
	ok, _, _, _ := verdict(in)
	return ok
}

// Queued says whether a request is waiting (for the trip or the town step).
func (p *PlayerRevive) Queued() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.req != nil && p.req.State == ReqQueued
}

// Start marks the queued request as running; false when none is queued.
func (p *PlayerRevive) Start() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.req == nil || p.req.State != ReqQueued {
		return false
	}
	p.req.State = ReqDoing
	return true
}

// Postpone puts a running request back to queued (a fight or a cancelled context; nothing was paid).
func (p *PlayerRevive) Postpone() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.req != nil && p.req.State == ReqDoing {
		p.req.State = ReqQueued
	}
}

// Finish settles a running request: done, or failed with the reason. A failure that is not about gold starts the
// cooldown (cool true).
func (p *PlayerRevive) Finish(ok bool, reason string, cool bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.req == nil || p.req.State != ReqDoing {
		return
	}
	if ok {
		p.req.State = ReqDone
		return
	}
	p.req.State, p.req.Reason = ReqFailed, reason
	if cool {
		p.coolEnd = p.now().Add(PlayerReviveCooldown)
	}
}

// OffView is the reply while the policy is not running or nothing was read yet: no merc, the button off with why.
func OffView(why string) PlayerView {
	return PlayerView{Schema: ViewSchema, State: "none", Button: Button{Action: ActionNone, Label: "Revive mercenary", Reason: why}}
}
