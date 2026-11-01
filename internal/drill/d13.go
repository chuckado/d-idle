package drill

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// FroDur is one Frozen Armor duration measurement with no bot in the loop: the skill is cast with `cmd skill 40`, the
// state list (`debug states`, buffs=) is read at every step, nobody recasts. Times are game milliseconds from the cast
// command (Step 1 is 2 ticks of 40 ms, so the resolution is 80 ms).
type FroDur struct {
	Err         string
	Level       int
	BuffsOn     string // the buffs= text at the first sample where the armor was on, "40:<level>,"
	OnAfterMS   int    // first sample with the armor on (-1 never)
	GoneAfterMS int    // first sample with it off again (-1 not within the trace)
	ExpectMS    int    // Skills.txt: (3000 + 300 (L-1)) frames x 40 ms
	CostPoints  float64
}

// FroBot is the bot's own buff path (Fighter.Buff, called every sample as the bot does when it decides to buff) over a
// simulated stretch of game time at one skill level.
type FroBot struct {
	Err           string
	Level         int
	SimMS         int
	CastsMS       []int // game ms from the start at which each Buff call produced a cast ("buff skill=Frozen Armor")
	ClockCastsMS  []int // the same instants on the bot's own clock (env clock, the Pacer), from the first
	FirstOnMS     int   // game ms of the first sample with the armor on (-1 never)
	Samples       int
	MaxStepMS     int // longest game time between two samples after the first cast, the resolution of the "never down"
	DownSamples   int // samples with the armor absent after it was first seen on
	DownMS        int // game ms those samples stand for (each sample's interval)
	ManaRefills   int
	LimitMS       int // what buffLimit says: min(buffRecastSec, duration - buffMarginSec) at this level
	StateReadable bool
	Lines         []string
}

func (b FroBot) String() string {
	var iv []string
	for i := 1; i < len(b.CastsMS); i++ {
		iv = append(iv, fmt.Sprintf("%.1f", float64(b.CastsMS[i]-b.CastsMS[i-1])/1000))
	}
	return fmt.Sprintf("level %d, %d s simulated, %d samples (longest gap between samples %d ms), casts at game s %v (intervals %s s; bot clock intervals %v s), first state-on sample at %d ms, down samples %d (%d ms), mana refills %d, bot's limit %d ms, Player.States readable by the bot: %v",
		b.Level, b.SimMS/1000, b.Samples, b.MaxStepMS, secs(b.CastsMS), strings.Join(iv, " "), clockIv(b.ClockCastsMS), b.FirstOnMS, b.DownSamples, b.DownMS, b.ManaRefills, b.LimitMS, b.StateReadable)
}

func secs(ms []int) []string {
	var o []string
	for _, m := range ms {
		o = append(o, fmt.Sprintf("%.1f", float64(m)/1000))
	}
	return o
}

func clockIv(ms []int) []string {
	var o []string
	for i := 1; i < len(ms); i++ {
		o = append(o, fmt.Sprintf("%.1f", float64(ms[i]-ms[i-1])/1000))
	}
	return o
}

// ManaRow is one sample of the M1 trace: game ticks since the cast command was queued.
type ManaRow struct {
	Tick, Raw, PMode, Missiles int
}

// ManaTrace is one M1 trial: Fire Bolt through `debug cast`, mana (`dump stats` raw[8], 256ths) read at every step.
// HitAfter is the number of steps after the cast command at which the interruption was applied (-1 none).
type ManaTrace struct {
	Err       string
	HitAfter  int
	Rows      []ManaRow
	Start     int // raw mana at the command, 256ths
	Cost      int // the verb's own mana figure is points before the cast; the cost is read from the drop
	PayTick   int // first row whose mana is at least 300 256ths under the row before it (-1 none)
	MissTick  int // first row with a player missile listed (-1 none)
	Hit       string
	HitLifeOK bool
	Final     int
}

func (t ManaTrace) String() string {
	var b []string
	for _, r := range t.Rows {
		b = append(b, fmt.Sprintf("+%d:%d/m%d/n%d", r.Tick, r.Raw, r.PMode, r.Missiles))
	}
	return fmt.Sprintf("hit after %d steps (%s): command mana %d, paid at tick +%d, missile first listed at +%d, final %d | %s", t.HitAfter, t.Hit, t.Start, t.PayTick, t.MissTick, t.Final, strings.Join(b, " "))
}

// D13Lab is the instance as the DRILLS13 drills use it.
type D13Lab interface {
	FieldLab
	FroDuration(ctx context.Context, r *Run, level int) FroDur
	FroBot(ctx context.Context, r *Run, level int, simMS int) FroBot
	ManaTrace(ctx context.Context, r *Run, hitAfter int) ManaTrace
}

// D13 returns the DRILLS13 group (letter Q).
func D13(l D13Lab) []Drill {
	lim := 25 * time.Minute
	fro := func(level int) func(ctx context.Context, r *Run) Outcome {
		return func(ctx context.Context, r *Run) Outcome {
			d := l.FroDuration(ctx, r, level)
			o := Outcome{Intended: fmt.Sprintf("level %d: the armor state switches on after a cast and lasts %d ms; the bot's Buff recasts at buffLimit before it ends, so the armor is never down in a simulated 600 s", level, 3000*40+300*(level-1)*40)}
			if d.Err != "" {
				o.Verdict, o.Reason = Blocked, "duration measure: "+d.Err
				return o
			}
			b := l.FroBot(ctx, r, level, 600000)
			o.Observed = fmt.Sprintf("DURATION (no bot): state on after %d ms (buffs=%q), gone after %d ms, table says %d ms, cast cost %.1f points || BOT: %s", d.OnAfterMS, d.BuffsOn, d.GoneAfterMS, d.ExpectMS, d.CostPoints, b)
			o.Sig = fmt.Sprintf("on=%v gone=%v casts=%d down=%d", d.OnAfterMS >= 0, d.GoneAfterMS, len(b.CastsMS), b.DownSamples)
			switch {
			case b.Err != "":
				o.Verdict, o.Reason = Blocked, "bot run: "+b.Err
			case d.OnAfterMS < 0:
				o.Verdict, o.Reason = Fail, "the armor state never showed after a cast (debug states buffs empty): the bit-10 read is not supported by this instance"
			case d.GoneAfterMS < 0:
				o.Verdict, o.Reason = Blocked, "the armor did not expire within the trace"
			case len(b.CastsMS) < 4:
				o.Verdict, o.Reason = Fail, fmt.Sprintf("the bot cast only %d times in %d s", len(b.CastsMS), b.SimMS/1000)
			case b.DownSamples > 0:
				o.Verdict, o.Reason = Fail, fmt.Sprintf("the armor was absent in %d samples (%d ms) after it had been on", b.DownSamples, b.DownMS)
			default:
				o.Verdict, o.Reason = Pass, fmt.Sprintf("duration %d ms measured (table %d ms); the bot recast at %s s (limit %d ms) and the armor was never absent in %d samples (gap between samples at most %d ms)", d.GoneAfterMS, d.ExpectMS, strings.Join(secs(b.CastsMS), " "), b.LimitMS, b.Samples, b.MaxStepMS)
			}
			return o
		}
	}
	return []Drill{
		{ID: "D-fro1", Primitive: "P10", Title: "fixrecast: Frozen Armor level 1: state after a cast, real duration, the bot's recast before it ends (simulated 600 s)", Limit: lim, Do: fro(1)},
		{ID: "D-fro2", Primitive: "P10", Title: "fixrecast: Frozen Armor level 2 (132 s): the same, recast at about 126 s", Limit: lim, Do: fro(2)},
		{ID: "D-m1", Primitive: "P26", Title: "M1 (sk-mana-trace D1/D3): when Fire Bolt's mana leaves on d2d relative to the command and the action frame; does an interrupted cast cost mana (record)", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome {
			var rows []string
			var trs []ManaTrace
			for _, hit := range []int{-1, -1, 0, 1, 2, 3, 4, 5, 6} {
				t := l.ManaTrace(ctx, r, hit)
				if t.Err != "" {
					return Blocks(fmt.Sprintf("trial hit=%d: %s", hit, t.Err))
				}
				trs = append(trs, t)
				rows = append(rows, t.String())
			}
			o := Outcome{Intended: "record only: the tick on which mana is paid and the tick of the missile for an undisturbed Fire Bolt, and the mana of casts interrupted by a forced hit (cheat debug hit) at each step"}
			o.Observed = strings.Join(rows, " || ")
			o.Sig = fmt.Sprintf("pay=%d miss=%d", trs[0].PayTick, trs[0].MissTick)
			o.Verdict, o.Reason = Pass, fmt.Sprintf("RECORD: undisturbed cast pays at tick +%d and lists the missile at +%d (ticks of 40 ms after the command was queued; the sampler steps 2 ticks at a time)", trs[0].PayTick, trs[0].MissTick)
			return o
		}},
	}
}
