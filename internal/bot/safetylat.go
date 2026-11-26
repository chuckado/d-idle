package bot

import (
	"context"
	"fmt"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/gclock"
)

// Safety latency (Fable review 2026-11-25 P3 a): the time from the moment a safety verdict is taken (the guard saw
// Fighter.Safety return something) to the moment the input for it left the bot. Before this nothing measured it; the
// 02:51 death on 2026-11-26 had `lastSafety=chicken` and three untouched potions, with no way to tell a slow reflex from
// a lost one. Drinks count to the end of Fighter.Act (the key is the whole action); portal, chicken and crowd moves count
// to the start of Fighter.Act, because the cast that follows is the action's own work, not the wait for the loop.

// safetyLat holds the open verdict and the statistics of the current minute and of the last death window.
type safetyLat struct {
	verdictAt time.Time    // set by the guard when it takes a verdict, taken by carryOut or the reflex
	hardA     SafetyAction // a portal, chicken or crowd cancel the action has not answered yet (the reflex deadline)
	hardAt    time.Time
	forced    SafetyAction // the guard carried hardA out itself after the deadline; the loop must not do it again

	n, sumMs, maxMs int // since the last minute line
	slow            int // sends over safetySlowMs since the last minute line
	reflex          int // drinks the safety check sent itself since the last minute line (bot.reflexDrink)
	lastMs          int
	lastAction      SafetyAction
	lastAt          time.Time
	total           int
}

// safetySlowMs is the latency above which a send counts as slow in the minute line. Two guard ticks at the default
// bot.watchMs plus the potion gap; a reading above it means the action did not return in time.
const safetySlowMs = 300

// noteVerdict opens a measurement. Called by the guard before it cancels the action or sends the reflex.
func (r *Runner) noteVerdict(now time.Time) {
	r.obMu.Lock()
	r.ob.sl.verdictAt = now
	r.obMu.Unlock()
}

// noteSent closes the measurement of action a at time now. A send without an open verdict (hazard, paused fight) has
// nothing to measure and is not counted.
func (r *Runner) noteSent(a SafetyAction, now time.Time, reflex bool) {
	r.obMu.Lock()
	defer r.obMu.Unlock()
	l := &r.ob.sl
	if l.verdictAt.IsZero() {
		return
	}
	ms := int(now.Sub(l.verdictAt) / time.Millisecond)
	if ms < 0 {
		ms = 0
	}
	l.verdictAt = time.Time{}
	l.n++
	l.total++
	l.sumMs += ms
	l.maxMs = max(l.maxMs, ms)
	if ms > safetySlowMs {
		l.slow++
	}
	if reflex {
		l.reflex++
	}
	l.lastMs, l.lastAction, l.lastAt = ms, a, now
}

// minuteKV is the minute-line view; the caller holds obMu. Nothing when no safety action was sent in the minute, so
// a quiet line is unchanged.
func (l *safetyLat) minuteKV() []any {
	if l.n == 0 {
		return nil
	}
	kv := []any{"safetyN", l.n, "safetyLatAvgMs", l.sumMs / l.n, "safetyLatMaxMs", l.maxMs, "safetySlow", l.slow, "reflexDrinks", l.reflex}
	l.n, l.sumMs, l.maxMs, l.slow, l.reflex = 0, 0, 0, 0, 0
	return kv
}

// lastWithin returns the last measured latency when it was sent within window before now.
func (l *safetyLat) lastWithin(now time.Time, window time.Duration) (ms int, a SafetyAction, ok bool) {
	if l.lastAt.IsZero() || now.Sub(l.lastAt) > window {
		return 0, SafetyNone, false
	}
	return l.lastMs, l.lastAction, true
}

// ReflexDrinker is the optional Fighter capability behind bot.reflexDrink: send a belt drink from the guard goroutine
// without stopping the running action (combat.Fighter.ReflexDrink). handled is false when only the action can do it
// (the pack path opens a panel); the guard then cancels as before.
type ReflexDrinker interface {
	ReflexDrink(ctx context.Context, a SafetyAction) (handled bool, err error)
}

// reflexDrink sends the drink for a itself when the knob is on and the fighter can. It returns the action the guard must
// cancel with when it did not handle it: the same a, or the town portal when the potion burst rule turns the drink into
// one (earlyExit, run here once because the loop's carryOut will not see a drink that was sent).
func (r *Runner) reflexDrink(snap *d2.Snapshot, a SafetyAction) (SafetyAction, bool) {
	if r.e.Tune.Bot.ReflexDrink == 0 || (a != SafetyDrinkLife && a != SafetyDrinkRejuv) {
		return a, false
	}
	rd, ok := r.d.Fighter.(ReflexDrinker)
	if !ok {
		return a, false
	}
	if a2 := r.earlyExit(snap, a); a2 != a {
		return a2, false
	}
	r.safetyBusy.Add(1)
	defer r.safetyBusy.Add(-1)
	ctx, cancel := gclock.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	handled, err := rd.ReflexDrink(ctx, a)
	if !handled {
		if err != nil {
			r.e.Log.Debug("reflex drink not sent, cancelling the action instead", "action", a.String(), "err", err)
		}
		return a, false
	}
	if a == SafetyDrinkRejuv {
		r.zonePotion()
	}
	r.noteSent(a, r.e.Clock.Now(), true)
	if p := snap.Player; p != nil {
		r.e.Log.Info(fmt.Sprintf("guard: reflex drank %s at life %d%% (state=%s)", potionName(a), p.Life*100/max(p.MaxLife, 1), r.State()))
	}
	return a, true
}

// hardAction is a safety action that stops the running action and is carried out by the loop afterwards.
func hardAction(a SafetyAction) bool {
	switch a {
	case SafetyTownPortal, SafetyChicken, SafetyCrowdAbandon, SafetyCrowdRetreat:
		return true
	}
	return false
}

// armHard notes a cancel for a hard action; the guard gives the action bot.reflexDeadlineMs to return.
func (r *Runner) armHard(a SafetyAction, now time.Time) {
	if !hardAction(a) || r.e.Tune.Bot.ReflexDeadlineMs <= 0 {
		return
	}
	r.obMu.Lock()
	r.ob.sl.hardA, r.ob.sl.hardAt, r.ob.sl.forced = a, now, SafetyNone
	r.obMu.Unlock()
}

// disarmHard is called when the action has returned: the loop carries the safety action out.
func (r *Runner) disarmHard() {
	r.obMu.Lock()
	r.ob.sl.hardA = SafetyNone
	r.obMu.Unlock()
}

// forceIfLate carries the armed hard action out from the guard when the cancelled action has not returned within the
// deadline. It reports true when there is nothing left to wait for (nothing armed, or the action was forced).
func (r *Runner) forceIfLate(now time.Time) bool {
	r.obMu.Lock()
	l := &r.ob.sl
	a, at := l.hardA, l.hardAt
	if a == SafetyNone {
		r.obMu.Unlock()
		return true
	}
	if now.Sub(at) < time.Duration(r.e.Tune.Bot.ReflexDeadlineMs)*time.Millisecond {
		r.obMu.Unlock()
		return false
	}
	r.obMu.Unlock()
	r.forceNow()
	return true
}

// forceNow carries the armed hard action out from the guard (the deadline has passed).
func (r *Runner) forceNow() {
	r.obMu.Lock()
	l := &r.ob.sl
	a, at := l.hardA, l.hardAt
	if a == SafetyNone {
		r.obMu.Unlock()
		return
	}
	l.hardA, l.forced = SafetyNone, a
	r.obMu.Unlock()
	now := r.e.Clock.Now()
	r.e.Log.Warn("reflex: the cancelled action did not return in time, the guard carries out the safety action itself", "action", a.String(), "waitedMs", now.Sub(at).Milliseconds())
	if err := r.carryOut(context.Background(), a); err != nil {
		r.e.Log.Warn("reflex: forced safety action failed", "action", a.String(), "err", err)
	}
}

// takeForced reports (once) that the guard already carried out the safety action.
func (r *Runner) takeForced() bool {
	r.obMu.Lock()
	defer r.obMu.Unlock()
	f := r.ob.sl.forced != SafetyNone
	r.ob.sl.forced = SafetyNone
	return f
}

// holdForHard keeps the guard alive after a hard cancel until the action returns or the deadline passes (game.exe path;
// the lockstep watch calls forceIfLate itself). It waits on a timer, not on Clock.Sleep: a sleep on a virtual clock
// advances it, and the wait of a guard that is about to be released must not move the time the rest of the bot sees
// (TestReactionTimeDelaysInterrupt was flaky under -race for exactly that).
func (r *Runner) holdForHard(done <-chan struct{}) {
	r.obMu.Lock()
	armed := r.ob.sl.hardA != SafetyNone
	r.obMu.Unlock()
	if !armed {
		return
	}
	hctx, cancel := gclock.WithTimeout(context.Background(), time.Duration(r.e.Tune.Bot.ReflexDeadlineMs)*time.Millisecond)
	defer cancel()
	select {
	case <-done:
	case <-hctx.Done():
		r.forceNow()
	}
}
