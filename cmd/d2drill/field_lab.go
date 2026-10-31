package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2dsrc"
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
	if op.Found == "" { // presets only list the rooms near the player: find the object in `debug objects`, stand beside it, read the presets again
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
