package main

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/drill"
	"d2lod-bot/internal/input"
)

// d11Setup is the common start of the DRILLS11 fights: alive, in the field level, the navigator's map built, the level's
// natural monsters gone, a known life. The caller defers d11Leave.
func (l *lab) d11Setup(ctx context.Context, r *drill.Run, level, maxLife, walk int) error {
	if err := l.Goto(ctx, r, level); err != nil {
		return fmt.Errorf("goto: %w", err)
	}
	if sw, err := l.be.Env.Snapshot(); err == nil && sw.Player != nil {
		wctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		_ = l.navr.GoTo(wctx, input.Point{X: sw.Player.X + walk, Y: sw.Player.Y})
		cancel()
	}
	if _, err := r.Cheat(ctx, "debug kill", "clear the level's natural monsters: the spawned ones are the fight under test"); err != nil {
		return err
	}
	_ = l.Tick(ctx, 60)
	for _, v := range []string{fmt.Sprintf("debug stat 7 %d", maxLife*256), fmt.Sprintf("debug stat 6 %d", maxLife*256)} {
		if _, err := r.Cheat(ctx, v, fmt.Sprintf("life %d: a known life for the fight under test", maxLife)); err != nil {
			return err
		}
	}
	_ = l.Tick(ctx, 10)
	return nil
}

func (l *lab) d11Leave(ctx context.Context, r *drill.Run) {
	_, _ = r.Cheat(ctx, "debug goto 1", "back to the Rogue Encampment after the fight: the spawned monsters stay in the field")
	_ = l.Tick(ctx, 25)
}

// d11Adjacent counts living monsters within 3 subtiles of the player (the crowd rule's adjacency) and all listed alive.
func (l *lab) d11Adjacent(ctx context.Context) (adj, alive int, life, maxLife int) {
	return l.d11Within(ctx, 3)
}

// d11att counts the distances (whole subtiles) at which living monsters stood while in an attack mode (A1, A2), over a drill.
var d11att = map[int]int{}

// d11all is the same for every living monster in any mode; d11each the distance of each, sorted, at the last read.
var d11all = map[int]int{}
var d11each string

// d11min[k] is the smallest distance (subtiles, one decimal) at which the k+1th nearest living monster stood over the reads of a drill.
var d11min = [3]float64{1e9, 1e9, 1e9}
var d11reads int

func d11AttStr() string {
	hist := func(m map[int]int) string {
		var k []int
		for d := range m {
			k = append(k, d)
		}
		sort.Ints(k)
		var b []string
		for _, d := range k {
			b = append(b, fmt.Sprintf("%d:%d", d, m[d]))
		}
		return strings.Join(b, " ")
	}
	return fmt.Sprintf("%s] [over %d reads the 1st, 2nd and 3rd nearest monster were at best %.1f, %.1f, %.1f subtiles away] [all living monsters ", hist(d11att), d11reads, d11min[0], d11min[1], d11min[2]) + "" + hist(d11all) + "] [last read " + d11each
}

// d11Within is d11Adjacent with a radius in subtiles.
func (l *lab) d11Within(ctx context.Context, radius float64) (adj, alive int, life, maxLife int) {
	sn, err := l.be.Env.Snapshot()
	if err != nil || sn.Player == nil {
		return
	}
	life, maxLife = sn.Player.Life, sn.Player.MaxLife
	ms, err := l.be.Client.Monsters(ctx)
	lv, lerr := l.be.Client.Level(ctx) // the monsters verb gives level-relative cells, the snapshot act subtiles: compare in cells
	if err != nil || lerr != nil {
		return
	}
	var ds []int
	var fds []float64
	defer func() {
		sort.Ints(ds)
		d11each = fmt.Sprint(ds)
		sort.Float64s(fds)
		d11reads++
		for k := 0; k < 3 && k < len(fds); k++ {
			d11min[k] = min(d11min[k], fds[k])
		}
	}()
	for _, m := range ms.Items {
		if m.HP.Cur <= 0 || m.Mode == "DD" || m.Mode == "DT" {
			continue
		}
		alive++
		d := math.Hypot(5*(m.CellX-lv.X), 5*(m.CellY-lv.Y))
		d11all[int(d+0.5)]++
		fds = append(fds, math.Round(d*10)/10)
		ds = append(ds, int(d+0.5))
		if strings.HasPrefix(m.Mode, "A") {
			d11att[int(d+0.5)]++
		}
		if d <= radius {
			adj++
		}
	}
	return
}

// d11OneScroll empties the belt and the pack and puts exactly one Town Portal scroll in the pack.
func (l *lab) d11OneScroll(ctx context.Context, r *drill.Run) error {
	if _, err := r.Cheat(ctx, "debug belt clear", "no potion in the belt: the fight under test has none to drink"); err != nil {
		return err
	}
	if err := l.ClearPack(ctx, r); err != nil {
		return err
	}
	if err := l.GiveGround(ctx, r, "tsc"); err != nil {
		return err
	}
	_ = l.Tick(ctx, 30)
	g, err := l.be.Client.Ground(ctx)
	if err != nil {
		return err
	}
	for _, it := range g.Items { // only the scroll: potions and gold left on the ground by earlier kills stay there
		if it.Code != "tsc" {
			continue
		}
		if err := l.be.Client.CmdPickup(ctx, it.Unit); err != nil {
			return err
		}
		_, _ = l.be.Client.Step(ctx, 60)
		break
	}
	_ = l.Tick(ctx, 10)
	if v, err := l.PView(ctx); err != nil || v.Scrolls != 1 {
		return fmt.Errorf("want exactly one scroll in the pack, have %d (err %v)", v.Scrolls, err)
	}
	return nil
}

// d11Mon is the melee monster of the portal drills: a Zombie does not scatter after a hit, so several end up adjacent at once (a Fallen darts in and out: 1 of 3 within 3 subtiles in 600 ticks).
const d11Mon = "zombie1"

// d11Spawn spawns n zombies and steps until at least want of them are adjacent (1200 ticks at most).
func (l *lab) d11Spawn(ctx context.Context, r *drill.Run, n, want int, radius float64) (adj int) {
	clear(d11att)
	clear(d11all)
	d11min, d11reads = [3]float64{1e9, 1e9, 1e9}, 0
	if _, err := r.Cheat(ctx, fmt.Sprintf("debug spawn %s %d near", d11Mon, n), "melee monsters near the throwaway: the fight under test (docs/forkcheats.md)"); err != nil {
		return -1
	}
	for i := 0; i < 120; i++ {
		_ = l.Tick(ctx, 10)
		if a, _, _, _ := l.d11Within(ctx, radius); a >= want {
			return a
		}
		l.d11Lure(ctx)
	}
	a, _, _, _ := l.d11Within(ctx, radius)
	if a < want && radius < 8 { // the rule's adjacency was never reached: carry on until the same number of monsters is within melee reach (8 subtiles)
		for i := 0; i < 60; i++ {
			if w, _, _, _ := l.d11Within(ctx, 8); w >= want {
				break
			}
			_ = l.Tick(ctx, 10)
			l.d11Lure(ctx)
		}
		a, _, _, _ = l.d11Within(ctx, radius)
	}
	return a
}

// d11Lure walks the character toward the nearest living monster that is not yet within 6 subtiles: a spawned zombie
// idles until something comes close (measured: most stay 14 to 20 subtiles away for 1200 ticks), then it follows.
func (l *lab) d11Lure(ctx context.Context) {
	ms, err := l.be.Client.Monsters(ctx)
	lv, lerr := l.be.Client.Level(ctx)
	if err != nil || lerr != nil {
		return
	}
	best, bx, by := 1e9, 0.0, 0.0
	for _, m := range ms.Items {
		if m.HP.Cur <= 0 || m.Mode == "DD" || m.Mode == "DT" {
			continue
		}
		if d := 5 * math.Hypot(m.CellX-lv.X, m.CellY-lv.Y); d > 6 && d < best {
			best, bx, by = d, m.CellX, m.CellY
		}
	}
	if best < 1e9 {
		_ = l.be.Client.CmdMove(ctx, bx, by)
	}
}

func (l *lab) PortalMelee(ctx context.Context, r *drill.Run, adj int, act bool) (pi drill.PMInfo) {
	pi.Where = fmt.Sprintf("Blood Moor, zombies spawned (waiting for %d within 3 subtiles; the crowd rule starts at 8 within 8)", adj)
	if err := l.d11Setup(ctx, r, 2, 300, 8); err != nil {
		pi.Err = err.Error()
		return
	}
	defer l.d11Leave(ctx, r)
	if err := l.d11OneScroll(ctx, r); err != nil {
		pi.Err = "scroll: " + err.Error()
		return
	}
	n := adj
	if adj >= 3 {
		n = 2 * adj // more monsters, more chance that 3 are within 3 subtiles at one read
	}
	if got := l.d11Spawn(ctx, r, n, adj, 3); got < 0 {
		pi.Err = "spawn failed"
		return
	}
	if err := l.SetLife(ctx, r, 102); err != nil {
		pi.Err = "life: " + err.Error()
		return
	}
	_ = l.Tick(ctx, 1)
	sn, err := l.be.Env.Snapshot()
	if err != nil || sn.Player == nil {
		pi.Err = "snapshot"
		return
	}
	pi.Adjacent, pi.Hostiles, pi.Life, pi.MaxLife = l.d11Adjacent(ctx)
	pi.Within8, _, _, _ = l.d11Within(ctx, 8)
	pi.AttackDists = d11AttStr()
	pi.Life = sn.Player.Life
	hp, mp := l.beltKinds(sn)
	pi.BeltPotions = hp + mp
	v, _ := l.PView(ctx)
	pi.ScrollsBefore = v.Scrolls
	a := l.rig.Fighter.Safety(sn)
	pi.Decision = a.String()
	if a == combat.TownPortal && act {
		pi.Acted = true
		if err := l.rig.Fighter.Act(ctx, a); err != nil {
			pi.ActErr = err.Error()
		}
		_ = l.Tick(ctx, 20)
	}
	v, _ = l.PView(ctx)
	pi.ScrollsAfter = v.Scrolls
	pi.PortalOpen = v.Near > 0
	if s2, e := l.be.Env.Snapshot(); e == nil && s2.Player != nil {
		pi.LifeAfter = s2.Player.Life
	}
	return
}

// CastTrace casts one town portal scroll with a monster in melee range, stepping one tick at a time from the click and
// writing the timeline of events (tick offsets from the click): player mode changes, life drops (hits), the nearest
// monster's mode changes, the scroll leaving the pack, the portal object. trial moves the standing place east so an
// earlier trial's portal does not hide this one; sync clicks when the monster's attack animation (A1 or A2) starts, so
// its blow lands inside the cast window; delay steps are waited after that start, so the blow lands earlier in the cast
// (the blow hit comes 4 steps after its animation starts: delay 3 puts it 1 step after the click).
func (l *lab) CastTrace(ctx context.Context, r *drill.Run, life, trial int, sync bool, delay int) (ci drill.CastInfo) {
	ci.PortalTick, ci.ScrollTick = -1, -1
	if err := l.d11Setup(ctx, r, 2, life, 8*(trial+1)); err != nil {
		ci.Err = err.Error()
		return
	}
	defer l.d11Leave(ctx, r)
	if err := l.d11Spot(ctx, r, trial); err != nil {
		ci.Err = "spot: " + err.Error()
		return
	}
	if err := l.d11OneScroll(ctx, r); err != nil {
		ci.Err = "scroll: " + err.Error()
		return
	}
	v, _ := l.PView(ctx)
	ci.ScrollsBefore = v.Scrolls
	if len(v.ScrollIDs) == 0 {
		ci.Err = "no scroll"
		return
	}
	if got := l.d11Spawn(ctx, r, 1, 1, 8); got < 0 {
		ci.Err = "spawn failed"
		return
	}
	c := l.be.Client
	if err := c.Key(ctx, "i"); err != nil {
		ci.Err = err.Error()
		return
	}
	defer func() { _ = c.Key(ctx, "i"); _ = l.Tick(ctx, 5) }()
	_ = l.Tick(ctx, 18)
	x, y, err := c.ItemAt(ctx, v.ScrollIDs[0])
	if err != nil {
		ci.Err = "itemat: " + err.Error()
		return
	}
	_ = c.Hover(ctx, x, y)
	_ = l.Tick(ctx, 4)
	nearest := func() (string, float64) { // the nearest living monster's mode and distance in subtiles
		ms, e := c.Monsters(ctx)
		lv, e2 := c.Level(ctx)
		if e != nil || e2 != nil {
			return "-", -1
		}
		best, bm := 1e9, "-"
		for _, m := range ms.Items {
			if m.HP.Cur > 0 && m.Mode != "DD" && m.Mode != "DT" {
				if d := 5 * math.Hypot(m.CellX-lv.X, m.CellY-lv.Y); d < best {
					best, bm = d, m.Mode
				}
			}
		}
		return bm, best
	}
	if sync { // wait for the start of a blow (the mode turns to an attack mode), at most 80 ticks
		pm, _ := nearest()
		for i := 0; i < 80; i++ {
			_ = l.Tick(ctx, 1)
			m, _ := nearest()
			if strings.HasPrefix(m, "A") && !strings.HasPrefix(pm, "A") {
				ci.Synced = true
				break
			}
			pm = m
		}
	}
	if sync && delay > 0 {
		_ = l.Tick(ctx, delay)
	}
	ci.Delay = delay
	_, d0 := nearest()
	ci.Adjacent, _, ci.LifeStart, _ = l.d11Within(ctx, 8)
	ci.Dist = d0
	oldID, oldPos := map[int]bool{}, map[[2]int]bool{} // an earlier portal can stand on the same cell or reuse the id
	if pv, err := l.PView(ctx); err == nil {
		for k, id := range pv.PortalIDs {
			oldID[id] = true
			oldPos[pv.PortalPos[k]] = true
		}
	}
	t0, ms0 := l.tick(), l.gameMS()
	if err := c.RClick(ctx, x, y); err != nil {
		ci.Err = "rclick: " + err.Error()
		return
	}
	ci.Rights = 1
	last, prevMode, prevMon := ci.LifeStart, -99, ""
	var ev []string
	add := func(i int, f string, a ...any) { ev = append(ev, fmt.Sprintf("+%d "+f, append([]any{i}, a...)...)) }
	for i := 1; i <= 60; i++ {
		_ = l.Tick(ctx, 1)
		sn, err := l.be.Env.Snapshot()
		if err != nil || sn.Player == nil {
			continue
		}
		if m, _ := nearest(); m != prevMon {
			add(i, "monster %s", m)
			prevMon = m
		}
		if sn.Player.Life < last {
			ci.Hits++
			ci.HitTicks = append(ci.HitTicks, i)
			add(i, "HIT life %d->%d", last, sn.Player.Life)
		}
		last = sn.Player.Life
		if sn.Player.Mode != prevMode {
			add(i, "player mode %d", sn.Player.Mode)
			prevMode = sn.Player.Mode
		}
		pv, err := l.PView(ctx)
		if err != nil {
			continue
		}
		if ci.ScrollTick < 0 && pv.Scrolls < ci.ScrollsBefore {
			ci.ScrollTick = i
			add(i, "scroll left the pack")
		}
		if ci.PortalTick < 0 {
			for k, id := range pv.PortalIDs {
				if !oldID[id] || !oldPos[pv.PortalPos[k]] {
					ci.PortalTick, ci.PortalMS, ci.PortalStep = l.tick()-t0, l.gameMS()-ms0, i
					add(i, "PORTAL object")
				}
			}
		}
		if ci.PortalTick >= 0 && i > 30 {
			break
		}
	}
	ci.Timeline = strings.Join(ev, "; ")
	ci.LifeEnd = last
	pv, _ := l.PView(ctx)
	ci.ScrollsAfter = pv.Scrolls
	return
}

// EmptyPlan runs Fighter.Clear in 6 s rounds with the mana set to 0 before each round, counting the out-of-mana
// back-offs and the failed walks in the log.
func (l *lab) EmptyPlan(ctx context.Context, r *drill.Run, level, n int, total time.Duration) (ei drill.EmptyPlanInfo) {
	ei.Where = fmt.Sprintf("area %d, %d zombies", level, n)
	if err := l.d11Setup(ctx, r, level, 1000, 8); err != nil {
		ei.Err = err.Error()
		return
	}
	defer l.d11Leave(ctx, r)
	if err := l.GiveSkills(ctx, r); err != nil {
		ei.Err = "skills: " + err.Error()
		return
	}
	if _, err := l.rig.BindSkills(); err != nil {
		ei.Err = "bind: " + err.Error()
		return
	}
	if _, err := r.Cheat(ctx, "debug belt clear", "no potion in the belt: the out-of-mana fight falls through to melee, not to a potion"); err != nil {
		ei.Err = err.Error()
		return
	}
	if where, err := l.d11WallBack(ctx, r); err != nil {
		ei.Err = "wall: " + err.Error()
		return
	} else {
		ei.Where += ", " + where
	}
	if _, err := r.Cheat(ctx, fmt.Sprintf("debug spawn zombie1 %d near", n), "a pack of more than 8 near the throwaway (not a weak pack, so the out-of-mana ladder reaches the retreat step): the fight under test (docs/forkcheats.md)"); err != nil {
		ei.Err = err.Error()
		return
	}
	_ = l.Tick(ctx, 5)
	ei.Spawned, _ = l.monstersAlive(ctx)
	mark := l.rec.Mark()
	for ei.Elapsed < total {
		if a, _, _, _ := l.d11Adjacent(ctx); a < 0 {
			break
		}
		if alive, _ := l.monstersAlive(ctx); alive == 0 {
			break
		}
		if err := l.SetMana(ctx, r, 0); err != nil {
			ei.Err = err.Error()
			break
		}
		if err := l.SetLife(ctx, r, 1000); err != nil {
			ei.Err = err.Error()
			break
		}
		res, err := l.rig.Fighter.Clear(ctx, combat.ClearOpts{Timeout: 6 * time.Second})
		ei.Rounds++
		if res.Elapsed == 0 {
			_ = l.Tick(ctx, 10) // a Clear that returned at once must not spin: lockstep time moves only on ticks
			ei.Elapsed += time.Second
		}
		if ei.Rounds > 400 {
			break
		}
		ei.Kills += res.Kills
		ei.Elapsed += res.Elapsed
		if err != nil {
			ei.Err = err.Error()
			break
		}
		l.rig.Fighter.ForgetIgnored()
	}
	for _, ln := range l.rec.Since(mark) {
		switch {
		case strings.Contains(ln, "out of mana: backing off"):
			ei.Backoffs++
			if len(ei.Samples) < 2 {
				ei.Samples = append(ei.Samples, ln)
			}
		case strings.Contains(ln, "out of mana: waiting"):
			ei.Waits++
		case strings.Contains(ln, "out of mana: back-off cap"):
			ei.CapMelee++
		case strings.Contains(ln, "combat move: already there"):
			ei.AlreadyThere++
		case strings.Contains(ln, "move failed"):
			ei.MoveFailed++
			if strings.Contains(ln, "empty path") {
				ei.EmptyPath++
			}
			if strings.Contains(ln, "level changed during the walk") {
				ei.LevelChanged++
			}
			if len(ei.Samples) < 5 {
				ei.Samples = append(ei.Samples, ln)
			}
		}
	}
	return
}

// d11WallBack puts the character (logged `debug warp`) on the nearest cell with a wall 3 subtiles behind it and 15 free
// subtiles in front, so a back-off of 12 subtiles away from a pack in front aims into the wall (the goal then snaps).
// The grid rows are level-relative subtiles, row = y, column = x (docs/d2d-backend.md); warp takes whole cells.
func (l *lab) d11WallBack(ctx context.Context, r *drill.Run) (string, error) {
	rows, err := l.be.Client.Grid(ctx, 5)
	if err != nil || len(rows) == 0 {
		return "", fmt.Errorf("no grid: %v", err)
	}
	lv, err := l.be.Client.Level(ctx)
	if err != nil {
		return "", err
	}
	at := func(x, y int) byte {
		if y < 0 || y >= len(rows) || x < 0 || x >= len(rows[y]) {
			return '#'
		}
		return rows[y][x]
	}
	px, py := int(5*lv.X), int(5*lv.Y)
	best, bd := [3]int{}, 1<<30
	dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for cy := py/5 - 12; cy <= py/5+12; cy++ {
		for cx := px/5 - 12; cx <= px/5+12; cx++ {
			x, y := 5*cx+2, 5*cy+2
			if at(x, y) != '.' {
				continue
			}
			for di, d := range dirs {
				ok := at(x-3*d[0], y-3*d[1]) == '#' && at(x-d[0], y-d[1]) == '.'
				for k := 1; ok && k <= 15; k++ {
					ok = at(x+k*d[0], y+k*d[1]) == '.'
				}
				if ok {
					if dd := (x-px)*(x-px) + (y-py)*(y-py); dd < bd {
						bd, best = dd, [3]int{cx, cy, di}
					}
				}
			}
		}
	}
	if bd == 1<<30 {
		return "", fmt.Errorf("no wall-backed cell within 12 cells of (%d,%d)", px, py)
	}
	if err := l.Warp(ctx, r, best[0], best[1]); err != nil {
		return "", err
	}
	_ = l.Tick(ctx, 10)
	d := dirs[best[2]]
	return fmt.Sprintf("warped to cell (%d,%d), wall 3 subtiles behind toward (%d,%d), open ahead", best[0], best[1], -d[0], -d[1]), nil
}

func (l *lab) gameMS() int {
	c, err := l.be.Client.Clock(context.Background())
	if err != nil {
		return -1
	}
	return c.GameMS
}

// d11Spot warps the character (logged `debug warp`) to the trial-th open spot of the level (a cell with 12 free
// subtiles all round, at least 8 cells from the arrival and from every earlier spot of this level), so an earlier
// trial's portal never stands at the same cell as this trial's (a new portal replaces the owner's old one, same id).
func (l *lab) d11Spot(ctx context.Context, r *drill.Run, trial int) error {
	rows, err := l.be.Client.Grid(ctx, 5)
	if err != nil || len(rows) == 0 {
		return fmt.Errorf("no grid: %v", err)
	}
	at := func(x, y int) bool { return y >= 0 && y < len(rows) && x >= 0 && x < len(rows[y]) && rows[y][x] == '.' }
	lv, err := l.be.Client.Level(ctx)
	if err != nil {
		return err
	}
	open := func(cx, cy int) bool {
		for dy := -12; dy <= 12; dy += 3 {
			for dx := -12; dx <= 12; dx += 3 {
				if !at(5*cx+2+dx, 5*cy+2+dy) {
					return false
				}
			}
		}
		return true
	}
	var used [][2]int
	used = append(used, [2]int{int(lv.X), int(lv.Y)})
	for t := 0; t <= trial; t++ {
		found := false
		for cy := 0; cy < len(rows)/5 && !found; cy++ {
			for cx := 0; cx < len(rows[0])/5 && !found; cx++ {
				ok := open(cx, cy)
				for _, u := range used {
					if ok && math.Hypot(float64(cx-u[0]), float64(cy-u[1])) < 8 {
						ok = false
					}
				}
				if ok {
					used = append(used, [2]int{cx, cy})
					found = true
				}
			}
		}
		if !found {
			return fmt.Errorf("no open spot number %d", t+1)
		}
	}
	u := used[len(used)-1]
	if err := l.Warp(ctx, r, u[0], u[1]); err != nil {
		return err
	}
	return l.Tick(ctx, 10)
}
