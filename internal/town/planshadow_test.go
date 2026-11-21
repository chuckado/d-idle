package town_test

import (
	"context"
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/supply"
	"d2lod-bot/internal/town"
)

// Stage 1 shadow: the plan names what a hero with an empty belt needs and would buy, for a mana build and a melee build,
// and the old shop does exactly what it did (the knob is read only by the trace).
func TestPlanShadowTwoBuilds(t *testing.T) {
	for _, tc := range []struct {
		name  string
		class d2.Class
		mana  bool
	}{{"sorceress", d2.Sorceress, true}, {"paladin", d2.Paladin, false}} {
		r, _ := withdrawRig(t, 1500, 0, 20, false)
		r.W.Player.Class = tc.class
		r.Env.Tune.Town.Plan = 1
		pl, err := r.T.PlanShadow(town.ChorePlan{Profile: "monastery_road"})
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if got := pl.Need.Lines[supply.Healing].Min; got != 4 {
			t.Errorf("%s: healing min %d, want 4", tc.name, got)
		}
		if has := pl.Need.Lines[supply.Mana].Min > 0; has != tc.mana {
			t.Errorf("%s: mana line present %v, want %v (%s)", tc.name, has, tc.mana, pl.Trace())
		}
		if !strings.Contains(pl.Trace(), "would buy healing") {
			t.Errorf("%s: trace %q lacks a healing buy", tc.name, pl.Trace())
		}
	}
}

func TestPlanShadowUnknownProfileIsAnError(t *testing.T) {
	r, _ := withdrawRig(t, 100, 0, 5, false)
	if _, err := r.T.PlanShadow(town.ChorePlan{Profile: "nope"}); err == nil || !strings.Contains(err.Error(), `"nope"`) {
		t.Fatalf("want an error naming the profile, got %v", err)
	}
}

// Knob 1 changes nothing the chores do: same purchases and gold as knob 0.
func TestPlanStage1ChangesNoBehaviour(t *testing.T) {
	run := func(stage int) town.ChoreReport {
		r, _ := withdrawRig(t, 600, 1418, 11, false)
		r.Env.Tune.Town.Plan = stage
		rep, err := r.T.Chores(context.Background(), town.ChorePlan{Potions: true, Scrolls: true, Gold: true})
		if err != nil {
			t.Fatal(err)
		}
		return rep
	}
	a, b := run(0), run(1)
	if a.GoldAfter != b.GoldAfter || a.Spent != b.Spent || len(a.Bought) != len(b.Bought) {
		t.Fatalf("stage 1 changed the visit: %+v vs %+v", a, b)
	}
}
