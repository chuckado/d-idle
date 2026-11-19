package bot

import (
	"context"
	"testing"
)

// Holdout seeds 149 and 163: the hero fled Cave Level 1 (area 9) by portal, the model then picked Black Marsh (6); the
// town portal leads to 9, so BackToField came out "wanted 6, in 9", in the zone that had just hurt it. The trip must
// walk from town instead. Both classes: the rule reads no class.
func TestReturnToFieldWalksPastBurnedPortal(t *testing.T) {
	for _, cl := range reaskClasses {
		t.Run(cl.String(), func(t *testing.T) {
			g, _ := reaskRig(t, cl, nil)
			g.w.set(func(w *world) { w.level = 12; w.area = 9 })
			g.r.poll()
			g.w.set(func(w *world) { w.area = 1 })
			g.r.poll() // 9 -> town alive: the portal now leads to 9
			g.r.zoneCostAt("chicken", 9)
			if err := g.r.returnToField(context.Background(), 6); err != nil {
				t.Fatal(err)
			}
			if g.w.count("back 6") != 0 {
				t.Fatalf("took the portal that lands in the burned zone: %v", g.w.log())
			}
			if g.w.count("wp 6")+g.w.count("exit 6") == 0 && g.r.latest().Player.Area != 6 {
				t.Fatalf("no walk: %v", g.w.log())
			}
		})
	}
}

// A clean landing zone, or a target equal to it, still uses the portal.
func TestReturnToFieldPortalWhenLandingIsClean(t *testing.T) {
	for _, cl := range reaskClasses {
		t.Run(cl.String(), func(t *testing.T) {
			g, _ := reaskRig(t, cl, nil)
			g.w.set(func(w *world) { w.level = 12; w.area = 9 })
			g.r.poll()
			g.w.set(func(w *world) { w.area = 1 })
			g.r.poll()
			if err := g.r.returnToField(context.Background(), 6); err != nil {
				t.Fatal(err)
			}
			if g.w.count("back 6") != 1 {
				t.Fatalf("calls: %v", g.w.log())
			}
		})
	}
}
