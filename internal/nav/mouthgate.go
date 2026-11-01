package nav

import (
	"fmt"
	"math"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/input"
)

// FIXMOUTH4 (docs/fixmouth4.md): the cave-mouth guard is an invariant on every world click nav sends while a stay
// guard holds, not a filter per planner. fixmouth to fixmouth3 checked the plan at its start; the character was then
// moved by replans, the stuck ladder, fights inside a walk (Yield) and basic-attack swings, and the next unchecked click
// fired the trigger (V DRILLS18: a step-away replanned from 3.2 and its first click, 5.4 from the exit, fired it).
const (
	// mouthTriggerR: a click whose straight segment comes within this many subtiles of an unwanted exit may fire it.
	// V (DRILLS16 to 18 logs): click targets that fired the trigger lay 13.0 (Den), 12.1, 11.7 and 11.2 (Cave Level 1)
	// and 5.4 (Den) from the listed exit position; click targets 16.2 to 18 away never did. 15 covers 13.0 with 2 margin and stays below 16.2. The real
	// bound and shape of the trigger area (seed-dependent, docs/d2d-loop.md) are UNVERIFIED, as is whether the game's own
	// short path from a click bends into it.
	mouthTriggerR = 15.0
	// gateMinClip: a clipped click shorter than this is not worth sending; the gate holds instead.
	gateMinClip = 2.0
	// gateMaxRewrites rewritten clicks per walk, then the gate holds (the walk fails with ErrMouthGate): a goal behind
	// the mouth must not turn into a shuttle between the boundary and the away point.
	gateMaxRewrites = 3
)

// ErrMouthGate: the click would cross the trigger area of an unwanted exit and no safe replacement exists. It wraps
// ErrNoPath, so roam and explore drop the target and the Fighter counts a failed approach.
var ErrMouthGate = fmt.Errorf("%w: the click would cross the trigger area of an unwanted exit", ErrNoPath)

// mouthGate wraps the controller of a Navigator: every ClickWorld passes the gate (New installs it).
type mouthGate struct {
	input.Controller
	n *Navigator
}

func (g mouthGate) ClickWorld(p input.Point) error {
	q, err := g.n.gateClick(g.Controller, p)
	if err != nil {
		return err
	}
	return g.Controller.ClickWorld(q)
}

func exitPoint(p d2map.Point) input.Point { return input.Point{X: p.X, Y: p.Y} }

// gatedExits lists the unwanted exits of the current level while the gate is active: not in town, and the stay guard
// holds for the level or a warp just brought us here. The config mouthGuard is not consulted: overworld levels list
// exits on their edges and gating them would clip walks along the edge for no observed failure.
func (n *Navigator) gatedExits() (*d2map.LevelMap, []input.Point) {
	if n.w == nil {
		return nil, nil
	}
	lm := n.w.Current()
	if lm == nil || data.IsTown(lm.No) || !(n.arrival || n.stayLevel == lm.No) {
		return nil, nil
	}
	var out []input.Point
	for _, e := range lm.Exits {
		if e.Dest != n.wantExit {
			out = append(out, exitPoint(e.Pos))
		}
	}
	return lm, out
}

// MouthClickOK reports whether a click at to, sent while the character stands at from, keeps the cave-mouth invariant.
// The Fighter asks it before a swing (a left click walks the character to the monster).
func (n *Navigator) MouthClickOK(from, to input.Point) bool {
	_, exits := n.gatedExits()
	return clickOKAll(from, to, exits)
}

// mouthClickOK is the invariant for one exit. The click target must lie outside the bound: every logged fired click had
// its target inside it, outward ones included (V: from 3.2 the click at 5.4 fired, DRILLS18; from 3.2 the second
// step-away waypoint at 12.1 fired, DRILLS17), while long step-away clicks to targets 20 and more away never fired. From
// outside the bound the straight segment must stay outside too (the game walks roughly that line); from inside
// (arrival offset, after a swing or knockback) it may only move outward (leavesExit).
func mouthClickOK(from, to, ex input.Point) bool {
	if dist(to, ex) < mouthTriggerR {
		return false
	}
	d0 := dist(from, ex)
	if d0 >= mouthTriggerR {
		return segDist(ex, from, to) >= mouthTriggerR
	}
	return leavesExit(from, to, d2map.Point{X: ex.X, Y: ex.Y}, d0)
}

func clickOKAll(from, to input.Point, exits []input.Point) bool {
	for _, e := range exits {
		if !mouthClickOK(from, to, e) {
			return false
		}
	}
	return true
}

// gateClick returns the click to send instead of p: p itself, a clipped or away-rewritten click, or ErrMouthGate.
func (n *Navigator) gateClick(ctl input.Controller, p input.Point) (input.Point, error) {
	lm, exits := n.gatedExits()
	if len(exits) == 0 {
		return p, nil
	}
	from, err := ctl.Pos()
	if err != nil {
		from = n.lastPos
	}
	if clickOKAll(from, p, exits) {
		return p, nil
	}
	if n.gateRewrites >= gateMaxRewrites {
		n.log.Info("nav mouth gate: click held", "from", from, "to", p, "why", "rewrite cap", "cap", gateMaxRewrites)
		return input.Point{}, ErrMouthGate
	}
	q, how := n.gateRewrite(lm, from, p, exits)
	if how == "" {
		n.log.Info("nav mouth gate: click held", "from", from, "to", p, "why", "no safe click")
		return input.Point{}, ErrMouthGate
	}
	n.gateRewrites++
	n.log.Info("nav mouth gate: click rewritten", "from", from, "to", p, "as", q, "how", how, "rewrites", n.gateRewrites)
	return q, nil
}

// gateRewrite: from outside the bound the click is clipped at the boundary; from inside it goes along the away vector
// (exit to character), turned by up to 90 degrees, 20, 25 then 15 subtiles, onto a walkable cell (the target must end
// outside the bound). how is "" when nothing
// passes the invariant.
func (n *Navigator) gateRewrite(lm *d2map.LevelMap, from, to input.Point, exits []input.Point) (q input.Point, how string) {
	ex, best := exits[0], math.MaxFloat64
	for _, e := range exits {
		if d := dist(from, e); !mouthClickOK(from, to, e) && d < best {
			ex, best = e, d
		}
	}
	d0 := dist(from, ex)
	if d0 >= mouthTriggerR {
		if c, ok := clipOutside(from, to, ex); ok && dist(from, c) >= gateMinClip && clickOKAll(from, c, exits) {
			return c, "clip"
		}
		return input.Point{}, "" // no detour from outside: the walk fails and its caller picks another goal
	}
	if d0 < 0.5 {
		return input.Point{}, ""
	}
	var f *Field
	if lm != nil {
		f = n.fieldFor(lm)
	}
	ux, uy := float64(from.X-ex.X)/d0, float64(from.Y-ex.Y)/d0
	// first pass: a straight walkable line (the character walks it as clicked); second pass: any walkable target
	for pass := 0; pass < 2; pass++ {
		for _, l := range []float64{mouthAway, mouthAway + 5, 15} {
			for _, deg := range []float64{0, 30, -30, 60, -60, 90, -90} {
				a := deg * math.Pi / 180
				vx, vy := ux*math.Cos(a)-uy*math.Sin(a), ux*math.Sin(a)+uy*math.Cos(a)
				c := input.Point{X: from.X + int(math.Round(vx*l)), Y: from.Y + int(math.Round(vy*l))}
				if f != nil {
					x, y, ok := f.nearestWalkable(c.X, c.Y, 3)
					if !ok || pass == 0 && !f.los(from.X, from.Y, x, y, 0) {
						continue
					}
					c = input.Point{X: x, Y: y}
				}
				if clickOKAll(from, c, exits) {
					return c, "away"
				}
			}
		}
		if f == nil {
			break
		}
	}
	return input.Point{}, ""
}

// clipOutside returns the last point of the segment from to that stays outside the bound around ex (half a subtile
// of margin), rounded to a cell that still passes the invariant.
func clipOutside(from, to, ex input.Point) (input.Point, bool) {
	wx, wy := float64(from.X-ex.X), float64(from.Y-ex.Y)
	vx, vy := float64(to.X-from.X), float64(to.Y-from.Y)
	a := vx*vx + vy*vy
	if a == 0 {
		return input.Point{}, false
	}
	b := 2 * (wx*vx + wy*vy)
	c := wx*wx + wy*wy - mouthTriggerR*mouthTriggerR
	disc := b*b - 4*a*c
	if disc < 0 {
		return input.Point{}, false
	}
	l := math.Sqrt(a)
	t := math.Min(1, (-b-math.Sqrt(disc))/(2*a)-0.5/l)
	for ; t > 0; t -= 0.25 / l {
		q := input.Point{X: from.X + int(math.Round(t*vx)), Y: from.Y + int(math.Round(t*vy))}
		if mouthClickOK(from, q, ex) {
			return q, true
		}
	}
	return input.Point{}, false
}
