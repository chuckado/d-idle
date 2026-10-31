package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/d2dsrc"
	"d2lod-bot/internal/d2dtown"
	"d2lod-bot/internal/drill"
	"d2lod-bot/internal/input"
)

var _ drill.FieldLab = (*lab)(nil)

var numRe = map[string]*regexp.Regexp{}

func kv(line, key string) (int, bool) {
	re := numRe[key]
	if re == nil {
		re = regexp.MustCompile(`\b` + key + `=(-?\d+)`)
		numRe[key] = re
	}
	m := re.FindStringSubmatch(line)
	if m == nil {
		return 0, false
	}
	n, _ := strconv.Atoi(m[1])
	return n, true
}

// objectMode reads the mode column of `debug objects` for a preset index ("NU" untouched, "ON" operated).
func (l *lab) objectMode(ctx context.Context, r *drill.Run, idx int) string {
	out, err := r.Cheat(ctx, "debug objects", "read-only listing of the objects' mode column: the drill compares it before and after the bot's click")
	if err != nil {
		return "err " + err.Error()
	}
	for _, ln := range strings.Split(out, "\n") {
		f := strings.Fields(ln)
		if len(f) >= 7 && f[0] == strconv.Itoa(idx) {
			return f[len(f)-1]
		}
	}
	return "absent"
}

func (l *lab) OperateObject(ctx context.Context, r *drill.Run, level int, want string) drill.ObjOp {
	op := drill.ObjOp{Level: level, Want: want}
	if err := l.Goto(ctx, r, level); err != nil {
		op.InteractErr = "goto: " + err.Error()
		return op
	}
	_ = l.be.Ctl.CloseAll()
	ps, err := l.be.Client.Presets(ctx)
	if err != nil {
		op.InteractErr = "presets: " + err.Error()
		return op
	}
	var names []string
	for _, p := range ps.Items {
		names = append(names, p.Name)
		if op.Found == "" && !p.Hidden && strings.Contains(strings.ToLower(p.Name), want) {
			op.Found, op.Class = p.Name, p.ObjectID
			op.Names = strconv.Itoa(p.Index) // the index, replaced below by the names when nothing matched
		}
	}
	if level >= 30 { // Tristram has monsters: the throwaway must survive the walk to the cage
		_ = l.Invincible(ctx, r)
	}
	if op.Found == "" { // presets and `debug objects` only list the rooms near the player: look around the level from a few cells
		for _, at := range [][2]int{{22, 58}, {14, 45}, {10, 60}, {30, 60}, {40, 45}} {
			if op.Found != "" {
				break
			}
			if _, err := r.Cheat(ctx, fmt.Sprintf("debug warp %d %d", at[0], at[1]), "stand at a spot of the level so the rooms around it stream in: the bot would have walked there; the drill tests the click"); err != nil {
				break
			}
			_ = l.Tick(ctx, 40)
			out, _ := r.Cheat(ctx, "debug objects", "read-only listing: where an object of this kind stands on the level")
			for _, ln := range strings.Split(out, "\n") {
				f := strings.Fields(ln)
				if len(f) < 7 || !strings.Contains(strings.ToLower(f[3]), want) {
					continue
				}
				x, _ := strconv.ParseFloat(f[1], 64)
				y, _ := strconv.ParseFloat(f[2], 64)
				if _, err := r.Cheat(ctx, fmt.Sprintf("debug warp %d %d", int(x)+2, int(y)), "stand beside the object: the bot walks there in a real run, the drill tests the click"); err != nil {
					break
				}
				_ = l.Tick(ctx, 40)
				if ps2, err := l.be.Client.Presets(ctx); err == nil {
					for _, p := range ps2.Items {
						if !p.Hidden && p.X-x < 1 && x-p.X < 1 && p.Y-y < 1 && y-p.Y < 1 {
							op.Found, op.Class, op.Names = p.Name, p.ObjectID, strconv.Itoa(p.Index)
						}
					}
				}
				break
			}
		}
	}
	if op.Found == "" {
		op.Names = strings.Join(names, ",")
		return op
	}
	idx, _ := strconv.Atoi(op.Names)
	op.Names = ""
	s, err := l.be.Env.Snapshot()
	if err != nil {
		op.InteractErr = err.Error()
		return op
	}
	id := uint32(d2dsrc.NPCIDBase + idx)
	for _, u := range s.Objects {
		if u.ID == id {
			op.Listed = true
			tg := input.Target{Type: 2, ID: u.ID, Pos: input.Point{X: u.X, Y: u.Y}}
			op.Before = l.objectMode(ctx, r, idx)
			t0 := l.tick()
			if err := l.be.Ctl.Interact(tg); err != nil {
				op.InteractErr = err.Error()
			}
			_, _ = l.be.Client.Step(ctx, 250)
			op.Ticks = l.tick() - t0
			op.After = l.objectMode(ctx, r, idx)
			break
		}
	}
	_ = l.CloseAll(ctx)
	return op
}

func (l *lab) monstersAlive(ctx context.Context) (int, string) {
	ms, err := l.be.Client.Monsters(ctx)
	if err != nil {
		return 0, err.Error()
	}
	n := 0
	var parts []string
	for _, m := range ms.Items {
		if m.HP.Cur > 0 && m.Mode != "DD" && m.Mode != "DT" {
			n++
			parts = append(parts, fmt.Sprintf("%s#%d(%s %d/%d)", m.ClassID, m.Unit, m.Boss, m.HP.Cur, m.HP.Max))
		}
	}
	return n, strings.Join(parts, " ")
}

func (l *lab) Fight(ctx context.Context, r *drill.Run, s drill.FightSpec) drill.FightInfo {
	var fi drill.FightInfo
	fail := func(msg string, err error) drill.FightInfo {
		fi.Err = msg + ": " + err.Error()
		return fi
	}
	if sn, err := l.be.Env.Snapshot(); err == nil && sn.Player != nil && combat.IsDead(sn.Player) {
		if err := l.rig.Fighter.Respawn(ctx, 40*time.Second); err != nil { // Esc: the bot's own respawn, back in town
			return fail("respawn", err)
		}
		_ = l.Tick(ctx, 25)
	}
	if err := l.Goto(ctx, r, s.Level); err != nil {
		return fail("goto", err)
	}
	if sw, err := l.be.Env.Snapshot(); err == nil && sw.Player != nil { // the navigator builds the level's map at its first walk: take it before the monsters come (a Clear on a map-less arrival calls everything unreachable)
		wctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		_ = l.navr.GoTo(wctx, input.Point{X: sw.Player.X + 8, Y: sw.Player.Y})
		cancel()
	}
	if s.Kill {
		if _, err := r.Cheat(ctx, "debug kill", "clear the level's natural monsters (their experience is a side effect): the spawned ones are the fight under test, a crowd would trigger the chicken"); err != nil {
			return fail("kill", err)
		}
		_ = l.Tick(ctx, 60)
	}
	defer func() { // leave the field: monsters left behind would kill the next drill's character
		_, _ = r.Cheat(ctx, "debug goto 1", "back to the Rogue Encampment after the fight: the spawned monsters stay in the field")
		_ = l.Tick(ctx, 25)
	}()
	if s.Life > 0 {
		for _, v := range []string{fmt.Sprintf("debug stat 7 %d", s.Life*256), fmt.Sprintf("debug stat 6 %d", s.Life*256)} {
			if _, err := r.Cheat(ctx, v, fmt.Sprintf("life %d: the fight must not kill the throwaway (or, at 40, be the stock life of a fresh Sorceress)", s.Life)); err != nil {
				return fail("life", err)
			}
		}
		_ = l.Tick(ctx, 10)
	}
	if err := l.GiveSkills(ctx, r); err != nil {
		return fail("skills", err)
	}
	n, err := l.rig.BindSkills()
	if err != nil {
		return fail("bind", err)
	}
	fi.Bound = n
	if s.Potions {
		if _, err := l.ArmBelt(ctx, r); err != nil {
			return fail("belt", err)
		}
	}
	if s.Mana > 0 {
		if err := l.SetMana(ctx, r, s.Mana); err != nil {
			return fail("mana", err)
		}
	}
	for _, sp := range s.Spawns {
		if _, err := r.Cheat(ctx, "debug spawn "+sp, "a monster near the throwaway: the fight under test (docs/forkcheats.md)"); err != nil {
			return fail("spawn", err)
		}
	}
	_ = l.Tick(ctx, 5)
	fi.Spawned, fi.Mons = l.monstersAlive(ctx)
	if s.LifeAt > 0 || s.ManaAt > 0 { // a decision with monsters in the snapshot: no Clear (a chicken would leave the game), the Fighter's own Safety
		if s.LifeAt > 0 {
			if err := l.SetLife(ctx, r, s.LifeAt); err != nil {
				return fail("life", err)
			}
		}
		if s.ManaAt > 0 {
			if err := l.SetMana(ctx, r, s.ManaAt); err != nil {
				return fail("mana", err)
			}
		}
		l.be.Env.Clock.Sleep(1500 * time.Millisecond)
		_ = l.Tick(ctx, 30)
		snap, err := l.be.Env.Snapshot()
		if err != nil {
			return fail("snapshot", err)
		}
		_, mp0 := l.beltKinds(snap)
		fi.ManaStart, fi.LifeEnd = snap.Player.Mana, snap.Player.Life
		act := l.rig.Fighter.Safety(snap)
		fi.Decisions = append(fi.Decisions, fmt.Sprintf("safety at life %d mana %d with %d hostile monsters listed: %s", snap.Player.Life, snap.Player.Mana, fi.Spawned, act.String()))
		switch act {
		case combat.Chicken:
			fi.Err = "chicken decided (not run: it would leave the game)"
		case combat.SafetyNone:
			fi.Err = "none"
		default:
			if aerr := l.rig.Fighter.Act(ctx, act); aerr != nil {
				fi.Err = act.String() + " act err " + aerr.Error()
			} else {
				fi.Err = act.String() + " done"
			}
			_ = l.Tick(ctx, 60)
			if s2, e2 := l.be.Env.Snapshot(); e2 == nil && s2.Player != nil {
				fi.InTown = s2.Player.InTown()
				fi.LifeEnd, fi.ManaEnd = s2.Player.Life, s2.Player.Mana
				_, mp1 := l.beltKinds(s2)
				fi.Decisions = append(fi.Decisions, fmt.Sprintf("belt mana potions %d -> %d, mana %d -> %d", mp0, mp1, fi.ManaStart, s2.Player.Mana))
				if act == combat.DrinkMana && mp1 == mp0-1 && s2.Player.Mana > fi.ManaStart {
					fi.ManaDrinks++
				}
			}
		}
		fi.Alive, _ = l.monstersAlive(ctx)
		if !s.ThenClear {
			return fi
		}
	}
	if s.SampleMissiles { // does the bot's snapshot list the shamans' missiles at all? (idle character, 24 samples of 5 ticks)
		raw, snapMax := 0, 0
		for i := 0; i < 24; i++ {
			_ = l.Tick(ctx, 5)
			if ml, err := l.be.Client.Missiles(ctx); err == nil {
				n := 0
				for _, m := range ml.Items {
					if m.Speed > 0 && m.OwnerType == 1 {
						n++
					}
				}
				raw = max(raw, n)
			}
			if sn, err := l.be.Env.Snapshot(); err == nil {
				snapMax = max(snapMax, len(sn.Missiles))
			}
		}
		fi.Missiles = snapMax
		fi.Decisions = append(fi.Decisions, fmt.Sprintf("idle sampling: debug missiles hostile moving rows max %d, Snapshot.Missiles max %d", raw, snapMax))
	}
	sn0, err := l.be.Env.Snapshot()
	if err != nil || sn0.Player == nil {
		return fail("snapshot", errors.New("no player"))
	}
	fi.ManaStart, fi.LifeMin = sn0.Player.Mana, sn0.Player.Life
	_ = sn0
	fi.Decisions = append(fi.Decisions, fmt.Sprintf("pre-clear player mode %d area %d", sn0.Player.Mode, sn0.Player.Area))
	if rr, e := l.be.Client.Do(ctx, "dump player"); e == nil && rr != nil {
		fi.Decisions = append(fi.Decisions, "dump player: "+strings.Join(rr.Lines, " ")+rr.Rest)
	}
	cmd0 := l.skillCmds()
	mark := l.rec.Mark()
	to := s.Clear
	if to == 0 {
		to = 60 * time.Second
	}
	var res combat.ClearResult
	var cerr error
	if s.ManaMid > 0 { // engage first with full mana, then drop the mana by cheat in the middle of the fight
		r0, e0 := l.rig.Fighter.Clear(ctx, combat.ClearOpts{Timeout: 8 * time.Second})
		res.Kills += r0.Kills
		fi.Decisions = append(fi.Decisions, fmt.Sprintf("engaged: first Clear %v kills %d err %v", r0.Elapsed.Round(time.Second), r0.Kills, e0))
		if err := l.SetMana(ctx, r, s.ManaMid); err != nil {
			return fail("mana", err)
		}
		fi.ManaStart = s.ManaMid
	}
	for round := 0; round < 3; round++ { // the first call can run before the level map is built (everything "unreachable"): ask again
		var r1 combat.ClearResult
		r1, cerr = l.rig.Fighter.Clear(ctx, combat.ClearOpts{Timeout: to})
		res.Kills += r1.Kills
		res.Avoided = max(res.Avoided, r1.Avoided)
		res.Skipped = max(res.Skipped, r1.Skipped)
		res.Ignored = r1.Ignored
		res.Potions += r1.Potions
		res.Elapsed += r1.Elapsed
		if cerr != nil || res.Kills > 0 || r1.TimedOut {
			break
		}
		_ = l.Tick(ctx, 100)
		l.rig.Fighter.ForgetIgnored()
	}
	fi.Log = l.rec.Since(mark)
	fi.CmdSkill = l.skillCmds() - cmd0
	fi.Kills, fi.Avoided, fi.Skipped, fi.Ignored, fi.Potions, fi.Elapsed = res.Kills, res.Avoided, res.Skipped, len(res.Ignored), res.Potions, res.Elapsed
	for _, ig := range res.Ignored {
		fi.Ignores = append(fi.Ignores, fmt.Sprint(ig))
	}
	if cerr != nil {
		fi.Err = cerr.Error()
		if sn, e := l.be.Env.Snapshot(); e == nil && sn.Player != nil {
			fi.Err += fmt.Sprintf(" (player mode %d life %d/%d state %v)", sn.Player.Mode, sn.Player.Life, sn.Player.MaxLife, sn.State)
		}
	}
	telem := false
	for _, ln := range fi.Log {
		if !strings.HasPrefix(ln, "combat decision") && len(fi.LogHead) < 14 {
			fi.LogHead = append(fi.LogHead, ln)
		}
		switch {
		case strings.HasPrefix(ln, "combat fight area=") || strings.HasPrefix(ln, "fight area="):
			c, _ := kv(ln, "casts")
			d, _ := kv(ln, "dodges")
			k, _ := kv(ln, "kites")
			a, _ := kv(ln, "approaches")
			fi.Casts, fi.Dodges, fi.Kites, fi.Approaches = fi.Casts+c, fi.Dodges+d, fi.Kites+k, fi.Approaches+a
			telem = true
			if m, ok := kv(ln, "minlife"); ok {
				fi.LifeMin = m // percent
			}
		case strings.HasPrefix(ln, "combat decision"):
			fi.Decisions = append(fi.Decisions, strings.TrimPrefix(ln, "combat decision "))
		case strings.HasPrefix(ln, "combat potion "):
			fi.PotionLines = append(fi.PotionLines, ln)
			if strings.Contains(ln, "kind=mana") {
				fi.ManaDrinks++
			}
		}
		if strings.Contains(strings.ToLower(ln), "champion") {
			fi.Champion = ln
		}
	}
	if !telem {
		fi.Casts = countLog(fi.Decisions, "act=cast ")
	}
	if d := countLog(fi.Log, "combat dodging"); d > fi.Dodges {
		fi.Dodges = d
	}
	if fi.Champion == "" {
		fi.Champion = "no log line names it"
	}
	if sn, e := l.be.Env.Snapshot(); e == nil && sn.Player != nil {
		fi.LifeEnd, fi.ManaEnd, fi.InTown = sn.Player.Life, sn.Player.Mana, sn.Player.InTown()
	}
	fi.Alive, _ = l.monstersAlive(ctx)
	return fi
}

// skillCmds is the number of `cmd skill` verbs sent on the control client so far.
func (l *lab) skillCmds() int {
	for _, v := range l.be.Client.Trace() {
		if v.Verb == "cmd skill" {
			return int(v.N)
		}
	}
	return 0
}

func (l *lab) ToArea(ctx context.Context, level int) (int, string) {
	wctx, cancel := context.WithTimeout(ctx, 4*time.Minute)
	defer cancel()
	msg := ""
	if err := l.navr.ToArea(wctx, level); err != nil {
		msg = err.Error()
	}
	area := 0
	if s, err := l.be.Env.Snapshot(); err == nil && s.Player != nil {
		area = s.Player.Area
	}
	return area, msg
}

func (l *lab) TouchWaypoint(ctx context.Context, reach float64) (bool, []int, string) {
	n, _ := l.wpNavigator()
	n.DrainActivated()
	t, err := n.ActivateWaypointNear(ctx, reach)
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	return t, n.DrainActivated(), msg
}

func countLog(lines []string, sub string) int {
	n := 0
	for _, l := range lines {
		if strings.Contains(l, sub) {
			n++
		}
	}
	return n
}

func (l *lab) GivePackUnid(ctx context.Context, r *drill.Run, code string, quality int) error {
	_, err := r.Cheat(ctx, fmt.Sprintf("debug give %s pack q%d unid", code, quality), "an unidentified magic "+code+" in the pack: the identify and sell-price drills need one and the game has no other way to make it")
	if err != nil {
		return err
	}
	return l.Tick(ctx, 10)
}

func (l *lab) Claim(ctx context.Context, r *drill.Run) (bool, string, int, int, error) {
	if err := l.Goto(ctx, r, 1); err != nil {
		return false, "", 0, 0, err
	}
	if _, err := r.Cheat(ctx, "debug quest set 2 0x0002", "Blood Raven's record with the reward pending: the claim drill needs the state the quest would leave"); err != nil {
		return false, "", 0, 0, err
	}
	_ = l.Tick(ctx, 10)
	_ = l.be.Ctl.CloseAll()
	savedP, savedN := d2dctl.Pause, d2dctl.Now
	l.be.Clock.Install(ctx) // the bot's own wiring: a game-time Pause under lockstep
	l.be.Clock.Sleep(40 * time.Millisecond) // warm the new Pacer (see ApproachKashya)
	defer func() { d2dctl.Pause, d2dctl.Now = savedP, savedN }()
	res, err := d2dtown.ClaimRogue(ctx, l.be.Client)
	_ = l.CloseAll(ctx)
	return res.Granted, res.Merc, res.BurialBefore, res.BurialAfter, err
}

func (l *lab) AkaraReward(ctx context.Context, r *drill.Run) (int, int, string, error) {
	if err := l.Goto(ctx, r, 1); err != nil {
		return 0, 0, "", err
	}
	if _, err := r.Cheat(ctx, "debug quest set 1 0x0002", "the Den of Evil record with the reward pending: the Akara reward drill needs the state the quest would leave"); err != nil {
		return 0, 0, "", err
	}
	_ = l.Tick(ctx, 10)
	sk := func() int {
		if s, err := l.be.Env.Snapshot(); err == nil && s.Player != nil {
			return s.Player.SkillPoints
		}
		return -1
	}
	b := sk()
	if oi := l.OpenNPC(ctx, 148); oi.Err != "" {
		_ = l.CloseAll(ctx)
		return b, sk(), "", fmt.Errorf("open Akara: %s", oi.Err)
	}
	_ = l.Tick(ctx, 40)
	_ = l.CloseAll(ctx)
	_ = l.Tick(ctx, 20)
	out, _ := r.Cheat(ctx, "debug quest list", "read-only: the quest records after the talk")
	return b, sk(), strings.ReplaceAll(strings.TrimSpace(out), "\n", "; "), nil
}

func (l *lab) Pin(ctx context.Context, r *drill.Run, level int, within time.Duration) drill.PinInfo {
	var pi drill.PinInfo
	if err := l.Goto(ctx, r, level); err != nil {
		pi.Err = err.Error()
		return pi
	}
	if err := l.Invincible(ctx, r); err != nil {
		pi.Err = err.Error()
		return pi
	}
	s0, err := l.be.Env.Snapshot()
	if err != nil || s0.Player == nil {
		pi.Err = "no player"
		return pi
	}
	// the navigator loads the level's collision map at its first walk: take that walk before the pocket exists, so the
	// map does not know the wall and only the stall ladder can find out (a map that knows it answers "no path" at once)
	wctx0, cancel0 := context.WithTimeout(ctx, 40*time.Second)
	_ = l.navr.GoTo(wctx0, input.Point{X: s0.Player.X + 10, Y: s0.Player.Y})
	cancel0()
	if sw, err := l.be.Env.Snapshot(); err == nil && sw.Player != nil {
		s0 = sw
	}
	if _, err := r.Cheat(ctx, "debug pin on", "wall the throwaway into a collision pocket: the stall ladder drill (P25)"); err != nil {
		pi.Err = err.Error()
		return pi
	}
	pi.Pinned = true
	defer func() {
		if _, err := r.Cheat(ctx, "debug pin off", "take the pocket down again"); err == nil {
			pi.Freed = true
		}
	}()
	_ = l.Tick(ctx, 10)
	mark, t0 := l.rec.Mark(), l.tick()
	wctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	target := input.Point{X: s0.Player.X + 60, Y: s0.Player.Y}
	var werr error
	for l.tick()-t0 < int(within/(40*time.Millisecond)) { // game time: the ladder gets 60 s, tried in short walks
		werr = l.navr.GoTo(wctx, target)
		if s, err := l.be.Env.Snapshot(); err == nil && s.Player != nil && (s.Player.Area != level || math.Hypot(float64(s.Player.X-s0.Player.X), float64(s.Player.Y-s0.Player.Y)) > 30) {
			break
		}
		if wctx.Err() != nil {
			break
		}
	}
	pi.Ticks = l.tick() - t0
	if werr != nil {
		pi.Err = werr.Error()
	}
	if s, err := l.be.Env.Snapshot(); err == nil && s.Player != nil {
		pi.Area = s.Player.Area
		pi.Moved = math.Hypot(float64(s.Player.X-s0.Player.X), float64(s.Player.Y-s0.Player.Y)) / 5
	}
	for _, ln := range l.rec.Since(mark) {
		if strings.Contains(ln, "stuck") || strings.Contains(ln, "failsafe") || strings.Contains(ln, "ladder") || strings.Contains(ln, "portal") || strings.Contains(ln, "waypoint") {
			pi.Rungs = append(pi.Rungs, ln)
		}
	}
	return pi
}

func (l *lab) Survive(ctx context.Context, r *drill.Run) error {
	if sn, err := l.be.Env.Snapshot(); err == nil && sn.Player != nil && combat.IsDead(sn.Player) {
		if err := l.rig.Fighter.Respawn(ctx, 40*time.Second); err != nil {
			return err
		}
		_ = l.Tick(ctx, 25)
	}
	return l.Invincible(ctx, r)
}

func (l *lab) PlaceAtMouth(ctx context.Context, r *drill.Run, outer, cave, side int) (string, string) {
	rep, err := l.be.Client.Do(ctx, "debug warps")
	if err != nil || rep == nil {
		return "", "debug warps: " + fmt.Sprint(err)
	}
	var ex, ey float64
	found := false
	for _, ln := range rep.Lines {
		f := strings.Fields(ln)
		if len(f) >= 4 && f[3] == strconv.Itoa(cave) {
			ex, _ = strconv.ParseFloat(f[1], 64)
			ey, _ = strconv.ParseFloat(f[2], 64)
			found = true
		}
	}
	if !found {
		return "", fmt.Sprintf("no exit to level %d listed by debug warps (%s)", cave, strings.Join(rep.Lines, "; "))
	}
	// offsets in cells from the exit position, tried in this order for the asked side first
	all := [][2]float64{{0, 1.5}, {1.5, 0}, {0, -1.5}, {-1.5, 0}, {1, 1}, {-1, 1}}
	order := append([][2]float64{all[side%len(all)]}, all...)
	for _, o := range order {
		x, y := ex+o[0], ey+o[1]
		if _, err := r.Cheat(ctx, fmt.Sprintf("debug warp %d %d", int(x), int(y)), "stand beside a cave mouth inside its trigger area: the arrival the mouth bounce fix is about; the bot's own walk then leaves"); err != nil {
			return "", err.Error()
		}
		_ = l.Tick(ctx, 15)
		s, err := l.be.Env.Snapshot()
		if err != nil || s.Player == nil {
			continue
		}
		if s.Player.Area != outer {
			// the warp tile fired: back out and try another offset
			if _, err := r.Cheat(ctx, fmt.Sprintf("debug goto %d", outer), "the placement entered the cave: arrive in the outer level again"); err != nil {
				return "", err.Error()
			}
			_ = l.Tick(ctx, 25)
			continue
		}
		d := math.Hypot(float64(s.Player.X)/5-ex, float64(s.Player.Y)/5-ey)
		if d <= 1.8 {
			return fmt.Sprintf("exit to %d at (%.1f,%.1f), standing %.1f cells from it (offset %.1f,%.1f)", cave, ex, ey, d, o[0], o[1]), ""
		}
	}
	return "", "no offset left the character in the outer level inside the trigger range"
}

func (l *lab) Warriv(ctx context.Context, r *drill.Run) (int, string, error) {
	if err := l.Goto(ctx, r, 1); err != nil {
		return 0, "", err
	}
	if _, err := r.Cheat(ctx, "debug quest set 6 0x0001", "Andariel's record done: the Warriv travel drill needs the state the quest would leave"); err != nil {
		return 0, "", err
	}
	_ = l.Tick(ctx, 10)
	_ = l.be.Ctl.CloseAll()
	npcs, err := l.be.Client.NPCs(ctx)
	if err != nil {
		return 0, "", err
	}
	idx := -1
	for _, n := range npcs.Items {
		if strings.EqualFold(n.Name, "Warriv") {
			idx = n.Index
		}
	}
	if idx < 0 {
		return 0, "", errors.New("no Warriv in npcs")
	}
	savedP, savedN := d2dctl.Pause, d2dctl.Now
	l.be.Clock.Install(ctx)
	l.be.Clock.Sleep(40 * time.Millisecond)
	defer func() { d2dctl.Pause, d2dctl.Now = savedP, savedN }()
	lines, err := l.be.Client.OpenNPCMenu(ctx, idx, 40*time.Second)
	var parts []string
	for _, ml := range lines {
		parts = append(parts, ml.Text)
	}
	menu := strings.Join(parts, " | ")
	if err != nil {
		return 0, menu, err
	}
	if err := l.be.Client.PickMenuLine(ctx, "go east"); err != nil {
		_ = l.CloseAll(ctx)
		return 0, menu, err
	}
	_ = l.Tick(ctx, 200)
	area := 0
	if s, err := l.be.Env.Snapshot(); err == nil && s.Player != nil {
		area = s.Player.Area
	}
	_ = l.CloseAll(ctx)
	return area, menu, nil
}

func (l *lab) Warp(ctx context.Context, r *drill.Run, x, y int) error {
	_, err := r.Cheat(ctx, fmt.Sprintf("debug warp %d %d", x, y), "put the character at a known cell of the level: the walk the bot would have made to get there is not under test")
	return err
}
