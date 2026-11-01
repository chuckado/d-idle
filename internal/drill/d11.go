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
	Adjacent, Hostiles          int // monsters within 3 subtiles / listed alive
	Life, MaxLife               int
	BeltPotions                 int
	ScrollsBefore, ScrollsAfter int
	Decision                    string // Fighter.Safety
	Acted                       bool   // the decision was carried out (only the portal case; a chicken leaves the game and is never run)
	ActErr                      string
	PortalOpen                  bool // a portal object near the player after the act
	LifeAfter                   int
	Where                       string
}

// CastTick is one tick of a cast trace.
type CastTick struct {
	T, Life, Mode int
	Portal        bool
}

// CastInfo is one D-castcut trial: a town portal scroll cast with one melee monster adjacent.
type CastInfo struct {
	Err                         string
	Adjacent                    int
	LifeStart, LifeEnd          int
	Hits                        int // ticks on which the life dropped during the trace
	PortalTick, PortalMS        int // ticks and game milliseconds from the click to the portal object, -1 none within the trace
	ScrollsBefore, ScrollsAfter int
	ModeSeq                     string // player mode, run length encoded "mode x ticks"
	Rights                      int
}

// D11Lab is the instance as the DRILLS11 drills use it.
type D11Lab interface {
	FieldLab
	EmptyPlan(ctx context.Context, r *Run, level, n int, total time.Duration) EmptyPlanInfo
	PortalMelee(ctx context.Context, r *Run, adj int, act bool) PMInfo
	CastTrace(ctx context.Context, r *Run, life int) CastInfo
}

// D11 returns the DRILLS11 group (letter P).
func D11(l D11Lab) []Drill {
	lim := 10 * time.Minute
	return []Drill{
		{ID: "D-emptyplan", Primitive: "P22", Title: "fixemptyplan: a pack, mana drained to 0 before every round: the out-of-mana back-off leaves no 'no path: empty path' line", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome { return dEmptyPlan(ctx, r, l) }},
		{ID: "D-portalmelee", Primitive: "P10", Title: "fixportalmelee: 3 melee monsters adjacent, 34 percent life, empty belt, one scroll: Safety chickens and the scroll stays", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome {
			i := l.PortalMelee(ctx, r, 3, false)
			o := Outcome{Intended: "adjacent >= 3, no potion, one scroll: Safety is Chicken (not TownPortal), scroll count unchanged"}
			o.Observed = i.String()
			o.Sig = fmt.Sprintf("%s|scroll %d->%d", i.Decision, i.ScrollsBefore, i.ScrollsAfter)
			switch {
			case i.Err != "":
				o.Verdict, o.Reason = Blocked, i.Err
			case i.Adjacent < 3 || i.BeltPotions != 0 || i.ScrollsBefore != 1:
				o.Verdict, o.Reason = Blocked, "setup not reached: "+i.String()
			case i.Decision != "Chicken" || i.ScrollsAfter != 1:
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
			case i.Decision != "TownPortal":
				o.Verdict, o.Reason = Fail, "decision "+i.Decision
			default:
				o.Verdict, o.Reason = Pass, fmt.Sprintf("TownPortal decided with 1 adjacent; act err %q, scrolls %d->%d, portal open %v, life after %d", i.ActErr, i.ScrollsBefore, i.ScrollsAfter, i.PortalOpen, i.LifeAfter)
			}
			return o
		}},
		{ID: "D-castcut", Primitive: "P10", Title: "D2.d unknown: does a monster hit cut a town portal cast, and how many ticks does a cast take (record)", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome {
			var rows []string
			cut, done, blocked := 0, 0, 0
			for i := 0; i < 4; i++ {
				c := l.CastTrace(ctx, r, 300)
				rows = append(rows, fmt.Sprintf("trial %d: %s", i+1, c.String()))
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
				o.Verdict, o.Reason = Pass, fmt.Sprintf("RECORD: %d trials completed the cast, %d did not produce a portal, %d blocked", done, cut, blocked)
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
	return fmt.Sprintf("%s: adjacent %d of %d hostiles, life %d/%d, belt potions %d, scrolls %d -> %d, decision %s, acted %v (err %q), portal open %v, life after %d, err %q",
		i.Where, i.Adjacent, i.Hostiles, i.Life, i.MaxLife, i.BeltPotions, i.ScrollsBefore, i.ScrollsAfter, i.Decision, i.Acted, i.ActErr, i.PortalOpen, i.LifeAfter, i.Err)
}

func (c CastInfo) String() string {
	return fmt.Sprintf("adjacent %d, life %d -> %d, hit ticks %d, portal after %d ticks = %d game ms (-1 none), scrolls %d -> %d, rights %d, player modes [%s], err %q",
		c.Adjacent, c.LifeStart, c.LifeEnd, c.Hits, c.PortalTick, c.PortalMS, c.ScrollsBefore, c.ScrollsAfter, c.Rights, c.ModeSeq, c.Err)
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
	o.Sig = fmt.Sprintf("empty=%d", ep)
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
