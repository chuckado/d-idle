package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/drill"
)

// binName is the d2d binary of this run (set in main), named in every DRILLS26 result.
var binName string

const fireBolt = 36

// ArmorFight: Frozen Armor level 1 and a stronger Fire Bolt (logged `debug points`), Blood Moor, a zombie pack that is
// topped up whenever fewer than 3 are alive, the bot's Fighter.Buff once, then Fighter.Clear rounds for simMS game ms.
func (l *lab) ArmorFight(ctx context.Context, r *drill.Run, simMS int) (o drill.ArmorFight) {
	o.SimMS, o.Binary = simMS, filepath.Base(binName)
	if err := l.d11Setup(ctx, r, 2, 200, 6); err != nil {
		o.Err = "setup: " + err.Error()
		return
	}
	defer l.d11Leave(ctx, r)
	if err := l.froLearn(ctx, r, 1); err != nil {
		o.Err = err.Error()
		return
	}
	if _, err := r.Cheat(ctx, fmt.Sprintf("debug points %d 8", fireBolt), "D26-A: a Fire Bolt that kills a zombie pack in reasonable time (skill levels, throwaway)"); err != nil {
		o.Err = err.Error()
		return
	}
	_ = l.Tick(ctx, 10)
	if n, err := l.rig.BindSkills(); err != nil || n == 0 {
		o.Err = fmt.Sprintf("BindSkills: %d, %v", n, err)
		return
	}
	l.froClear(ctx, r)
	sd := l.skillDebug(r)
	if _, err := sd.ManaFull(ctx, "D26-A: full mana for the opening cast"); err != nil {
		o.Err = err.Error()
		return
	}
	spawn := func() {
		if _, err := r.Cheat(ctx, "debug spawn "+d11Mon+" 4 near", "D26-A: top up the plain melee pack under test (docs/forkcheats.md)"); err == nil {
			o.Spawned += 4
		}
	}
	spawn()
	_ = l.Tick(ctx, 20)
	isCast := func(ln string) bool { return strings.HasPrefix(ln, "combat buff skill=Frozen Armor") }
	g0, _, err := l.gameClock(ctx)
	if err != nil {
		o.Err = err.Error()
		return
	}
	mk := l.rec.Mark()
	if err := l.rig.Fighter.Buff(ctx); err != nil {
		o.Err = "first Buff: " + err.Error()
		return
	}
	for _, ln := range l.rec.Since(mk) {
		if isCast(ln) {
			o.CastsMS = append(o.CastsMS, 0)
			o.Lines = append(o.Lines, "opening: "+ln)
		}
	}
	o.MinLife = 1 << 30
	prev, first, stuck := 0, false, 0
	for {
		g, _, err := l.gameClock(ctx)
		if err != nil {
			o.Err = err.Error()
			return
		}
		el := g - g0
		if el >= simMS {
			break
		}
		b, _ := l.buffsText(ctx)
		o.Samples++
		o.MaxGapMS = max(o.MaxGapMS, el-prev)
		if froOn(b) {
			first = true
			o.UpSamples++
			o.UpMS += el - prev
		} else if first {
			o.DownMS += el - prev
		}
		prev = el
		sn, err := l.be.Env.Snapshot()
		if err == nil && sn.Player != nil {
			if combat.IsDead(sn.Player) {
				o.Deaths++
				if err := l.ensureAlive(ctx, r); err != nil {
					o.Err = err.Error()
					return
				}
			} else {
				o.MinLife = min(o.MinLife, sn.Player.Life)
				o.MaxLife = max(o.MaxLife, sn.Player.MaxLife)
				if sn.Player.Mana < 10 {
					_, _ = sd.ManaFull(ctx, "D26-A: keep casts affordable in a long fight")
				}
				if sn.Player.MaxLife > 0 && sn.Player.Life*4 < sn.Player.MaxLife {
					if err := l.SetLife(ctx, r, sn.Player.MaxLife); err == nil {
						o.Refills++
					}
				}
			}
		}
		if alive, _ := l.monstersAlive(ctx); alive < 3 {
			spawn()
			_ = l.Tick(ctx, 10)
		}
		mk := l.rec.Mark()
		res, cerr := l.rig.Fighter.Clear(ctx, combat.ClearOpts{Timeout: 6 * time.Second})
		o.Rounds++
		o.Kills += res.Kills
		o.Potions += res.Potions
		if res.Elapsed == 0 || cerr != nil {
			if stuck++; stuck > 50 {
				o.Err = fmt.Sprintf("Clear returns at once (%v) 50 times", cerr)
				return
			}
			_ = l.Tick(ctx, 10)
		} else {
			stuck = 0
		}
		gg, _, _ := l.gameClock(ctx)
		for _, ln := range l.rec.Since(mk) {
			if isCast(ln) {
				o.CastsMS = append(o.CastsMS, gg-g0)
				o.FightCastMS = append(o.FightCastMS, gg-g0)
				o.Lines = append(o.Lines, fmt.Sprintf("t=%d ms: %s", gg-g0, ln))
			}
		}
		l.rig.Fighter.ForgetIgnored()
	}
	return
}
