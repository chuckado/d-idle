package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"d2lod-bot/internal/data"
	"d2lod-bot/internal/drill"
	"d2lod-bot/internal/input"
	"d2lod-bot/internal/nav"
)

// DRILLS22 (card drills22): the live proof of card fixtelekey. The leg is the bot's own Navigator.ExitTo with its
// Teleport capability wired as cmd/d2bot/telewire.go wires it (that function is in package main of another command, so
// d22wire repeats its four lines: TeleportKeyFn = Fighter.TeleportKey, TeleportOK from data.CanTeleportIn and the
// clear-line rule). Hops are counted from the navigator's own log lines ("nav teleport" at Debug, "nav teleport did
// not move" at Info) through the recorder. Tags: V = read from the log of the run, S = read from code, U = not measured.

const (
	d22Rocky    = 41
	d22Dry      = 42
	d22Lair     = 73
	d22Arcane   = 74
	d22Lut      = 40
	d22ManaPts  = 150
	d22ClvlWant = 18
)

// d22wire installs the Teleport capability on the navigator: wired=true is the production wiring, wired=false is the
// deliberately unwired variant (key but no level rule) used to provoke refused hops in a town.
func (l *lab) d22wire(nv *nav.Navigator, wired bool) {
	nv.TeleportKey = ""
	nv.TeleportKeyFn = l.rig.Fighter.TeleportKey
	nv.TeleportOK = nil
	if !wired || l.db == nil {
		return
	}
	nv.TeleportOK = func(level int) bool {
		a := l.db.AreaByID(level)
		return a == nil || data.CanTeleportIn(a) && a.TeleportRule() == data.TeleportAllowed
	}
}

type d22leg struct {
	Err                    error
	Hops, Refused          int
	ManaB, ManaA           float64
	Ticks                  int
	Wall                   time.Duration
	StartLvl, EndLvl       int
	StartX, StartY, EX, EY float64
	Lines                  []string
	Key                    string
}

func (g d22leg) String() string {
	return fmt.Sprintf("key=%q level %d->%d (%.0f,%.0f)->(%.0f,%.0f) hops=%d refused=%d mana %.1f->%.1f ticks=%d wall=%s err=%v",
		g.Key, g.StartLvl, g.EndLvl, g.StartX, g.StartY, g.EX, g.EY, g.Hops, g.Refused, g.ManaB, g.ManaA, g.Ticks, g.Wall.Round(time.Millisecond), g.Err)
}

func (l *lab) d22state(ctx context.Context) (lvl int, x, y, mana float64, err error) {
	s, err := l.be.Env.Snapshot()
	if err != nil || s.Player == nil {
		return 0, 0, 0, 0, fmt.Errorf("snapshot: %v", err)
	}
	raw, err := l.rawStats(ctx)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return s.Player.Area, float64(s.Player.X), float64(s.Player.Y), float64(raw[8]) / 256, nil
}

// d22prep warps to the level, clears the natural monsters, gives life, level 18 and a big mana pool.
func (l *lab) d22prep(ctx context.Context, r *drill.Run, level int, learn bool) error {
	if err := l.Goto(ctx, r, level); err != nil {
		return err
	}
	for _, v := range []struct{ verb, why string }{
		{"debug kill", "DRILLS22: clear the level's natural monsters: the leg measures navigation, not a fight"},
		{fmt.Sprintf("debug stat 12 %d", d22ClvlWant), "DRILLS22: character level 18, the level of the fixtelekey claim"},
		{fmt.Sprintf("debug stat 9 %d", d22ManaPts*256), "DRILLS22: max mana a clvl 18 Sorceress would have, so the leg is not mana-starved"},
		{fmt.Sprintf("debug stat 7 %d", 2000*256), "DRILLS22: life for the leg"},
		{fmt.Sprintf("debug stat 6 %d", 2000*256), "DRILLS22: life for the leg"},
	} {
		if _, err := r.Cheat(ctx, v.verb, v.why); err != nil {
			return err
		}
	}
	if learn {
		if _, err := r.Cheat(ctx, "debug points 54 1", "DRILLS22: Teleport learnt (level 1) at clvl 18 for the fixtelekey proof"); err != nil {
			return err
		}
	}
	if _, err := l.skillDebug(r).ManaFull(ctx, "DRILLS22: full mana at the start of the leg"); err != nil {
		return err
	}
	_ = l.Tick(ctx, 30)
	if _, err := l.rig.BindSkills(); err != nil {
		return err
	}
	return nil
}

// d22run runs ExitTo(to) from the current level and returns the counts.
func (l *lab) d22run(ctx context.Context, nv *nav.Navigator, to int, max time.Duration) d22leg {
	var g d22leg
	g.Key = nv.TeleportKeyFn()
	lvl, x, y, mana, err := l.d22state(ctx)
	if err != nil {
		g.Err = err
		return g
	}
	g.StartLvl, g.StartX, g.StartY, g.ManaB = lvl, x, y, mana
	mark := l.rec.Mark()
	t0, w0 := l.tick(), time.Now()
	cctx, cancel := context.WithTimeout(ctx, max)
	g.Err = nv.ExitTo(cctx, to)
	cancel()
	g.Ticks, g.Wall = l.tick()-t0, time.Since(w0)
	g.Lines = l.rec.Since(mark)
	for _, ln := range g.Lines {
		switch {
		case strings.HasPrefix(ln, "nav teleport did not move"):
			g.Refused++
		case strings.HasPrefix(ln, "nav teleport "):
			g.Hops++
		}
	}
	lvl, x, y, mana, err = l.d22state(ctx)
	if err == nil {
		g.EndLvl, g.EX, g.EY, g.ManaA = lvl, x, y, mana
	}
	return g
}

func (l *lab) d22dump(g d22leg, tag string) string {
	var hop []string
	for _, ln := range g.Lines {
		if strings.HasPrefix(ln, "nav teleport") {
			hop = append(hop, ln)
		}
	}
	if len(hop) > 12 {
		hop = append(hop[:6:6], hop[len(hop)-6:]...)
	}
	for _, h := range hop {
		fmt.Printf("D22 %s hop line: %s\n", tag, h)
	}
	fmt.Printf("D22 %s: %s\n", tag, g)
	return g.String()
}

func (l *lab) D22(ctx context.Context, r *drill.Run, id string) (o drill.Outcome) {
	var err error
	switch id {
	case "D22-C":
		o, err = l.d22C(ctx, r)
	case "D22-A":
		o, err = l.d22A(ctx, r)
	case "D22-B":
		o, err = l.d22B(ctx, r)
	case "D22-T":
		o, err = l.d22T(ctx, r)
	case "D22-R":
		o, err = l.d22R(ctx, r)
	default:
		err = fmt.Errorf("unknown drill %s", id)
	}
	if err != nil {
		return drill.Blocks(err.Error())
	}
	return o
}

func (l *lab) d22C(ctx context.Context, r *drill.Run) (drill.Outcome, error) {
	if err := l.d22prep(ctx, r, d22Rocky, false); err != nil {
		return drill.Outcome{}, err
	}
	l.d22wire(l.navr, true)
	g := l.d22run(ctx, l.navr, d22Dry, 150*time.Second)
	obs := l.d22dump(g, "control")
	o := drill.Outcome{Intended: "no Teleport learned: key empty, no hop, the leg walks", Observed: obs, Sig: fmt.Sprintf("hops=%d key=%q", g.Hops, g.Key)}
	switch {
	case g.Key != "" || g.Hops != 0 || g.Refused != 0:
		o.Verdict, o.Reason = drill.Fail, "the control hopped or had a key"
	case g.Err != nil:
		o.Verdict, o.Reason = drill.Fail, "the walking control failed: "+g.Err.Error()
	default:
		o.Verdict, o.Reason = drill.Pass, fmt.Sprintf("walked %d ticks, no key, no hops, no error", g.Ticks)
	}
	return o, nil
}

func (l *lab) d22A(ctx context.Context, r *drill.Run) (drill.Outcome, error) {
	if err := l.d22prep(ctx, r, d22Rocky, true); err != nil {
		return drill.Outcome{}, err
	}
	l.d22wire(l.navr, true)
	g := l.d22run(ctx, l.navr, d22Dry, 150*time.Second)
	obs := l.d22dump(g, "teleport")
	o := drill.Outcome{Intended: "Teleport learned at clvl 18 in Act 2: the leg hops (>0), refused hops are few, mana is paid per hop", Observed: obs, Sig: fmt.Sprintf("hops=%d refused=%d", g.Hops, g.Refused)}
	paid := g.ManaB - g.ManaA
	switch {
	case g.Key == "":
		o.Verdict, o.Reason = drill.Fail, "no Teleport key after learning and BindSkills"
	case g.Hops == 0:
		o.Verdict, o.Reason = drill.Fail, "no hop fired on an Act 2 leg"
	case g.Err != nil:
		o.Verdict, o.Reason = drill.Fail, "the leg failed: "+g.Err.Error()
	case g.Refused > 3:
		o.Verdict, o.Reason = drill.Fail, fmt.Sprintf("%d refused hops", g.Refused)
	default:
		o.Verdict, o.Reason = drill.Pass, fmt.Sprintf("%d hops, %d refused, mana net %.1f, %d ticks", g.Hops, g.Refused, paid, g.Ticks)
	}
	return o, nil
}

func (l *lab) d22B(ctx context.Context, r *drill.Run) (drill.Outcome, error) {
	if err := l.d22prep(ctx, r, d22Rocky, true); err != nil {
		return drill.Outcome{}, err
	}
	l.d22wire(l.navr, true)
	s, err := l.be.Env.Snapshot()
	if err != nil || s.Player == nil {
		return drill.Outcome{}, fmt.Errorf("snapshot: %v", err)
	}
	var sks []string
	for _, sk := range s.Player.Skills {
		sks = append(sks, fmt.Sprintf("%d:%d", sk.ID, sk.Level))
	}
	key := l.rig.Fighter.TeleportKey()
	states, _ := l.be.Client.Do(ctx, "debug states")
	// the hop through the key: one ExitTo hop count on a short leg is not needed, a single CastAt through the navigator's input is
	_, x0, y0, m0, _ := l.d22state(ctx)
	castErr := error(nil)
	if key != "" {
		castErr = l.be.Env.In.CastAt(key, input.Point{X: int(x0) + 12, Y: int(y0)})
		_ = l.Tick(ctx, 14)
	}
	_, x1, y1, m1, _ := l.d22state(ctx)
	moved := cdist(x0, y0, x1, y1)
	obs := fmt.Sprintf("TeleportKey=%q HasTeleport=%v skills=[%s] debug states: %s; CastAt(%q) 12 cells east: err=%v moved %.1f cells, mana %.1f->%.1f", key, l.navr.HasTeleport(), strings.Join(sks, " "), strings.TrimSpace(states.Text()), key, castErr, moved, m0, m1)
	fmt.Println("D22 bind:", obs)
	o := drill.Outcome{Intended: "key is the bind of skill 54, the cast through it hops and pays mana", Observed: obs, Sig: key}
	if key == fmt.Sprintf("sk%d", 54) && castErr == nil && moved >= 2 && m1 < m0 {
		o.Verdict, o.Reason = drill.Pass, fmt.Sprintf("bind %s, cast moved %.1f cells and paid %.1f mana", key, moved, m0-m1)
	} else if key != "sk54" {
		o.Verdict, o.Reason = drill.Fail, "TeleportKey is not sk54"
	} else {
		o.Verdict, o.Reason = drill.Fail, "the cast through the bind did not hop (blocked cell or not bound in the input)"
	}
	return o, nil
}

func (l *lab) d22T(ctx context.Context, r *drill.Run) (drill.Outcome, error) {
	var parts []string
	// rules by level (S: Levels.txt through data.CanTeleportIn and TeleportRule)
	l.d22wire(l.navr, true)
	for _, lv := range []int{d22Lut, d22Rocky, d22Lair, d22Arcane} {
		parts = append(parts, fmt.Sprintf("level %d allowed=%v", lv, l.navr.TeleportOK(lv)))
	}
	ok := !l.navr.TeleportOK(d22Lut) && l.navr.TeleportOK(d22Rocky) && !l.navr.TeleportOK(d22Lair)
	// town leg: Lut Gholein to Rocky Waste with the wired navigator and Teleport learned (the earlier drills learnt it)
	if err := l.d22prep(ctx, r, d22Lut, true); err != nil {
		return drill.Outcome{}, err
	}
	g := l.d22run(ctx, l.navr, d22Rocky, 100*time.Second)
	obs := l.d22dump(g, "town")
	parts = append(parts, obs)
	// the Summoner's level: reachable by goto only; a Teleport stand point needs the boss runner (not reachable from a drill)
	if _, err := l.Cheat(ctx, "drills:D22-T", fmt.Sprintf("debug goto %d", d22Arcane), "DRILLS22: is the Arcane Sanctuary reachable by goto (no boss fight is forced)"); err != nil {
		parts = append(parts, "goto 74: "+err.Error())
	} else {
		_ = l.Tick(ctx, 30)
		lv, _, _, _, _ := l.d22state(ctx)
		parts = append(parts, fmt.Sprintf("goto 74 -> area %d, HasTeleport=%v", lv, l.navr.HasTeleport()))
	}
	fmt.Println("D22 town/rules:", strings.Join(parts, " || "))
	o := drill.Outcome{Intended: "no hop in town (level rule), Duriel's Lair left to walking, Summoner level allows hops", Observed: strings.Join(parts, " || "), Sig: fmt.Sprintf("hops=%d", g.Hops)}
	if ok && g.Hops == 0 && g.Refused == 0 {
		o.Verdict, o.Reason = drill.Pass, "town leg: 0 hops, 0 refused"
	} else {
		o.Verdict, o.Reason = drill.Fail, fmt.Sprintf("rules ok=%v town hops=%d refused=%d err=%v", ok, g.Hops, g.Refused, g.Err)
	}
	return o, nil
}

func (l *lab) d22R(ctx context.Context, r *drill.Run) (drill.Outcome, error) {
	if err := l.d22prep(ctx, r, d22Lut, true); err != nil {
		return drill.Outcome{}, err
	}
	l.d22wire(l.navr, false)
	defer l.d22wire(l.navr, true)
	g := l.d22run(ctx, l.navr, d22Rocky, 100*time.Second)
	obs := l.d22dump(g, "unwired-town")
	o := drill.Outcome{Intended: "the game refuses every hop in a town: the navigator gives up after a few (teleportBadHops 3) and walks", Observed: obs, Sig: fmt.Sprintf("hops=%d refused=%d", g.Hops, g.Refused)}
	paid := g.ManaB - g.ManaA
	if g.Refused <= 6 && !errors.Is(g.Err, context.DeadlineExceeded) {
		o.Verdict, o.Reason = drill.Pass, fmt.Sprintf("%d refused hops, %d landed, mana net %.1f, leg err=%v", g.Refused, g.Hops, paid, g.Err)
	} else {
		o.Verdict, o.Reason = drill.Fail, fmt.Sprintf("%d refused hops (bound 6), err=%v", g.Refused, g.Err)
	}
	return o, nil
}
