package main

import (
	"context"
	"fmt"
	"math"
	"strings"

	"d2lod-bot/internal/drill"
)

// DRILLS15 part 2: the mana (M), missile (K) and armor (E) drills.

// d15playerMissiles returns the player-owned missile rows now.
func (l *lab) d15playerMissiles(ctx context.Context) (n int, ids map[int][2]float64, names string) {
	ids = map[int][2]float64{}
	ms, err := l.be.Client.Missiles(ctx)
	if err != nil {
		return
	}
	var nm []string
	for _, m := range ms.Items {
		if m.OwnerType == 0 {
			n++
			ids[m.Unit] = [2]float64{m.X, m.Y}
			nm = append(nm, m.Name)
		}
	}
	return n, ids, strings.Join(nm, ",")
}

func (l *lab) d15M2(ctx context.Context, r *drill.Run) (drill.Outcome, error) {
	if err := l.d15prep(ctx, r, skFireBolt, 1); err != nil {
		return drill.Outcome{}, err
	}
	id, mx, my, err := l.d15spawn(ctx, r, "fallen1", "M2: a monster whose blow lands during the cast")
	if err != nil {
		return drill.Outcome{}, err
	}
	if _, err := r.Cheat(ctx, fmt.Sprintf("debug mon set %d life 5000 dmg 20 20", id), "M2: a blow above max life / 12 (a get-hit), a monster that stays alive"); err != nil {
		return drill.Outcome{}, err
	}
	l.d15warp(ctx, r, mx+1, my)
	type trial struct {
		hitTick, missTick, payTick int
		startMode                  int
		start, end                 int
		lifeLost                   int
	}
	var ts []trial
	var cutPaid, cutNoMiss, cutTrials, missBeforeHit, noHit, skipped int
	for i := 0; i < 40; i++ {
		_, _ = r.Cheat(ctx, "debug stat 6 25600", "M2: full life before each trial")
		_ = l.d15mana(ctx, r, -1, "M2: full mana before each trial")
		l.d15settle(ctx)
		if x, y, ok := l.d15monster(ctx, id); ok {
			if p, err := l.d15read(ctx); err == nil && cdist(p.X, p.Y, x, y) > 2.5 {
				l.d15warp(ctx, r, x+1, y)
			}
		}
		_ = l.Tick(ctx, 1)
		b, err := l.d15read(ctx)
		if err != nil {
			return drill.Outcome{}, err
		}
		tx, ty := b.X+8, b.Y
		if x, y, ok := l.d15monster(ctx, id); ok {
			tx, ty = x, y
		}
		if _, err := l.skillDebug(r).CastAt(ctx, skFireBolt, tx, ty, "M2: Fire Bolt next to a hitting monster"); err != nil {
			return drill.Outcome{}, err
		}
		t := trial{hitTick: -1, missTick: -1, payTick: -1, startMode: b.PMode, start: b.Mana}
		prev := b.Mana
		for k := 1; k <= 14; k++ {
			if _, err := l.be.Client.Step(ctx, 1); err != nil {
				return drill.Outcome{}, err
			}
			p, _ := l.d15read(ctx)
			n, _, _ := l.d15playerMissiles(ctx)
			if t.hitTick < 0 && p.PMode == 4 {
				t.hitTick = k
			}
			if t.missTick < 0 && n > 0 {
				t.missTick = k
			}
			if t.payTick < 0 && p.Mana <= prev-300 {
				t.payTick = k
			}
			prev = p.Mana
			t.end = p.Mana
			t.lifeLost = b.Life - p.Life
		}
		ts = append(ts, t)
		switch {
		case t.startMode == 4:
			skipped++
		case t.hitTick < 0:
			noHit++
		case t.missTick < 0 || t.hitTick < t.missTick:
			cutTrials++
			if t.missTick < 0 {
				cutNoMiss++
				if t.payTick >= 0 {
					cutPaid++
				}
			}
		default:
			missBeforeHit++
		}
	}
	var rows []string
	for i, t := range ts {
		if i < 14 || t.hitTick >= 0 && t.missTick < 0 {
			rows = append(rows, fmt.Sprintf("#%d start-mode %d hit@%d missile@%d paid@%d lifelost %.1f", i, t.startMode, t.hitTick, t.missTick, t.payTick, float64(t.lifeLost)/256))
		}
		if len(rows) > 30 {
			break
		}
	}
	o := drill.Outcome{Intended: "game.exe: a get-hit before the action frame cancels the cast: no missile and NO mana paid (pay is at the action frame, sk-mana-trace D3)", Observed: fmt.Sprintf("%d casts: no blow %d, blow after the missile %d, blow before the missile (cut) %d (of those no missile %d, mana paid %d), started in hit recovery %d | %s", len(ts), noHit, missBeforeHit, cutTrials, cutNoMiss, cutPaid, skipped, strings.Join(rows, " ; "))}
	o.Sig = fmt.Sprintf("cut=%d paid=%d", cutNoMiss, cutPaid)
	switch {
	case cutNoMiss == 0:
		return drill.Blocks(fmt.Sprintf("no cast was cut by a real blow in %d casts (the monster's A1 rarely lands inside the 8-tick window)", len(ts))), nil
	case cutPaid == cutNoMiss:
		o.Verdict, o.Reason = drill.Fail, fmt.Sprintf("%d of %d cut casts (real monster blow, dmg 20) lost the mana with no missile (trace D3: game.exe pays nothing)", cutPaid, cutNoMiss)
	case cutPaid == 0:
		o.Verdict, o.Reason = drill.Pass, fmt.Sprintf("%d cut casts, no mana paid", cutNoMiss)
	default:
		o.Verdict, o.Reason = drill.Fail, fmt.Sprintf("%d of %d cut casts paid mana (game.exe pays none)", cutPaid, cutNoMiss)
	}
	return o, nil
}

func (l *lab) d15M3(ctx context.Context, r *drill.Run) (drill.Outcome, error) {
	if err := l.d15prep(ctx, r, skTeleport, 25); err != nil {
		return drill.Outcome{}, err
	}
	s, err := l.d15read(ctx)
	if err != nil {
		return drill.Outcome{}, err
	}
	dx, dy, n := l.d15dir(ctx, s.X, s.Y, 6)
	if n < 4 {
		return drill.Outcome{}, fmt.Errorf("no free line of 4 cells")
	}
	var obs []string
	moved := map[int]bool{}
	for _, raw := range []int{128, 0, 255} {
		l.d15warp(ctx, r, s.X, s.Y)
		if err := l.d15mana(ctx, r, raw, fmt.Sprintf("M3: mana %d (256ths) for the level 25 Teleport floor", raw)); err != nil {
			return drill.Outcome{}, err
		}
		_ = l.Tick(ctx, 2)
		b, a, err := l.d15cast(ctx, r, skTeleport, s.X+dx*4, s.Y+dy*4, fmt.Sprintf("M3: level 25 Teleport with mana %d/256", raw))
		if err != nil {
			return drill.Outcome{}, err
		}
		mv := cdist(a.X, a.Y, b.X, b.Y)
		moved[raw] = mv > 0.5
		obs = append(obs, fmt.Sprintf("mana %d/256 (%.2f points): moved %.2f cells, mana after %d/256 (14 ticks of regen add about 42)", b.Mana, float64(b.Mana)/256, mv, a.Mana))
	}
	_, _ = r.Cheat(ctx, "debug points 54 1", "M3: Teleport back to level 1 after the floor test")
	o := drill.Outcome{Intended: "game.exe: from level 25 the cost formula is <= 0: the cast works at any mana (check passes at 0 <= mana), pays 1 point only when it has it; d2d (trace D2): refuses without a full point", Observed: strings.Join(obs, " || ")}
	o.Sig = fmt.Sprint(moved)
	if moved[128] && moved[0] {
		o.Verdict, o.Reason = drill.Pass, "the level 25 hop works at half a point and at 0 mana"
	} else {
		o.Verdict, o.Reason = drill.Fail, fmt.Sprintf("level 25 Teleport refused at low mana (moved: half point %v, 0 mana %v, 255/256 %v): d2d has no minmana floor (trace D2), game.exe casts", moved[128], moved[0], moved[255])
	}
	return o, nil
}

func (l *lab) d15M4(ctx context.Context, r *drill.Run) (drill.Outcome, error) {
	if err := l.d15prep(ctx, r, -1, 0); err != nil {
		return drill.Outcome{}, err
	}
	var obs []string
	deltas := map[string]int{}
	for _, warm := range []bool{false, true} {
		if warm {
			if _, err := r.Cheat(ctx, "debug points 37 1", "M4: Warmth 1 for the regen read"); err != nil {
				return drill.Outcome{}, err
			}
		}
		if _, err := r.Cheat(ctx, fmt.Sprintf("debug stat 9 %d", 45*256), "M4: 45 points of max mana (the trace's ene 45 case)"); err != nil {
			return drill.Outcome{}, err
		}
		if err := l.d15mana(ctx, r, 0, "M4: empty mana before 30 s"); err != nil {
			return drill.Outcome{}, err
		}
		_ = l.Tick(ctx, 2)
		b, _ := l.d15read(ctx)
		for i := 0; i < 3; i++ {
			if err := l.Tick(ctx, 249); err != nil {
				return drill.Outcome{}, err
			}
		}
		a, _ := l.d15read(ctx)
		raw, _ := l.rawStats(ctx)
		d := a.Mana - b.Mana
		key := "no Warmth"
		if warm {
			key = "Warmth 1"
		}
		deltas[key] = d
		obs = append(obs, fmt.Sprintf("%s: max mana raw %d, mana %d -> %d over 750 ticks (30 s): %+d/256 = %.2f points", key, raw[9], b.Mana, a.Mana, d, float64(d)/256))
	}
	_, _ = r.Cheat(ctx, fmt.Sprintf("debug stat 9 %d", 35*256), "M4: max mana back to the character's 35 points")
	_ = l.d15mana(ctx, r, -1, "M4: full mana after the test")
	o := drill.Outcome{Intended: "game.exe: +2250/256 (8.8 points) in 30 s at 45 max mana (3 per frame); Warmth 1 unchanged; the bot's model says 13.5 points", Observed: strings.Join(obs, " || ")}
	o.Sig = fmt.Sprint(deltas)
	if d := deltas["no Warmth"]; d >= 2200 && d <= 2300 && deltas["Warmth 1"] == d {
		o.Verdict, o.Reason = drill.Pass, fmt.Sprintf("d2d regenerates %+d/256 in 750 ticks, the game.exe figure (2250), Warmth 1 changes nothing", d)
	} else {
		o.Verdict, o.Reason = drill.Fail, fmt.Sprintf("d2d regen over 750 ticks: %d/256 without Warmth, %d with (game.exe 2250 for both)", deltas["no Warmth"], deltas["Warmth 1"])
	}
	return o, nil
}

func (l *lab) d15M6(ctx context.Context, r *drill.Run) (drill.Outcome, error) {
	if err := l.d15prep(ctx, r, skBlizzard, 1); err != nil {
		return drill.Outcome{}, err
	}
	s, err := l.d15read(ctx)
	if err != nil {
		return drill.Outcome{}, err
	}
	dx, dy, n := l.d15dir(ctx, s.X, s.Y, 10)
	if n < 6 {
		return drill.Outcome{}, fmt.Errorf("no free line of 6 cells")
	}
	tx, ty := s.X+dx*float64(n), s.Y+dy*float64(n)
	var obs []string
	second := map[int]bool{}
	for _, gap := range []int{25, 40, 60, 100} {
		l.d15warp(ctx, r, s.X, s.Y)
		_ = l.Tick(ctx, 60)
		if err := l.d15mana(ctx, r, -1, "M6: full mana for the first Blizzard"); err != nil {
			return drill.Outcome{}, err
		}
		l.d15settle(ctx)
		b1, _ := l.d15read(ctx)
		if _, err := l.skillDebug(r).CastAt(ctx, skBlizzard, tx, ty, fmt.Sprintf("M6: first Blizzard, the second %d ticks later", gap)); err != nil {
			return drill.Outcome{}, err
		}
		_ = l.Tick(ctx, gap-1)
		st, _ := l.be.Client.Do(ctx, "debug states")
		p, _ := l.d15read(ctx)
		if err := l.d15mana(ctx, r, -1, "M6: full mana for the second Blizzard (the first one's cost is not the question)"); err != nil {
			return drill.Outcome{}, err
		}
		before, _ := l.d15read(ctx)
		if _, err := l.skillDebug(r).CastAt(ctx, skBlizzard, tx, ty, fmt.Sprintf("M6: second Blizzard %d ticks after the first", gap)); err != nil {
			return drill.Outcome{}, err
		}
		paid, mode10 := false, false
		for k := 0; k < 20; k++ {
			_, _ = l.be.Client.Step(ctx, 1)
			q, _ := l.d15read(ctx)
			if q.PMode == 10 {
				mode10 = true
			}
			if q.Mana <= before.Mana-1000 {
				paid = true
			}
		}
		second[gap] = paid
		obs = append(obs, fmt.Sprintf("gap %d ticks (%d ms): first cost %.1f points; at the second cast mode %d, states %q; second cast started (mode 10 seen) %s, mana paid %s", gap, 40*gap, float64(b1.Mana-0)/256-float64(p.Mana)/256+0, p.PMode, strings.TrimSpace(st.Text()), yn(mode10), yn(paid)))
	}
	o := drill.Outcome{Intended: "game.exe: the second cast is refused until 45 frames after the first's action frame (skilldelay state 0x79); d2d (trace D6) casts both", Observed: strings.Join(obs, " || ")}
	o.Sig = fmt.Sprint(second)
	if !second[25] && second[100] {
		o.Verdict, o.Reason = drill.Pass, "the second Blizzard is refused inside the delay and cast after it"
	} else {
		o.Verdict, o.Reason = drill.Fail, fmt.Sprintf("second cast paid at gap 25: %v, 40: %v, 60: %v, 100: %v (game.exe refuses the early ones): d2d does not enforce the skill delay", second[25], second[40], second[60], second[100])
	}
	return o, nil
}

// d15K1: one missile skill, a point 12 cells away on a clear line; reach = the farthest the missile gets from the caster.
func (l *lab) d15K1(ctx context.Context, r *drill.Run) (drill.Outcome, error) {
	if err := l.d15prep(ctx, r, -1, 0); err != nil {
		return drill.Outcome{}, err
	}
	s, err := l.d15read(ctx)
	if err != nil {
		return drill.Outcome{}, err
	}
	dx, dy, n := l.d15dir(ctx, s.X, s.Y, 16)
	if n < 14 {
		return drill.Outcome{}, fmt.Errorf("no free line of 14 cells (best %d)", n)
	}
	type sk struct {
		id     int
		name   string
		expect float64
	}
	var obs, sig []string
	ok := true
	for _, k := range []sk{{skFireBolt, "Fire Bolt", 46.9}, {skIceBolt, "Ice Bolt", 28.1}, {skLightning, "Lightning", 35.2}} {
		if _, err := r.Cheat(ctx, fmt.Sprintf("debug points %d 1", k.id), "K1: "+k.name+" level 1 for the reach read"); err != nil {
			return drill.Outcome{}, err
		}
		l.d15warp(ctx, r, s.X, s.Y)
		_ = l.d15mana(ctx, r, -1, "K1: full mana")
		l.d15settle(ctx)
		b, _ := l.d15read(ctx)
		if _, err := l.skillDebug(r).CastAt(ctx, k.id, b.X+dx*12, b.Y+dy*12, "K1: "+k.name+" at a point 12 cells (60 subtiles) away on a clear line"); err != nil {
			return drill.Outcome{}, err
		}
		first, last := -1, -1
		maxd := 0.0
		var fx, fy float64
		names := ""
		var firstRaw string
		for t := 1; t <= 120; t++ {
			_, _ = l.be.Client.Step(ctx, 1)
			ms, _ := l.be.Client.Missiles(ctx)
			any := false
			for _, m := range ms.Items {
				if m.OwnerType != 0 {
					continue
				}
				any = true
				if first < 0 {
					first, fx, fy, names, firstRaw = t, m.X, m.Y, m.Name, m.Raw
				}
				maxd = math.Max(maxd, cdist(m.X, m.Y, b.X, b.Y))
			}
			if any {
				last = t
			} else if first >= 0 {
				break
			}
		}
		reach := 5 * maxd
		okk := first > 0 && math.Abs(reach-k.expect) <= 3
		ok = ok && okk
		obs = append(obs, fmt.Sprintf("%s: first row at +%d ticks (%.2f,%.2f) = %.1f subtiles from the caster, last row at +%d (flight %d ticks), farthest %.1f subtiles (trace %.1f), name %q, first row %q", k.name, first, fx, fy, 5*cdist(fx, fy, b.X, b.Y), last, last-first+1, reach, k.expect, names, firstRaw))
		sig = append(sig, fmt.Sprintf("%s %.0f", k.name, reach))
	}
	o := drill.Outcome{Intended: "game.exe: farthest 46.9 / 28.1 / 35.2 subtiles from the start (missile range tables, V from the trace's computed D)", Observed: strings.Join(obs, " || "), Sig: strings.Join(sig, ",")}
	if ok {
		o.Verdict, o.Reason = drill.Pass, "all three reaches within 3 subtiles of the trace figures (the missile rows are read per tick: the farthest sampled position)"
	} else {
		o.Verdict, o.Reason = drill.Fail, "at least one reach differs from the trace by more than 3 subtiles (see the numbers)"
	}
	return o, nil
}

func (l *lab) d15K2(ctx context.Context, r *drill.Run) (drill.Outcome, error) {
	if err := l.d15prep(ctx, r, -1, 0); err != nil {
		return drill.Outcome{}, err
	}
	s, err := l.d15read(ctx)
	if err != nil {
		return drill.Outcome{}, err
	}
	dx, dy, n := l.d15dir(ctx, s.X, s.Y, 10)
	if n < 6 {
		return drill.Outcome{}, fmt.Errorf("no free line of 6 cells")
	}
	var obs, sig []string
	ok := true
	for _, lv := range []int{1, 5, 20} {
		if _, err := r.Cheat(ctx, fmt.Sprintf("debug points %d %d", skCharged, lv), fmt.Sprintf("K2: Charged Bolt level %d", lv)); err != nil {
			return drill.Outcome{}, err
		}
		l.d15warp(ctx, r, s.X, s.Y)
		_ = l.d15mana(ctx, r, -1, "K2: full mana")
		l.d15settle(ctx)
		b, _ := l.d15read(ctx)
		if _, err := l.skillDebug(r).CastAt(ctx, skCharged, b.X+dx*float64(n), b.Y+dy*float64(n), fmt.Sprintf("K2: Charged Bolt level %d at a point %d cells away", lv, n)); err != nil {
			return drill.Outcome{}, err
		}
		seen := map[int][2]float64{}
		first, last, peak := -1, -1, 0
		for t := 1; t <= 100; t++ {
			_, _ = l.be.Client.Step(ctx, 1)
			c, ids, _ := l.d15playerMissiles(ctx)
			for id, p := range ids {
				if _, in := seen[id]; !in {
					seen[id] = p
				}
			}
			if c > 0 {
				if first < 0 {
					first = t
				}
				last = t
				peak = max(peak, c)
			} else if first >= 0 {
				break
			}
		}
		want := min(24, 3+lv-1)
		far := 0.0
		for _, p := range seen {
			far = math.Max(far, 5*cdist(p[0], p[1], b.X, b.Y))
		}
		obs = append(obs, fmt.Sprintf("level %d: %d distinct bolts (trace %d), at most %d alive at once, first seen +%d, last alive +%d (%d frames after the first), first-seen positions at most %.1f subtiles from the caster", lv, len(seen), want, peak, first, last, last-first+1, far))
		sig = append(sig, fmt.Sprint(len(seen)))
		ok = ok && len(seen) == want
	}
	o := drill.Outcome{Intended: "game.exe: min(24, 3 + (level - 1)) bolts = 3, 7, 22 from the caster's subtile, all gone by 77 frames", Observed: strings.Join(obs, " || "), Sig: strings.Join(sig, ",")}
	if ok {
		o.Verdict, o.Reason = drill.Pass, "bolt counts 3, 7, 22 at levels 1, 5, 20"
	} else {
		o.Verdict, o.Reason = drill.Fail, "a bolt count differs from min(24, 3 + level - 1) (the count is of distinct missile unit ids seen at one-tick sampling; short-lived bolts could be missed: see the numbers)"
	}
	return o, nil
}

func (l *lab) d15E1(ctx context.Context, r *drill.Run) (drill.Outcome, error) {
	if err := l.d15prep(ctx, r, skEnergySh, 1); err != nil {
		return drill.Outcome{}, err
	}
	s, err := l.d15read(ctx)
	if err != nil {
		return drill.Outcome{}, err
	}
	var obs []string
	if _, err := l.skillDebug(r).StatesClear(ctx, skEnergySh, "E1: start without the shield"); err != nil {
		_ = err // "skill not in states" when it is not on
	}
	if _, err := r.Cheat(ctx, "debug stat 7 25600", "E1: 100 points of max life so a 50 point hit is not lethal"); err != nil {
		return drill.Outcome{}, err
	}
	_, _ = r.Cheat(ctx, "debug stat 6 25600", "E1: full life")
	_ = l.d15mana(ctx, r, -1, "E1: full mana for the Energy Shield cast")
	l.d15settle(ctx)
	if _, err := l.skillDebug(r).CastAt(ctx, skEnergySh, s.X, s.Y, "E1: Energy Shield on the caster"); err != nil {
		return drill.Outcome{}, err
	}
	_ = l.Tick(ctx, 25)
	b0, _ := l.buffsText(ctx)
	obs = append(obs, "after the cast buffs="+b0)
	if err := l.d15mana(ctx, r, 768, "E1: 3 points of mana: the 50 point hit's absorb is capped by mana (3 x 16 / 32 = 1.5 points)"); err != nil {
		return drill.Outcome{}, err
	}
	_ = l.Tick(ctx, 1)
	p0, _ := l.d15read(ctx)
	out, err := r.Cheat(ctx, "debug hit 50", "E1: a 50 point physical blow on the shielded caster")
	if err != nil {
		return drill.Outcome{}, err
	}
	p1, _ := l.d15read(ctx)
	b1, _ := l.buffsText(ctx)
	obs = append(obs, fmt.Sprintf("hit 50 (%s): life %.2f -> %.2f (lost %.2f points), mana %.2f -> %.2f, buffs=%s", strings.TrimSpace(out), float64(p0.Life)/256, float64(p1.Life)/256, float64(p0.Life-p1.Life)/256, float64(p0.Mana)/256, float64(p1.Mana)/256, b1))
	_ = l.Tick(ctx, 20)
	_, _ = r.Cheat(ctx, "debug stat 6 25600", "E1: full life for the second hit")
	_ = l.d15mana(ctx, r, -1, "E1: mana full again: does the shield still absorb")
	_ = l.Tick(ctx, 2)
	q0, _ := l.d15read(ctx)
	_, _ = r.Cheat(ctx, "debug hit 50", "E1: second 50 point blow after the refill")
	q1, _ := l.d15read(ctx)
	b2, _ := l.buffsText(ctx)
	obs = append(obs, fmt.Sprintf("after mana full, hit 50: life lost %.2f points, mana %.2f -> %.2f, buffs=%s", float64(q0.Life-q1.Life)/256, float64(q0.Mana)/256, float64(q1.Mana)/256, b2))
	lost := float64(p0.Life-p1.Life) / 256
	o := drill.Outcome{Intended: "game.exe: absorbs 1.5 points of a 50 point hit at 3 points of mana (life lost 48.5), mana 0, the state removed; mana full afterwards: no shield until recast. d2d (trace D3): absorbs 1, mana 1 left, buffs stay", Observed: strings.Join(obs, " || ")}
	o.Sig = fmt.Sprintf("lost=%.1f buffs=%s/%s", lost, b1, b2)
	stateKept := strings.Contains(b1, "58:")
	switch {
	case !strings.Contains(b0, "58:"):
		return drill.Blocks("the Energy Shield state never showed after the cast: " + b0), nil
	case math.Abs(lost-48.5) < 0.3 && !stateKept && p1.Mana < 128:
		o.Verdict, o.Reason = drill.Pass, "life lost 48.5, mana drained, shield state removed (game.exe)"
	default:
		o.Verdict, o.Reason = drill.Fail, fmt.Sprintf("d2d: life lost %.2f points (game.exe 48.5), mana left %.2f points (game.exe 0), state kept %v (game.exe removes it)", lost, float64(p1.Mana)/256, stateKept)
	}
	return o, nil
}
