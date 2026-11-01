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
	if _, err := r.Cheat(ctx, fmt.Sprintf("debug points %d 3", fireBolt), "D26-A: a Fire Bolt that kills a zombie pack in reasonable time (skill levels, throwaway)"); err != nil {
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
		o.Spawned += 4
		l.d11Spawn(ctx, r, 4, 2, 8) // spawns, then lures until two are in reach: a spawned monster idles until she is close
	}
	spawn()
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
	var lastSpawn bool
	var lastClear int
	var lastRep time.Duration
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
		if el-prev > 15000 {
			o.Lines = append(o.Lines, fmt.Sprintf("t=%d ms: sample gap %d ms (previous round: spawn %v, Clear %d ms, reported %v)", el, el-prev, lastSpawn, lastClear, lastRep))
		}
		if froOn(b) {
			first = true
			o.UpSamples++
			o.UpMS += el - prev
		} else if first {
			o.DownMS += el - prev
			o.Lines = append(o.Lines, fmt.Sprintf("t=%d ms: armor absent at this sample (previous sample %d ms)", el, prev))
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
		spawned := false
		if alive, _ := l.monstersAlive(ctx); alive == 0 {
			// the bot buffs between fights before it walks to the next pack: do the same before the lure (it checks the clock itself)
			bmk := l.rec.Mark()
			_ = l.rig.Fighter.Buff(ctx)
			for _, ln := range l.rec.Since(bmk) {
				if isCast(ln) {
					gb, _, _ := l.gameClock(ctx)
					o.CastsMS = append(o.CastsMS, gb-g0)
					o.FightCastMS = append(o.FightCastMS, gb-g0)
					bt, _ := l.buffsText(ctx)
					o.Lines = append(o.Lines, fmt.Sprintf("t=%d ms: %s (between packs: buffs=%q)", gb-g0, ln, bt))
				}
				if strings.HasPrefix(ln, "combat buff cut by a hit") {
					o.Lines = append(o.Lines, fmt.Sprintf("between packs: %s", ln))
				}
			}
			spawn()
			spawned = true
		}
		t0, _, _ := l.gameClock(ctx)
		mk := l.rec.Mark()
		res, cerr := l.rig.Fighter.Clear(ctx, combat.ClearOpts{Timeout: 2 * time.Second})
		o.Rounds++
		t1, _, _ := l.gameClock(ctx)
		lastSpawn, lastClear, lastRep = spawned, t1-t0, res.Elapsed
		o.Kills += res.Kills
		o.Potions += res.Potions
		if res.Elapsed == 0 || cerr != nil {
			if stuck++; stuck > 50 {
				o.Err = fmt.Sprintf("Clear returns at once (%v) 50 times", cerr)
				return
			}
			l.d11Lure(ctx) // nothing in reach: walk toward the nearest idle zombie
			_ = l.Tick(ctx, 20)
		} else {
			stuck = 0
		}
		gg, _, _ := l.gameClock(ctx)
		for _, ln := range l.rec.Since(mk) {
			if strings.HasPrefix(ln, "combat buff cut by a hit") {
				_, _, pm, _ := l.playerCell(ctx)
				o.Lines = append(o.Lines, fmt.Sprintf("t=%d ms: %s (mode now %d)", gg-g0, ln, pm))
			}
			if isCast(ln) {
				o.CastsMS = append(o.CastsMS, gg-g0)
				o.FightCastMS = append(o.FightCastMS, gg-g0)
				bt, _ := l.buffsText(ctx)
				mana := -1
				if sn, err := l.be.Env.Snapshot(); err == nil && sn.Player != nil {
					mana = sn.Player.Mana
				}
				_, _, pm, _ := l.playerCell(ctx)
				o.Lines = append(o.Lines, fmt.Sprintf("t=%d ms: %s (read at the end of the round: buffs=%q mana=%d mode=%d)", gg-g0, ln, bt, mana, pm))
			}
		}
		l.rig.Fighter.ForgetIgnored()
	}
	for i := 1; i < len(o.CastsMS); i++ { // the armor lasts 120.3 s from a cast (V, drills13): a longer gap between casts had it off
		o.GapBoundMS += max(0, o.CastsMS[i]-o.CastsMS[i-1]-120320)
	}
	return
}

// ArmorIdle (DRILLS28 D28-0): no monsters; cast Frozen Armor (the command the bot's key press sends) at 60 s and 114 s
// of game time and read `debug states` and the player mode at one-second steps from 114 s to 130 s.
func (l *lab) ArmorIdle(ctx context.Context, r *drill.Run) (o drill.ArmorIdle) {
	o.Binary = filepath.Base(binName)
	if err := l.d11Setup(ctx, r, 2, 200, 6); err != nil {
		o.Err = "setup: " + err.Error()
		return
	}
	defer l.d11Leave(ctx, r)
	if err := l.froLearn(ctx, r, 1); err != nil {
		o.Err = err.Error()
		return
	}
	l.froClear(ctx, r)
	sd := l.skillDebug(r)
	g0, _, err := l.gameClock(ctx)
	if err != nil {
		o.Err = err.Error()
		return
	}
	cast := func() bool {
		if _, err := sd.ManaFull(ctx, "D28-0: known mana for the cast"); err != nil {
			o.Err = err.Error()
			return false
		}
		x, y, _, err := l.playerCell(ctx)
		if err != nil {
			o.Err = err.Error()
			return false
		}
		g, _, _ := l.gameClock(ctx)
		if err := l.be.Client.CastCell(ctx, frostArmor, x, y); err != nil {
			o.Err = "cast: " + err.Error()
			return false
		}
		o.CastsMS = append(o.CastsMS, g-g0)
		return true
	}
	sample := func(note string) {
		g, _, _ := l.gameClock(ctx)
		el := g - g0
		b, _ := l.buffsText(ctx)
		_, _, pm, _ := l.playerCell(ctx)
		on := froOn(b)
		o.AtMS = append(o.AtMS, el)
		o.OnAt = append(o.OnAt, on)
		if el >= 121000 && el <= 125000 {
			o.Seen++
		}
		o.Timeline = append(o.Timeline, fmt.Sprintf("t=%.1f buffs=%q mode=%d %s", float64(el)/1000, b, pm, note))
	}
	o.Window = true
	stepTo := func(ms int, coarse int) bool {
		for {
			g, _, err := l.gameClock(ctx)
			if err != nil {
				o.Err = err.Error()
				return false
			}
			if g-g0 >= ms {
				return true
			}
			if err := l.Tick(ctx, coarse); err != nil {
				o.Err = err.Error()
				return false
			}
		}
	}
	if !stepTo(60000, 25) || !cast() {
		return
	}
	for i := 0; i < 3; i++ {
		_ = l.Tick(ctx, 13)
		sample("after first cast")
	}
	if !stepTo(113000, 25) {
		return
	}
	sample("before second cast")
	if !stepTo(114000, 5) || !cast() {
		return
	}
	for {
		_ = l.Tick(ctx, 25) // about 1 s
		sample("")
		g, _, _ := l.gameClock(ctx)
		el := g - g0
		if el >= 121000 && el <= 125000 && !o.OnAt[len(o.OnAt)-1] {
			o.Window = false
		}
		if el >= 131000 {
			break
		}
	}
	return
}
