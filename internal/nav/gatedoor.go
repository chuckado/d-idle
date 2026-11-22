package nav

import (
	"context"
	"errors"
	"math"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2map"
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
	// The map's object presets, not the snapshot: the snapshot lists the objects near the character only, and the gate
	// is usually far away when the cut-off shows (V live: the hero stood 80 subtiles off).
	var at input.Point
	best, found := gateDoorReach, false
	for _, p := range o.lm.Presets {
		if p.Kind != d2map.PresetObject || !n.isDoorClass(p.ID) || n.doorTried[p.Pos] >= maxDoorTries {
			continue
		}
		for _, c := range cands {
			if d := dist(input.Point{X: p.Pos.X, Y: p.Pos.Y}, c.at); d < best {
				at, best, found = input.Point{X: p.Pos.X, Y: p.Pos.Y}, d, true
			}
		}
	}
	// the snapshot's objects count too (they carry the unit id, and a door the map did not list is still a door)
	for i := range o.snap.Objects {
		u := &o.snap.Objects[i]
		pt := d2map.Point{X: u.X, Y: u.Y}
		if !n.isDoorClass(u.Class) || n.doorTried[pt] >= maxDoorTries {
			continue
		}
		for _, c := range cands {
			if d := dist(input.Point{X: u.X, Y: u.Y}, c.at); d < best {
				at, best, found = input.Point{X: u.X, Y: u.Y}, d, true
			}
		}
	}
	if !found {
		if n.approachGate(ctx, o, levelNo, cands[0]) {
			return true
		}
		n.logNoDoor(o, levelNo, cands)
		return false
	}
	if n.doorTried == nil {
		n.doorTried = map[d2map.Point]int{}
	}
	n.doorTried[d2map.Point{X: at.X, Y: at.Y}]++
	n.log.Info("nav: the gate is cut off by a door, opening it", "to", levelNo, "door", at, "from", o.pos)
	err := n.walk(ctx, func(*obs) (input.Point, float64, error) { return at, 4, nil }, walkOpts{})
	cur, oerr := n.observeWait(ctx)
	if err != nil && !errors.Is(err, ErrAlreadyThere) && (oerr != nil || dist(cur.pos, at) > 8) {
		// a walk onto the door's own blocked cell ends stuck beside it, which is close enough to use it
		n.log.Info("nav: could not reach the gate door", "err", err)
		return true // tried: the counter above bounds the retries, the caller observes again
	}
	tg := input.Target{Type: uint32(d2.TypeObject), Pos: at}
	if oerr == nil {
		bestD := 6.0
		for i := range cur.snap.Objects {
			u := &cur.snap.Objects[i]
			if d := dist(input.Point{X: u.X, Y: u.Y}, at); d < bestD && n.isDoorClass(u.Class) {
				tg, bestD = input.Target{Type: uint32(d2.TypeObject), ID: u.ID, Pos: input.Point{X: u.X, Y: u.Y}}, d
			}
		}
	}
	if err := n.in.Interact(tg); err != nil {
		n.log.Info("nav door interact failed", "err", err)
	}
	n.clk.Sleep(900 * time.Millisecond)
	n.invalidate()
	return true
}

// approachGate walks to the reachable cell nearest to a cut-off gate, at most twice per level pair: the map lists the
// objects of the rooms that are up, and a door at the far end of the level is not up while the hero stands at the
// other end (V live, Monastery Gate: 5 objects listed from the arrival point, the nearest 42 subtiles off, the door
// at the gate not among them). Once the rooms there load, the next pass finds the door. It reports whether it walked.
func (n *Navigator) approachGate(ctx context.Context, o *obs, levelNo int, g gate) bool {
	key := [2]int{o.level, levelNo}
	if n.doorApproach == nil {
		n.doorApproach = map[[2]int]int{}
	}
	if n.doorApproach[key] >= 2 {
		return false
	}
	cell, ok := n.nearestReachableTo(o, g.at)
	if !ok || dist(cell, g.at) > dist(o.pos, g.at)-10 { // no reachable point nearer to the gate: nothing to approach
		return false
	}
	n.doorApproach[key]++
	n.log.Info("nav: the gate is cut off, walking to the nearest reachable cell to look for its door", "to", levelNo, "gate", g.at, "cell", cell, "from", o.pos)
	if err := n.walk(ctx, func(*obs) (input.Point, float64, error) { return cell, 3, nil }, walkOpts{}); err != nil && !errors.Is(err, ErrAlreadyThere) {
		n.log.Info("nav: could not walk to the cut-off gate", "err", err)
	}
	return true
}

// nearestReachableTo is the cell the character can walk to on the known grid that is nearest to target and has some
// clearance from walls (a plan to the target itself fails, it is cut off, and a failed plan leaves no waypoints).
func (n *Navigator) nearestReachableTo(o *obs, target input.Point) (input.Point, bool) {
	n.planMu.Lock()
	defer n.planMu.Unlock()
	f := n.fieldFor(o.lm)
	f.ClearDyn()
	sx, sy, ok := f.nearestWalkable(o.pos.X, o.pos.Y, 8)
	if !ok {
		return input.Point{}, false
	}
	hops := floodHops(f, sx, sy)
	best, bestD, found := input.Point{}, math.MaxFloat64, false
	for i, h := range hops {
		if h < 0 {
			continue
		}
		x, y := f.pt(int32(i))
		if f.Clearance(x, y) < 1.5 {
			continue
		}
		if d := dist(input.Point{X: x, Y: y}, target); d < bestD {
			best, bestD, found = input.Point{X: x, Y: y}, d, true
		}
	}
	return best, found
}

// logNoDoor says once per level pair why no door was found beside the cut-off gate (what the map and the snapshot list).
func (n *Navigator) logNoDoor(o *obs, levelNo int, cands []gate) {
	key := [2]int{o.level, levelNo}
	if n.doorLogged == nil {
		n.doorLogged = map[[2]int]bool{}
	}
	if n.doorLogged[key] {
		return
	}
	n.doorLogged[key] = true
	nearest := func(pts []input.Point) float64 {
		best := -1.0
		for _, p := range pts {
			for _, c := range cands {
				if d := dist(p, c.at); best < 0 || d < best {
					best = d
				}
			}
		}
		return best
	}
	var presets, snaps []input.Point
	for _, p := range o.lm.Presets {
		if p.Kind == d2map.PresetObject {
			presets = append(presets, input.Point{X: p.Pos.X, Y: p.Pos.Y})
		}
	}
	for i := range o.snap.Objects {
		snaps = append(snaps, input.Point{X: o.snap.Objects[i].X, Y: o.snap.Objects[i].Y})
	}
	n.log.Info("nav: cut-off gate, no door beside it", "to", levelNo, "gate", cands[0].at, "presetObjects", len(presets), "nearestPreset", nearest(presets), "snapObjects", len(snaps), "nearestSnap", nearest(snaps))
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
