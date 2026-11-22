package nav

import (
	"context"
	"errors"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/input"
)

// maxDoorTries bounds how often ExitTo opens the same gate door: a door that does not clear the way is not clicked again.
const maxDoorTries = 2

// gateDoorReach is how near (subtiles) a door must stand to a gate cell to be the thing that closes it.
const gateDoorReach = 12.0

// openGateDoor: a gate to levelNo is on the map but no plan reaches it, and a door object stands beside it. V live
// (card mapverify, probe tools/probes/strip.py): the 2 cells of the Monastery Gate border toward the Outer Cloister
// sit behind an object (class 24, collision blocked until it is used), cut off from the rest of the grid; S that using
// it opens the cell. Walk next to the door, use it once, and let the planner re-read the grid. It reports whether it
// acted (the caller then observes again); false when there is no such door or it was tried enough.
func (n *Navigator) openGateDoor(ctx context.Context, o *obs, levelNo int) bool {
	cands := n.gateCands(o, levelNo)
	if len(cands) == 0 {
		return false
	}
	var door *d2.Unit
	best := gateDoorReach
	for i := range o.snap.Objects {
		u := &o.snap.Objects[i]
		if !n.isDoorClass(u.Class) || n.doorTried[u.ID] >= maxDoorTries {
			continue
		}
		for _, c := range cands {
			if d := dist(input.Point{X: u.X, Y: u.Y}, c.at); d < best {
				door, best = u, d
			}
		}
	}
	if door == nil {
		return false
	}
	if n.doorTried == nil {
		n.doorTried = map[uint32]int{}
	}
	n.doorTried[door.ID]++
	at := input.Point{X: door.X, Y: door.Y}
	n.log.Info("nav: the gate is cut off by a door, opening it", "to", levelNo, "door", door.Class, "at", at, "from", o.pos)
	err := n.walk(ctx, func(*obs) (input.Point, float64, error) { return at, 4, nil }, walkOpts{})
	if err != nil && !errors.Is(err, ErrAlreadyThere) {
		n.log.Info("nav: could not reach the gate door", "err", err)
		return true // tried: the counter above bounds the retries, the caller observes again
	}
	if err := n.in.Interact(input.Target{Type: uint32(d2.TypeObject), ID: door.ID, Pos: at}); err != nil {
		n.log.Info("nav door interact failed", "err", err)
	}
	n.clk.Sleep(900 * time.Millisecond)
	n.invalidate()
	return true
}

func (n *Navigator) isDoorClass(class int) bool {
	if n.doorFn != nil {
		return n.doorFn(class)
	}
	if n.e == nil || n.e.Data == nil {
		return false
	}
	ob := n.e.Data.ObjectByID(class)
	return ob != nil && ob.IsDoor
}
