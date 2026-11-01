package drill

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// DRILLS16 (card drills16): the live proof of fixbuff (docs/fixbuff.md): the `debug states` line on the production lab
// binary, Player.States filled by d2dsrc, the armor recast inside a long fight, and the state/skill ids of the other
// armors.

// BuffRaw is item (1) and (2): the raw `debug states` replies and the bit-10 samples around one Frozen Armor cast.
type BuffRaw struct {
	Err        string
	Before     string // raw reply before the cast
	After      string // raw reply 10 ticks after `cmd skill 40`
	ParsedOK   bool   // Client.Buffs parsed {40, level}
	Parsed     string
	Samples    int // snapshots read while the armor was on
	Bit10      int // of those, with Player.States bit 10 set
	AfterClear string
	ClearBit10 int // snapshots after `debug states clear 40` with the bit still set (must be 0)
	ClearN     int
}

// LongFight is item (3): a fight past 120 s at level 1 with no need-buff edge.
type LongFight struct {
	Err         string
	SimMS       int
	Rounds      int
	CastsMS     []int // game ms (from the first cast) of every buff cast seen in the log
	FightCastMS []int // those that happened while a Clear round was running (not the first, manual one)
	Samples     int
	Down        int // samples with the armor absent after it was first on
	DownMS      int
	MaxGapMS    int
	Bit10Up     int // samples where buffs said armor on and Player.States bit 10 was set
	Bit10Miss   int // samples where buffs said on and the bit was clear
	MonLife     string
	Lines       []string
}

// ArmorID is one row of item (4).
type ArmorID struct {
	Skill      int
	Via        string
	Reply      string // the verb's reply
	Line       string // raw `debug states` reply after it
	Parsed     string
	States     string // Player.States bits set among 0..127 in the snapshot
	Err        string
	CleanAfter bool
}

func (a ArmorID) String() string {
	return fmt.Sprintf("skill %d via %s: reply %q; debug states %q; parsed %s; Player.States bits %s; clear ok %v; err %q", a.Skill, a.Via, a.Reply, a.Line, a.Parsed, a.States, a.CleanAfter, a.Err)
}

// D16Lab is the instance as the DRILLS16 drills use it.
type D16Lab interface {
	FieldLab
	BuffRaw(ctx context.Context, r *Run) BuffRaw
	LongFight(ctx context.Context, r *Run, simMS int) LongFight
	ArmorIDs(ctx context.Context, r *Run) []ArmorID
}

// D16 returns the DRILLS16 group (letter R).
func D16(l D16Lab) []Drill {
	lim := 25 * time.Minute
	return []Drill{
		{ID: "R-bs1", Primitive: "P10", Title: "fixbuff B3: the raw `debug states` reply on the production lab binary, Client.Buffs parse, Player.States bit 10 after a Frozen Armor cast", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome {
			b := l.BuffRaw(ctx, r)
			o := Outcome{Intended: "`debug states` lists buffs=40:1, Client.Buffs parses it, Player.States has bit 10 at every sample while the armor is up and not after the clear"}
			if b.Err != "" {
				return Blocks(b.Err)
			}
			o.Observed = fmt.Sprintf("before: %q | after cast: %q | parsed: %s (ok %v) | up samples %d, bit 10 set in %d | after clear: %q, bit 10 still set in %d of %d", b.Before, b.After, b.Parsed, b.ParsedOK, b.Samples, b.Bit10, b.AfterClear, b.ClearBit10, b.ClearN)
			o.Sig = fmt.Sprintf("parsed=%v bit=%v clear=%v", b.ParsedOK, b.Bit10 == b.Samples && b.Samples > 0, b.ClearBit10 == 0)
			switch {
			case !b.ParsedOK:
				o.Verdict, o.Reason = Fail, "Client.Buffs did not parse the reply to {skill 40}"
			case b.Samples == 0 || b.Bit10 != b.Samples:
				o.Verdict, o.Reason = Fail, fmt.Sprintf("Player.States bit 10 set in %d of %d samples with the armor up", b.Bit10, b.Samples)
			case b.ClearBit10 != 0:
				o.Verdict, o.Reason = Fail, fmt.Sprintf("bit 10 stayed set in %d of %d samples after the clear", b.ClearBit10, b.ClearN)
			default:
				o.Verdict, o.Reason = Pass, fmt.Sprintf("raw line %q parsed; bit 10 set in %d of %d samples, clear in %d of %d after the clear", b.After, b.Bit10, b.Samples, b.ClearN-b.ClearBit10, b.ClearN)
			}
			return o
		}},
		{ID: "R-bs3", Primitive: "P10", Title: "fixbuff B1: Frozen Armor level 1 is recast from inside a fight longer than 120 s (no need-buff edge)", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome {
			f := l.LongFight(ctx, r, 300000)
			o := Outcome{Intended: "a monster that cannot die keeps the bot fighting for 300 s; the armor is recast from the fight loop before the 120 s expiry, armor absent in 0 samples"}
			if f.Err != "" {
				return Blocks(f.Err)
			}
			o.Observed = fmt.Sprintf("%d s simulated in %d Clear rounds; casts at game s %v (inside a Clear round: %v); %d samples (longest gap %d ms), armor absent in %d (%d ms); bit 10 set in %d of the up samples, clear in %d; monster %s | log: %s",
				f.SimMS/1000, f.Rounds, secs(f.CastsMS), secs(f.FightCastMS), f.Samples, f.MaxGapMS, f.Down, f.DownMS, f.Bit10Up, f.Bit10Miss, f.MonLife, strings.Join(f.Lines[:min3(len(f.Lines))], " ; "))
			o.Sig = fmt.Sprintf("casts=%d down=%d", len(f.CastsMS), f.Down)
			switch {
			case len(f.FightCastMS) == 0:
				o.Verdict, o.Reason = Fail, "no armor cast came from inside the fight"
			case f.Down > 0:
				o.Verdict, o.Reason = Fail, fmt.Sprintf("the armor was absent in %d samples (%d ms)", f.Down, f.DownMS)
			default:
				o.Verdict, o.Reason = Pass, fmt.Sprintf("%d casts from inside the fight at game s %v, armor never absent in %d samples", len(f.FightCastMS), secs(f.FightCastMS), f.Samples)
			}
			return o
		}},
		{ID: "R-bs4", Primitive: "P10", Title: "fixbuff B2: which skill ids the live `debug states` reports for Shiver Armor, Energy Shield, Chilling Armor, Bone Armor (record)", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome {
			rows := l.ArmorIDs(ctx, r)
			var s []string
			for _, a := range rows {
				s = append(s, a.String())
			}
			o := Outcome{Intended: "record only: the buffs= entry and the state bit each armor produces", Observed: strings.Join(s, " || ")}
			o.Verdict, o.Reason = Pass, fmt.Sprintf("RECORD: %d rows", len(rows))
			return o
		}},
	}
}
