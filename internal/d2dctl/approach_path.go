package d2dctl

import (
	"math"
	"sort"
)

// The map route of ApproachNPC (fixstall2, DRILLS6 F7). A blind side step of 3.5 cells does not get around a camp
// obstacle that is wider than that or an NPC that wanders into the way; `debug grid 5` gives the whole level's
// collision (V live: 5*W by 5*H rows, '.' free, '#' blocked, d2d-backend.md), so the way is planned on it: a
// breadth-first search over subtiles from the player to the free cell nearest the NPC, other NPCs counted as
// blocked, then walked leg by leg with "cmd move". Whether the grid carries the camp fires and stalls is UNVERIFIED
// (the live proof is the H.pause recipe in docs/fixstall2.md).

// approachStandOff is how near (subtiles) to the NPC the route ends: the interact verb walks the last bit itself.
const approachStandOff = 6

// npcBlockRadius is the radius (subtiles) around another NPC the route keeps clear.
const npcBlockRadius = 4

// legLen is the subtile spacing of the "cmd move" waypoints of a route.
const legLen = 12

type subPt struct{ x, y int }

// planAround returns waypoints (in cells) from (fx,fy) to the free cell near (tx,ty), or nil when no route exists.
// rows is the `debug grid 5` reply, others the cells of NPCs to keep clear of (the target excluded).
//
// Clearance is adaptive (fixstall3, DRILLS7 F10): clearance 1 (every route cell has free neighbours) keeps clear of
// wall edges and is tried first; when the player stands in a gap 1 to 2 subtiles wide (the H.pause stall cells beside
// the camp block) no clearance-1 route exists, and clearance 0 (cell centres only, diagonal steps still need both
// side cells) is the fallback.
func planAround(rows []string, fx, fy, tx, ty float64, others [][2]float64) []([2]float64) {
	if wps := planAroundClear(rows, fx, fy, tx, ty, others, 1); wps != nil {
		return wps
	}
	if wps := planAroundClear(rows, fx, fy, tx, ty, others, 0); wps != nil {
		return wps
	}
	// The player stands in a pocket of the grid (an isolated free subtile beside an object, DRILLS8 stall cell
	// 26.1,20.1): plan from the nearest free subtile that has a route and step there first (UNVERIFIED live).
	var cands []subPt
	cx, cy := int(5*fx), int(5*fy)
	for dy := -pocketRadius; dy <= pocketRadius; dy++ {
		for dx := -pocketRadius; dx <= pocketRadius; dx++ {
			if (dx != 0 || dy != 0) && cy+dy >= 0 && cy+dy < len(rows) && cx+dx >= 0 && cx+dx < len(rows[cy+dy]) && rows[cy+dy][cx+dx] == '.' {
				cands = append(cands, subPt{cx + dx, cy + dy})
			}
		}
	}
	sort.Slice(cands, func(i, j int) bool {
		return math.Hypot(float64(cands[i].x-cx), float64(cands[i].y-cy)) < math.Hypot(float64(cands[j].x-cx), float64(cands[j].y-cy))
	})
	for _, c := range cands {
		x, y := float64(c.x)/5+0.1, float64(c.y)/5+0.1
		if wps := planAroundClear(rows, x, y, tx, ty, others, 0); wps != nil {
			return append([][2]float64{{x, y}}, wps...)
		}
	}
	return nil
}

// pocketRadius is how far (subtiles) planAround looks for a free subtile when the player's own is boxed in.
const pocketRadius = 4

func planAroundClear(rows []string, fx, fy, tx, ty float64, others [][2]float64, clear int) []([2]float64) {
	h := len(rows)
	if h == 0 {
		return nil
	}
	w := 0
	for _, r := range rows {
		w = max(w, len(r))
	}
	free := func(x, y int) bool {
		return x >= 0 && y >= 0 && y < h && x < len(rows[y]) && rows[y][x] == '.'
	}
	blockedNPC := func(x, y int) bool {
		for _, o := range others {
			if math.Hypot(float64(x)-5*o[0], float64(y)-5*o[1]) <= npcBlockRadius {
				return true
			}
		}
		return false
	}
	// clearance 1: the player is wider than a point, a route along a wall edge stalls (U)
	ok := func(x, y int) bool {
		if blockedNPC(x, y) {
			return false
		}
		for dy := -clear; dy <= clear; dy++ {
			for dx := -clear; dx <= clear; dx++ {
				if !free(x+dx, y+dy) {
					return false
				}
			}
		}
		return true
	}
	start := subPt{int(5 * fx), int(5 * fy)}
	if !free(start.x, start.y) {
		return nil
	}
	target := [2]float64{5 * tx, 5 * ty}
	prev := map[subPt]subPt{start: start}
	queue := []subPt{start}
	best, bestD := start, math.Hypot(float64(start.x)-target[0], float64(start.y)-target[1])
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		if d := math.Hypot(float64(p.x)-target[0], float64(p.y)-target[1]); d < bestD {
			best, bestD = p, d
		}
		if bestD <= approachStandOff {
			break
		}
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				q := subPt{p.x + dx, p.y + dy}
				if _, seen := prev[q]; seen || (dx == 0 && dy == 0) || !ok(q.x, q.y) {
					continue
				}
				if dx != 0 && dy != 0 && !(ok(p.x+dx, p.y) && ok(p.x, p.y+dy)) { // no corner cutting
					continue
				}
				prev[q] = p
				queue = append(queue, q)
			}
		}
	}
	if best == start {
		return nil
	}
	var path []subPt // start excluded, in walking order
	for p := best; p != start; p = prev[p] {
		path = append(path, p)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return legWaypoints(free, [2]float64{5 * fx, 5 * fy}, path, clear)
}

// legWaypoints thins the step-by-step path into "cmd move" waypoints. A leg is walked as a STRAIGHT line (the verb
// stops at the first object), so a waypoint may only be as far as the line from the previous one stays on free
// cells (with the route's clearance); a bend of the route always gets a waypoint of its own. fixstall4 (DRILLS8 F13):
// the old fixed 12-step spacing let the first chord of the H.pause route cut across the corner of the camp block
// (VERIFIED offline on the real grid fixture: the chord from the stall cell to the first waypoint crossed blocked
// subtiles; that the live leg stopped there is UNVERIFIED, no live log of the leg itself).
func legWaypoints(free func(x, y int) bool, start [2]float64, path []subPt, clear int) [][2]float64 {
	var wps [][2]float64
	anchor, anchorIdx := start, -1
	emit := func(i int) {
		wps = append(wps, [2]float64{float64(path[i].x)/5 + 0.1, float64(path[i].y)/5 + 0.1})
		anchor, anchorIdx = centre(path[i]), i
	}
	for i := 0; i < len(path); i++ {
		if i == len(path)-1 {
			emit(i)
			break
		}
		if i+1-anchorIdx > legLen || !lineFree(free, anchor, centre(path[i+1]), clear) {
			emit(i)
		}
	}
	return wps
}

// lineMargin is the half width (subtiles) a straight leg keeps free around its line.
const lineMargin = 0.25

// centre is the continuous position (subtile units) of the middle of a subtile.
func centre(p subPt) [2]float64 { return [2]float64{float64(p.x) + 0.5, float64(p.y) + 0.5} }

// lineFree reports whether the straight segment a to b (continuous subtile positions) touches only free cells, each
// with `clear` free neighbours around it (sampled every quarter subtile).
func lineFree(free func(x, y int) bool, a, b [2]float64, clear int) bool {
	n := max(int(4*math.Max(math.Abs(b[0]-a[0]), math.Abs(b[1]-a[1]))), 1)
	for k := 0; k <= n; k++ {
		px := a[0] + (b[0]-a[0])*float64(k)/float64(n)
		py := a[1] + (b[1]-a[1])*float64(k)/float64(n)
		m := lineMargin // the walker is not a point: a quarter subtile each side, except at the start (it may touch a wall)
		if math.Hypot(px-a[0], py-a[1]) < 1 {
			m = 0
		}
		for _, ox := range []float64{-m, m} {
			for _, oy := range []float64{-m, m} {
				cx, cy := int(math.Floor(px+ox)), int(math.Floor(py+oy))
				for dy := -clear; dy <= clear; dy++ {
					for dx := -clear; dx <= clear; dx++ {
						if !free(cx+dx, cy+dy) {
							return false
						}
					}
				}
			}
		}
	}
	return true
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// routeCells is the length (cells) of the polyline from (fx,fy) through the waypoints.
func routeCells(fx, fy float64, wps [][2]float64) float64 {
	d := 0.0
	for _, w := range wps {
		d += math.Hypot(w[0]-fx, w[1]-fy)
		fx, fy = w[0], w[1]
	}
	return d
}
