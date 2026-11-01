package drill

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// EmptyPlanInfo is one D-emptyplan fight: Fighter.Clear in short rounds with the mana set to 0 before each one, so the
// out-of-mana back-off runs (docs/fixemptyplan.md). The counts are of the fight's own log lines.
type EmptyPlanInfo struct {
	Err                                       string
	Rounds, Kills, Spawned                    int
	Backoffs, MoveFailed, EmptyPath, CapMelee int // "out of mana: backing off", "move failed", failures whose err holds "empty path", "back-off cap, melee"
	Waits, AlreadyThere                       int // "out of mana: waiting"; "combat move: already there" (the quiet tick of a snapped goal)
	Elapsed                                   time.Duration
	Samples                                   []string // the first failing lines, else the first back-off lines
	Where                                     string
}

// PMInfo is one D-portalmelee observation: adj melee monsters next to the character at about 34 percent life, an empty
// belt and exactly one scroll, then the bot's Safety decision.
type PMInfo struct {
	Err                         string
	Adjacent, Within8, Hostiles int // monsters within 3 subtiles (the rule's adjacency) / within 8 (melee reach, measured) / listed alive
	Life, MaxLife               int
	BeltPotions                 int
	ScrollsBefore, ScrollsAfter int
	Decision                    string // Fighter.Safety
	Acted                       bool   // the decision was carried out (only the portal case; a chicken leaves the game and is never run)
	ActErr                      string
	PortalOpen                  bool // a portal object near the player after the act
	LifeAfter                   int
	Where                       string
	AttackDists                 string // "distance:samples" of monsters in an attack mode (whole subtiles)
}

// CastTick is one tick of a cast trace.
type CastTick struct {
	T, Life, Mode int
	Portal        bool
}

// CastInfo is one D-castcut trial: a town portal scroll cast with one melee monster near.
type CastInfo struct {
	Err                         string
	Adjacent                    int     // monsters within 8 subtiles at the click
	Dist                        float64 // the nearest monster's distance in subtiles at the click
	Synced                      bool    // the click came when the monster's blow animation began
	Delay                       int     // steps waited after that start before the click
	LifeStart, LifeEnd          int
	Hits                        int   // ticks on which the life dropped after the click
	HitTicks                    []int // those ticks, counted from the click
	PortalTick, PortalMS        int   // ticks and game milliseconds from the click to the portal object, -1 none within the trace
	PortalStep                  int   // the step (one Step call after the click) on which the portal object was first listed
	ScrollTick                  int   // tick on which the scroll left the pack, -1 never
	ScrollsBefore, ScrollsAfter int
	Timeline                    string // the events, tick offsets from the click
	AttackDists                 string
	Rights                      int
}

// D11Lab is the instance as the DRILLS11 drills use it.
type D11Lab interface {
	FieldLab
	EmptyPlan(ctx context.Context, r *Run, level, n int, total time.Duration) EmptyPlanInfo
	PortalMelee(ctx context.Context, r *Run, adj int, act bool) PMInfo
	CastTrace(ctx context.Context, r *Run, life, trial int, sync bool, delay int) CastInfo
}

// D11 returns the DRILLS11 group (letter P).
func D11(l D11Lab) []Drill {
	lim := 10 * time.Minute
	return []Drill{
		{ID: "D-emptyplan", Primitive: "P22", Title: "fixemptyplan: a pack, mana drained to 0 before every round: the out-of-mana back-off leaves no 'no path: empty path' line", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome { return dEmptyPlan(ctx, r, l) }},
		{ID: "D-portalmelee", Primitive: "P10", Title: "fixportalmelee: 3 Zombies adjacent, 34 percent life, empty belt, one scroll: Safety chickens and the scroll stays", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome {
			i := l.PortalMelee(ctx, r, 3, false)
			o := Outcome{Intended: "adjacent >= 3, no potion, one scroll: Safety is Chicken (not TownPortal), scroll count unchanged"}
			o.Observed = i.String()
			o.Sig = fmt.Sprintf("%s|scroll %d->%d", i.Decision, i.ScrollsBefore, i.ScrollsAfter)
			switch {
			case i.Err != "":
				o.Verdict, o.Reason = Blocked, i.Err
			case i.BeltPotions != 0 || i.ScrollsBefore != 1 || i.Within8 < 3:
				o.Verdict, o.Reason = Blocked, "setup not reached: "+i.String()
			case i.Adjacent < 3 && i.Decision == "town-portal" && i.ScrollsAfter == 1:
				o.Verdict, o.Reason = Fail, fmt.Sprintf("PRODUCT: %d monsters in melee reach (within 8 subtiles) but only %d within the rule's 3: Safety is town-portal at life %d/%d, the rule does not fire (the live geometry never put 3 within 3 subtiles)", i.Within8, i.Adjacent, i.Life, i.MaxLife)
			case i.Decision != "chicken" || i.ScrollsAfter != 1:
				o.Verdict, o.Reason = Fail, "decision "+i.Decision+", scrolls "+fmt.Sprint(i.ScrollsBefore, "->", i.ScrollsAfter)
			default:
				o.Verdict, o.Reason = Pass, fmt.Sprintf("%d adjacent at life %d/%d: Chicken, scroll kept (1)", i.Adjacent, i.Life, i.MaxLife)
			}
			return o
		}},
		{ID: "D-portalmelee.ctl", Primitive: "P10", Title: "fixportalmelee control: ONE monster adjacent, same life, empty belt, one scroll: the bot casts the portal as before", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome {
			i := l.PortalMelee(ctx, r, 1, true)
			o := Outcome{Intended: "one adjacent: Safety is TownPortal and the cast is carried out (scroll spent or portal open)"}
			o.Observed = i.String()
			o.Sig = i.Decision
			switch {
			case i.Err != "":
				o.Verdict, o.Reason = Blocked, i.Err
			case i.Adjacent != 1 || i.BeltPotions != 0 || i.ScrollsBefore != 1:
				o.Verdict, o.Reason = Blocked, "setup not reached: "+i.String()
			case i.Decision != "town-portal":
				o.Verdict, o.Reason = Fail, "decision "+i.Decision
			default:
				o.Verdict, o.Reason = Pass, fmt.Sprintf("TownPortal decided with 1 adjacent; act err %q, scrolls %d->%d, portal open %v, life after %d", i.ActErr, i.ScrollsBefore, i.ScrollsAfter, i.PortalOpen, i.LifeAfter)
			}
			return o
		}},
		{ID: "D-castcut", Primitive: "P10", Title: "D2.d unknown: does a monster hit cut a town portal cast, and how many ticks does a cast take (record)", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome {
			var rows []string
			cut, done, blocked := 0, 0, 0
			before, same, during := 0, 0, 0 // hits that landed before the portal object appeared, on the same step, and after it but inside the cast animation (to step 7)
			for i, tr := range []struct {
				life, delay int
				sync        bool
			}{{300, 0, false}, {300, 0, true}, {300, 1, true}, {300, 2, true}, {300, 3, true}, {300, 2, true}, {40, 3, true}, {40, 3, true}, {40, 2, true}, {40, 2, true}, {300, 3, true}, {300, 3, true}, {300, 3, true}, {300, 2, true}, {300, 2, true}, {300, 2, true}} {
				c := l.CastTrace(ctx, r, tr.life, i, tr.sync, tr.delay)
				rows = append(rows, fmt.Sprintf("trial %d (max life %d): %s", i+1, tr.life, c.String()))
				for _, h := range c.HitTicks {
					switch {
					case c.PortalTick < 0:
					case h < c.PortalStep:
						before++
					case h == c.PortalStep:
						same++
					case h <= 7:
						during++
					}
				}
				switch {
				case c.Err != "" || c.Adjacent < 1:
					blocked++
				case c.PortalTick >= 0:
					done++
				default:
					cut++
				}
			}
			o := Outcome{Intended: "record only: ticks from click to portal object with one monster adjacent, whether it completes under hits"}
			o.Observed = strings.Join(rows, " || ")
			o.Sig = "record"
			if done+cut == 0 {
				o.Verdict, o.Reason = Blocked, "no trial reached an adjacent monster: "+o.Observed
			} else {
				o.Verdict, o.Reason = Pass, fmt.Sprintf("RECORD: %d trials completed the cast (portal object listed), %d did not produce a portal, %d blocked; monster hits that landed before the portal object %d, on the same step %d, later inside the cast animation %d", done, cut, blocked, before, same, during)
			}
			return o
		}},
	}
}

func (i EmptyPlanInfo) String() string {
	return fmt.Sprintf("%s: %d rounds, %d spawned, %d kills; back-offs %d, move failed %d (of which empty path %d), back-off cap melee %d, waits %d, already-there ticks %d; %v err %q; samples: %s",
		i.Where, i.Rounds, i.Spawned, i.Kills, i.Backoffs, i.MoveFailed, i.EmptyPath, i.CapMelee, i.Waits, i.AlreadyThere, i.Elapsed.Round(time.Second), i.Err, strings.Join(i.Samples, " / "))
}

func (i PMInfo) String() string {
	return fmt.Sprintf("%s: adjacent %d (within 8: %d) of %d hostiles, life %d/%d, belt potions %d, scrolls %d -> %d, decision %s, acted %v (err %q), portal open %v, life after %d, attack-mode distances (subtile:samples) [%s], err %q",
		i.Where, i.Adjacent, i.Within8, i.Hostiles, i.Life, i.MaxLife, i.BeltPotions, i.ScrollsBefore, i.ScrollsAfter, i.Decision, i.Acted, i.ActErr, i.PortalOpen, i.LifeAfter, i.AttackDists, i.Err)
}

func (c CastInfo) String() string {
	return fmt.Sprintf("monsters within 8: %d, nearest %.1f subtiles, click synced to a blow %v (+%d steps), life %d -> %d, hits after the click %d at ticks %v, portal object after %d ticks = %d game ms (-1 none), scroll left the pack at tick %d (-1 never), scrolls %d -> %d, timeline [%s], err %q",
		c.Adjacent, c.Dist, c.Synced, c.Delay, c.LifeStart, c.LifeEnd, c.Hits, c.HitTicks, c.PortalTick, c.PortalMS, c.ScrollTick, c.ScrollsBefore, c.ScrollsAfter, c.Timeline, c.Err)
}

// dEmptyPlan: PASS when the log has no "empty path" and there were back-offs; BLOCKED when none ran.
func dEmptyPlan(ctx context.Context, r *Run, l D11Lab) Outcome {
	var all []EmptyPlanInfo
	var obs []string
	for _, lv := range []struct{ level, n int }{{2, 12}, {9, 12}} {
		i := l.EmptyPlan(ctx, r, lv.level, lv.n, 100*time.Second)
		all = append(all, i)
		obs = append(obs, i.String())
	}
	o := Outcome{Intended: "no 'empty path' in the log, back-offs happened, the fight fell through (kills or melee) instead of stalling"}
	o.Observed = strings.Join(obs, " || ")
	bo, mf, ep, ki := 0, 0, 0, 0
	errs := ""
	for _, i := range all {
		bo += i.Backoffs
		mf += i.MoveFailed
		ep += i.EmptyPath
		ki += i.Kills
		errs += i.Err
	}
	o.Sig = o.Observed // the whole observation: both repeats must agree, so a run-2 difference shows in the report
	switch {
	case ep > 0:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("%d 'empty path' lines (back-offs %d, move failed %d)", ep, bo, mf)
	case bo == 0:
		o.Verdict, o.Reason = Blocked, "no out-of-mana back-off ran, the recipe was not exercised: "+o.Observed
	default:
		o.Verdict, o.Reason = Pass, fmt.Sprintf("back-offs %d, failed walks %d (%.0f percent), empty path 0, kills %d", bo, mf, 100*float64(mf)/float64(bo), ki)
	}
	return o
}
