package bot

import (
	"time"
)

// Safety latency (Fable review 2026-11-25 P3 a): the time from the moment a safety verdict is taken (the guard saw
// Fighter.Safety return something) to the moment the input for it left the bot. Before this nothing measured it; the
// 02:51 death on 2026-11-26 had `lastSafety=chicken` and three untouched potions, with no way to tell a slow reflex from
// a lost one. Drinks count to the end of Fighter.Act (the key is the whole action); portal, chicken and crowd moves count
// to the start of Fighter.Act, because the cast that follows is the action's own work, not the wait for the loop.

// safetyLat holds the open verdict and the statistics of the current minute and of the last death window.
type safetyLat struct {
	verdictAt time.Time // set by the guard when it takes a verdict, taken by carryOut or the reflex

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
