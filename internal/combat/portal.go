package combat

import "d2lod-bot/internal/d2"

// Stats reads one unit's stat list (d2.StateReader.UnitStats; on d2d the dump's item extras, on Game.exe the stat list).
type Stats func(addr uint32) (d2.Stats, error)

// ItemClasser is the slice of Catalog the portal pick needs (the bot supplies its own adapter).
type ItemClasser interface{ Item(class int) ItemClass }

// PortalQty is the casts left in one portal item. A scroll is one cast. A tome's count is its quantity stat (70), and
// an EMPTY tome has no stat list at all (V live on i4 per internal/town/shop.go qty; Fresha i5 2026-10-30 clicked an
// empty tome 159 times; d2d: d2denv.UnitStats sets stat 70 only when the dump's qty > 0). An unreadable stat list
// counts as empty, like the town's scrollCount. UNVERIFIED: that Game.exe also omits stat 70 at 0 charges (the town
// code relies on the same assumption). With no stats source at all a tome counts as usable.
func PortalQty(u *d2.Unit, cat ItemClasser, stats Stats) int {
	if !cat.Item(u.Class).Tome {
		return 1
	}
	if stats == nil {
		return 1
	}
	st, err := stats(u.Addr)
	if err != nil || !st.Has(statQuantity) {
		return 0
	}
	return int(st.Get(statQuantity))
}

// PortalSource is the first scroll or non-empty tome in the main pack: the one item every portal cast (combat, town
// trip, kit check) must use. Nil when the pack has no portal source; an empty tome is never one.
func PortalSource(snap *d2.Snapshot, cat ItemClasser, stats Stats) *d2.Unit {
	if snap == nil || snap.Inventory == nil {
		return nil
	}
	for i := range snap.Inventory.Items {
		u := &snap.Inventory.Items[i]
		if u.Item == nil || u.Item.Where != d2.ContInventory || !cat.Item(u.Class).Portal {
			continue
		}
		if PortalQty(u, cat, stats) > 0 {
			return u
		}
	}
	return nil
}
