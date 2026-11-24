package main

import (
	"context"
	"testing"

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
		(*r.snap).Objects[0].Mode = 1
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
				if task.Classify(err) != task.KindDanger || in.clicks != 0 || walks != objReachTries {
					t.Fatalf("far: err %v kind %v clicks %d walks %d", err, task.Classify(err), in.clicks, walks)
				}
			})
		}
	}
}
