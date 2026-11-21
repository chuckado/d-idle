package combat

import (
	"time"

	"d2lod-bot/internal/d2"
)

// The expectation ledger (card dlv-expect, combat plan section 3.4): every paid cast records what should be
// observable and by when; later ticks settle it. Three consecutive failures of one (skill, reason) demote the skill
// for a back-off time. Modes of combat.deliveryExpect: 0 off (nothing recorded), 1 log only (record, settle, count and
// log, decisions unchanged), 2 demote (a demoted skill is skipped by the attack and buff passes).
//
// Only the reason "mana" exists today: an expectation is a closure, so the other reasons of the plan (state present,
// census grew, target lost life, position changed) join with the drivers that know what to look for. The ad-hoc
// memories it will absorb (capSeen, curse tries, leftFail, buff cut) are untouched in this card (U: they stay until
// mode 2 has been seen live).
const (
	expectOff    = 0
	expectLog    = 1
	expectDemote = 2

	expectReasonMana = "mana"
	expectStrikes    = 3 // consecutive failures that demote
)

// Back-off of a demoted (skill, reason): the first demotion lasts expectBackoff, each next one doubles, up to the cap (U).
const (
	expectBackoff    = 20 * time.Second
	expectBackoffMax = 5 * time.Minute
)

type expectKey struct{ skill, reason string }

// expectation is one pending cast. check returns done when the outcome is known; ok then says met. cut means the
// cast was interrupted by a hit, which says nothing about the skill.
type expectation struct {
	key      expectKey
	deadline time.Time
	check    func(p *d2.Player) (done, ok, cut bool)
}

// ExpectCounters are the telemetry counters of one skill.
type ExpectCounters struct{ OK, NoEffect, Cut, Demotions int }

type expectLedger struct {
	pending  []expectation
	fails    map[expectKey]int
	until    map[expectKey]time.Time
	backoff  map[expectKey]time.Duration
	counters map[string]*ExpectCounters
}

func (l *expectLedger) counter(skill string) *ExpectCounters {
	if l.counters == nil {
		l.counters = map[string]*ExpectCounters{}
	}
	c := l.counters[skill]
	if c == nil {
		c = &ExpectCounters{}
		l.counters[skill] = c
	}
	return c
}

// ExpectCounters returns a copy of the per-skill counters (telemetry, tests).
func (f *Fighter) ExpectCounters() map[string]ExpectCounters {
	out := make(map[string]ExpectCounters, len(f.expect.counters))
	for n, c := range f.expect.counters {
		out[n] = *c
	}
	return out
}

func (f *Fighter) expectMode() int { return f.tune().DeliveryExpect }

// expectRecord notes the expectation of a cast just sent. A skill that costs nothing has nothing to observe by mana.
func (f *Fighter) expectRecord(snap *d2.Snapshot, sp SkillSpec) {
	if f.expectMode() == expectOff || sp.Cost <= 0 {
		return
	}
	mana, cost := snap.Player.Mana, sp.Cost
	window := time.Duration(max(castTicks(sp.Name, f.tune().CastFCR, f.D2D), 10)+4) * tickMs * time.Millisecond
	f.expect.pending = append(f.expect.pending, expectation{
		key:      expectKey{sp.Name, expectReasonMana},
		deadline: f.now().Add(window),
		check: func(p *d2.Player) (bool, bool, bool) {
			if mana-p.Mana >= cost {
				return true, true, false
			}
			if p.Mode == modeGetHit {
				return true, false, true
			}
			return false, false, false
		},
	})
}

// expectTick settles the pending expectations against the fresh player; the fight loop calls it once per tick.
func (f *Fighter) expectTick(p *d2.Player) {
	if f.expectMode() == expectOff || len(f.expect.pending) == 0 {
		return
	}
	l, now := &f.expect, f.now()
	keep := l.pending[:0]
	for _, e := range l.pending {
		done, ok, cut := e.check(p)
		if !done && now.Before(e.deadline) {
			keep = append(keep, e)
			continue
		}
		f.expectSettle(e.key, done && ok, done && cut, now)
	}
	l.pending = keep
}

func (f *Fighter) expectSettle(k expectKey, ok, cut bool, now time.Time) {
	l := &f.expect
	c := l.counter(k.skill)
	switch {
	case ok:
		c.OK++
		delete(l.fails, k)
		delete(l.backoff, k)
	case cut:
		c.Cut++
	default:
		c.NoEffect++
		if l.fails == nil {
			l.fails, l.until, l.backoff = map[expectKey]int{}, map[expectKey]time.Time{}, map[expectKey]time.Duration{}
		}
		l.fails[k]++
		f.log().Debug("delivery expect failed", "skill", k.skill, "reason", k.reason, "strikes", l.fails[k])
		if l.fails[k] >= expectStrikes {
			b := l.backoff[k]
			b = min(max(2*b, expectBackoff), expectBackoffMax)
			l.backoff[k] = b
			l.until[k] = now.Add(b)
			l.fails[k] = 0
			c.Demotions++
			f.log().Info("delivery expect demoted", "skill", k.skill, "reason", k.reason, "backoffMs", b.Milliseconds(), "mode", f.expectMode())
		}
	}
}

// expectDemoted reports whether the skill is demoted now; only mode 2 acts on it.
func (f *Fighter) expectDemoted(name string) bool {
	if f.expectMode() != expectDemote {
		return false
	}
	now := f.now()
	for k, t := range f.expect.until {
		if k.skill == name && now.Before(t) {
			return true
		}
	}
	return false
}
