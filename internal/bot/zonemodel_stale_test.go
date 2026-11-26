package bot

import (
	"context"
	"strings"
	"testing"
)

// Questy bed 2026-11-26: four deaths at one spot of the Dark Wood. Each chicken noted the pack (stale pack), the trip home came back
// through the town portal, which drops the hero where it left, into the same pack. A portal that leads where a stale pack still
// stands is not taken: the hero walks in (the other end of the level). Two classes. Must fail without the stale-pack check.
func TestReturnToFieldWalksPastAStalePack(t *testing.T) {
	for _, cl := range reaskClasses {
		t.Run(cl.String(), func(t *testing.T) {
			g, _ := reaskRig(t, cl, nil)
			g.w.set(func(w *world) { w.level = 12; w.area = 9 })
			g.r.poll()
			g.w.set(func(w *world) { w.area = 1 })
			g.r.poll() // 9 -> town alive: the portal now leads to 9
			g.r.obMu.Lock()
			g.r.ob.stale = stalePack{area: 9, hostiles: 12, at: g.r.e.Clock.Now()}
			g.r.obMu.Unlock()
			if err := g.r.returnToField(context.Background(), 9, false); err != nil {
				t.Fatal(err)
			}
			if g.w.count("back 9") != 0 {
				t.Fatalf("took the portal into the pack: %v", g.w.log())
			}
		})
	}
}

// Questy bed 2026-11-26: the corpse in the Dark Wood was skipped for "1 healing potions < 2 for the underground_passage road", then
// the game was reset with the corpse unreachable. A corpse decision and run are exempt from the early road guard. Two classes.
func TestCorpseRunIsExemptFromTheEarlyRoad(t *testing.T) {
	for _, class := range reaskClasses {
		t.Run(class.String(), func(t *testing.T) {
			g, _ := reaskRig(t, class, nil)
			g.r.e.Tune.Route.EarlyRoute = 1
			g.w.set(func(w *world) { w.level, w.area = 4, 1 })
			g.r.poll()
			s := g.r.latest()
			if g.r.earlyRouteWhy(s, 5) == "" {
				t.Skip("the rig's hero is not held by the early road")
			}
			rec := &recovery{rec: DeathRecord{Time: g.clock.Now(), Area: 5, AreaName: "Dark Wood", CorpseValue: 500}, area: 5}
			_, why := g.r.retrievalDecision(rec, s)
			if strings.Contains(why, "route guard") {
				t.Fatalf("the corpse decision was refused by the early road: %q", why)
			}
		})
	}
}
