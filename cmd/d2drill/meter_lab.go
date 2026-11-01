package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"d2lod-bot/internal/d2dcombat"
	"d2lod-bot/internal/drill"
	"d2lod-bot/internal/env"
)

// gtClock is the D-meter ground truth: it wraps the clock under the rig's Meter and, after every sleep, reads the raw
// `monsters` and `state` verbs per unit id (not the snapshot the Meter folds), so the totals are computed twice by
// different code on the same game ticks.
type gtClock struct {
	inner env.Clock
	l     *lab
	ctx   context.Context

	mu       sync.Mutex
	first    map[int]int
	last     map[int]int
	dead     map[int]bool
	class    map[int]string
	taken    int
	t0       d2dcombat.Totals // the Meter's totals when the window opened: Enable forgets the last sample, not the totals
	lastLife int
	have     bool
	samples  int
}

func newGT(l *lab, ctx context.Context, inner env.Clock) *gtClock {
	return &gtClock{inner: inner, l: l, ctx: ctx, first: map[int]int{}, last: map[int]int{}, dead: map[int]bool{}, class: map[int]string{}}
}

func (g *gtClock) Now() time.Time { return g.inner.Now() }

func (g *gtClock) Sleep(d time.Duration) {
	g.inner.Sleep(d)
	g.sample()
}

func (g *gtClock) sample() {
	ms, merr := g.l.be.Client.Monsters(g.ctx)
	st, _, serr := g.l.be.Client.State(g.ctx)
	g.mu.Lock()
	defer g.mu.Unlock()
	g.samples++
	if merr == nil {
		for _, m := range ms.Items {
			if m.HP.Max <= 0 || g.dead[m.Unit] {
				continue
			}
			if m.HP.Cur <= 0 || m.Mode == "DD" || m.Mode == "DT" {
				if _, seen := g.first[m.Unit]; seen {
					g.dead[m.Unit] = true
				}
				continue
			}
			if _, ok := g.first[m.Unit]; !ok {
				g.first[m.Unit] = m.HP.Cur
				g.class[m.Unit] = m.ClassID
			}
			g.last[m.Unit] = m.HP.Cur
		}
	}
	if serr == nil && st.Life.Max > 0 {
		if g.have && g.lastLife > st.Life.Cur {
			g.taken += g.lastLife - st.Life.Cur
		}
		g.lastLife, g.have = st.Life.Cur, true
	}
}

// startMeter installs the ground truth under the Meter and enables it; call finishMeter after the Clear loop.
func (l *lab) startMeter(ctx context.Context) (*gtClock, env.Clock) {
	old := l.rig.Meter.Clk
	g := newGT(l, ctx, old)
	l.rig.Meter.Clk = g
	l.rig.Meter.Enable()
	g.t0 = l.rig.Meter.Totals()
	return g, old
}

func (l *lab) finishMeter(g *gtClock, old env.Clock, clearKills int) *drill.MeterCheck {
	l.rig.Meter.Clk = old
	t := l.rig.Meter.Totals()
	t.Damage -= g.t0.Damage
	t.ManaSpent -= g.t0.ManaSpent
	t.Deaths -= g.t0.Deaths
	t.Samples -= g.t0.Samples
	t.DamageDealt -= g.t0.DamageDealt
	t.Kills -= g.t0.Kills
	g.mu.Lock()
	defer g.mu.Unlock()
	mc := &drill.MeterCheck{Totals: t.String(), Kills: t.Kills, Dealt: t.DamageDealt, Taken: t.Damage, DealtOK: t.DealtOK, GTTaken: g.taken, Units: len(g.first), Samples: g.samples, ClearKills: clearKills}
	ids := make([]int, 0, len(g.first))
	for id := range g.first {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	for _, id := range ids {
		f, la := g.first[id], g.last[id]
		mc.SpawnHP += f
		drops := f - la
		mc.GTDealtDrops += drops
		full := drops
		st := "alive"
		if g.dead[id] {
			mc.GTDead++
			full += la
			st = "dead"
		}
		mc.GTDealtFull += full
		mc.PerUnit = append(mc.PerUnit, fmt.Sprintf("#%d %s hp %d->%d %s", id, g.class[id], f, la, st))
	}
	return mc
}
