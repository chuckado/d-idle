package bot

import "testing"

// BELTFIX5, Gate 1 20261103-0338, minute lines 331-339 (clvl 11, Black Marsh chicken at 04:05:53): the purse held 126
// gold, then 6, with 0 healing potions, while the stash held 1418 (merc log: gold=1424 against purse 6). The pack-gold
// test switched the potion chore off ("Recovery: farming gold", 7 virtual minutes unhealed). Banked gold counts when no
// healing potion is left; Town.withdraw (internal/town) takes it out.
func TestBankedGoldCountsWithNoHealingPotion(t *testing.T) {
	for _, row := range []struct{ purse, bank int }{{126, 1418}, {6, 1418}, {3, 1418}, {0, 60}} {
		for _, inTown := range []bool{true, false} {
			if !inTown && row.bank < 300 {
				continue // a field trip needs the trip-budget floor; the town visit only needs one potion's gold
			}
			g := newRig(t)
			area := 1
			if !inTown {
				area = 6
			}
			g.w.set(func(w *world) { w.level, w.area, w.gold0, w.stash0 = 11, area, row.purse, row.bank })
			g.w.belt()
			g.r.loop.broke.on = true
			s, _ := g.w.snapshot()
			plan, emerg, ok := g.r.choresDue(s)
			if !ok || !plan.Potions || !emerg {
				t.Errorf("purse %d bank %d town %v: want an emergency potion chore, got ok=%v %+v emergency=%v", row.purse, row.bank, inTown, ok, plan, emerg)
			}
		}
	}
}

// The bank does not start trips for a character that still has a potion, and an empty bank changes nothing.
func TestBankedGoldIgnoredWithPotionOrNothingBanked(t *testing.T) {
	g := newRig(t)
	g.w.set(func(w *world) { w.level, w.area, w.gold0, w.stash0 = 11, 1, 6, 0 })
	g.w.belt()
	g.r.loop.broke.on = true
	s, _ := g.w.snapshot()
	if plan, emerg, ok := g.r.choresDue(s); ok && (plan.Potions || emerg) {
		t.Errorf("6 gold, empty bank: no potion chore, got %+v emergency=%v", plan, emerg)
	}
}
