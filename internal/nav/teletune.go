package nav

import (
	"math"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/input"
)

// TELETUNE: when a hop is worth its time and mana (docs/teletune.md). A hop is declined, and the caller walks, when
//  1. the walk to the landing cell is shorter than the break-even leg (the hop costs more time than the walk), or
//  2. the mana left after it would fall below the reserve kept for fights.
// Nothing here waits for mana: a hop that mana does not allow is a walk (the regeneration wait of DRILLS22 is gone).

// teleHopTicksNominal is the lockstep ticks one hop takes: the settle sleep (700 ms) plus the cast animation and one
// observation. The default leg uses 64 (DRILLS22: 448 ticks / 7 hops), about twice this.
const teleHopTicksNominal = 33.0 // 17.5 settle + about 13 cast frames (S, U) + 2.5 poll

// RunSubPerTick is the running speed (V live, DRILLS: about 0.5625 subtiles per tick).
const RunSubPerTick = 0.5625

// BreakEvenLeg is the walking distance (subtiles) at which walking takes as long as one hop of hopTicks ticks.
func BreakEvenLeg(subPerTick, hopTicks float64) float64 { return subPerTick * hopTicks }

// hopFacts are the inputs of one hop decision. WalkLen is the A* length of the walk to the landing cell in
// subtiles (+Inf when no known walk leads there: a wall or void in between).
type hopFacts struct {
	WalkLen             float64
	Mana, MaxMana, Cost int
	ReservePct, MinLeg  float64
}

// hopWorth says whether to hop and, when not, why ("leg": too short, "mana": reserve). MaxMana 0 means the stat is
// unknown: only the leg rule applies.
func hopWorth(h hopFacts) (bool, string) {
	if h.WalkLen < h.MinLeg {
		return false, "leg"
	}
	if h.MaxMana > 0 {
		reserve := int(math.Ceil(h.ReservePct / 100 * float64(h.MaxMana)))
		if h.Mana-h.Cost < reserve {
			return false, "mana"
		}
	}
	return true, ""
}

// TeleLegTicks models one leg of walkSub subtiles crossed with hops of hopSub subtiles each, hopTicks per hop, then
// walking the rest. Mana regenerates at regenPerSec and a hop waits when the pool is short (the old behaviour). It
// is the model behind the DRILLS22 table in docs/teletune.md; the code itself uses hopWorth.
func TeleLegTicks(walkSub, hopSub, hopTicks float64, hops int, mana, cost, regenPerSec, framesPerSec float64) float64 {
	t := 0.0
	left := walkSub
	for i := 0; i < hops && left > 0; i++ {
		if mana < cost {
			w := (cost - mana) / regenPerSec * framesPerSec
			t += w
			mana += w * regenPerSec / framesPerSec
		}
		mana -= cost
		t += hopTicks
		mana += hopTicks * regenPerSec / framesPerSec
		left -= hopSub
	}
	if left > 0 {
		t += left / RunSubPerTick
	}
	return t
}

// teleCost is the mana one cast needs: the skill data at the learned level when the bot wired TeleportCost, else
// the TeleportMana field, else the level 1 upper bound.
func (n *Navigator) teleCost(p *d2.Player) int {
	if n.TeleportCost != nil && p != nil {
		if c := n.TeleportCost(p); c > 0 {
			return c
		}
	}
	return n.teleMana()
}

// walkLen is the A* length from the character to c over known walkable cells, +Inf without a known walk.
func (n *Navigator) walkLen(o *obs, c input.Point) float64 {
	f := n.fieldFor(o.lm)
	if f.W == 0 {
		return math.Inf(1)
	}
	sx, sy, ok := f.nearestWalkable(o.pos.X, o.pos.Y, 8)
	if !ok || !f.Walkable(c.X, c.Y) {
		return math.Inf(1)
	}
	f.passMask = fWalk
	cells, partial := n.pl.Plan(f, sx, sy, c.X, c.Y, 0, 1)
	if partial || len(cells) == 0 {
		return math.Inf(1)
	}
	var sum float64
	px, py := f.pt(cells[0])
	for _, i := range cells[1:] {
		x, y := f.pt(i)
		sum += math.Hypot(float64(x-px), float64(y-py))
		px, py = x, y
	}
	return sum
}

// hopDeclined applies hopWorth to a candidate landing cell and logs a decline once per reason change.
func (n *Navigator) hopDeclined(o *obs, cell input.Point) bool {
	t := n.tune()
	p := o.snap.Player
	ok, why := hopWorth(hopFacts{
		WalkLen: n.walkLen(o, cell), Mana: p.Mana, MaxMana: p.MaxMana, Cost: n.teleCost(p),
		ReservePct: t.TeleReservePct, MinLeg: t.TeleMinLeg,
	})
	if ok {
		return false
	}
	if n.teleWhy != why {
		n.teleWhy = why
		n.log.Info("nav teleport declined, walking", "why", why, "mana", p.Mana, "max", p.MaxMana)
	}
	return true
}
