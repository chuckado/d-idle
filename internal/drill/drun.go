package drill

import (
	"context"
	"fmt"
	"strings"
)

// RunLeg is one fixed leg of about 40 subtiles outside town, sampled once per tick (change fixrun, docs/fixrun.md).
type RunLeg struct {
	Name       string
	Dir        int // the direction index tried (the same start cell and direction in every leg)
	Err        string
	Ticks      int      // ticks from the move command to the end of the leg
	Lag        int      // ticks before the first movement
	Moving     int      // ticks in which the character moved
	Path       float64  // summed subtiles
	Net        float64  // straight-line subtiles start to end
	PerTick    float64  // Path / Moving (about 0.5625 running, 0.375 walking: UNVERIFIED for this class)
	StamStart  string   // raw stamina cur/max before the move
	StamEnd    string   // after the leg
	Believed   bool     // the controller's run belief after the move
	Keys       int      // run-key presses so far
	Resets     int      // belief resets so far
	WalkingRaw string   // the raw `walking=` field of the state, last read
	Moves      []string // the first per-tick deltas (subtiles), for the record
}

// RunLab is the optional part of the lab for D-run.
type RunLab interface {
	// Leg runs one leg on level 2 after a logged `debug goto 2`: a move command through the bot's controller (run
	// policy included), a step of one tick at a time.
	Leg(ctx context.Context, r *Run, name string) RunLeg
	// Restart leaves and enters the game through the backend's cycler (the drill restart path, OnInGame included).
	Restart(ctx context.Context) error
	// SetRunPolicy switches the controller's run policy (off: the bot never presses the run key: the walking baseline).
	SetRunPolicy(on bool)
}

func (l RunLeg) String() string {
	return fmt.Sprintf("%s: dir %d, %.1f subtiles in %d moving ticks (%d total, first move at tick %d) = %.4f per tick, net %.1f; stamina %s -> %s; runBelieved=%v runKeys=%d runResets=%d walkingRaw=%q; first deltas %s; err %q",
		l.Name, l.Dir, l.Path, l.Moving, l.Ticks, l.Lag, l.PerTick, l.Net, l.StamStart, l.StamEnd, l.Believed, l.Keys, l.Resets, l.WalkingRaw, strings.Join(l.Moves, ","), l.Err)
}

// DRun is the fixrun live proof: the same leg in a fresh game, again in the same game, after a restart, again.
func DRun(l RunLab) Drill {
	return Drill{ID: "D-run", Primitive: "P22", Title: "fixrun: a 40-subtile leg outside town in a fresh game, again after a game restart: equal speed means the run reset works", Limit: 10 * 60e9, Do: func(ctx context.Context, r *Run) Outcome {
		a := l.Leg(ctx, r, "fresh game")
		a2 := l.Leg(ctx, r, "same game again")
		rerr := l.Restart(ctx)
		var b, b2, w RunLeg
		if rerr == nil {
			b = l.Leg(ctx, r, "after restart")
			b2 = l.Leg(ctx, r, "after restart, again")
			// the baseline: a fresh game and a controller that never presses the run key = the game's own start state
			if werr := l.Restart(ctx); werr == nil {
				l.SetRunPolicy(false)
				w = l.Leg(ctx, r, "baseline: fresh game, no run key")
				l.SetRunPolicy(true)
				_ = l.Restart(ctx) // leave a fresh game with the policy on and the belief reset
			}
		}
		kind := func(v float64) string {
			switch {
			case v >= 0.5 && v < 0.65:
				return "running"
			case v >= 0.3 && v < 0.45:
				return "walking"
			}
			return "unclear"
		}
		obs := fmt.Sprintf("%s | %s | restart err %v | %s | %s | %s | walking baseline / running ratio %.2f", a, a2, rerr, b, b2, w, w.PerTick/a.PerTick)
		o := Outcome{Intended: "the same leg moves at the same speed in a fresh game and after a restart, and runs (stamina drains, about 0.5625 subtiles per tick)", Observed: obs,
			Sig: fmt.Sprintf("%s %s %s %s", kind(a.PerTick), kind(a2.PerTick), kind(b.PerTick), kind(b2.PerTick))}
		switch {
		case rerr != nil:
			o.Verdict, o.Reason = Fail, "restart: "+rerr.Error()
		case a.Err != "" || b.Err != "" || a.Moving < 10 || b.Moving < 10:
			o.Verdict, o.Reason = Blocked, "no leg could be measured: "+a.Err+" "+b.Err
		default:
			ratio := b.PerTick / a.PerTick
			if ratio < 0.85 || ratio > 1.15 {
				o.Verdict, o.Reason = Fail, fmt.Sprintf("before and after a restart differ: %.4f vs %.4f per tick (ratio %.2f)", a.PerTick, b.PerTick, ratio)
			} else if kind(a.PerTick) != "running" {
				o.Verdict, o.Reason = Fail, fmt.Sprintf("equal before and after (%.4f vs %.4f) but not running: the bot walks (%s)", a.PerTick, b.PerTick, kind(a.PerTick))
			} else {
				o.Verdict, o.Reason = Pass, fmt.Sprintf("equal before and after a restart (%.4f vs %.4f per tick), running", a.PerTick, b.PerTick)
			}
		}
		return o
	}}
}
