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
func (l *lab) d11Setup(ctx context.Context, r *drill.Run, level, maxLife int) error {
	if err := l.Goto(ctx, r, level); err != nil {
		return fmt.Errorf("goto: %w", err)
	}
	if sw, err := l.be.Env.Snapshot(); err == nil && sw.Player != nil {
		wctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		_ = l.navr.GoTo(wctx, input.Point{X: sw.Player.X + 8, Y: sw.Player.Y})
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
	return hist(d11att) + "] [all living monsters " + hist(d11all) + "] [last read " + d11each
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
	defer func() {
		sort.Ints(ds)
		d11each = fmt.Sprint(ds)
	}()
	for _, m := range ms.Items {
		if m.HP.Cur <= 0 || m.Mode == "DD" || m.Mode == "DT" {
			continue
		}
		alive++
		d := math.Hypot(5*(m.CellX-lv.X), 5*(m.CellY-lv.Y))
		d11all[int(d+0.5)]++
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

// d11Squeeze warps the character next to the want nearest monsters (logged `debug warp` to the cell of their centroid)
// until at least want of them are within 3 subtiles, at most 8 tries: the attackers' natural distance is larger than the
// bot's adjacency (finding of DRILLS11), so the rule's branch is reached by placing the character.
func (l *lab) d11Squeeze(ctx context.Context, r *drill.Run, want int) (tries int) {
	for ; tries < 8; tries++ {
		if a, _, _, _ := l.d11Within(ctx, 3); a >= want {
			return tries
		}
		ms, err := l.be.Client.Monsters(ctx)
		lv, lerr := l.be.Client.Level(ctx)
		if err != nil || lerr != nil {
			return tries
		}
		type md struct{ x, y, d float64 }
		var near []md
		for _, m := range ms.Items {
			if m.HP.Cur > 0 && m.Mode != "DD" && m.Mode != "DT" {
				near = append(near, md{m.CellX, m.CellY, math.Hypot(m.CellX-lv.X, m.CellY-lv.Y)})
			}
		}
		sort.Slice(near, func(i, j int) bool { return near[i].d < near[j].d })
		if len(near) < want {
			return tries
		}
		var cx, cy float64
		for _, m := range near[:want] {
			cx += m.x
			cy += m.y
		}
		cx, cy = cx/float64(want), cy/float64(want)
		if err := l.Warp(ctx, r, int(cx+0.5)-tries%2, int(cy+0.5)); err != nil {
			return tries
		}
		_ = l.Tick(ctx, 1)
	}
	return tries
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
	if _, err := r.Cheat(ctx, fmt.Sprintf("debug spawn %s %d near", d11Mon, n), "melee monsters near the throwaway: the fight under test (docs/forkcheats.md)"); err != nil {
		return -1
	}
	for i := 0; i < 120; i++ {
		_ = l.Tick(ctx, 10)
		if a, _, _, _ := l.d11Within(ctx, radius); a >= want {
			return a
		}
	}
	a, _, _, _ := l.d11Within(ctx, radius)
	return a
}

func (l *lab) PortalMelee(ctx context.Context, r *drill.Run, adj int, act bool) (pi drill.PMInfo) {
	pi.Where = fmt.Sprintf("Blood Moor, %d zombie", adj)
	if err := l.d11Setup(ctx, r, 2, 300); err != nil {
		pi.Err = err.Error()
		return
	}
	defer l.d11Leave(ctx, r)
	if err := l.d11OneScroll(ctx, r); err != nil {
		pi.Err = "scroll: " + err.Error()
		return
	}
	if got := l.d11Spawn(ctx, r, adj, adj, 3); got < 0 {
		pi.Err = "spawn failed"
		return
	}
	if got, _, _, _ := l.d11Within(ctx, 3); got < adj {
		pi.Squeezed = l.d11Squeeze(ctx, r, adj)
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

func (l *lab) CastTrace(ctx context.Context, r *drill.Run, life int) (ci drill.CastInfo) {
	ci.PortalTick = -1
	if err := l.d11Setup(ctx, r, 2, life); err != nil {
		ci.Err = err.Error()
		return
	}
	defer l.d11Leave(ctx, r)
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
	ci.Adjacent, _, ci.LifeStart, _ = l.d11Within(ctx, 8)
	old := map[int]bool{} // portal objects by id: an earlier trial's portal can stand on the same cell
	if pv, err := l.PView(ctx); err == nil {
		for _, id := range pv.PortalIDs {
			old[id] = true
		}
	}
	t0, ms0 := l.tick(), l.gameMS()
	if err := c.RClick(ctx, x, y); err != nil {
		ci.Err = "rclick: " + err.Error()
		return
	}
	ci.Rights = 1
	last, run, prev := ci.LifeStart, 0, -1
	var seq, mseq []string
	mprev, mrun := "", 0
	flushM := func() {
		if mrun > 0 {
			mseq = append(mseq, fmt.Sprintf("%s x%d", mprev, mrun))
		}
	}
	flush := func() {
		if run > 0 {
			seq = append(seq, fmt.Sprintf("%d x%d", prev, run))
		}
	}
	for i := 1; i <= 70; i++ {
		_ = l.Tick(ctx, 1)
		sn, err := l.be.Env.Snapshot()
		if err != nil || sn.Player == nil {
			continue
		}
		if ms, e := l.be.Client.Monsters(ctx); e == nil { // the nearest living monster's mode, run length encoded
			best, bm := 1e9, "-"
			lv, _ := l.be.Client.Level(ctx)
			for _, m := range ms.Items {
				if m.HP.Cur > 0 && m.Mode != "DD" && m.Mode != "DT" {
					if d := math.Hypot(m.CellX-lv.X, m.CellY-lv.Y); d < best {
						best, bm = d, m.Mode
					}
				}
			}
			if bm != mprev {
				flushM()
				mprev, mrun = bm, 0
			}
			mrun++
		}
		if sn.Player.Life < last {
			ci.Hits++
		}
		last = sn.Player.Life
		if sn.Player.Mode != prev {
			flush()
			prev, run = sn.Player.Mode, 0
		}
		run++
		if ci.PortalTick < 0 {
			if pv, err := l.PView(ctx); err == nil {
				for _, id := range pv.PortalIDs {
					if !old[id] {
						ci.PortalTick = l.tick() - t0
						ci.PortalMS = l.gameMS() - ms0
					}
				}
			}
		}
		if ci.PortalTick >= 0 && i > 45 {
			break
		}
	}
	flush()
	flushM()
	ci.ModeSeq = strings.Join(seq, ", ")
	ci.AttackDists = d11AttStr()
	ci.MonSeq = strings.Join(mseq, ", ")
	ci.LifeEnd = last
	pv, _ := l.PView(ctx)
	ci.ScrollsAfter = pv.Scrolls
	return
}

// EmptyPlan runs Fighter.Clear in 6 s rounds with the mana set to 0 before each round, counting the out-of-mana
// back-offs and the failed walks in the log.
func (l *lab) EmptyPlan(ctx context.Context, r *drill.Run, level, n int, total time.Duration) (ei drill.EmptyPlanInfo) {
	ei.Where = fmt.Sprintf("area %d, %d zombies", level, n)
	if err := l.d11Setup(ctx, r, level, 1000); err != nil {
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
