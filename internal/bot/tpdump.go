package bot

import (
	"fmt"
	"strings"
	"sync"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2"
)

// tpStatQuantity mirrors combat's private statQuantity (stack count, V live per docs/classes/smoke-amazon.md).
const tpStatQuantity = 70

// tpDumper rate-limits the inv-tp-dump line (docs/hastpdump.md): one line per area entry and one per HasTP change.
type tpDumper struct {
	mu      sync.Mutex
	seen    bool
	area    int
	verdict bool
}

// due reports whether a line is owed and why; it records the new state either way.
func (d *tpDumper) due(area int, verdict bool) (reason string, ok bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	switch {
	case !d.seen:
		reason = "first"
	case d.area != area:
		reason = "area"
	case d.verdict != verdict:
		reason = "verdict"
	}
	d.seen, d.area, d.verdict = true, area, verdict
	return reason, reason != ""
}

// dumpTP logs the inventory as hasPortalSource saw it. Same inputs as the verdict: the snapshot, the catalog and the
// UnitStats reader; no behaviour change. No-op with a nil Log or without an inventory.
func (r *Runner) dumpTP(s *d2.Snapshot, verdict bool, cat combat.ItemClasser, st combat.Stats, src *d2.Unit) {
	if r.e == nil || r.e.Log == nil || s == nil || s.Player == nil {
		return
	}
	reason, ok := r.tpd.due(s.Player.Area, verdict)
	if !ok {
		return
	}
	var items []string
	var portals []string
	if s.Inventory != nil {
		for i := range s.Inventory.Items {
			u := &s.Inventory.Items[i]
			if u.Item == nil {
				continue
			}
			items = append(items, fmt.Sprintf("%d@%v", u.Class, u.Item.Where))
			if !cat.Item(u.Class).Portal {
				continue
			}
			ic := cat.Item(u.Class)
			kind := "scroll"
			if ic.Tome {
				kind = "tome"
			}
			raw := "nostats"
			if st != nil {
				if m, err := st(u.Addr); err != nil {
					raw = "err:" + err.Error()
				} else if m.Has(tpStatQuantity) {
					raw = fmt.Sprint(m.Get(tpStatQuantity))
				} else {
					raw = "no70"
				}
			}
			portals = append(portals, fmt.Sprintf("%s class=%d where=%v addr=%#x counted=%d raw70=%s",
				kind, u.Class, u.Item.Where, u.Addr, combat.PortalQty(u, cat, st), raw))
		}
	}
	decided := "none: no portal item in ContInventory with a positive count"
	if src != nil {
		decided = fmt.Sprintf("class=%d addr=%#x", src.Class, src.Addr)
	}
	r.e.Log.Info("inv-tp-dump", "why", reason, "area", s.Player.Area, "hastp", verdict, "decided_by", decided,
		"statsReader", st != nil, "portal_items", strings.Join(portals, "; "), "items", strings.Join(items, " "))
}
