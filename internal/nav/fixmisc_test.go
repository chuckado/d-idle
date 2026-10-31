package nav

import (
	"errors"
	"testing"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/input"
)

// DRILLS5 F4: a position jump (debug goto, town portal or waypoint arrival, restart) puts the player where the map
// tracker has not read yet. PathTo is pure planning, so before Sync every point there is unreachable; after Sync the
// rooms around the new position are loaded and the path exists.
func TestSyncAfterSameLevelJump(t *testing.T) {
	s := newSim(openField(0, 0, 400, 60), input.Point{X: 5, Y: 30})
	n := s.navigator()
	if err := n.Sync(); err != nil {
		t.Fatal(err)
	}
	s.px, s.py = 300.5, 30.5 // jump without a walk
	to := input.Point{X: 310, Y: 30}
	if _, err := n.PathTo(to); err == nil {
		t.Fatal("before Sync the jumped-to rooms are unknown: the pure planner must refuse (documents the F4 cause)")
	}
	if err := n.Sync(); err != nil {
		t.Fatal(err)
	}
	if _, err := n.PathTo(to); err != nil {
		t.Fatalf("after Sync: %v", err)
	}
}

func TestSyncAfterLevelChangeJump(t *testing.T) {
	a := &simLevel{no: 1, rect: d2map.Rect{X: 0, Y: 0, W: 100, H: 60}}
	b := &simLevel{no: 2, rect: d2map.Rect{X: 300, Y: 0, W: 100, H: 60}}
	s := newSim(openField(0, 0, 400, 60), input.Point{X: 5, Y: 30}, a, b)
	n := s.navigator()
	if err := n.Sync(); err != nil {
		t.Fatal(err)
	}
	s.px, s.py = 350.5, 30.5 // town portal / waypoint arrival in another level
	if err := n.Sync(); err != nil {
		t.Fatal(err)
	}
	if _, err := n.PathTo(input.Point{X: 360, Y: 30}); err != nil {
		t.Fatalf("after Sync in the new level: %v", err)
	}
}

// A restart: the first observation after it is a loading screen, then the game. Sync reports ErrNotInGame while loading
// (the caller keeps its old behaviour) and works on the next call.
func TestSyncDuringLoadingThenReady(t *testing.T) {
	s := newSim(openField(0, 0, 100, 60), input.Point{X: 5, Y: 30})
	n := s.navigator()
	s.loadTicks = 1
	if err := n.Sync(); !errors.Is(err, ErrNotInGame) {
		t.Fatalf("loading: %v", err)
	}
	if err := n.Sync(); err != nil {
		t.Fatal(err)
	}
	if _, err := n.PathTo(input.Point{X: 20, Y: 30}); err != nil {
		t.Fatal(err)
	}
}
