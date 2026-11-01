package nav

import (
	"context"
	"errors"
	"testing"
	"time"

	"d2lod-bot/internal/input"
)

// corridorSim: a one-cell-wide vertical corridor; the player stands in it, the cell east of the player is rock.
func corridorSim() *sim {
	g := asciiTruth(0, 0,
		"##.###",
		"##.###",
		"##.###",
		"##.###",
		"##.###")
	s := newSim(g, input.Point{X: 2, Y: 2})
	s.knownSrc = g
	s.levels[0].known = g
	s.levels[0].lm.Grid = g
	return s
}

// fixemptyplan F4 (docs/fixemptyplan.md): a blocked goal whose nearest walkable cell is the player's own cell. The plan
// is empty without an error; PathTo used to return (nil, nil) and GoTo "no path: empty path".
func TestBlockedGoalSnapsOntoStart(t *testing.T) {
	s := corridorSim()
	n := s.navigator()
	goal := input.Point{X: 5, Y: 2} // rock, beyond the arrival tolerance

	res, err := planOn(t, s.levels[0].known, input.Point{X: 2, Y: 2}, goal, nil)
	if err != nil || !res.snapped || len(res.wps) != 0 {
		t.Fatalf("plan: err=%v snapped=%v wps=%d (the cause: an empty result, no error)", err, res.snapped, len(res.wps))
	}

	wps, err := n.PathTo(goal)
	if wps != nil || !errors.Is(err, ErrAlreadyThere) {
		t.Fatalf("PathTo = %v, %v; want nil, ErrAlreadyThere", wps, err)
	}
	if !errors.Is(err, ErrNoPath) {
		t.Fatal("older callers match ErrNoPath: keep it")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = n.GoTo(ctx, goal)
	if !errors.Is(err, ErrAlreadyThere) {
		t.Fatalf("GoTo = %v; want ErrAlreadyThere", err)
	}
}

// Edge cases that must stay as they were: a walkable goal elsewhere has a path, the player's own walkable cell is not
// an error from the planner's point of view of a real walk, a far blocked goal snaps to a different cell.
func TestSnappedStartEdgeCases(t *testing.T) {
	cases := []struct {
		name         string
		goal         input.Point
		alreadyThere bool
		wantPath     bool
	}{
		{"blocked beside the player", input.Point{X: 5, Y: 2}, true, false},
		{"walkable cell down the corridor", input.Point{X: 2, Y: 4}, false, true},
		{"blocked goal that snaps to another cell", input.Point{X: 5, Y: 4}, false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			n := corridorSim().navigator()
			wps, err := n.PathTo(c.goal)
			if got := errors.Is(err, ErrAlreadyThere); got != c.alreadyThere {
				t.Fatalf("already-there = %v (err %v)", got, err)
			}
			if c.wantPath && (err != nil || len(wps) == 0) {
				t.Fatalf("want a path: %v %v", wps, err)
			}
		})
	}
}
