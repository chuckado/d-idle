package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/drill"
)

// DRILLS13 (card drills13): the Frozen Armor recast proof (card fixrecast) and the mana timing trace M1
// (docs/research/sk-mana-trace.md). Skill-drill verbs go through d2dctl.SkillDebug, which ledgers them itself.

const frostArmor = 40

func (l *lab) skillDebug(r *drill.Run) *d2dctl.SkillDebug {
	return &d2dctl.SkillDebug{C: l.be.Client, Ledger: l.led, Actor: "drills:" + r.Drill, Instance: l.inst, Character: l.char}
}

func (l *lab) gameClock(ctx context.Context) (ms, tick int, err error) {
	ci, err := l.be.Client.Clock(ctx)
	return ci.GameMS, ci.Tick, err
}

// buffsText is the read-only `debug states` listing's buffs= field ("40:1," while Frozen Armor level 1 is on).
func (l *lab) buffsText(ctx context.Context) (string, error) {
	rep, err := l.be.Client.Do(ctx, "debug states")
	if err != nil {
		return "", err
	}
	t := rep.Text()
	i := strings.Index(t, "buffs=")
	if i < 0 {
		return "", fmt.Errorf("no buffs= in %q", t)
	}
	t = t[i+len("buffs="):]
	if j := strings.IndexAny(t, " \n"); j >= 0 {
		t = t[:j]
	}
	return t, nil
}

func froOn(buffs string) bool {
	for _, b := range strings.Split(buffs, ",") {
		if strings.HasPrefix(b, fmt.Sprint(frostArmor, ":")) {
			return true
		}
	}
	return false
}

var rawRe = regexp.MustCompile(`"raw":\s*(\[[^\]]*\])`)

// rawStats is `dump stats` raw[]: raw[6] life, raw[8] mana, in 256ths.
func (l *lab) rawStats(ctx context.Context) ([]int, error) {
	rep, err := l.be.Client.Do(ctx, "dump stats")
	if err != nil {
		return nil, err
	}
	m := rawRe.FindStringSubmatch(rep.Text())
	if m == nil {
		return nil, fmt.Errorf("no raw[] in dump stats")
	}
	var v []int
	if err := json.Unmarshal([]byte(m[1]), &v); err != nil || len(v) < 9 {
		return nil, fmt.Errorf("raw[]: %v (%d values)", err, len(v))
	}
	return v, nil
}

func (l *lab) playerCell(ctx context.Context) (x, y float64, pmode int, err error) {
	ps, err := l.be.Client.Players(ctx)
	if err != nil || len(ps.Items) == 0 {
		return 0, 0, 0, fmt.Errorf("players: %v", err)
	}
	return ps.Items[0].X, ps.Items[0].Y, ps.Items[0].PMode, nil
}

// froLearn gives Frozen Armor points until its base level is want.
func (l *lab) froLearn(ctx context.Context, r *drill.Run, want int) error {
	for i := 0; i < 5; i++ {
		sn, err := l.be.Env.Snapshot()
		if err != nil || sn.Player == nil {
			return fmt.Errorf("snapshot: %v", err)
		}
		have := 0
		for _, s := range sn.Player.Skills {
			if s.ID == frostArmor {
				have = s.Level
			}
		}
		if have >= want {
			return nil
		}
		if _, err := r.Cheat(ctx, fmt.Sprintf("debug points %d 1", frostArmor), fmt.Sprintf("learn Frozen Armor level %d: the recast proof needs the skill (fixrecast live recipe)", have+1)); err != nil {
			return err
		}
		_ = l.Tick(ctx, 10)
	}
	return fmt.Errorf("Frozen Armor did not reach level %d", want)
}

// froClear removes the armor state if it is on (a logged test action).
func (l *lab) froClear(ctx context.Context, r *drill.Run) {
	if b, err := l.buffsText(ctx); err == nil && froOn(b) {
		_, _ = l.skillDebug(r).StatesClear(ctx, frostArmor, "fixrecast proof: start from no armor")
		_ = l.Tick(ctx, 5)
	}
}

func (l *lab) FroDuration(ctx context.Context, r *drill.Run, level int) (o drill.FroDur) {
	o = drill.FroDur{Level: level, OnAfterMS: -1, GoneAfterMS: -1, ExpectMS: (3000 + 300*(level-1)) * 40}
	if err := l.froLearn(ctx, r, level); err != nil {
		o.Err = err.Error()
		return
	}
	l.froClear(ctx, r)
	sd := l.skillDebug(r)
	if _, err := sd.ManaFull(ctx, "fixrecast proof: known mana for the Frozen Armor cast"); err != nil {
		o.Err = err.Error()
		return
	}
	_ = l.Tick(ctx, 5)
	x, y, _, err := l.playerCell(ctx)
	if err != nil {
		o.Err = err.Error()
		return
	}
	m0, _ := l.rawStats(ctx)
	g0, _, err := l.gameClock(ctx)
	if err != nil {
		o.Err = err.Error()
		return
	}
	if err := l.be.Client.CastCell(ctx, frostArmor, x, y); err != nil { // cmd skill 40 x y, the command the bot's key press sends
		o.Err = "cast: " + err.Error()
		return
	}
	for {
		if _, err := l.be.Client.Step(ctx, 1); err != nil {
			o.Err = err.Error()
			return
		}
		g, _, _ := l.gameClock(ctx)
		el := g - g0
		if o.OnAfterMS < 0 {
			if b, err := l.buffsText(ctx); err == nil && froOn(b) {
				o.OnAfterMS, o.BuffsOn = el, b
				if m1, err := l.rawStats(ctx); err == nil && m0 != nil {
					o.CostPoints = float64(m0[8]-m1[8]) / 256 // not corrected for the regen of the ticks in between
				}
			}
		} else if el > o.ExpectMS-4000 { // 80 ms steps only around the end
			if b, err := l.buffsText(ctx); err == nil && !froOn(b) {
				o.GoneAfterMS = el
				break
			}
		}
		if o.OnAfterMS >= 0 && el < o.ExpectMS-4000 { // coarse: 24 steps of 2 ticks at a time
			if _, err := l.be.Client.Step(ctx, 24); err != nil {
				o.Err = err.Error()
				return
			}
			continue
		}
		if el > o.ExpectMS+30000 || (o.OnAfterMS < 0 && el > 5000) {
			break
		}
	}
	l.syncPacer()
	return
}

func (l *lab) FroBot(ctx context.Context, r *drill.Run, level int, simMS int) (o drill.FroBot) {
	o = drill.FroBot{Level: level, SimMS: simMS, FirstOnMS: -1}
	if err := l.froLearn(ctx, r, level); err != nil {
		o.Err = err.Error()
		return
	}
	if n, err := l.rig.BindSkills(); err != nil || n == 0 {
		o.Err = fmt.Sprintf("BindSkills: %d, %v", n, err)
		return
	}
	l.froClear(ctx, r)
	sd := l.skillDebug(r)
	if _, err := sd.ManaFull(ctx, "fixrecast proof: full mana before the simulated stretch"); err != nil {
		o.Err = err.Error()
		return
	}
	_ = l.Tick(ctx, 5)
	c := l.tune.Combat
	frames := 3000 + 300*(level-1)
	o.LimitMS = min(c.BuffRecastSec*1000, frames*40-c.BuffMarginSec*1000)
	g0, _, err := l.gameClock(ctx)
	if err != nil {
		o.Err = err.Error()
		return
	}
	c0 := l.be.Env.Clock.Now()
	prev, last := 0, -1
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
		b, err := l.buffsText(ctx)
		if err != nil {
			o.Err = err.Error()
			return
		}
		up := froOn(b)
		o.Samples++
		if len(o.CastsMS) > 0 {
			o.MaxStepMS = max(o.MaxStepMS, el-prev)
		}
		if up && o.FirstOnMS < 0 {
			o.FirstOnMS = el
		}
		if up && !o.StateReadable {
			if sn, err := l.be.Env.Snapshot(); err == nil && sn.Player != nil && sn.Player.States.Has(10) {
				o.StateReadable = true
			}
		}
		if o.FirstOnMS >= 0 && !up {
			o.DownSamples++
			o.DownMS += el - prev
		}
		prev = el
		if sn, err := l.be.Env.Snapshot(); err == nil && sn.Player != nil && sn.Player.Mana < 10 {
			if _, err := sd.ManaFull(ctx, "fixrecast proof: mana for the next buff (the stretch is longer than the regen)"); err == nil {
				o.ManaRefills++
			}
		}
		mark := l.rec.Mark()
		cs := l.be.Env.Clock.Now()
		if err := l.rig.Fighter.Buff(ctx); err != nil {
			o.Err = "Buff: " + err.Error()
			return
		}
		for _, ln := range l.rec.Since(mark) {
			if strings.HasPrefix(ln, "buff skill=Frozen Armor") {
				o.CastsMS = append(o.CastsMS, el)
				o.ClockCastsMS = append(o.ClockCastsMS, int(cs.Sub(c0).Milliseconds()))
				last = el
				o.Lines = append(o.Lines, ln)
			}
		}
		step := 24 // coarse: about 1 s
		if last >= 0 && el-last > o.LimitMS-4000 && el-last < o.LimitMS+16000 {
			step = 1 // fine around the recast and the real end: about 120 ms
		}
		_ = l.Tick(ctx, step)
	}
	return
}

var d13m1Ready bool

func (l *lab) ManaTrace(ctx context.Context, r *drill.Run, hitAfter int) (t drill.ManaTrace) {
	t = drill.ManaTrace{HitAfter: hitAfter, PayTick: -1, MissTick: -1, Hit: "no interruption"}
	if !d13m1Ready {
		if err := l.d11Setup(ctx, r, 2, 100, 6); err != nil {
			t.Err = "setup: " + err.Error()
			return
		}
		d13m1Ready = true
	}
	sd := l.skillDebug(r)
	for _, v := range []string{"debug stat 6 25600"} { // life 100
		if _, err := r.Cheat(ctx, v, "M1: full life before the trial (the interruption trials take a hit)"); err != nil {
			t.Err = err.Error()
			return
		}
	}
	if _, err := sd.ManaFull(ctx, "M1: known mana for the action-frame read"); err != nil {
		t.Err = err.Error()
		return
	}
	_ = l.Tick(ctx, 40)
	x, y, _, err := l.playerCell(ctx)
	if err != nil {
		t.Err = err.Error()
		return
	}
	row := func(tick0 int) (drill.ManaRow, error) {
		_, tk, err := l.gameClock(ctx)
		if err != nil {
			return drill.ManaRow{}, err
		}
		raw, err := l.rawStats(ctx)
		if err != nil {
			return drill.ManaRow{}, err
		}
		_, _, pm, err := l.playerCell(ctx)
		if err != nil {
			return drill.ManaRow{}, err
		}
		ms, err := l.be.Client.Missiles(ctx)
		n := 0
		if err == nil {
			for _, m := range ms.Items {
				if m.OwnerType == 0 {
					n++
				}
			}
		}
		return drill.ManaRow{Tick: tk - tick0, Raw: raw[8], PMode: pm, Missiles: n}, nil
	}
	_, tick0, _ := l.gameClock(ctx)
	r0, err := row(tick0)
	if err != nil {
		t.Err = err.Error()
		return
	}
	t.Start = r0.Raw
	if _, err := sd.CastAt(ctx, 36, x+15, y, "M1: Fire Bolt through the normal cast path, mana read every step"); err != nil {
		t.Err = err.Error()
		return
	}
	t.Rows = append(t.Rows, r0)
	hit := func() {
		if _, err := r.Cheat(ctx, "debug hit 1", "M1: a forced get-hit as the interruption of the cast (a cheat; a real monster blow is not used: castcut showed weak hits do not cut)"); err != nil {
			t.Err = err.Error()
		}
		t.Hit = fmt.Sprintf("debug hit 1 (cheat) after %d steps", hitAfter)
	}
	if hitAfter == 0 {
		hit()
	}
	for i := 1; i <= 10 && t.Err == ""; i++ {
		if _, err := l.be.Client.Step(ctx, 1); err != nil {
			t.Err = err.Error()
			return
		}
		rw, err := row(tick0)
		if err != nil {
			t.Err = err.Error()
			return
		}
		t.Rows = append(t.Rows, rw)
		if i == hitAfter {
			hit()
		}
	}
	for i := 1; i < len(t.Rows); i++ {
		if t.PayTick < 0 && t.Rows[i].Raw <= t.Rows[i-1].Raw-300 {
			t.PayTick = t.Rows[i].Tick
		}
		if t.MissTick < 0 && t.Rows[i].Missiles > 0 {
			t.MissTick = t.Rows[i].Tick
		}
	}
	t.Final = t.Rows[len(t.Rows)-1].Raw
	_ = l.Tick(ctx, 60)
	return
}
