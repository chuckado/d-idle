package main

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/input"
	"d2lod-bot/internal/nav"
	"d2lod-bot/internal/quest"
)

// V (L0 fixtures docs/probes/cainl0/testdata, five seeds): the tree room 161 origins; the tree stands at origin + 0.5
// and is listed only within about 13 to 18 cells.
var treeRooms = [][2]int{{44, 44}, {52, 12}, {60, 68}, {52, 60}, {36, 28}}

const treeVisible = 13 * 5 // subtiles

type treeWorld struct {
	lm      *d2map.LevelMap
	player  input.Point
	tree    input.Point
	goTos   []input.Point
	sweeps  int
	hints   []*input.Point
	clicked bool
}

func newTreeWorld(room [2]int) *treeWorld {
	c := d2map.Point{X: room[0]*5 + 2, Y: room[1]*5 + 2}
	lm := &d2map.LevelMap{RoomRows: []d2map.RoomRow{{Type: 100, Pos: d2map.Point{X: 5, Y: 5}}, {Type: 161, Pos: c}}}
	return &treeWorld{lm: lm, player: input.Point{X: 72, Y: 93}, tree: input.Point{X: c.X, Y: c.Y}}
}

func (w *treeWorld) near() bool {
	dx, dy := w.player.X-w.tree.X, w.player.Y-w.tree.Y
	return dx*dx+dy*dy <= treeVisible*treeVisible
}
func (w *treeWorld) target() (input.Point, bool) {
	return roomTargetIn(w.lm, quest.ObjRoomType[quest.ObjTreeInifuss], w.player)
}
func (w *treeWorld) goTo(_ context.Context, p input.Point) error {
	w.goTos = append(w.goTos, p)
	w.player = p
	return nil
}
func (w *treeWorld) sweep(_ context.Context, hint *input.Point) error {
	w.sweeps++
	w.hints = append(w.hints, hint)
	return context.DeadlineExceeded
}
func (w *treeWorld) look() (*d2.Unit, error) {
	if w.near() {
		return &d2.Unit{Class: quest.ObjTreeInifuss, X: w.tree.X, Y: w.tree.Y}, nil
	}
	return nil, errors.New("object class 30 not found")
}

// The must-fail test of CAIN-7: with the tree at a random room the bot walks to the room centre (origin + 0.5) and the
// tree is found, with no coverage sweep. Before the fix the same level only ever swept.
func TestTreeFoundByWalkingToItsRoom(t *testing.T) {
	for _, room := range treeRooms {
		w := newTreeWorld(room)
		clk := &exploreClock{}
		known := func() bool { return w.near() }
		u, err := exploreForObject(context.Background(), clk, &exploreTally{}, exploreKey{2, 30},
			roomWalk(w.target, w.goTo, w.sweep, known), w.look)
		if err != nil || u == nil {
			t.Fatalf("room %v: %v", room, err)
		}
		want := input.Point{X: room[0]*5 + 2, Y: room[1]*5 + 2}
		if len(w.goTos) != 1 || w.goTos[0] != want || w.sweeps != 0 {
			t.Fatalf("room %v: walked %v sweeps %d, want one walk to %v", room, w.goTos, w.sweeps, want)
		}
	}
}

// Without room rows (an older d2d) the sweep runs and an unfound object keeps the explore budget semantics.
func TestNoRoomRowFallsBackToTheSweep(t *testing.T) {
	w := newTreeWorld(treeRooms[0])
	w.lm.RoomRows = nil
	clk := &exploreClock{}
	_, err := exploreForObject(context.Background(), clk, &exploreTally{}, exploreKey{2, 30},
		roomWalk(w.target, w.goTo, w.sweep, w.near), w.look)
	if w.sweeps != 1 || len(w.goTos) != 0 {
		t.Fatalf("sweeps %d walks %v", w.sweeps, w.goTos)
	}
	if err == nil {
		t.Fatal("tree far away must not be found")
	}
}

// Standing in the room already but no tree: the room is not asked for again, the sweep takes over.
func TestAtTheRoomTheSweepTakesOver(t *testing.T) {
	w := newTreeWorld(treeRooms[1])
	w.player = w.tree
	if _, ok := w.target(); ok {
		t.Fatal("already at the room centre")
	}
}

// A second class and room type: the table is data, an object with no room type is not routed.
func TestObjRoomTypeOnlyForListedClasses(t *testing.T) {
	lm := &d2map.LevelMap{RoomRows: []d2map.RoomRow{{Type: 161, Pos: d2map.Point{X: 500, Y: 500}}}}
	if _, ok := roomTargetIn(lm, quest.ObjRoomType[quest.ObjCainGibbet], input.Point{}); ok {
		t.Fatal("class without a room type routed (type 0 matched a row)")
	}
}

// The 2026-11-19 sandbox failure: the room lies beyond the known map, GoTo ends at the nearest known point with
// ErrNoPath, 16 cells short of the room (tree not listed yet). The sweep must then explore toward the room.
func TestUnreachableRoomSweepsTowardIt(t *testing.T) {
	w := newTreeWorld(treeRooms[4])
	room, _ := w.target()
	stop := input.Point{X: room.X - 16*5, Y: room.Y}
	goTo := func(_ context.Context, p input.Point) error {
		w.goTos = append(w.goTos, p)
		w.player = stop
		return fmt.Errorf("nav: no path: %w", nav.ErrNoPath)
	}
	if w.near() {
		t.Fatal("setup: stop must be out of listing range")
	}
	_, _ = exploreForObject(context.Background(), &exploreClock{}, &exploreTally{}, exploreKey{5, 30},
		roomWalk(w.target, goTo, w.sweep, w.near), w.look)
	if w.sweeps != 1 || w.hints[0] == nil || *w.hints[0] != room {
		t.Fatalf("sweeps %d hints %v, want one sweep toward %v", w.sweeps, w.hints, room)
	}
}
