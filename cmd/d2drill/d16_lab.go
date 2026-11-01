package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/drill"
)

// DRILLS16 (card drills16): live proof of fixbuff. Verbs: the bare read-only `debug states`, the ledgered SkillDebug
// verbs (states set/clear, mana full, mon set) and `debug points` / `debug spawn` / `debug kill` through the logged cheats.

func (l *lab) buffsRaw(ctx context.Context) string {
	rep, err := l.be.Client.Do(ctx, "debug states")
	if err != nil {
		return "ERR " + err.Error()
	}
	return strings.TrimSpace(rep.Text())
}

func buffsStr(bs []d2dctl.Buff) string {
	var s []string
	for _, b := range bs {
		s = append(s, fmt.Sprintf("%d:%d", b.Skill, b.Level))
	}
	return "[" + strings.Join(s, " ") + "]"
}

// stateBits lists the state numbers 0..127 set in the snapshot's Player.States.
func (l *lab) stateBits() (string, bool) {
	sn, err := l.be.Env.Snapshot()
	if err != nil || sn.Player == nil {
		return "snapshot error", false
	}
	var b []string
	for i := 0; i < 128; i++ {
		if sn.Player.States.Has(i) {
			b = append(b, fmt.Sprint(i))
		}
	}
	return "{" + strings.Join(b, ",") + "}", sn.Player.States.Has(10)
}

func (l *lab) BuffRaw(ctx context.Context, r *drill.Run) (o drill.BuffRaw) {
	if err := l.froLearn(ctx, r, 1); err != nil {
		o.Err = err.Error()
		return
	}
	l.froClear(ctx, r)
	sd := l.skillDebug(r)
	if _, err := sd.ManaFull(ctx, "fixbuff proof: known mana for the Frozen Armor cast"); err != nil {
		o.Err = err.Error()
		return
	}
	_ = l.Tick(ctx, 5)
	o.Before = l.buffsRaw(ctx)
	x, y, _, err := l.playerCell(ctx)
	if err != nil {
		o.Err = err.Error()
		return
	}
	if err := l.be.Client.CastCell(ctx, frostArmor, x, y); err != nil {
		o.Err = "cast: " + err.Error()
		return
	}
	_ = l.Tick(ctx, 10)
	o.After = l.buffsRaw(ctx)
	bs, err := l.be.Client.Buffs(ctx)
	o.Parsed = buffsStr(bs)
	for _, b := range bs {
		if b.Skill == frostArmor && b.Level == 1 {
			o.ParsedOK = err == nil
		}
	}
	for i := 0; i < 12; i++ { // several samples spread over 30 s of game time
		if _, bit := l.stateBits(); true {
			o.Samples++
			if bit {
				o.Bit10++
			}
		}
		_ = l.Tick(ctx, 60)
	}
	if _, err := sd.StatesClear(ctx, frostArmor, "fixbuff proof: remove the armor, bit 10 must drop"); err != nil {
		o.Err = err.Error()
		return
	}
	_ = l.Tick(ctx, 5)
	o.AfterClear = l.buffsRaw(ctx)
	for i := 0; i < 5; i++ {
		o.ClearN++
		if _, bit := l.stateBits(); bit {
			o.ClearBit10++
		}
		_ = l.Tick(ctx, 10)
	}
	return
}

func ip(v int) *int { return &v }

func (l *lab) LongFight(ctx context.Context, r *drill.Run, simMS int) (o drill.LongFight) {
	o.SimMS = simMS
	if err := l.d11Setup(ctx, r, 2, 100, 6); err != nil {
		o.Err = "setup: " + err.Error()
		return
	}
	defer l.d11Leave(ctx, r)
	if err := l.froLearn(ctx, r, 1); err != nil {
		o.Err = err.Error()
		return
	}
	if n, err := l.rig.BindSkills(); err != nil || n == 0 {
		o.Err = fmt.Sprintf("BindSkills: %d, %v", n, err)
		return
	}
	l.froClear(ctx, r)
	sd := l.skillDebug(r)
	if _, err := sd.ManaFull(ctx, "fixbuff proof: full mana for the first cast"); err != nil {
		o.Err = err.Error()
		return
	}
	if got := l.d11Spawn(ctx, r, 1, 1, 8); got < 0 {
		o.Err = "spawn failed"
		return
	}
	ms, err := l.be.Client.Monsters(ctx)
	unit := -1
	for _, m := range ms.Items {
		if err == nil && m.HP.Cur > 0 && m.Mode != "DD" && m.Mode != "DT" {
			unit = m.Unit
			break
		}
	}
	if unit < 0 {
		o.Err = "no live monster to hold"
		return
	}
	if rep, err := sd.MonSet(ctx, unit, d2dctl.MonEdit{Life: ip(1000000), MaxLife: ip(1000000), Dmg: &[2]int{0, 0}}, "fixbuff proof: a monster that cannot die and deals no damage keeps the bot in one fight past 120 s"); err != nil {
		o.Err = "mon set: " + err.Error()
		return
	} else {
		o.MonLife = rep
	}
	isCast := func(ln string) bool { return strings.HasPrefix(ln, "combat buff skill=Frozen Armor") }
	mark := l.rec.Mark()
	g0, _, err := l.gameClock(ctx)
	if err != nil {
		o.Err = err.Error()
		return
	}
	if err := l.rig.Fighter.Buff(ctx); err != nil {
		o.Err = "first Buff: " + err.Error()
		return
	}
	for _, ln := range l.rec.Since(mark) {
		if isCast(ln) {
			o.CastsMS = append(o.CastsMS, 0)
			o.Lines = append(o.Lines, "manual: "+ln)
		}
	}
	if len(o.CastsMS) == 0 {
		o.Err = "the first Buff made no cast"
		return
	}
	prev, last, first := 0, 0, false
	for stuck := 0; ; {
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
		up := froOn(b)
		o.Samples++
		o.MaxGapMS = max(o.MaxGapMS, el-prev)
		if up {
			first = true
			if _, bit := l.stateBits(); bit {
				o.Bit10Up++
			} else {
				o.Bit10Miss++
			}
		} else if first {
			o.Down++
			o.DownMS += el - prev
		}
		prev = el
		if sn, err := l.be.Env.Snapshot(); err == nil && sn.Player != nil && sn.Player.Mana < 10 {
			_, _ = sd.ManaFull(ctx, "fixbuff proof: the fight is long, keep the casts affordable")
		}
		to := 6 * time.Second
		if d := el - last; d > 100000 && d < 130000 {
			to = time.Second // fine sampling around the recast and the real end
		}
		mk := l.rec.Mark()
		res, cerr := l.rig.Fighter.Clear(ctx, combat.ClearOpts{Timeout: to})
		o.Rounds++
		if res.Elapsed == 0 || cerr != nil {
			if stuck++; stuck > 50 {
				o.Err = fmt.Sprintf("Clear returns at once (%v) 50 times: the bot is not in the fight", cerr)
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
				last = gg - g0
				o.Lines = append(o.Lines, fmt.Sprintf("t=%d ms: %s", gg-g0, ln))
			}
		}
	}
	return
}

func (l *lab) ArmorIDs(ctx context.Context, r *drill.Run) (rows []drill.ArmorID) {
	sd := l.skillDebug(r)
	l.froClear(ctx, r)
	_, _ = sd.StatesClear(ctx, -1, "fixbuff proof: start the armor-id record with no states")
	_ = l.Tick(ctx, 5)
	read := func(a *drill.ArmorID) {
		a.Line = l.buffsRaw(ctx)
		bs, err := l.be.Client.Buffs(ctx)
		a.Parsed = buffsStr(bs)
		if err != nil {
			a.Err += " buffs: " + err.Error()
		}
		a.States, _ = l.stateBits()
	}
	for _, sk := range []int{40, 50, 58, 60, 68} {
		a := drill.ArmorID{Skill: sk, Via: "debug states set"}
		rep, err := sd.StatesSet(ctx, sk, 1, fmt.Sprintf("fixbuff proof: which `debug states` entry and state bit skill %d produces", sk))
		a.Reply = rep
		if err != nil {
			a.Err = err.Error()
		}
		_ = l.Tick(ctx, 5)
		read(&a)
		if _, err := sd.StatesClear(ctx, sk, "fixbuff proof: remove it again"); err == nil {
			a.CleanAfter = true
		}
		_ = l.Tick(ctx, 5)
		rows = append(rows, a)
	}
	// The same through a real cast: learn the skill (debug points sets the level), full mana, cmd skill.
	for _, sk := range []int{50, 58, 60, 68} {
		a := drill.ArmorID{Skill: sk, Via: "cmd skill after debug points"}
		if _, err := r.Cheat(ctx, fmt.Sprintf("debug points %d 1", sk), fmt.Sprintf("learn skill %d level 1 to cast it for the armor-id record", sk)); err != nil {
			a.Err = "points: " + err.Error()
			rows = append(rows, a)
			continue
		}
		_ = l.Tick(ctx, 10)
		_, _ = sd.ManaFull(ctx, "fixbuff proof: mana for the cast")
		_ = l.Tick(ctx, 3)
		x, y, _, _ := l.playerCell(ctx)
		if err := l.be.Client.CastCell(ctx, sk, x, y); err != nil {
			a.Err = "cast: " + err.Error()
		}
		_ = l.Tick(ctx, 20)
		read(&a)
		if _, err := sd.StatesClear(ctx, sk, "fixbuff proof: remove it again"); err == nil {
			a.CleanAfter = true
		}
		_ = l.Tick(ctx, 5)
		rows = append(rows, a)
	}
	_, _ = sd.StatesClear(ctx, -1, "fixbuff proof: leave no states on the throwaway")
	return
}
