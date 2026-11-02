package bot

import (
	"sync"
	"time"
)

// Debounce of the town-portal verdict (docs/soakstall.md, povertyflip and tpglitch). The tome's stat 70 drops out of
// the dump while the tome is full (V: 95 provable cases in the inv-tp-dump lines of gate1r2-20261101, 2.6 s median,
// 63 s at most), so "no portal source" only counts once it has held for tpDebounceSpan of virtual time. Gaining a
// source is taken at once. It sits at the single source of the verdict (hasPortalSource) so the kit gate, the poverty
// check and every other reader see the same value.
const tpDebounceSpan = 2 * time.Minute // virtual; above the longest drop-out seen

type tpDebounce struct {
	mu          sync.Mutex
	set, val    bool // val is the accepted verdict once set
	emptySince  time.Time
	emptyActive bool
	looks       int // false looks in the current run of them
}

// tpEvent is what one look decided worth a log line: kind "hold" (a run of false looks ended in true before the span
// ran out: the debounce kept the verdict) or "flip" (the run reached the span: the verdict became false). Zero kind: nothing.
type tpEvent struct {
	kind  string
	looks int           // false looks in the run
	held  time.Duration // virtual time from the first false look to this one
}

// look returns the accepted verdict after seeing raw: the first look is taken as is, true is accepted at once, false
// only after every look for tpDebounceSpan said false. The event is non-zero once per false run, when it ends.
func (d *tpDebounce) look(raw bool, now time.Time) (bool, tpEvent) {
	d.mu.Lock()
	defer d.mu.Unlock()
	var ev tpEvent
	if !d.set || raw {
		if d.set && d.val && d.emptyActive {
			ev = tpEvent{"hold", d.looks, now.Sub(d.emptySince)}
		}
		d.set, d.val, d.emptyActive, d.looks = true, raw, false, 0
		return raw, ev
	}
	if !d.val {
		return false, ev
	}
	if !d.emptyActive {
		d.emptyActive, d.emptySince, d.looks = true, now, 0
	}
	d.looks++
	if now.Sub(d.emptySince) >= tpDebounceSpan {
		d.val, d.emptyActive = false, false
		ev = tpEvent{"flip", d.looks, now.Sub(d.emptySince)}
	}
	return d.val, ev
}
