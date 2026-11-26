package town_test

import (
	"context"
	"errors"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/gear/geartest"
	"d2lod-bot/internal/town"
)

// The 02:51 death of 2026-11-26 (Fable review 3.1): a trip in its chores stage while the hero stood in Cold Plains; the
// shop chore walked 21 s toward a town NPC that was not there. Every chore now verifies "in town" before it starts, and a
// step that finds itself outside a town ends the chores with ErrNotInTown instead of running the rest.
func TestChoresAbortWhenTheHeroIsOutsideTown(t *testing.T) {
	geartest.Need(t)
	for _, class := range []d2.Class{d2.Sorceress, d2.Barbarian} { // not a class rule; both walk the same town
		t.Run(class.String(), func(t *testing.T) {
			r := newRig(t)
			w := r.W
			dressed(w)
			w.Player.Class = class
			w.Player.Life = 12 // heal and shop are due
			calls, flipAt := 0, 6
			r.T.UseSource(town.Source{Snapshot: func() (*d2.Snapshot, error) {
				calls++
				if calls == flipAt {
					w.Player.Area = 3 // Cold Plains: the portal was taken, the hero is in the field
				}
				return w.Snapshot()
			}, Items: w.AllItems, Owner: w.Owner, Stats: w.UnitStats})
			rep, err := r.T.Chores(context.Background(), town.FullPlan())
			if !errors.Is(err, town.ErrNotInTown) {
				t.Fatalf("want ErrNotInTown once the hero is outside town, got %v (steps %+v)", err, rep.Steps)
			}
			ran := 0
			for _, s := range rep.Steps {
				if s.Ran && s.Name != "close" {
					ran++
				}
			}
			if ran > 1 {
				t.Fatalf("%d chores ran after the hero left town: %+v", ran, rep.Steps)
			}
		})
	}
}
