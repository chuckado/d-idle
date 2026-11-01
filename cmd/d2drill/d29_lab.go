package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/drill"
)

// holdLeadMS is how long before the bot's due time (114 s after a cast at level 1: armor duration minus the data-driven
// margin) the pack is brought in; the hold lasts until the recast is done and a few seconds longer.
// holdMon and holdPack: the monster of the pack (D29_MON overrides it: a fast attacker lands more hits than the plain
// zombie, which hit the character about once in 20 s in the first runs) and the number spawned per top-up.
var (
	holdMon  = "fallen1"
	holdPack = 8
)

const (
	holdLeadMS = 22000
	holdDueMS  = 114000
	holdPostMS = 6000
)

// ArmorHold (D29-C): the D26-A setup, then no monsters until the armor is about to be due; a pack of zombies is
// spawned and kept adjacent (no Clear rounds, so nothing is killed) and the bot's own Fighter.Buff is called in a tight
// loop. A hit during a cast makes the bot log `buff cut by a hit` and retry at once. Outside the hold windows the
// pack is killed with Clear rounds and the character idles.
func (l *lab) ArmorHold(ctx context.Context, r *drill.Run, simMS int) (o drill.ArmorHold) {
	o.SimMS, o.Binary = simMS, filepath.Base(binName)
	if m := os.Getenv("D29_MON"); m != "" {
		holdMon = m
	}
	save := d11Mon
	d11Mon = holdMon
	defer func() { d11Mon = save }()
	if err := l.d11Setup(ctx, r, 2, 200, 6); err != nil {
		o.Err = "setup: " + err.Error()
		return
	}
	defer l.d11Leave(ctx, r)
	if err := l.froLearn(ctx, r, 1); err != nil {
		o.Err = err.Error()
		return
	}
	if _, err := r.Cheat(ctx, fmt.Sprintf("debug points %d 3", fireBolt), "D29-C: a Fire Bolt that kills the pack between hold windows (skill levels, throwaway)"); err != nil {
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
	if _, err := sd.ManaFull(ctx, "D29-C: full mana for the opening cast"); err != nil {
		o.Err = err.Error()
		return
	}
	g0, _, err := l.gameClock(ctx)
	if err != nil {
		o.Err = err.Error()
		return
	}
	now := func() int { g, _, _ := l.gameClock(ctx); return g - g0 }
	isCast := func(ln string) bool { return strings.HasPrefix(ln, "combat buff skill=Frozen Armor") }
	mode := func() (int, int) {
		sn, err := l.be.Env.Snapshot()
		if err != nil || sn.Player == nil {
			return -1, 0
		}
		return sn.Player.Mode, sn.Player.Life
	}
	mk := l.rec.Mark()
	if err := l.rig.Fighter.Buff(ctx); err != nil {
		o.Err = "first Buff: " + err.Error()
		return
	}
	lastCast := 0
	for _, ln := range l.rec.Since(mk) {
		if isCast(ln) {
			o.Casts++
			o.Lines = append(o.Lines, "opening: "+ln)
		}
	}
	o.MinLife = 1 << 30
	o.Modes = map[int]int{}
	prev, first, downRun := 0, false, 0
	sample := func() {
		el := now()
		b, _ := l.buffsText(ctx)
		o.Samples++
		o.PollMS = max(o.PollMS, el-prev)
		if froOn(b) {
			first, downRun = true, 0
		} else if first {
			o.DownMS += el - prev
			downRun += el - prev
			o.MaxDownMS = max(o.MaxDownMS, downRun)
			o.Lines = append(o.Lines, fmt.Sprintf("t=%d ms: armor absent at this sample (previous sample %d ms)", el, prev))
		}
		prev = el
		m, life := mode()
		o.ModeReads++
		o.Modes[m]++
		if m == 4 {
			o.Hits++
		}
		if life > 0 {
			o.MinLife = min(o.MinLife, life)
		}
	}
	pendingCut := false
	holdUntil := 0
	// holdBuff is one call of the bot's Buff pass with everything recorded; it also runs inside the spawn and lure loop,
	// where 20 to 40 s can pass (the bot buffs between its own steps, so the hold must not starve the pass).
	holdBuff := func() {
		adj, _, _, _ := l.d11Within(ctx, 8)
		mi, _ := mode()
		at := now()
		bmk := l.rec.Mark()
		_ = l.rig.Fighter.Buff(ctx)
		end := now()
		mo, life := mode()
		rc := drill.ArmorRecast{AtMS: at, EndMS: end, ModeIn: mi, ModeOut: mo, Adjacent: adj, Life: life}
		for _, ln := range l.rec.Since(bmk) {
			if isCast(ln) {
				rc.Cast = true
				o.Casts++
				if pendingCut {
					o.CutRetries++
					pendingCut = false
				}
				o.Lines = append(o.Lines, fmt.Sprintf("t=%d ms: %s", at, ln))
			}
			if strings.HasPrefix(ln, "combat buff cut by a hit") {
				rc.Cut = true
				o.Cuts++
				pendingCut = true
				o.Lines = append(o.Lines, fmt.Sprintf("t=%d ms: %s (mode %d>%d)", at, ln, mi, mo))
			}
		}
		b, _ := l.buffsText(ctx)
		rc.StateOn = froOn(b)
		if rc.Cast || rc.Cut {
			o.Recasts = append(o.Recasts, rc)
		}
		switch {
		case rc.Cast && !rc.Cut:
			lastCast, holdUntil = end, end+holdPostMS // a clean cast ends the hold shortly after
		case rc.Cut:
			holdUntil = end + 10000 // keep the pack on her until the retry has been seen
		}
	}
	spawnHook = func(context.Context) {
		if el := now(); el >= lastCast+holdDueMS-holdLeadMS-2000 {
			holdBuff()
			sample()
		}
	}
	defer func() { spawnHook = nil }()
	for now() < simMS {
		sample()
		el := now()
		sn, err := l.be.Env.Snapshot()
		if err == nil && sn.Player != nil {
			if combat.IsDead(sn.Player) {
				o.Deaths++
				if err := l.ensureAlive(ctx, r); err != nil {
					o.Err = err.Error()
					return
				}
			} else {
				if sn.Player.Mana < 10 {
					_, _ = sd.ManaFull(ctx, "D29-C: keep casts affordable")
				}
				if sn.Player.MaxLife > 0 && sn.Player.Life*100 < 35*sn.Player.MaxLife {
					if err := l.SetLife(ctx, r, sn.Player.MaxLife); err == nil {
						o.Refills++
					}
				}
			}
		}
		holding := el >= lastCast+holdDueMS-holdLeadMS || el < holdUntil
		if !holding {
			if alive, _ := l.monstersAlive(ctx); alive > 0 {
				_, _ = l.rig.Fighter.Clear(ctx, combat.ClearOpts{Timeout: 2 * time.Second})
				l.rig.Fighter.ForgetIgnored()
			} else {
				_ = l.Tick(ctx, 25)
			}
			continue
		}
		alive, _ := l.monstersAlive(ctx)
		if alive < 3 {
			o.Spawned += holdPack
			l.d11SpawnHold(ctx, r)
		}
		holdBuff()
		_ = l.Tick(ctx, 5)
	}
	sample()
	return
}

// d11SpawnHold spawns a pack next to the character without waiting for a lure: the monsters come to her in a few ticks
// (a spawned monster idles until she is close, so this also walks toward the nearest one while it waits).
func (l *lab) d11SpawnHold(ctx context.Context, r *drill.Run) {
	l.d11Spawn(ctx, r, holdPack, 3, 8)
}
