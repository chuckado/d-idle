package main

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"d2lod-bot/internal/drill"
)

// DRILLS15 (card drills15): the first live skill-smoke drills from the traces sk-teleport, sk-mana, sk-missile and
// sk-armor. Every verb that changes state goes through r.Cheat or d2dctl.SkillDebug (ledgered). The reads are `dump`,
// `debug players`, `monsters`, `debug missiles`, `debug blocked`, `debug states`. Positions are level-relative cells
// (a subtile is a fifth of a cell). Tags: V = read from the logs of the run, S = read from code, U = not measured.

const (
	skFireBolt  = 36
	skWarmth    = 37
	skCharged   = 38
	skIceBolt   = 45
	skLightning = 49
	skTeleport  = 54
	skEnergySh  = 58
	skBlizzard  = 59
)

type d15st struct {
	X, Y        float64
	PMode, Mana int // mana raw, 256ths
	Life        int // raw
}

func (s d15st) String() string {
	return fmt.Sprintf("(%.2f,%.2f) mode %d mana %.1f life %.1f", s.X, s.Y, s.PMode, float64(s.Mana)/256, float64(s.Life)/256)
}

func cdist(ax, ay, bx, by float64) float64 { return math.Hypot(ax-bx, ay-by) }

func (l *lab) d15read(ctx context.Context) (d15st, error) {
	x, y, pm, err := l.playerCell(ctx)
	if err != nil {
		return d15st{}, err
	}
	raw, err := l.rawStats(ctx)
	if err != nil {
		return d15st{}, err
	}
	return d15st{X: x, Y: y, PMode: pm, Mana: raw[8], Life: raw[6]}, nil
}

// d15settle steps until the player is neither casting (10) nor in hit recovery (4), at most 40 ticks.
func (l *lab) d15settle(ctx context.Context) {
	for i := 0; i < 40; i++ {
		_, _, pm, err := l.playerCell(ctx)
		if err != nil || (pm != 10 && pm != 4) {
			return
		}
		_, _ = l.be.Client.Step(ctx, 1)
	}
}

var mercRe = regexp.MustCompile(`"merc":\{[^}]*"x":(-?[0-9.]+),"y":(-?[0-9.]+)`)

func (l *lab) d15merc(ctx context.Context) (x, y float64, ok bool) {
	rep, err := l.be.Client.Do(ctx, "dump")
	if err != nil {
		return
	}
	m := mercRe.FindStringSubmatch(rep.Text())
	if m == nil {
		return
	}
	x, _ = strconv.ParseFloat(m[1], 64)
	y, _ = strconv.ParseFloat(m[2], 64)
	return x, y, true
}

func (l *lab) d15blocked(ctx context.Context, x, y float64) bool {
	b, err := l.be.Client.Blocked(ctx, x, y)
	return err != nil || b
}

// d15dir picks the direction (8 of them) with the longest free line from (x,y), checked every half cell, up to max cells.
func (l *lab) d15dir(ctx context.Context, x, y float64, max int) (dx, dy float64, n int) {
	best := -1.0
	for _, d := range [][2]float64{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {0.7071, 0.7071}, {-0.7071, 0.7071}, {0.7071, -0.7071}, {-0.7071, -0.7071}} {
		free := 0.0
		for s := 0.5; s <= float64(max); s += 0.5 {
			if l.d15blocked(ctx, x+d[0]*s, y+d[1]*s) {
				break
			}
			free = s
		}
		if free > best {
			best, dx, dy = free, d[0], d[1]
		}
		if free >= float64(max) {
			break
		}
	}
	return dx, dy, int(best)
}

// d15prep puts the character in Blood Moor with the natural monsters gone, the skill learnt at the given level and mana full.
func (l *lab) d15prep(ctx context.Context, r *drill.Run, skill, level int) error {
	if err := l.d11Setup(ctx, r, 2, 100, 6); err != nil {
		return err
	}
	if skill >= 0 {
		if _, err := r.Cheat(ctx, fmt.Sprintf("debug points %d %d", skill, level), fmt.Sprintf("DRILLS15: skill %d at level %d for the drill (debug points sets the level)", skill, level)); err != nil {
			return err
		}
	}
	return l.d15mana(ctx, r, -1, "known mana at the start of the trial")
}

// d15mana sets mana: raw < 0 is `debug mana full`, else `debug stat 8 <raw/256 points>` (stat 8 takes points x 256).
func (l *lab) d15mana(ctx context.Context, r *drill.Run, raw int, why string) error {
	if raw < 0 {
		_, err := l.skillDebug(r).ManaFull(ctx, "DRILLS15: "+why)
		return err
	}
	_, err := r.Cheat(ctx, fmt.Sprintf("debug stat 8 %d", raw), "DRILLS15: "+why)
	return err
}

func (l *lab) d15warp(ctx context.Context, r *drill.Run, x, y float64) {
	_, _ = r.Cheat(ctx, fmt.Sprintf("debug warp %d %d", int(math.Round(x)), int(math.Round(y))), "DRILLS15: back to the start cell of the next trial")
	_ = l.Tick(ctx, 10)
	l.d15settle(ctx)
}

// d15cast casts a skill at a cell through the ledgered `debug cast` and reads the player before and 14 ticks later.
func (l *lab) d15cast(ctx context.Context, r *drill.Run, skill int, tx, ty float64, why string) (b, a d15st, err error) {
	l.d15settle(ctx)
	if b, err = l.d15read(ctx); err != nil {
		return
	}
	if _, err = l.skillDebug(r).CastAt(ctx, skill, tx, ty, why); err != nil {
		return
	}
	if err = l.Tick(ctx, 14); err != nil {
		return
	}
	a, err = l.d15read(ctx)
	return
}

func yn(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func (l *lab) D15(ctx context.Context, r *drill.Run, id string) (o drill.Outcome) {
	var err error
	switch id {
	case "D15-T1":
		o, err = l.d15T1(ctx, r)
	case "D15-T2":
		o, err = l.d15T2(ctx, r)
	case "D15-T3":
		o, err = l.d15T3(ctx, r)
	case "D15-T4":
		o, err = l.d15T4(ctx, r)
	case "D15-T5":
		o, err = l.d15T5(ctx, r)
	case "D15-T7":
		o, err = l.d15T7(ctx, r)
	case "D15-M2":
		o, err = l.d15M2(ctx, r)
	case "D15-M3":
		o, err = l.d15M3(ctx, r)
	case "D15-M4":
		o, err = l.d15M4(ctx, r)
	case "D15-M6":
		o, err = l.d15M6(ctx, r)
	case "D15-K1":
		o, err = l.d15K1(ctx, r)
	case "D15-K2":
		o, err = l.d15K2(ctx, r)
	case "D15-E1":
		o, err = l.d15E1(ctx, r)
	default:
		err = fmt.Errorf("unknown drill %s", id)
	}
	if err != nil {
		return drill.Blocks(err.Error())
	}
	return o
}

func (l *lab) d15T1(ctx context.Context, r *drill.Run) (drill.Outcome, error) {
	if err := l.Goto(ctx, r, 1); err != nil {
		return drill.Outcome{}, err
	}
	if _, err := r.Cheat(ctx, "debug points 54 1", "T1: Teleport learnt for the town cast"); err != nil {
		return drill.Outcome{}, err
	}
	if err := l.d15mana(ctx, r, -1, "T1: full mana, the cast in town must leave it alone"); err != nil {
		return drill.Outcome{}, err
	}
	_ = l.Tick(ctx, 5)
	s, err := l.d15read(ctx)
	if err != nil {
		return drill.Outcome{}, err
	}
	dx, dy, n := l.d15dir(ctx, s.X, s.Y, 8)
	if n < 4 {
		return drill.Outcome{}, fmt.Errorf("no free line of 4 cells in town (best %d)", n)
	}
	var obs []string
	moved, paid := false, false
	for _, d := range []float64{3, 4} {
		tx, ty := s.X+dx*math.Min(d, float64(n)), s.Y+dy*math.Min(d, float64(n))
		b, a, err := l.d15cast(ctx, r, skTeleport, tx, ty, "T1: Teleport in town (game.exe refuses with no mana)")
		if err != nil {
			return drill.Outcome{}, err
		}
		mv := cdist(a.X, a.Y, b.X, b.Y)
		obs = append(obs, fmt.Sprintf("target %.1f cells away: before %s, after 14 ticks %s, moved %.2f cells, mana change %+.2f points", cdist(tx, ty, b.X, b.Y), b, a, mv, float64(a.Mana-b.Mana)/256))
		moved = moved || mv > 0.5
		paid = paid || a.Mana < b.Mana
	}
	o := drill.Outcome{Intended: "game.exe: refused in a town room at the start (FUN_0056f640), the character stays and no mana is taken", Observed: strings.Join(obs, " || ")}
	o.Sig = fmt.Sprintf("moved=%v paid=%v", moved, paid)
	if !moved && !paid {
		o.Verdict, o.Reason = drill.Pass, "refused in town on d2d: no move, no mana taken (equal to game.exe, trace D: V read)"
	} else {
		o.Verdict, o.Reason = drill.Fail, fmt.Sprintf("d2d differs from game.exe in town: moved=%v mana paid=%v", moved, paid)
	}
	return o, nil
}

// d15hops trials: for each distance the character is put back at the start, mana made full, and one Teleport cast.
func (l *lab) d15T2(ctx context.Context, r *drill.Run) (drill.Outcome, error) {
	if err := l.d15prep(ctx, r, skTeleport, 1); err != nil {
		return drill.Outcome{}, err
	}
	s, err := l.d15read(ctx)
	if err != nil {
		return drill.Outcome{}, err
	}
	dx, dy, n := l.d15dir(ctx, s.X, s.Y, 26)
	var obs []string
	landed := map[int]bool{}
	for _, d := range []int{3, 6, 9, 12, 15, 18, 24} {
		if d > n {
			obs = append(obs, fmt.Sprintf("%d cells: no free line (%d free)", d, n))
			continue
		}
		l.d15warp(ctx, r, s.X, s.Y)
		_ = l.d15mana(ctx, r, -1, "T2: full mana for each hop")
		_ = l.Tick(ctx, 3)
		tx, ty := s.X+dx*float64(d), s.Y+dy*float64(d)
		b, a, err := l.d15cast(ctx, r, skTeleport, tx, ty, fmt.Sprintf("T2: Teleport hop %d cells (%d subtiles) on open ground", d, 5*d))
		if err != nil {
			return drill.Outcome{}, err
		}
		land := cdist(a.X, a.Y, tx, ty) < 1.0
		landed[d] = land
		obs = append(obs, fmt.Sprintf("%d cells (%d subtiles) from (%.1f,%.1f) to (%.1f,%.1f): landed at the target %s, now (%.2f,%.2f), mana %.1f -> %.1f", d, 5*d, b.X, b.Y, tx, ty, yn(land), a.X, a.Y, float64(b.Mana)/256, float64(a.Mana)/256))
	}
	o := drill.Outcome{Intended: "game.exe: lands when the target is in the caster's room or its near list (40 subtiles = 8 cells on 8x8-tile rooms, caster's place decides); refuses at 25 cells (125 subtiles), mana kept", Observed: strings.Join(obs, " || ")}
	o.Sig = fmt.Sprint(landed)
	far := landed[24]
	switch {
	case len(landed) == 0:
		return drill.Blocks("no free line for any hop"), nil
	case landed[9] && !far:
		o.Verdict, o.Reason = drill.Pass, "the 9-cell hop lands and the 24-cell hop does not (matches the trace's room rule)"
	case landed[9] && far:
		o.Verdict, o.Reason = drill.Fail, "the 9-cell hop lands (as game.exe allows) and the 24-cell hop ALSO lands: d2d has no room limit (trace D1), game.exe refuses it"
	default:
		o.Verdict, o.Reason = drill.Fail, "the 9-cell hop did not land (the trace allows it when the room list touches)"
	}
	return o, nil
}

func (l *lab) d15T3(ctx context.Context, r *drill.Run) (drill.Outcome, error) {
	if err := l.d15prep(ctx, r, skTeleport, 1); err != nil {
		return drill.Outcome{}, err
	}
	s, err := l.d15read(ctx)
	if err != nil {
		return drill.Outcome{}, err
	}
	type cell struct{ x, y float64 }
	var cands []cell
	for rad := 3.0; rad <= 5.0; rad += 1 {
		for a := 0.0; a < 360; a += 15 {
			x, y := s.X+rad*math.Cos(a*math.Pi/180), s.Y+rad*math.Sin(a*math.Pi/180)
			if l.d15blocked(ctx, x, y) {
				cands = append(cands, cell{x, y})
			}
		}
	}
	if len(cands) == 0 {
		return drill.Outcome{}, fmt.Errorf("no blocked cell 3 to 5 cells from (%.1f,%.1f)", s.X, s.Y)
	}
	var picks []cell
	for _, c := range cands { // spread the picks
		ok := true
		for _, p := range picks {
			if cdist(c.x, c.y, p.x, p.y) < 3 {
				ok = false
			}
		}
		if ok && len(picks) < 3 {
			picks = append(picks, c)
		}
	}
	var obs []string
	moved := 0
	for _, c := range picks {
		l.d15warp(ctx, r, s.X, s.Y)
		_ = l.d15mana(ctx, r, -1, "T3: full mana for each blocked-target hop")
		_ = l.Tick(ctx, 3)
		b, a, err := l.d15cast(ctx, r, skTeleport, c.x, c.y, "T3: Teleport at a blocked cell (game.exe snaps to the nearest free spot)")
		if err != nil {
			return drill.Outcome{}, err
		}
		mv := cdist(a.X, a.Y, b.X, b.Y)
		if mv > 0.5 {
			moved++
		}
		obs = append(obs, fmt.Sprintf("blocked target (%.2f,%.2f) %.1f cells away: moved %.2f cells to (%.2f,%.2f), %.2f cells from the target, mana %.1f -> %.1f (cost %.1f)", c.x, c.y, cdist(c.x, c.y, b.X, b.Y), mv, a.X, a.Y, cdist(a.X, a.Y, c.x, c.y), float64(b.Mana)/256, float64(a.Mana)/256, float64(b.Mana-a.Mana)/256))
	}
	o := drill.Outcome{Intended: "game.exe: lands at the nearest free spot of the ring search (mask 0x1c09), mana paid at the action frame", Observed: strings.Join(obs, " || ")}
	o.Sig = fmt.Sprintf("moved %d of %d", moved, len(picks))
	if moved == len(picks) {
		o.Verdict, o.Reason = drill.Pass, "every blocked target ends in a hop to a nearby spot (snap), as game.exe does"
	} else {
		o.Verdict, o.Reason = drill.Fail, fmt.Sprintf("d2d did not snap in %d of %d blocked-target hops (game.exe snaps); the mana column says whether the refused hop was paid", len(picks)-moved, len(picks))
	}
	return o, nil
}

var spawnID = regexp.MustCompile(`#(\d+)`)

func (l *lab) d15spawn(ctx context.Context, r *drill.Run, name, why string) (int, float64, float64, error) {
	out, err := r.Cheat(ctx, "debug spawn "+name+" 1 near", why)
	if err != nil {
		return 0, 0, 0, err
	}
	m := spawnID.FindStringSubmatch(out)
	if m == nil {
		return 0, 0, 0, fmt.Errorf("no #id in %q", out)
	}
	id, _ := strconv.Atoi(m[1])
	_ = l.Tick(ctx, 2)
	x, y, ok := l.d15monster(ctx, id)
	if !ok {
		return id, 0, 0, fmt.Errorf("monster #%d not listed", id)
	}
	return id, x, y, nil
}

func (l *lab) d15monster(ctx context.Context, id int) (x, y float64, ok bool) {
	ms, err := l.be.Client.Monsters(ctx)
	if err != nil {
		return
	}
	for _, m := range ms.Items {
		if m.Unit == id {
			return m.CellX, m.CellY, true
		}
	}
	return
}

func (l *lab) d15T4(ctx context.Context, r *drill.Run) (drill.Outcome, error) {
	if err := l.d15prep(ctx, r, skTeleport, 1); err != nil {
		return drill.Outcome{}, err
	}
	s, err := l.d15read(ctx)
	if err != nil {
		return drill.Outcome{}, err
	}
	var obs []string
	var closest []float64
	for _, mode := range []string{"point", "unit"} {
		l.d15warp(ctx, r, s.X, s.Y)
		_, _ = r.Cheat(ctx, "debug kill", "T4: only the monster under test in the level")
		_ = l.Tick(ctx, 30)
		id, mx, my, err := l.d15spawn(ctx, r, "zombie1", "T4: a monster to teleport at")
		if err != nil {
			return drill.Outcome{}, err
		}
		_, _ = r.Cheat(ctx, fmt.Sprintf("debug mon set %d life 5000", id), "T4: the monster stays alive and in place for the read")
		_ = l.d15mana(ctx, r, -1, "T4: full mana")
		l.d15settle(ctx)
		b, _ := l.d15read(ctx)
		if mode == "point" {
			if _, err = l.skillDebug(r).CastAt(ctx, skTeleport, mx, my, "T4: Teleport at the monster's cell"); err != nil {
				return drill.Outcome{}, err
			}
		} else {
			if _, err = l.skillDebug(r).CastAtUnit(ctx, skTeleport, id, "T4: Teleport at the monster as a unit target"); err != nil {
				return drill.Outcome{}, err
			}
		}
		_ = l.Tick(ctx, 14)
		a, _ := l.d15read(ctx)
		mx2, my2, _ := l.d15monster(ctx, id)
		d := cdist(a.X, a.Y, mx2, my2)
		closest = append(closest, d)
		obs = append(obs, fmt.Sprintf("%s target: monster #%d at (%.2f,%.2f) when cast, (%.2f,%.2f) after; player (%.2f,%.2f) -> (%.2f,%.2f), moved %.2f cells, %.2f cells from the monster, mana %.1f -> %.1f", mode, id, mx, my, mx2, my2, b.X, b.Y, a.X, a.Y, cdist(a.X, a.Y, b.X, b.Y), d, float64(b.Mana)/256, float64(a.Mana)/256))
	}
	o := drill.Outcome{Intended: "game.exe: lands next to the monster, never on it (the landing spot's mask includes units, plus footprint), for a point and for a unit target", Observed: strings.Join(obs, " || ")}
	o.Sig = fmt.Sprintf("%.1f %.1f", closest[0], closest[1])
	onIt := closest[0] < 0.8 || closest[1] < 0.8
	switch {
	case onIt:
		o.Verdict, o.Reason = drill.Fail, fmt.Sprintf("the character ends %.2f / %.2f cells (point / unit) from the monster: on it (trace D3: d2d lands on monsters); game.exe lands beside it", closest[0], closest[1])
	default:
		o.Verdict, o.Reason = drill.Pass, fmt.Sprintf("the character ends %.2f / %.2f cells from the monster, not on it", closest[0], closest[1])
	}
	return o, nil
}

func (l *lab) d15T5(ctx context.Context, r *drill.Run) (drill.Outcome, error) {
	if err := l.d15prep(ctx, r, skTeleport, 1); err != nil {
		return drill.Outcome{}, err
	}
	if _, err := r.Cheat(ctx, "debug merc hire 0", "T5: a merc to see whether d2d moves it with a Teleport hop (no gold, throwaway)"); err != nil {
		return drill.Outcome{}, err
	}
	_ = l.Tick(ctx, 60)
	s, err := l.d15read(ctx)
	if err != nil {
		return drill.Outcome{}, err
	}
	dx, dy, n := l.d15dir(ctx, s.X, s.Y, 8)
	if n < 4 {
		return drill.Outcome{}, fmt.Errorf("no free line of 4 cells (best %d)", n)
	}
	mx0, my0, ok := l.d15merc(ctx)
	if !ok {
		return drill.Outcome{}, fmt.Errorf("no merc in dump after the hire")
	}
	hop := 3.0
	tx, ty := s.X+dx*hop, s.Y+dy*hop
	b, _ := l.d15read(ctx)
	if _, err = l.skillDebug(r).CastAt(ctx, skTeleport, tx, ty, "T5: a 3-cell (15 subtile) hop with the merc beside the player"); err != nil {
		return drill.Outcome{}, err
	}
	var rows []string
	landTick, jump := -1, false
	prev := cdist(mx0, my0, b.X, b.Y)
	rows = append(rows, fmt.Sprintf("+0 merc-player %.2f", prev))
	mx, my := mx0, my0
	for i := 1; i <= 45; i++ {
		if _, err := l.be.Client.Step(ctx, 1); err != nil {
			return drill.Outcome{}, err
		}
		p, _ := l.d15read(ctx)
		mx, my, _ = l.d15merc(ctx)
		d := cdist(mx, my, p.X, p.Y)
		if landTick < 0 && cdist(p.X, p.Y, b.X, b.Y) > 1 {
			landTick = i
			jump = d < 2.0 && prev > 2.0
		}
		if i <= 12 || i%6 == 0 {
			rows = append(rows, fmt.Sprintf("+%d player (%.2f,%.2f) merc (%.2f,%.2f) gap %.2f", i, p.X, p.Y, mx, my, d))
		}
		prev = d
	}
	moves := cdist(mx, my, mx0, my0)
	_, _ = r.Cheat(ctx, "debug merc kill", "T5: end of the merc test, the character is left without a merc")
	o := drill.Outcome{Intended: "game.exe: the merc (warp pets) is moved to the landing spot with the owner at once (read of FUN_00650be0; not confirmed by play)", Observed: fmt.Sprintf("merc start (%.2f,%.2f), player landed at tick +%d; gap player-merc at the sample before the landing %.2f; merc moved %.2f cells in 45 ticks || %s", mx0, my0, landTick, cdist(mx0, my0, b.X, b.Y), moves, strings.Join(rows, " | "))}
	o.Sig = fmt.Sprintf("land=%d jump=%v", landTick, jump)
	switch {
	case landTick < 0:
		return drill.Blocks("the character did not hop (no landing seen)"), nil
	case jump:
		o.Verdict, o.Reason = drill.Pass, "the merc stood next to the player on the first sample after the landing: d2d moved it with her"
	default:
		o.Verdict, o.Reason = drill.Fail, fmt.Sprintf("the merc did not jump with the owner (the gap on the landing tick was larger than 2 cells): d2d does not warp the merc (trace D7); it ran %.2f cells afterwards", moves)
	}
	return o, nil
}

func (l *lab) d15T7(ctx context.Context, r *drill.Run) (drill.Outcome, error) {
	if err := l.d15prep(ctx, r, skTeleport, 1); err != nil {
		return drill.Outcome{}, err
	}
	s, err := l.d15read(ctx)
	if err != nil {
		return drill.Outcome{}, err
	}
	dx, dy, n := l.d15dir(ctx, s.X, s.Y, 26)
	if n < 24 {
		return drill.Outcome{}, fmt.Errorf("no free line of 24 cells (best %d)", n)
	}
	var obs []string
	kept := false
	for _, d := range []float64{25, 4} {
		if d > float64(n) {
			continue
		}
		l.d15warp(ctx, r, s.X, s.Y)
		if err := l.d15mana(ctx, r, 24*256, "T7: exactly the 24 points of a level 1 Teleport"); err != nil {
			return drill.Outcome{}, err
		}
		_ = l.Tick(ctx, 3)
		tx, ty := s.X+dx*d, s.Y+dy*d
		b, a, err := l.d15cast(ctx, r, skTeleport, tx, ty, fmt.Sprintf("T7: a %.0f-cell hop with exactly 24 points", d))
		if err != nil {
			return drill.Outcome{}, err
		}
		mv := cdist(a.X, a.Y, b.X, b.Y)
		obs = append(obs, fmt.Sprintf("%.0f cells (%.0f subtiles): moved %.2f cells, mana %.2f -> %.2f points (14 ticks of regen included, +%.2f)", d, 5*d, mv, float64(b.Mana)/256, float64(a.Mana)/256, 14*3.0/256))
		if d == 25 && mv < 0.5 {
			kept = a.Mana >= b.Mana-256
		}
	}
	o := drill.Outcome{Intended: "game.exe: a hop it refuses (target outside the near rooms) pays nothing: mana stays 24; d2d (trace D4) pays 24", Observed: strings.Join(obs, " || ")}
	o.Sig = fmt.Sprint(kept)
	if len(obs) == 0 {
		return drill.Blocks("no free line"), nil
	}
	if strings.Contains(obs[0], "moved 0.") && kept {
		o.Verdict, o.Reason = drill.Pass, "the 25-cell hop was refused and the mana was kept"
	} else if strings.Contains(obs[0], "moved 0.") {
		o.Verdict, o.Reason = drill.Fail, "the 25-cell hop was refused and the mana was paid (trace D4: game.exe keeps it)"
	} else {
		o.Verdict, o.Reason = drill.Fail, "the 25-cell hop was NOT refused on d2d (no room limit, trace D1), so there is no refusal to price: the mana is spent on a hop game.exe would refuse"
	}
	return o, nil
}
