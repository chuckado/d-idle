package main

import (
	"context"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/invdesk"
)

// lockstepWorld changes only when the bot's clock sleeps, like d2d under --lockstep: the verb is queued, the item reads
// identified a few steps later.
type lockstepWorld struct {
	class d2.Class
	sent  bool
	steps int
}

func (w *lockstepWorld) Sleep(time.Duration) {
	if w.sent {
		w.steps++
	}
}
func (w *lockstepWorld) Now() time.Time { return time.Unix(1, 0) }

func (w *lockstepWorld) Snapshot() (*d2.Snapshot, error) {
	fl := uint32(0)
	if w.steps >= 3 {
		fl = d2.FlagIdentified
	}
	inv := &d2.Inventory{Items: []d2.Unit{{ID: 9, Type: d2.TypeItem, Item: &d2.Item{Where: d2.ContInventory, Flags: fl}}}}
	return &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Class: w.class}, Inventory: inv}, nil
}

func (w *lockstepWorld) Caps(context.Context) ([]string, error) { return []string{CapIdentify}, nil }
func (w *lockstepWorld) ItemMove(context.Context, uint32, string, int, int) (bool, error) {
	return false, nil
}
func (w *lockstepWorld) ItemIdentify(context.Context, uint32) (bool, error) {
	w.sent = true
	return false, nil
}
func (w *lockstepWorld) ItemDrop(context.Context, uint32) error { return nil }
func (w *lockstepWorld) GoldStash(context.Context, int, bool) (int, int, error) {
	return 0, 0, nil
}

const CapIdentify = invdesk.CapItemIdentify

// F3 (pgvault live): the queue path settles on the bot goroutine, which is also what steps a lockstep game. The mover
// must wait on the bot clock, not the wall clock, or the verb's change never shows ("answered ok but the game did not
// show the change"). Two classes; the same test with a wall-clock pause fails.
func TestItemMoverSettlesOnTheBotClock(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		w := &lockstepWorld{class: class}
		e := &env.Env{Clock: w, Read: stateOnly{w, nil}}
		m := newItemMover(e, w, func() bool { return true }, func() bool { return false })
		m.Settles = 5 // the default 25 polls would also pass; 5 virtual steps is what the world needs
		done, err := m.Identify(context.Background(), 9)
		if err != nil || !done {
			t.Fatalf("%v: identify done=%v err=%v (steps %d)", class, done, err, w.steps)
		}
	}
}

type stateOnly struct {
	w *lockstepWorld
	d2.StateReader
}

func (s stateOnly) Snapshot() (*d2.Snapshot, error) { return s.w.Snapshot() }
