package main

import (
	"context"
	"d2lod-bot/internal/d2dsrc"
	"d2lod-bot/internal/data"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/input"
	"d2lod-bot/internal/quest"
	"d2lod-bot/internal/task"
)

type reachIn struct {
	input.Controller
	snap    **d2.Snapshot
	clicks  int
	flipsOn bool
}

func (r *reachIn) CloseAll() error { return nil }
func (r *reachIn) Interact(input.Target) error {
	r.clicks++
	if r.flipsOn {
		o := (*r.snap).Objects[0]
		o.Mode = 1
		(*r.snap).Objects = []d2.Unit{o} // a new slice: the unit the adapter holds keeps its old mode
	}
	return nil
}

// V lab 2026-11-23 (necro.act1-tools): the walk to the Malus stand ended 20 subtiles away (stuck at a door, hostiles near)
// and the eight clicks of the stand were used up from there in about a second. A walk that ends out of reach is not a
// click: it is retried, then reported as danger (clear the way), and no click is sent. Both classes: the check is class
// generic. Must fail while click ignores the distance.
func TestStandClickNeedsReach(t *testing.T) {
	for _, cl := range []d2.Class{d2.Necromancer, d2.Paladin} {
		for _, tc := range []struct {
			name   string
			px     int
			wantOK bool
		}{{"far", 20, false}, {"near", 3, true}} {
			t.Run(cl.String()+"/"+tc.name, func(t *testing.T) {
				snap := &d2.Snapshot{Player: &d2.Player{Class: cl, Area: 28, X: 100 + tc.px, Y: 100},
					Objects: []d2.Unit{{Type: d2.TypeObject, Class: quest.ObjMalusStand, ID: 7, X: 100, Y: 100}}}
				in := &reachIn{snap: &snap, flipsOn: true}
				walks := 0
				o := objectAdapter{e: &env.Env{Read: cainReader{snap: &snap}, Clock: cainClock{}, In: in},
					goUnit: func(context.Context, int, uint32) error { walks++; return nil }}
				err := o.Interact(context.Background(), "Malus stand")
				if tc.wantOK {
					if err != nil || in.clicks != 1 {
						t.Fatalf("near: err %v clicks %d", err, in.clicks)
					}
					return
				}
				if task.Classify(err) != task.KindNotFound || in.clicks != 0 || walks != objReachTries {
					t.Fatalf("far: err %v kind %v clicks %d walks %d", err, task.Classify(err), in.clicks, walks)
				}
			})
		}
	}
}

type tickClock struct{ t *time.Time }

func (c tickClock) Now() time.Time        { return *c.t }
func (c tickClock) Sleep(d time.Duration) { *c.t = c.t.Add(d) }

// An object that never changes in reach is a not-found failure the executor retries (MaxFails), not a timeout: a timeout is
// extended only while the exploration advances and otherwise skips an optional step on the first try (V lab 2026-11-23
// necro stand_click_afar: "never changed after 8 clicks", then skipped, no retry). Both classes.
func TestStandThatNeverChangesIsRetryable(t *testing.T) {
	for _, cl := range []d2.Class{d2.Necromancer, d2.Paladin} {
		t.Run(cl.String(), func(t *testing.T) {
			now := time.Unix(0, 0)
			snap := &d2.Snapshot{Player: &d2.Player{Class: cl, Area: 28, X: 102, Y: 100},
				Objects: []d2.Unit{{Type: d2.TypeObject, Class: quest.ObjMalusStand, ID: 7, X: 100, Y: 100}}}
			in := &reachIn{snap: &snap}
			o := objectAdapter{e: &env.Env{Read: cainReader{snap: &snap}, Clock: tickClock{&now}, In: in},
				goUnit: func(context.Context, int, uint32) error { return nil }}
			err := o.Interact(context.Background(), "Malus stand")
			if task.Classify(err) != task.KindNotFound || in.clicks != 8 {
				t.Fatalf("err %v kind %v clicks %d", err, task.Classify(err), in.clicks)
			}
		})
	}
}

// placerIn is a controller with the d2d's own placement verbs.
type placerIn struct {
	input.Controller
	index int
	item  uint32
	calls int
}

func (p *placerIn) PlaceItem(index int, item uint32) error {
	p.index, p.item = index, item
	p.calls++
	return nil
}

// On the d2d backend the Staff goes into the orifice through the controller's PlaceItem (V devp28: interact, hand, place), with
// the object's preset index and the pack item's id: no hover scan, no real game input. Must fail with "needs the real game
// input (hover scan)" without the branch. Two builds.
func TestPlaceItemUsesTheD2DVerbs(t *testing.T) {
	for _, cl := range []d2.Class{d2.Paladin, d2.Sorceress} {
		t.Run(cl.String(), func(t *testing.T) {
			snap := &d2.Snapshot{Player: &d2.Player{Class: cl, Area: 66, X: 100, Y: 100},
				Objects:   []d2.Unit{{Type: d2.TypeObject, Class: 152, ID: uint32(d2dsrc.NPCIDBase + 131), X: 100, Y: 100}},
				Inventory: &d2.Inventory{Items: []d2.Unit{{Class: 1, ID: 77, Item: &d2.Item{Where: d2.ContInventory}}}}}
			db := &data.DB{Items: []*data.ItemBase{{Code: "xxx"}, {Code: "hst"}}}
			in := &placerIn{}
			o := objectAdapter{e: &env.Env{Read: cainReader{snap: &snap}, Clock: cainClock{}, In: in, Data: db},
				goUnit: func(context.Context, int, uint32) error { return nil }}
			if err := o.placeItemOnce(context.Background(), 152, "hst"); err != nil {
				t.Fatal(err)
			}
			if in.calls != 1 || in.index != 131 || in.item != 77 {
				t.Fatalf("calls %d index %d item %d", in.calls, in.index, in.item)
			}
		})
	}
}
