package bot

import (
	"context"
	"errors"
	"time"

	"d2lod-bot/internal/town"
)

// portalRetryDue says whether an action that failed with err and was cut by safety action a should be retried as a town
// trip: a cast cancelled before the click (nothing spent) by a transient safety action (a drink or a retreat), once per
// stretch, and only when no trip is pending (a trip's next tick retries the cast by itself). Chicken, town portal and
// crowd abandon carry their own way out. PortalRetryWaitSec 0 turns it off. Pinned by TestPortalCancelRetry.
func portalRetryDue(err error, a SafetyAction, tripPending, used bool, waitSec int) bool {
	if waitSec <= 0 || tripPending || used || !errors.Is(err, town.ErrPortalCancelled) {
		return false
	}
	switch a {
	case SafetyDrinkLife, SafetyDrinkMana, SafetyDrinkRejuv, SafetyCrowdRetreat:
		return true
	}
	return false
}

// retryCancelledPortal runs after handleSafety: waits the bounded time for the fight to settle and queues one town trip
// back from the same area, so the hero walks the potion trip it set out on instead of re-planning into the crowd.
func (r *Runner) retryCancelledPortal(ctx context.Context, err error, a SafetyAction) {
	l := &r.loop
	if c, ok := r.d.Town.(interface{ TakePortalCancelled() bool }); ok && c.TakePortalCancelled() {
		err = town.ErrPortalCancelled
	}
	if !portalRetryDue(err, a, l.trip != nil, l.portalRetried, r.e.Tune.Bot.PortalRetryWaitSec) {
		return
	}
	s := r.latest()
	if s == nil || s.Player == nil || s.Player.InTown() {
		return
	}
	l.portalRetried = true
	r.e.Log.Info("town portal cast was cancelled before the click: retrying once", "area", s.Player.Area, "action", a, "waitSec", r.e.Tune.Bot.PortalRetryWaitSec)
	r.pause(ctx, time.Duration(r.e.Tune.Bot.PortalRetryWaitSec)*time.Second)
	l.trip = &trip{plan: r.fullChores("portal cast retry"), returnTo: s.Player.Area}
}
