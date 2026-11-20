package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/d2dctl/fake"
	"d2lod-bot/internal/d2dworld"
	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/input"
	"d2lod-bot/internal/quest"
)

// layoutRig is the fork's levelmap verb through the real wire (fake server), d2dworld with the knob closure and an
// objectAdapter whose walk is recorded; a walk within streaming range makes the tree appear in the snapshot.
type layoutRig struct {
	snap  *d2.Snapshot
	goTos []input.Point
	o     objectAdapter
	tree  input.Point
}

func newLayoutRig(t *testing.T, hero input.Point, knob bool, verb bool) *layoutRig {
	t.Helper()
	path, cleanup, err := fake.TempSocket()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	fix := fake.Sample()
	fix.Level = "1 30.70 13.70 57 41 0 0"
	fix.Presets = []string{"0\t119\t30.0\t10.0\t0\tWaypoint"}
	if verb {
		doc, _ := json.Marshal(&d2dctl.LevelMap{Level: 1, Unit: "cell",
			Rooms:   []d2dctl.LayoutRoom{{X: 40, Y: 30, W: 8, H: 8, Preset: true, Def: 161}},
			Objects: []d2dctl.LayoutObject{{Index: 9, Class: quest.ObjTreeInifuss, Name: "Tree of Inifuss", X: 44.5, Y: 34.5, Kind: "tree"}}})
		fix.LevelMap = string(doc)
	}
	srv, err := fake.New(path, fix, fake.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(srv.Close)
	c := d2dctl.New(path, d2dctl.Options{Timeout: 2 * time.Second})
	t.Cleanup(func() { c.Close() })
	w := d2dworld.New(d2dworld.Client{Client: c}, d2dworld.Options{LayoutRead: func() bool { return knob }})
	r := &layoutRig{snap: &d2.Snapshot{Player: &d2.Player{Area: 1, X: hero.X, Y: hero.Y}}, tree: input.Point{X: 222, Y: 172}}
	r.o = objectAdapter{
		e: &env.Env{Read: cainReader{snap: &r.snap}, Clock: cainClock{}},
		lmFn: func() *d2map.LevelMap {
			_, _ = w.Update()
			return w.Current()
		},
		goTo: func(_ context.Context, p input.Point) error {
			r.goTos = append(r.goTos, p)
			r.snap.Player.X, r.snap.Player.Y = p.X, p.Y
			dx, dy := p.X-r.tree.X, p.Y-r.tree.Y
			if dx*dx+dy*dy <= treeVisible*treeVisible { // the tree streams in within about 13 cells
				r.snap.Objects = []d2.Unit{{Type: d2.TypeObject, Class: quest.ObjTreeInifuss, X: r.tree.X, Y: r.tree.Y}}
			}
			return nil
		},
	}
	return r
}

// K1 consumer (ADR 0022): the hero starts far from the tree. With quests.layoutRead on the first planned walk goes to
// the tree's own position and the tree is found; the same for two hero classes (the object table is data, the walk does
// not depend on the class: two start spots stand for a Paladin and a Sorceress arriving from different warps).
func TestLayoutReadWalksStraightToTheTree(t *testing.T) {
	for name, hero := range map[string]input.Point{"Paladin": {X: 20, Y: 20}, "Sorceress": {X: 270, Y: 100}} {
		r := newLayoutRig(t, hero, true, true)
		u, err := r.o.findNear(context.Background(), quest.ObjTreeInifuss, nil)
		if err != nil || u == nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(r.goTos) != 1 || r.goTos[0] != r.tree {
			t.Fatalf("%s: walks %v, want one walk to %v", name, r.goTos, r.tree)
		}
	}
}

// Knob off, or a d2d without the verb: the preset is unknown, so findNear does not walk (find then explores as before).
func TestLayoutReadOffOrVerbAbsentDoesNotWalk(t *testing.T) {
	for name, tc := range map[string]struct{ knob, verb bool }{"knob off": {false, true}, "verb absent": {true, false}} {
		r := newLayoutRig(t, input.Point{X: 20, Y: 20}, tc.knob, tc.verb)
		if _, err := r.o.findNear(context.Background(), quest.ObjTreeInifuss, nil); err == nil {
			t.Fatalf("%s: found a tree nobody knew", name)
		}
		if len(r.goTos) != 0 {
			t.Fatalf("%s: walked %v", name, r.goTos)
		}
	}
}
