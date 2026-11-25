package bot

import (
	"context"
	"sync/atomic"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/goals"
)

// fakeTaker is the corpse click of the d2d adapter: in town after a fresh game it puts the gear back on the body.
type fakeTaker struct {
	g      *rig
	always bool         // the corpse lies in town from the start (a death the bot never tracked); else only after a fresh game
	taken  atomic.Bool // the corpse is gone once clicked
	calls  atomic.Int32 // read by the test goroutine while the bot goroutine clicks
}

func (f *fakeTaker) TakeCorpse(ctx context.Context, area int) (bool, error) {
	if area != 1 || !f.always && f.g.w.count("newgame") == 0 {
		return false, nil // in the field the corpse is gone with the game's ground: nothing to click
	}
	if !f.taken.CompareAndSwap(false, true) {
		return false, nil
	}
	f.calls.Add(1)
	f.g.w.call("takecorpse")
	f.g.w.set(func(w *world) { w.items = append(w.items, gearUnit()) })
	return true, nil
}

// User request 2026-11-25: a corpse that three tries in the field could not fetch is not given up (all four deaths so far
// were abandoned, the hero lay naked and wore junk). The bot starts a fresh game, the game moves the corpse to town (V d2d), and
// the first thing in the new game is the click on it. Two classes. Must fail without the reset: the old code abandoned it.
func TestCorpseResetAfterFailedTriesAndTakenInTown(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		t.Run(class.String(), func(t *testing.T) {
			var tk *fakeTaker
			g := newRig(t, func(c *Config, d *Deps, tu *env.Tuning) {})
			tk = &fakeTaker{g: g}
			g.r.d.CorpseTake = tk
			g.r.d.Cycler = fakeCycler{g.w}
			g.mover.exploreFn = blockExplore
			g.w.set(func(w *world) {
				w.class = class
				w.level = 3
				w.area = 2
				w.x, w.y = 300, 310
				w.items = append(w.items, gearUnit())
			})
			g.run()
			defer g.finish()
			g.r.Start(farm("Blood Moor"))
			g.waitFor("in field", func() bool { return g.w.count("explore") > 0 })
			die(g)
			g.waitFor("death", func() bool { return len(g.events(goals.EvDeath)) == 1 })
			respawn(g)
			g.w.set(func(w *world) { w.x, w.y = 300, 310 })
			g.waitFor("three tries", func() bool { return g.w.count("back 2") >= 3 })
			g.waitFor("fresh game", func() bool { return g.w.count("newgame") > 0 })
			g.waitFor("corpse taken in town", func() bool { return tk.calls.Load() > 0 })
			if g.w.count("back 2") != 3 {
				t.Fatalf("%d field tries, want 3: %s", g.w.count("back 2"), g.w.log())
			}
			if g.w.count("takecorpse") != 1 {
				t.Fatalf("corpse clicked %d times: %s", g.w.count("takecorpse"), g.w.log())
			}
		})
	}
}

// Without a taker (a backend that cannot click the corpse) nothing is reset: the corpse stays abandoned as before.
func TestCorpseNoResetWithoutTaker(t *testing.T) {
	g := newRig(t)
	g.r.d.Cycler = fakeCycler{g.w}
	g.mover.exploreFn = blockExplore
	g.w.set(func(w *world) { w.level = 3; w.area = 2; w.x, w.y = 300, 310; w.items = append(w.items, gearUnit()) })
	g.run()
	defer g.finish()
	g.r.Start(farm("Blood Moor"))
	g.waitFor("in field", func() bool { return g.w.count("explore") > 0 })
	die(g)
	g.waitFor("death", func() bool { return len(g.events(goals.EvDeath)) == 1 })
	respawn(g)
	g.w.set(func(w *world) { w.x, w.y = 300, 310 })
	g.waitFor("three tries", func() bool { return g.w.count("back 2") >= 3 })
	g.waitFor("work resumes", func() bool { return g.r.State() == StateClearing })
	if g.w.count("newgame") != 0 {
		t.Fatalf("reset without a corpse taker: %s", g.w.log())
	}
}

// V live 2026-11-25: five pieces lay in the Rogue Encampment for hours, a corpse of a death the bot never tracked. Standing in town
// the bot clicks any corpse that carries items, tracked or not. Two classes. Must fail without the standing sweep.
func TestStandingSweepTakesAnUntrackedCorpseInTown(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		g := newRig(t)
		tk := &fakeTaker{g: g, always: true}
		g.r.d.CorpseTake = tk
		g.mover.exploreFn = blockExplore
		g.w.set(func(w *world) { w.class = class; w.area = 1 })
		g.run()
		g.r.Start(farm("Blood Moor"))
		g.waitFor("corpse taken", func() bool { return tk.calls.Load() > 0 })
		g.finish()
		if g.w.count("takecorpse") != 1 {
			t.Fatalf("%v: corpse clicked %d times: %s", class, g.w.count("takecorpse"), g.w.log())
		}
	}
}

// V live 2026-11-24 20:52 EST: the record said corpse value 0 with five worn pieces. A worn piece counts at least wornPieceFloor
// even when the catalog prices it at nothing. Two classes.
func TestWornPiecesKeepTheCorpseValueAboveZero(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		g := newRig(t)
		cheap := gearUnit()
		cheap.Class, cheap.Item.Quality = 3, 2 // zero cost in the fake catalog
		g.mover.exploreFn = blockExplore
		g.w.set(func(w *world) { w.class = class; w.level = 3; w.area = 2; w.items = append(w.items, cheap) })
		g.run()
		defer g.finish()
		g.r.Start(farm("Blood Moor"))
		g.waitFor("in field", func() bool { return g.w.count("explore") > 0 })
		die(g)
		g.waitFor("death", func() bool { return len(g.events(goals.EvDeath)) == 1 })
		if v := g.r.DeathReport().Records[0].CorpseValue; v < wornPieceFloor {
			t.Fatalf("%v: corpse value %d with a worn piece", class, v)
		}
	}
}
