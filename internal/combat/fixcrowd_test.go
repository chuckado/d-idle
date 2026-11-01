package combat

import (
	"testing"

	"d2lod-bot/internal/d2"
)

func tg(dist float64) Target { return Target{Dist: dist} }

// DRILLS11 geometry (card fixcrowd): zombies at 1, 5 and 8 subtiles are in melee reach, far ones are not.
func TestCrowdAdjacentReach(t *testing.T) {
	const reach = 8
	for _, c := range []struct {
		name  string
		dists []float64
		adj   int
	}{
		{"drill geometry 1/5/8", []float64{1, 5, 8}, 3},
		{"best live read 1.0/2.4/5.7", []float64{1, 2.4, 5.7}, 3},
		{"far monsters are not adjacent", []float64{9, 12, 20}, 0},
		{"one adjacent, rest far", []float64{1, 9, 15}, 1},
		{"boundary 8.01", []float64{8.01}, 0},
	} {
		var hs []Target
		for _, d := range c.dists {
			hs = append(hs, tg(d))
		}
		if _, adj := Crowd(&d2.Player{}, hs, reach); adj != c.adj {
			t.Errorf("%s: adjacent %d want %d", c.name, adj, c.adj)
		}
	}
	// the old 3-subtile rule never saw the drill geometry as a crowd
	if _, adj := Crowd(&d2.Player{}, []Target{tg(1), tg(5), tg(8)}, 3); adj != 1 {
		t.Errorf("reach 3 control: adjacent %d want 1", adj)
	}
}

// The drill shape through the fake game: 3 zombies at 1, 5 and 8 subtiles, 34% life, no potion, scroll: Chicken; one
// adjacent (the control drill) stays TownPortal; the knob override restores the narrow rule.
func TestSafetyMeleeCutCastDrillGeometry(t *testing.T) {
	run := func(xs []float64, reach int) SafetyAction {
		h, _ := newPortalHarness(t)
		h.s.life = 34
		for i, x := range xs {
			h.s.addMon(uint32(10+i), clsFallen, 100+x, 100, 500, 5, 6)
		}
		h.f.e.Tune.Combat.CrowdAdjacent = reach
		snap, _ := h.s.snapshot()
		return h.f.Safety(snap)
	}
	if a := run([]float64{1, 5, 8}, 8); a != Chicken {
		t.Errorf("1/5/8 at reach 8: got %v want Chicken", a)
	}
	if a := run([]float64{1, 9, 14}, 8); a != TownPortal {
		t.Errorf("single adjacent (control): got %v want TownPortal", a)
	}
	if a := run([]float64{1, 12, 14}, 8); a != TownPortal {
		t.Errorf("far monsters: got %v want TownPortal", a)
	}
	if a := run([]float64{1, 5, 8}, 3); a != TownPortal {
		t.Errorf("knob override 3: got %v want TownPortal", a)
	}
}
