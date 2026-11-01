package main

import (
	"context"
	"fmt"
	"math"
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
	for _, m := range ms.Items {
		if m.HP.Cur <= 0 || m.Mode == "DD" || m.Mode == "DT" {
			continue
		}
		alive++
		if math.Hypot(5*(m.CellX-lv.X), 5*(m.CellY-lv.Y)) <= 3 {
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

// d11Mon is the melee monster of the portal drills: a Fallen is fast (a Zombie did not reach the character in 200 ticks, docs/forkcheats.md).
const d11Mon = "fallen1"

// d11Spawn spawns n zombies and steps until at least want of them are adjacent (600 ticks at most).
func (l *lab) d11Spawn(ctx context.Context, r *drill.Run, n, want int) (adj int) {
	if _, err := r.Cheat(ctx, fmt.Sprintf("debug spawn %s %d near", d11Mon, n), "melee monsters near the throwaway: the fight under test (docs/forkcheats.md)"); err != nil {
		return -1
	}
	for i := 0; i < 60; i++ {
		_ = l.Tick(ctx, 10)
		if a, _, _, _ := l.d11Adjacent(ctx); a >= want {
			return a
		}
	}
	a, _, _, _ := l.d11Adjacent(ctx)
	return a
}

func (l *lab) PortalMelee(ctx context.Context, r *drill.Run, adj int, act bool) (pi drill.PMInfo) {
	pi.Where = fmt.Sprintf("Blood Moor, %d fallen", adj)
	if err := l.d11Setup(ctx, r, 2, 100); err != nil {
		pi.Err = err.Error()
		return
	}
	defer l.d11Leave(ctx, r)
	if err := l.d11OneScroll(ctx, r); err != nil {
		pi.Err = "scroll: " + err.Error()
		return
	}
	if got := l.d11Spawn(ctx, r, adj, adj); got < 0 {
		pi.Err = "spawn failed"
		return
	}
	if err := l.SetLife(ctx, r, 34); err != nil {
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
	if got := l.d11Spawn(ctx, r, 1, 1); got < 0 {
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
	ci.Adjacent, _, ci.LifeStart, _ = l.d11Adjacent(ctx)
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
	var seq []string
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
	ci.ModeSeq = strings.Join(seq, ", ")
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
