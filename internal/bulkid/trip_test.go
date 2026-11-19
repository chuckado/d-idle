package bulkid

import (
	"context"
	"errors"
	"slices"
	"testing"

	"d2lod-bot/internal/env"
)

// cainWorld wraps fakeMover with identified flags. Cain identifies what is in the PACK only (S: not the stash).
type cainWorld struct {
	*fakeMover
	ident   map[uint32]bool
	present bool
	failID  bool
	moves   int
}

func (c *cainWorld) Move(ctx context.Context, id uint32, to string) error {
	c.moves++
	return c.fakeMover.Move(ctx, id, to)
}
func (c *cainWorld) Unidentified(context.Context) ([]Item, error) {
	var out []Item
	for _, it := range append(slices.Clone(c.pack), c.stash...) {
		if !c.ident[it.ID] {
			out = append(out, it)
		}
	}
	return out, nil
}
func (c *cainWorld) Present(context.Context) bool { return c.present }
func (c *cainWorld) IdentifyAll(context.Context) (int, error) {
	if c.failID {
		return 0, errors.New("Cain's Identify line was not reached")
	}
	n := 0
	for _, it := range c.pack {
		if !c.ident[it.ID] {
			c.ident[it.ID] = true
			n++
		}
	}
	return n, nil
}

func newTrip(t *testing.T, cw *cainWorld, tu env.IdentifyTuning, measured bool) *Trip {
	return &Trip{F: mustFilter(t), Tu: tu, Hero: func(context.Context) Hero { return Hero{CainOpen: true} },
		L: cw, Cain: cw, S: &Stager{M: cw, J: &Journal{}, Chunk: tu.Chunk}, Measured: measured}
}

// both builds: rings in the stash for one, an amulet and a charm for the other.
func tripWorld(build string) *cainWorld {
	f := &fakeMover{packCap: 40, stashCap: 100}
	mk := func(id uint32, ty, q, place string) Item {
		return Item{ID: id, Types: []string{ty}, Quality: q, Ilvl: 20, Value: 100, Place: place}
	}
	if build == "Sorceress" {
		f.stash = []Item{mk(1, "ring", "magic", PlaceStash), mk(2, "ring", "rare", PlaceStash), mk(3, "helm", "magic", PlaceStash)}
		f.pack = []Item{mk(10, "amul", "magic", PlacePack)}
	} else {
		f.stash = []Item{mk(1, "amul", "rare", PlaceStash), mk(2, "char", "magic", PlaceStash)}
		f.pack = []Item{mk(10, "ring", "magic", PlacePack), mk(11, "sword", "magic", PlacePack)}
	}
	return &cainWorld{fakeMover: f, ident: map[uint32]bool{}, present: true}
}

func TestTripIdentifiesStashItemsAndReturnsThem(t *testing.T) {
	for _, build := range []string{"Sorceress", "Paladin"} {
		cw := tripWorld(build)
		stashBefore := len(cw.stash)
		rep, err := newTrip(t, cw, env.DefaultIdentifyTuning(), false).Run(context.Background())
		if err != nil {
			t.Fatalf("%s: %v", build, err)
		}
		if rep.Staged == 0 || rep.Returned != rep.Staged || rep.Identified != rep.Planned {
			t.Fatalf("%s: %+v", build, rep)
		}
		if len(cw.stash) != stashBefore {
			t.Fatalf("%s: stash %d -> %d", build, stashBefore, len(cw.stash))
		}
		for _, it := range cw.stash { // only the filtered ones were touched
			if it.Types[0] == "helm" && cw.ident[it.ID] {
				t.Fatalf("%s: an unfiltered stash item was identified", build)
			}
		}
		if !cw.ident[1] {
			t.Fatalf("%s: stash item 1 not identified", build)
		}
	}
}

// Must-fail: an absent Cain means no move happens at all.
func TestAbsentCainMovesNothing(t *testing.T) {
	cw := tripWorld("Sorceress")
	cw.present = false
	rep, err := newTrip(t, cw, env.DefaultIdentifyTuning(), false).Run(context.Background())
	if err != nil || cw.moves != 0 || rep.Staged != 0 || rep.Reason != WhyNoCain {
		t.Fatalf("%+v moves %d %v", rep, cw.moves, err)
	}
}

func TestNeverManualAndMeasuredDoNothing(t *testing.T) {
	for _, c := range []struct {
		pol      int
		measured bool
	}{{env.IdentifyNever, false}, {env.IdentifyManual, false}, {env.IdentifyCainFree, true}} {
		cw := tripWorld("Paladin")
		tu := env.DefaultIdentifyTuning()
		tu.Policy = c.pol
		rep, _ := newTrip(t, cw, tu, c.measured).Run(context.Background())
		if cw.moves != 0 || len(cw.ident) != 0 || rep.Planned != 0 {
			t.Fatalf("%+v: moves %d ident %d", c, cw.moves, len(cw.ident))
		}
	}
}

// Must-fail: when Cain's dialog fails the stash items still go home.
func TestFailedDialogStillReturnsStashItems(t *testing.T) {
	cw := tripWorld("Sorceress")
	cw.failID = true
	total0 := total(cw.fakeMover)
	rep, err := newTrip(t, cw, env.DefaultIdentifyTuning(), false).Run(context.Background())
	if err == nil || rep.Returned != rep.Staged || len(cw.stash) != 3 || total(cw.fakeMover) != total0 {
		t.Fatalf("%+v %v stash %d", rep, err, len(cw.stash))
	}
}

func TestPerTripCapLimitsStaging(t *testing.T) {
	cw := tripWorld("Sorceress")
	tu := env.DefaultIdentifyTuning()
	tu.TripCap = 1
	rep, _ := newTrip(t, cw, tu, false).Run(context.Background())
	if rep.Planned != 1 {
		t.Fatalf("%+v", rep)
	}
}
