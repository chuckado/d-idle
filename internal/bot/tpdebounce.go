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
}

// look returns the accepted verdict after seeing raw: the first look is taken as is, true is accepted at once, false
// only after every look for tpDebounceSpan said false.
func (d *tpDebounce) look(raw bool, now time.Time) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.set || raw {
		d.set, d.val, d.emptyActive = true, raw, false
		return raw
	}
	if !d.val {
		return false
	}
	if !d.emptyActive {
		d.emptyActive, d.emptySince = true, now
	}
	if now.Sub(d.emptySince) >= tpDebounceSpan {
		d.val = false
	}
	return d.val
}
