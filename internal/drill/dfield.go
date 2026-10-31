package drill

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ObjOp is one field object operated the way the bot does it (snapshot object, Interact, the mode read back).
type ObjOp struct {
	Level       int
	Want        string // the preset name searched for
	Found       string // the preset name matched ("" none)
	Names       string // the preset names seen on the level (when nothing matched)
	Class       int    // objects.txt id of the row
	Listed      bool   // the bot's snapshot has the object (Snapshot.Objects, id = NPCIDBase + index)
	InteractErr string
	Before      string // `debug objects` mode column before and after ("NU" untouched, "ON" operated)
	After       string
	Ticks       int
}

// FightSpec is one scripted fight: a level, a spawn list, the character's life and mana set, the fight's time limit.
type FightSpec struct {
	Level          int
	Spawns         []string // arguments of `debug spawn`, e.g. "fallenshaman1 2 near"
	Life           int      // maximum and current life by cheat (0: leave the character as is)
	Mana           int      // current mana by cheat (0: leave)
	ManaAt         int      // like LifeAt for mana
	ManaMid        int      // mana set by cheat after a first 8 s Clear: the mana drink inside a fight
	ThenClear      bool     // after the decision, run Clear too (hit effect and kills)
	SampleMissiles bool     // before the fight, sample debug missiles and Snapshot.Missiles with the character idle
	Kill           bool     // debug kill after the arrival: the level's natural monsters are gone (they give experience)
	Potions        bool     // belt holds healing and mana potions
	Cold           bool     // no warm-up walk after the arrival and one Clear only: the first Clear on a map the navigator has not walked (fixmisc)
	Clear          time.Duration
	// LifeAt, when set, lowers the life to this value as soon as the first spawned monster is within reach (P29 retreat)
	LifeAt int
}

// FightInfo is what Fighter.Clear and the log of the call report.
type FightInfo struct {
	Err                                       string
	Kills, Avoided, Skipped, Ignored, Potions int
	Casts, Dodges, Kites, Approaches          int
	CmdSkill                                  int // `cmd skill` verbs sent during the call
	LifeMin, LifeEnd, ManaStart, ManaEnd      int
	Spawned, Alive                            int // monsters listed after the spawn / still alive after the call
	Elapsed                                   time.Duration
	Missiles                                  int // missiles of other owners seen listed (debug missiles) at the start of the fight, sampled once per 10 ticks before the call
	Log                                       []string
	Decisions                                 []string // "act: why" of the combat decision lines, in order
	Champion                                  string   // what the combat layer did with the champion (log), "" none
	InTown, Dead                              bool
	Bound                                     int
	PotionLines                               []string
	ManaDrinks                                int
	Ignores                                   []string
	Mons                                      string
	LogHead                                   []string // the first non-decision log lines of the call
}

func (f FightInfo) String() string {
	return f.base() + "; log: " + strings.Join(f.LogHead, " / ")
}

func (f FightInfo) base() string {
	return fmt.Sprintf("kills %d avoided %d skipped %d ignored %d potions %d; casts %d cmd-skill %d dodges %d kites %d approaches %d; life min %d end %d mana %d->%d; spawned %d alive %d in %v err %q town=%v dead=%v",
		f.Kills, f.Avoided, f.Skipped, f.Ignored, f.Potions, f.Casts, f.CmdSkill, f.Dodges, f.Kites, f.Approaches, f.LifeMin, f.LifeEnd, f.ManaStart, f.ManaEnd, f.Spawned, f.Alive, f.Elapsed.Round(time.Second), f.Err, f.InTown, f.Dead)
}

// FieldLab is the instance as the field drills (objects and fights) use it.
type FieldLab interface {
	HoverLab
	WPLab
	// ToArea is the bot's navigator ToArea (an exit walk); the area the character stands in afterwards.
	// Warp is the logged cheat `debug warp x y` (cells).
	Warp(ctx context.Context, r *Run, x, y int) error
	// Survive respawns a dead character (the bot's Respawn) and gives a big life: the walks cross monster levels without combat.
	Survive(ctx context.Context, r *Run) error
	// PlaceAtMouth puts the character (logged `debug warp`) just outside the trigger of the level's exit to cave, on the
	// side-th side, and describes where it stands; err is empty when it stands in the outer level within 1.8 cells of the exit.
	PlaceAtMouth(ctx context.Context, r *Run, outer, cave, side int) (desc, err string)
	ToArea(ctx context.Context, level int) (area int, err string)
	// TouchWaypoint is nav.ActivateWaypointNear on the current level with the given reach in subtiles; drained is what
	// the navigator hands the runner (DrainActivated).
	TouchWaypoint(ctx context.Context, reach float64) (touched bool, drained []int, err string)
	// OperateObject runs the bot's object path on the first level preset whose name contains want.
	OperateObject(ctx context.Context, r *Run, level int, want string) ObjOp
	// Fight spawns the monsters, binds skills and runs Fighter.Clear with the bot's own decisions.
	Fight(ctx context.Context, r *Run, s FightSpec) FightInfo
	// Quest steps for the D8 records (debug quest) are driven by QuestLab when the lab has them.
}

// F returns the field drills of DRILLS5: operating field objects and fighting with `debug spawn`.
func F(l FieldLab) []Drill {
	lim := 8 * time.Minute
	obj := func(id, title, level, want string) Drill {
		lv := 2
		fmt.Sscanf(level, "%d", &lv)
		return Drill{ID: id, Primitive: "AUDIT", Title: title, Limit: lim, Do: func(ctx context.Context, r *Run) Outcome { return fObj(ctx, r, l, lv, want) }}
	}
	fight := func(id, prim, title string, f func(context.Context, *Run, FieldLab) Outcome) Drill {
		return Drill{ID: id, Primitive: prim, Title: title, Limit: 10 * time.Minute, Do: func(ctx context.Context, r *Run) Outcome { return f(ctx, r, l) }}
	}
	var run []Drill
	if rl, ok := l.(RunLab); ok {
		run = append(run, DRun(rl))
	}
	return append(run, []Drill{
		fight("D-sync", "P27", "fixmisc: the first Fighter.Clear right after a `debug goto`, no warm-up walk, against a weak pack: it attacks (no `no path and no firing position`)", fSync),
		obj("H.obj.chest", "a chest of Blood Moor: listed in the snapshot, operated by the bot's Interact", "2", "chest"),
		obj("H.obj.shrine", "the shrine of Blood Moor: listed, operated by the bot's Interact", "2", "shrine"),
		{ID: "DW.m1", Primitive: "P23", Title: "fix-nav: standing in the Cave Level 1 mouth trigger (south side) in Cold Plains, waypoint to town: at most 2 bounces", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome {
			return fMouth(ctx, r, l, 3, 9, 0, "Cold Plains, mouth of Cave Level 1, south side")
		}},
		{ID: "DW.m2", Primitive: "P23", Title: "fix-nav: standing in the Cave Level 1 mouth trigger (east side) in Cold Plains, waypoint to town: at most 2 bounces", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome {
			return fMouth(ctx, r, l, 3, 9, 1, "Cold Plains, mouth of Cave Level 1, east side")
		}},
		{ID: "DW.a12", Primitive: "P23", Title: "fix-nav: an unactivated waypoint (Cold Plains) touched in passing is handed to the runner and works as a travel target", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome { return fTouch(ctx, r, l) }},
		fight("D5.dodge", "P27", "fix-missile: a ranged monster fires at the character: the fight loop dodges, one cast command per cast", fDodge),
		fight("D5.p27", "P27", "target selection: a pack with a champion at clvl 1, the bot's choice", fP27),
		fight("D5.p29", "P29", "retreat: a pack at low life, the bot's chicken and heal decision", fP29),
		fight("D5.p28", "P28", "mana potion in a fight, P26 hit effect on a target", fP28),
		obj("H.obj.cage", "Cain's gibbet in Tristram: listed (class 26), operated by the bot's Interact", "38", "gibbet"),
	}...)
}

func fObj(ctx context.Context, r *Run, l FieldLab, level int, want string) Outcome {
	op := l.OperateObject(ctx, r, level, want)
	intended := fmt.Sprintf("a %q of level %d is listed in Snapshot.Objects and Interact operates it (mode NU -> ON)", want, level)
	obs := fmt.Sprintf("preset %q class %d listed=%v; Interact err %q; mode %s -> %s; %d ticks", op.Found, op.Class, op.Listed, op.InteractErr, op.Before, op.After, op.Ticks)
	o := Outcome{Intended: intended, Observed: obs, Sig: fmt.Sprintf("found=%v listed=%v err=%v after=%s", op.Found != "", op.Listed, op.InteractErr != "", op.After)}
	switch {
	case op.Found == "":
		o.Verdict, o.Reason = Blocked, "no preset named "+want+" on the level; names seen: "+op.Names
	case !op.Listed:
		o.Verdict, o.Reason = Fail, "the object is not in the bot's snapshot (d2dsrc fieldObjects)"
	case op.InteractErr != "":
		o.Verdict, o.Reason = Fail, "Interact failed: "+op.InteractErr
	case want == "gibbet" && op.Class == 26 && op.After == op.Before:
		o.Verdict, o.Reason = Pass, "PARTIAL: listed with the class the bot expects (26, quest.ObjCainGibbet) and Interact accepted; the cage opens only with the Cain quest state, so its mode stays "+op.After+" (effect UNVERIFIED)"
	case op.Before == "ON" && op.After == "ON":
		o.Verdict, o.Reason = Pass, "listed and Interact accepted; the object was already operated by an earlier repeat (mode ON)"
	case op.After == op.Before:
		o.Verdict, o.Reason = Fail, "Interact returned but the object's mode did not change ("+op.After+")"
	default:
		o.Verdict, o.Reason = Pass, "listed and operated by the bot's own Interact"
	}
	return o
}

func countLines(lines []string, subs ...string) int {
	n := 0
	for _, l := range lines {
		ok := true
		for _, s := range subs {
			if !strings.Contains(l, s) {
				ok = false
			}
		}
		if ok {
			n++
		}
	}
	return n
}

// fDodge is the fix-missile and fix-missile2 live proof (D5 dodge): Fallen Shamans (firebolt-like shafire missiles) beside
// a character with plenty of life; the fight loop must dodge at least once and send about one cast command per cast.
func fDodge(ctx context.Context, r *Run, l FieldLab) Outcome {
	// a shaman's bolt is slow and the bot may kill it before one comes close: fight up to 3 rounds (a fresh arrival and
	// a fresh spawn each) until a dodge was seen; the counters add up
	var fi FightInfo
	rounds := 0
	for rounds < 3 {
		rounds++
		f := l.Fight(ctx, r, FightSpec{Level: 3, Kill: true, Spawns: []string{"fallenshaman1 4 near"}, SampleMissiles: rounds == 1, Life: 400, Potions: true, Clear: 90 * time.Second})
		fi.Casts += f.Casts
		fi.Dodges += f.Dodges
		fi.CmdSkill += f.CmdSkill
		fi.Kills += f.Kills
		fi.Spawned = max(fi.Spawned, f.Spawned)
		fi.LifeMin = f.LifeMin
		fi.Err = f.Err
		fi.Elapsed += f.Elapsed
		if rounds == 1 {
			fi.Decisions, fi.LogHead = f.Decisions, f.LogHead
			fi.Missiles = f.Missiles
		}
		fi.Log = append(fi.Log, f.Log...)
		if fi.Dodges > 0 && fi.Casts >= 5 {
			break
		}
	}
	per := 0.0
	if fi.Casts > 0 {
		per = float64(fi.CmdSkill) / float64(fi.Casts)
	}
	intended := "dodges > 0 with the d2d missile listing, cast commands per cast about 1.0 (CastGapMs 650)"
	obs := fmt.Sprintf("%d round(s); %s; cmd skill per cast %.2f; dodging lines %d; %s", rounds, fi.base(), per, countLines(fi.Log, "combat dodging"), strings.Join(firstN(fi.Decisions, 3), " | "))
	o := Outcome{Intended: intended, Observed: obs, Sig: fmt.Sprintf("dodge>0=%v perCast<=1.3=%v", fi.Dodges > 0, per > 0 && per <= 1.3)}
	switch {
	case fi.Spawned == 0:
		o.Verdict, o.Reason = Blocked, "the spawn listed no monster"
	case fi.Casts == 0:
		o.Verdict, o.Reason = Fail, "the fighter cast nothing"
	case fi.Dodges == 0:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("no dodge in %d rounds (casts %d)", rounds, fi.Casts)
	case per > 1.3:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("cast commands per cast %.2f (wanted about 1.0 or fewer): the bot resends casts", per)
	default:
		o.Verdict, o.Reason = Pass, fmt.Sprintf("%d dodges, %.2f cast commands per cast over %d casts (above 1.3 would be resends; below 1.0 the fighter counted casts it did not send as commands)", fi.Dodges, per, fi.Casts)
	}
	return o
}

func fP27(ctx context.Context, r *Run, l FieldLab) Outcome {
	fi := l.Fight(ctx, r, FightSpec{Level: 4, Kill: true, Spawns: []string{"zombie1 2 near", "fallen1 1 near champion"}, Life: 200, Potions: true, Clear: 90 * time.Second})
	intended := "the bot fights the nearest reachable threat and avoids or handles the champion by its danger table (recorded, no fixed verdict)"
	obs := fmt.Sprintf("%s; champion: %s; ignores [%s]; decisions %s", fi, fi.Champion, strings.Join(fi.Ignores, "; "), strings.Join(firstN(fi.Decisions, 12), " | "))
	o := Outcome{Intended: intended, Observed: obs, Sig: fmt.Sprintf("acted=%v", fi.Kills > 0 || fi.Avoided > 0)}
	switch {
	case fi.Spawned == 0:
		o.Verdict, o.Reason = Blocked, "the spawn listed no monster"
	case fi.Kills == 0 && fi.Avoided == 0:
		o.Verdict, o.Reason = Fail, "the fight neither killed nor avoided anything: "+fi.Err
	default:
		o.Verdict, o.Reason = Pass, fmt.Sprintf("kills %d avoided %d ignored %d (champion: %s)", fi.Kills, fi.Avoided, fi.Ignored, fi.Champion)
	}
	return o
}

func fP29(ctx context.Context, r *Run, l FieldLab) Outcome {
	fi := l.Fight(ctx, r, FightSpec{Level: 2, Spawns: []string{"zombie1 6 near"}, Life: 40, LifeAt: 9, Potions: true, Clear: 60 * time.Second})
	intended := "life at 9 of 40 with a pack in reach and an empty belt: Fighter.Safety decides a retreat (town portal or chicken) instead of fighting on"
	obs := fmt.Sprintf("%s; decisions %s", fi, strings.Join(firstN(fi.Decisions, 12), " | "))
	o := Outcome{Intended: intended, Observed: obs, Sig: fmt.Sprintf("acted=%v", fi.Err != "none" && fi.Err != "")}
	switch {
	case fi.Spawned == 0:
		o.Verdict, o.Reason = Blocked, "the spawn listed no monster"
	case strings.Contains(fi.Err, "drink-life done") || strings.Contains(fi.Err, "drink-rejuv done"):
		o.Verdict, o.Reason = Pass, "the bot decided to heal with the belt (it holds potions): "+fi.Err
	case strings.Contains(fi.Err, "town-portal") || strings.Contains(fi.Err, "chicken") || strings.Contains(fi.Err, "crowd") || fi.InTown:
		o.Verdict, o.Reason = Pass, "the bot decided to retreat: "+fi.Err
	default:
		o.Verdict, o.Reason = Fail, "no retreat at 9 of 40 life: "+fi.Err
	}
	return o
}

func fP28(ctx context.Context, r *Run, l FieldLab) Outcome {
	// up to 3 rounds (a fresh arrival and spawn each) until the fighter drinks a mana potion; the rounds are reported
	var fi FightInfo
	rounds, drank := 0, 0
	// the resource policy is poor-aware (internal/resource/potion.go: a poor character with a melee fallback and a weak
	// pack rests instead of drinking mana): the first half is rich, the control after it is poor
	if err := l.SetGold(ctx, r, 5000); err != nil {
		return setupFail(err)
	}
	for rounds < 3 {
		rounds++
		f := l.Fight(ctx, r, FightSpec{Level: 4, Spawns: []string{"zombie1 20 near"}, Life: 1000, Potions: true, Clear: 150 * time.Second})
		fi.Kills += f.Kills
		fi.Casts += f.Casts
		fi.CmdSkill += f.CmdSkill
		fi.ManaDrinks += f.ManaDrinks
		fi.Spawned = max(fi.Spawned, f.Spawned)
		fi.Err = f.Err
		fi.Elapsed += f.Elapsed
		if f.ManaDrinks > 0 {
			drank++
		}
		if rounds == 1 {
			fi.Decisions, fi.LogHead = f.Decisions, f.LogHead
		}
		if drank > 0 && fi.Kills > 0 {
			break
		}
	}
	poor := ""
	if err := l.SetGold(ctx, r, 0); err == nil {
		pf := l.Fight(ctx, r, FightSpec{Level: 4, Spawns: []string{"zombie1 20 near"}, Life: 1000, Potions: true, Clear: 150 * time.Second})
		poor = fmt.Sprintf("control with gold 0: %d mana drinks, %d kills", pf.ManaDrinks, pf.Kills)
	}
	intended := "a fight with a pack of 20 zombies drains the mana of a character with gold: the fighter drinks a belt mana potion in the fight ('combat potion kind=mana') and the casts kill the targets (P26 hit effect)"
	obs := fmt.Sprintf("%d round(s) with 5000 gold, %d with a mana drink; %s; mana drinks %d; %s; decisions %s", rounds, drank, fi.base(), fi.ManaDrinks, poor, strings.Join(firstN(fi.Decisions, 1), " | "))
	o := Outcome{Intended: intended, Observed: obs, Sig: fmt.Sprintf("kills>0=%v drank=%v", fi.Kills > 0, fi.ManaDrinks > 0)}
	switch {
	case fi.Spawned == 0:
		o.Verdict, o.Reason = Blocked, "the spawn listed no monster"
	case fi.Kills == 0:
		o.Verdict, o.Reason = Fail, "no kill: the hit effect of the cast is not seen: "+fi.Err
	case fi.ManaDrinks == 0:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("the mana ran down in %d fights and no mana potion was drunk (what it did instead: see the log)", rounds)
	default:
		o.Verdict, o.Reason = Pass, fmt.Sprintf("%d mana drink(s) in %d of %d fights, %d kills", fi.ManaDrinks, drank, rounds, fi.Kills)
	}
	return o
}

func firstN(s []string, n int) []string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// fMouth is DW.m1 and DW.m2: the character walks out of a cave into the level (the bot's own ToArea, so it arrives
// where d2d puts it, inside the exit's trigger), then takes the waypoint to town. The bounces are the log's "level
// changed during the walk" errors; the pass condition is town within the call and at most 2 of them.
func fMouth(ctx context.Context, r *Run, l FieldLab, outer, cave int, side int, what string) Outcome {
	if err := l.Survive(ctx, r); err != nil {
		return setupFail(err)
	}
	if err := l.Goto(ctx, r, 1); err != nil {
		return setupFail(err)
	}
	if err := l.ActivateAll(ctx, r); err != nil {
		return setupFail(err)
	}
	if err := l.Tick(ctx, 300); err != nil { // d2d refuses a waypoint trip within 10 s of an arrival (debug goto is one)
		return setupFail(err)
	}
	// the navigator learns the level's waypoint when it arrives by it (as the bot does on its way in), then the
	// character stands in the mouth
	if c0 := l.TakeWaypoint(ctx, outer); c0.Err != "" {
		return Blocks(what + ": the waypoint trip in failed: " + c0.Err)
	}
	if err := l.Tick(ctx, 40); err != nil {
		return setupFail(err)
	}
	v0, _ := l.WView(ctx)
	what += fmt.Sprintf(" (waypoint objects %v, arrival %d,%d)", v0.Waypoints, v0.X, v0.Y)
	desc, perr := l.PlaceAtMouth(ctx, r, outer, cave, side)
	if perr != "" {
		return Blocks(what + ": the character could not be put in the mouth's trigger area: " + perr)
	}
	if err := l.Tick(ctx, 30); err != nil {
		return setupFail(err)
	}
	// the caller retries like the goal loop does: a bounce into the cave ends the call (area = cave), the bot walks out
	// again (ToArea) and takes the waypoint again; the fix turns the step-away direction after a bounce
	bounces := 0
	var c WPCall
	var v WPView
	var logs []string
	var verr error
	for try := 0; try < 4; try++ {
		c = l.TakeWaypoint(ctx, 1)
		logs = append(logs, c.Log...)
		if err := l.Tick(ctx, 30); err != nil {
			return setupFail(err)
		}
		if v, verr = l.WView(ctx); verr != nil {
			return setupFail(verr)
		}
		if v.Area != cave {
			break
		}
		bounces++
		if area, e := l.ToArea(ctx, outer); area != outer {
			logs = append(logs, fmt.Sprintf("walk out of the cave ended in area %d (%s)", area, e))
			break
		}
		_ = l.Tick(ctx, 30)
	}
	c.Log = logs
	steps := countLines(c.Log, "stepping away from a cave mouth")
	intended := "waypoint to town from inside the mouth's trigger area: arrives in area 1, at most 2 bounces back into the cave"
	obs := fmt.Sprintf("%s: %s; arrived in %d, last verb %s, %d ticks, %d bounces into the cave, %d step-aways; log: %s", what, desc, v.Area, errOrOK(c.Err), c.Ticks, bounces, steps, strings.Join(firstN(c.Log, 10), " / "))
	o := Outcome{Intended: intended, Observed: obs, Sig: fmt.Sprintf("area=%d err=%v bounces<=2=%v", v.Area, c.Err != "", bounces <= 2)}
	switch {
	case v.Area != 1 || c.Err != "":
		o.Verdict, o.Reason = Fail, fmt.Sprintf("not in town (area %d, verb %s)", v.Area, errOrOK(c.Err))
	case bounces > 2:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("%d bounces (wanted at most 2)", bounces)
	default:
		o.Verdict, o.Reason = Pass, fmt.Sprintf("town in %d ticks with %d bounces and %d step-aways", c.Ticks, bounces, steps)
	}
	return o
}

// fTouch is DW.a12 at the navigator: Cold Plains' waypoint is not activated; ActivateWaypointNear touches it, the
// navigator hands the level to the runner (DrainActivated), and the waypoint then works as a travel target from town.
func fTouch(ctx context.Context, r *Run, l FieldLab) Outcome {
	if err := l.Survive(ctx, r); err != nil {
		return setupFail(err)
	}
	if err := l.Goto(ctx, r, 3); err != nil {
		return setupFail(err)
	}
	if err := l.Tick(ctx, 50); err != nil {
		return setupFail(err)
	}
	// Cold Plains' waypoint stands at about cell (23,69) on map seed 11 (seen by the DW.m1 arrival: objects [114 344] subtiles);
	// the character is put beside it, as if the bot had walked there, with the waypoint still not activated
	if err := l.Warp(ctx, r, 25, 69); err != nil {
		return setupFail(err)
	}
	if err := l.Tick(ctx, 40); err != nil {
		return setupFail(err)
	}
	touched, drained, terr := l.TouchWaypoint(ctx, 300)
	if err := l.Goto(ctx, r, 1); err != nil {
		return setupFail(err)
	}
	c := l.TakeWaypoint(ctx, 3)
	if err := l.Tick(ctx, 30); err != nil {
		return setupFail(err)
	}
	v, _ := l.WView(ctx)
	has := false
	for _, a := range drained {
		if a == 3 {
			has = true
		}
	}
	intended := "the touch in passing activates Cold Plains, DrainActivated lists 3, and a later waypoint trip to 3 works without 'no activated waypoint'"
	obs := fmt.Sprintf("touched=%v err %q drained %v; from town TakeWaypoint(3): %s, area %d; log: %s", touched, terr, drained, errOrOK(c.Err), v.Area, strings.Join(firstN(c.Log, 6), " / "))
	o := Outcome{Intended: intended, Observed: obs, Sig: fmt.Sprintf("touched=%v drained5=%v area=%d", touched, has, v.Area)}
	switch {
	case terr != "":
		o.Verdict, o.Reason = Fail, "the touch failed: "+terr
	case !touched:
		o.Verdict, o.Reason = Blocked, "the waypoint was not known or in reach at the arrival cell"
	case !has:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("the navigator did not hand level 3 over (drained %v)", drained)
	case c.Err != "" || v.Area != 3 || strings.Contains(strings.Join(c.Log, " "), "no activated waypoint"):
		o.Verdict, o.Reason = Fail, "the later trip to Cold Plains failed: "+c.Err
	default:
		o.Verdict, o.Reason = Pass, "touched in passing, handed over, and used as a travel target"
	}
	return o
}

// fSync is the fixmisc proof: DRILLS5 F4 saw the first Clear after a position jump call every monster unreachable.
func fSync(ctx context.Context, r *Run, l FieldLab) Outcome {
	f := l.Fight(ctx, r, FightSpec{Level: 2, Cold: true, Kill: true, Spawns: []string{"zombie1 3 near"}, Life: 400, Clear: 60 * time.Second})
	unreach := countLines(f.Log, "no path and no firing position")
	o := Outcome{Intended: "the first Clear after `debug goto` attacks the spawned zombies (kills > 0, no unreachable lines)",
		Observed: fmt.Sprintf("kills %d, cast commands %d, unreachable lines %d, avoided %d, skipped %d, ignored %d, %v, err %q; log head: %s", f.Kills, f.CmdSkill, unreach, f.Avoided, f.Skipped, f.Ignored, f.Elapsed.Round(time.Second), f.Err, strings.Join(firstN(f.LogHead, 5), " / ")),
		Sig:      fmt.Sprintf("kills>0 %v unreach>0 %v", f.Kills > 0, unreach > 0)}
	switch {
	case f.Spawned == 0:
		o.Verdict, o.Reason = Blocked, "no monster was spawned: "+f.Err
	case unreach > 0 || f.Kills == 0:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("the first Clear on a cold arrival did not attack: kills %d, unreachable lines %d", f.Kills, unreach)
	default:
		o.Verdict, o.Reason = Pass, fmt.Sprintf("the first Clear on a cold arrival attacked and killed %d of %d", f.Kills, f.Spawned)
	}
	return o
}
