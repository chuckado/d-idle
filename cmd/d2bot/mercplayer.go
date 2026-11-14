package main

import (
	"context"
	"strings"

	"d2lod-bot/internal/bot"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/merc"
	"d2lod-bot/internal/watch"
)

// Revive mercenary as a player command (card botmerc, docs/merc-revive.md): the glue between the watch routes, the
// request object (merc.PlayerRevive) and the bot's player-command machinery. The trip home is the errand trip of the
// inventory desk (bot.playerKinds), the revive itself is the EXISTING reviveStep, so the gold-only bar, the hourly cap
// and the reserve stay the policy's own. A survival action is never held back by it: the request is only looked at
// where the player's other commands are (the guard's safe moment).

// player returns the request object, created on first use.
func (w *mercWire) player() *merc.PlayerRevive {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.pr == nil {
		w.pr = merc.NewPlayerRevive(w.now)
	}
	return w.pr
}

func (w *mercWire) noteSnap(s *d2.Snapshot) {
	if s == nil || s.Player == nil {
		return
	}
	w.mu.Lock()
	w.lastSnap = s
	w.mu.Unlock()
}

// snap is the runner's newest snapshot (set by mercGear), else the one the trip poll showed.
func (w *mercWire) snap() *d2.Snapshot {
	w.mu.Lock()
	src, last := w.snapSrc, w.lastSnap
	w.mu.Unlock()
	if src != nil {
		if s := src(); s != nil && s.Player != nil {
			return s
		}
	}
	return last
}

// playerInputs are the facts of the revive decision, read from the same places as reviveStep.
func (w *mercWire) playerInputs() (merc.Inputs, bool) {
	s := w.snap()
	if s == nil {
		return merc.Inputs{}, false
	}
	st, err := w.State()
	if err != nil {
		return merc.Inputs{}, false
	}
	c := mercCtx(w, s)
	w.mu.Lock()
	quote := w.quoteFor(st.Level)
	w.mu.Unlock()
	return merc.Inputs{State: st, Ctx: c, Quote: quote}, true
}

// runPlayerRevive runs the queued request in town: the existing revive step decides and pays.
func (w *mercWire) runPlayerRevive(ctx context.Context) []string {
	pr := w.player()
	if !pr.Start() {
		return nil
	}
	s := w.snap()
	if s == nil || s.Player == nil || !s.Player.InTown() {
		pr.Postpone()
		return nil
	}
	what := w.reviveStep(ctx, s, false)
	line := "revive mercenary (your request): "
	switch {
	case ctx.Err() != nil:
		pr.Postpone()
		return []string{line + "postponed"}
	case what == "merc revived":
		pr.Finish(true, "", false)
	case what == "":
		if st, err := w.State(); err == nil && st.Known && st.Owned && st.Alive {
			pr.Finish(true, "", false) // the town visit revived him first
			what = "the mercenary is alive"
		} else {
			pr.Finish(false, "the mercenary could not be read", false)
			what = "the mercenary could not be read"
		}
	case what == "merc revive failed":
		pr.Finish(false, "the revive did not go through: the hireling could not be reached or the price changed", true)
	case strings.Contains(what, "backoff"):
		pr.Finish(false, "the last attempt failed a moment ago", true)
	default: // skipped for gold, the cap or no reviver: not a trip failure, so no cooldown
		pr.Finish(false, strings.TrimPrefix(what, "merc revive skipped: "), false)
	}
	return []string{line + what}
}

// mercGear adds the player's revive request to the inventory desk's player commands. It implements the optional parts
// of bot.PlayerGear (playerKinds, errandKinder) by passing the desk's own through.
type mercGear struct {
	inner bot.PlayerGear
	snap  func() *d2.Snapshot // the runner's newest snapshot
}

type kinds interface {
	NextKind() string
	ErrandsWaiting() int
}

// request returns the wiring and request when a revive request is waiting, else nil (the cheap path every tick).
func (g mercGear) request() (*mercWire, *merc.PlayerRevive) {
	w := latestMerc.Load()
	if w == nil {
		return nil, nil
	}
	w.mu.Lock()
	w.snapSrc = g.snap
	w.mu.Unlock()
	if pr := w.player(); pr.Queued() {
		return w, pr
	}
	return nil, nil
}

func (g mercGear) heroInTown(w *mercWire) bool {
	s := w.snap()
	return s != nil && s.Player != nil && s.Player.InTown()
}

func (g mercGear) Pending() int {
	n := 0
	if g.inner != nil {
		n = g.inner.Pending()
	}
	if w, _ := g.request(); w != nil && g.heroInTown(w) {
		n++
	}
	return n
}

func (g mercGear) RunPending(ctx context.Context) []string {
	var out []string
	if g.inner != nil {
		out = g.inner.RunPending(ctx)
	}
	if w, _ := g.request(); w != nil && g.heroInTown(w) && ctx.Err() == nil {
		out = append(out, w.runPlayerRevive(ctx)...)
	}
	return out
}

func (g mercGear) NextKind() string {
	if k, ok := g.inner.(kinds); ok {
		if n := k.NextKind(); n != "" {
			return n
		}
	}
	if w, _ := g.request(); w != nil && g.heroInTown(w) {
		return "revive_merc"
	}
	return ""
}

// ErrandsWaiting counts the revive request while the hero is in the field and the policy would pay: the errand trip then
// walks home at the next safe moment. A request short of gold waits where it is (no trip for nothing).
func (g mercGear) ErrandsWaiting() int {
	n := 0
	if k, ok := g.inner.(kinds); ok {
		n = k.ErrandsWaiting()
	}
	if w, pr := g.request(); w != nil && !g.heroInTown(w) {
		if in, ok := w.playerInputs(); ok && pr.Ready(in) {
			n++
		}
	}
	return n
}

func (g mercGear) ErrandKind() string {
	if k, ok := g.inner.(interface{ ErrandKind() string }); ok {
		if n := k.ErrandKind(); n != "" {
			return n
		}
	}
	return "revive_merc"
}

// mercAPIAdapter lets the watch routes reach the running merc wiring (late: the wiring exists once the town seam ran).
type mercAPIAdapter struct{}

var _ watch.MercDesk = mercAPIAdapter{}

func (mercAPIAdapter) view() (merc.PlayerView, *mercWire) {
	w := latestMerc.Load()
	if w == nil {
		return merc.OffView("the mercenary policy is not running (merc.enabled is 0)"), nil
	}
	in, ok := w.playerInputs()
	if !ok {
		return merc.OffView("waiting for the first look at the hero and the mercenary"), nil
	}
	return w.player().View(in), w
}

func (a mercAPIAdapter) View() any { v, _ := a.view(); return v }

func (a mercAPIAdapter) Request() (any, error) {
	v, w := a.view()
	if w == nil {
		return nil, &merc.Refusal{Code: "off", Reason: v.Button.Reason}
	}
	in, _ := w.playerInputs()
	return w.player().Request(in)
}

func (a mercAPIAdapter) Cancel() error {
	w := latestMerc.Load()
	if w == nil {
		return &merc.Refusal{Code: "off", Reason: "the mercenary policy is not running"}
	}
	return w.player().Cancel()
}

// mercAPI is what newFun's caller mounts: the routes are always there, the adapter answers `off` without a wiring.
func mercAPI() *watch.MercAPI {
	api := watch.NewMercAPI()
	api.Bind(mercAPIAdapter{})
	return api
}
