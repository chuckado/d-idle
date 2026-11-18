package resource

import "testing"

func TestCastGuardBlocked(t *testing.T) {
	known := CastGuard{Mode: 1, ExposureSec: 3, LifeFrac: 0.5, DmgKnown: true, Incoming: 10}
	for _, c := range []struct {
		name string
		g    CastGuard
		adj  int
		life int
		want bool
	}{
		{"old rule, 3 adjacent", CastGuard{}, 3, 100, true},
		{"old rule, 2 adjacent", CastGuard{}, 2, 100, false},
		{"new, 3 adjacent, safe sum", known, 3, 100, false},
		{"new, 1 adjacent, sum reaches half life", known, 1, 60, true},
		{"new, sum just under", known, 1, 61, false},
		{"new, terms missing, 3 adjacent: old rule", CastGuard{Mode: 1, ExposureSec: 3, LifeFrac: 0.5}, 3, 1000, true},
		{"new, terms missing, 2 adjacent: old rule", CastGuard{Mode: 1, ExposureSec: 3, LifeFrac: 0.5}, 2, 1000, false},
		{"new, no exposure: old rule", CastGuard{Mode: 1, LifeFrac: 0.5, DmgKnown: true}, 3, 1000, true},
	} {
		if got, why := c.g.Blocked(c.adj, c.life); got != c.want || (got && why == "") {
			t.Errorf("%s: got %v %q want %v", c.name, got, why, c.want)
		}
	}
}

func TestAbortFightGuardReason(t *testing.T) {
	tu := DefaultTuning()
	in := Input{Life: tu.Chicken(0), MaxLife: 100, Mana: 10, MaxMana: 10, HasTP: true, Hostiles: 5, Adjacent: 1, LastTripGold: -1,
		Skills: []Skill{staff}, CastGuard: CastGuard{Mode: 1, ExposureSec: 3, LifeFrac: 0.5, DmgKnown: true, Incoming: 100}}
	got := AbortFight(in, tu)
	if got.Kind != LeaveGame {
		t.Fatalf("got %v", got)
	}
}
