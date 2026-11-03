package bot

import (
	"strings"
	"testing"
	"time"
)

// BELTFIX3 (1), Gate 1 20261102 BELT A: the character stood in town six times with no healing potion and 332 gold and
// left each time without shopping, because lastTrip was stamped by the failed field attempt seconds before and the town
// visit used the full resupply cooldown.
func TestTownArrivalRestocksWithoutCooldown(t *testing.T) {
	g := newRig(t)
	g.mover.exploreFn = blockExplore
	g.w.set(func(w *world) { w.level, w.area, w.gold0 = 9, 1, 332 })
	g.w.belt() // no healing potion in the belt, none in the pack
	g.r.setLastTrip(g.e.Clock.Now().Add(-10 * time.Second))
	g.run()
	defer g.finish()
	g.r.Start(farm("Blood Moor"))
	g.waitFor("restock before leaving", func() bool { return g.w.count("chores") > 0 })
	g.waitFor("in field", func() bool { return g.w.count("explore") > 0 })
	log := g.w.log()
	if strings.Index(log, "chores") > strings.Index(log, "explore") {
		t.Fatalf("left town before shopping: %s", log)
	}
}

// In town with no healing potion but no gold for one the visit is not an emergency (nothing could be bought).
func TestTownArrivalNoGoldNoEmergency(t *testing.T) {
	g := newRig(t)
	g.w.set(func(w *world) { w.level, w.area, w.gold0 = 9, 1, 10 })
	g.w.belt()
	s, _ := g.w.snapshot()
	if _, emerg, ok := g.r.choresDue(s); ok && emerg {
		t.Error("10 gold buys no potion: no emergency")
	}
	g.w.set(func(w *world) { w.gold0 = 332 })
	s, _ = g.w.snapshot()
	if _, emerg, ok := g.r.choresDue(s); !ok || !emerg {
		t.Errorf("332 gold, no potion, in town: want an emergency, got ok=%v emergency=%v", ok, emerg)
	}
}
