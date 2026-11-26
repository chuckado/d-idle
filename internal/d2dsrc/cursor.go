package d2dsrc

import "d2lod-bot/internal/d2"

// cursor fills Inventory.Cursor. d2d's "dump items" does not list the item on the mouse and "dump ui" only gives its
// base code ("held"), so the unit is the one that was in the previous read and is missing now with that base code
// (V live in D2DLOOP: after a click on a pack cell held="tsc" and the tsc was gone from the list; without this the
// hand's Take never saw "cursor holds item N" and every equip failed). The unit stays the cursor item while the same
// code is held and the item is still absent from the lists.
//
// The two dumps are not one frame: the item list can lose the unit one read before "held" shows its code (live 2026-11-26,
// Questy: "held item lea was not seen in the previous read", equip failed twice, the item was on the mouse). A unit that has
// left the list therefore stays a candidate for cursorMemory reads, the most recently vanished first.
func (b *builder) cursor() {
	s, inv := b.s, b.snap.Inventory
	s.cmu.Lock()
	defer s.cmu.Unlock()
	carried := s.remember(inv.Items)
	defer func() { s.prev = carried }()
	if b.held == "" || inv.Cursor != nil {
		s.cur = nil
		return
	}
	row := b.row(b.held)
	here := map[uint32]bool{}
	for _, u := range inv.Items {
		here[u.ID] = true
	}
	if s.cur != nil && s.cur.Class == row && !here[s.cur.ID] {
		b.putCursor(*s.cur)
		return
	}
	best := -1
	for i, m := range s.prev {
		u := m.u
		if u.Class == row && !here[u.ID] && u.Item != nil && u.Item.Where != d2.ContCursor && (best < 0 || m.miss < s.prev[best].miss) {
			best = i
		}
	}
	if best >= 0 {
		u := s.prev[best].u
		s.cur = &u
		b.putCursor(u)
		return
	}
	s.cur = nil
	b.rp.unknown("inventory.cursor", "held item "+b.held+" was not seen in the previous read")
}

func (b *builder) putCursor(u d2.Unit) {
	it := *u.Item
	it.Where, it.Page, it.NodePage = d2.ContCursor, 0, 0
	u.Item, u.Mode = &it, d2.ItemModeCursor
	inv := b.snap.Inventory
	inv.Cursor = &u
	inv.Items = append(inv.Items, u)
	inv.ItemCount++
	b.rp.derived("inventory.cursor", "unit missing since the previous read with the held base code")
}

// cursorMemory is how many reads a vanished unit stays a candidate for the cursor item.
const cursorMemory = 6

type seenUnit struct {
	u    d2.Unit
	miss int // reads since the unit was last in the list (0: it is in the current list)
}

// remember is the new memory after a read with the given item list: the units of the list, then the earlier ones that left it
// within cursorMemory reads.
func (s *Source) remember(items []d2.Unit) []seenUnit {
	out := make([]seenUnit, 0, len(items)+len(s.prev))
	here := map[uint32]bool{}
	for _, u := range items {
		here[u.ID] = true
		out = append(out, seenUnit{u: u})
	}
	for _, m := range s.prev {
		if here[m.u.ID] || m.miss+1 > cursorMemory {
			continue
		}
		m.miss++
		out = append(out, m)
	}
	return out
}
