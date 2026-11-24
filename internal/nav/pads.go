package nav

import (
	"context"
	"sort"
	"time"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/input"
)

// teleportPadClass is the objects.txt row "Teleport Pad" of the Arcane Sanctuary (V lab, tools/probes/objectlist.py level 74
// seed 11: class 192, mode NU, two pads on the arrival platform). The sanctuary is platforms over a void; a pad carries the
// hero to a platform the walk planner cannot reach (docs/act2.md Y4). Where each pad leads is not in any table the bot reads
// (U), so the hop is a search: use the nearest pad not used yet, look again, walk when the goal became reachable.
const teleportPadClass = 192

const (
	padJump     = 12.0 // subtiles: a move of the hero this far in one poll is the pad's jump
	padWait     = 4 * time.Second
	padMaxHops  = 12
	padReach    = 8.0 // subtiles: the walk ends this near a pad and the clicks do the rest
	padFreeZone = 4.0 // pads this close to where the hero landed are the pair of the pad it took
)

type padKey struct{ level, x, y int }

// padsUsed marks the pads this game already took (or landed on), so a hop search never goes round in a circle.
func (n *Navigator) usePad(level int, p input.Point) {
	if n.padsUsed == nil {
		n.padsUsed = map[padKey]bool{}
	}
	n.padsUsed[padKey{level, p.X, p.Y}] = true
}

// padApproach is a point the planner reaches in full and that lies within padReach of the pad: the pad itself, else a cell
// of the ring around it (V lab, level 74: a pad cell reads blocked, its neighbours belong to the platform).
func (n *Navigator) padApproach(o *obs, p input.Point) (input.Point, bool) {
	cands := []input.Point{p}
	for _, d := range [][2]int{{4, 0}, {-4, 0}, {0, 4}, {0, -4}, {3, 3}, {-3, 3}, {3, -3}, {-3, -3}, {7, 0}, {-7, 0}, {0, 7}, {0, -7}} {
		cands = append(cands, input.Point{X: p.X + d[0], Y: p.Y + d[1]})
	}
	for _, c := range cands {
		if res, err := n.plan(o.lm, o.pos, c, n.tune().ArriveTolerance, nil, 0); err == nil && !res.partial && len(res.wps) > 0 && dist(res.wps[len(res.wps)-1], p) <= padReach {
			return c, true
		}
	}
	return input.Point{}, false
}

// hopPad walks to the nearest unused pad that can be walked to and stands on it until the hero jumps. It returns true after
// a jump, false when no pad is left or the jump did not come. goal is the point the caller could not reach.
func (n *Navigator) hopPad(ctx context.Context, o *obs, goal input.Point) bool {
	if n.padHops >= padMaxHops {
		return false
	}
	// The snapshot lists only objects near the hero; the level map's presets list the whole level (V: d2dworld levelmap
	// level 74 carries the pads), so the search reaches pads on platforms far from here.
	seen := map[input.Point]bool{}
	var pads []input.Point
	near := map[input.Point]input.Point{} // pad to the point the walk goes to
	add := func(p input.Point) {
		if !seen[p] && !n.padsUsed[padKey{o.level, p.X, p.Y}] {
			if at, ok := n.padApproach(o, p); ok {
				pads = append(pads, p)
				near[p] = at
			}
		}
		seen[p] = true
	}
	for i := range o.snap.Objects {
		if u := &o.snap.Objects[i]; u.Class == teleportPadClass {
			add(input.Point{X: u.X, Y: u.Y})
		}
	}
	if o.lm != nil {
		for _, pr := range o.lm.Presets {
			if pr.Kind == d2map.PresetObject && pr.ID == teleportPadClass {
				add(input.Point{X: pr.Pos.X, Y: pr.Pos.Y})
			}
		}
	}
	if len(pads) == 0 {
		if len(seen) > 0 {
			n.log.Info("nav teleport pad: none usable", "level", o.level, "known", len(seen), "used", len(n.padsUsed), "goal", goal)
		}
		return false
	}
	sort.Slice(pads, func(i, j int) bool { return dist(o.pos, pads[i]) < dist(o.pos, pads[j]) })
	pad := pads[0]
	n.usePad(o.level, pad)
	n.padHops++
	at := near[pad]
	n.log.Info("nav teleport pad", "level", o.level, "pad", pad, "goal", goal, "hop", n.padHops, "others", len(pads)-1)
	if err := n.walk(ctx, func(*obs) (input.Point, float64, error) { return at, 1.5, nil }, walkOpts{goal: &at}); err != nil && dist(n.posNow(), pad) > padReach {
		return false
	}
	end := n.clk.Now().Add(padWait)
	for n.clk.Now().Before(end) && ctx.Err() == nil {
		_ = n.in.ClickWorld(pad)
		n.clk.Sleep(5 * n.tune().Poll())
		cur, err := n.observeWait(ctx)
		if err != nil {
			return false
		}
		if dist(cur.pos, o.pos) > padJump || cur.level != o.level {
			for i := range cur.snap.Objects {
				u := &cur.snap.Objects[i]
				if p := (input.Point{X: u.X, Y: u.Y}); u.Class == teleportPadClass && dist(cur.pos, p) <= padFreeZone {
					n.usePad(cur.level, p) // the pair of the pad just taken
				}
			}
			return true
		}
	}
	return false
}

// posNow is the hero's position now (the origin when it cannot be read).
func (n *Navigator) posNow() input.Point {
	if o, err := n.observe(); err == nil {
		return o.pos
	}
	return input.Point{}
}
