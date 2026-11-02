package bot

import "time"

// Debounce of the town-portal verdict for the poverty check (docs/soakstall.md, povertyflip). Gate 1 showed HasTP
// flipping per area (82 of 86 "poverty mode: on" lines said no town portal while 131 scrolls were bought); the code
// shows no evident root cause (PortalSource only takes ContInventory items with a positive PortalQty; the per-area
// inv-tp-dump line of docs/hastpdump.md will tell), so a verdict must hold before poverty mode may see it.
const (
	tpDebounceLooks = 3
	tpDebounceSpan  = 2 * time.Minute // virtual
)

type tpDebounce struct {
	set, val  bool // val is the accepted verdict once set
	cand      bool
	n         int
	candSince time.Time
}

// look returns the accepted verdict after seeing raw: the first look is taken as is, a change needs tpDebounceLooks
// consecutive looks or tpDebounceSpan without a contrary look.
func (d *tpDebounce) look(raw bool, now time.Time) bool {
	if !d.set {
		d.set, d.val = true, raw
		return raw
	}
	if raw == d.val {
		d.n = 0
		return d.val
	}
	if d.n == 0 || d.cand != raw {
		d.cand, d.n, d.candSince = raw, 0, now
	}
	d.n++
	if d.n >= tpDebounceLooks || now.Sub(d.candSince) >= tpDebounceSpan {
		d.val, d.n = raw, 0
	}
	return d.val
}
