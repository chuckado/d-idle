package d2dctl

import (
	"math"
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
func planAround(rows []string, fx, fy, tx, ty float64, others [][2]float64) []([2]float64) {
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
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
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
	var path []subPt
	for p := best; p != start; p = prev[p] {
		path = append(path, p)
	}
	var wps [][2]float64
	for i := len(path) - 1; i >= 0; i-- {
		if n := len(path) - 1 - i; n%legLen == legLen-1 || i == 0 {
			wps = append(wps, [2]float64{float64(path[i].x)/5 + 0.1, float64(path[i].y)/5 + 0.1})
		}
	}
	return wps
}
