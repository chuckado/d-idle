package drill

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// DRILLS29 (card drills29): the Frozen Armor recast with a pack held in melee at the due time, to provoke a hit-cut
// cast and prove the armorcut retry live. Harness code only: the bot's own Fighter.Buff does every cast.

// ArmorRecast is one recast attempt (one Fighter.Buff call that cast, or was cut).
type ArmorRecast struct {
	AtMS     int // game ms (from the start) just before the Buff call
	EndMS    int // game ms after it returned
	Cast     bool
	Cut      bool // the bot logged `buff cut by a hit`
	ModeIn   int  // player mode (bot snapshot) before the call
	ModeOut  int  // and after it
	Adjacent int  // monsters within 8 subtiles at the call
	StateOn  bool // state read after the call
	Life     int
}

// ArmorHold is the result of one hold run.
type ArmorHold struct {
	Err        string
	Binary     string
	SimMS      int
	Recasts    []ArmorRecast
	Casts      int
	Cuts       int
	CutRetries int // a cast that followed a cut with no other cast between
	Samples    int
	PollMS     int // longest interval between two state reads
	DownMS     int // game ms of the samples with the armor absent after it was first on
	MaxDownMS  int // longest absent run (sample to sample)
	Hits       int // reads of the player in hit recovery (mode 4) across the run
	ModeReads  int
	Modes      map[int]int // reads per player mode (bot snapshot; -1 = idle or unmapped)
	MinLife    int
	Refills    int
	Deaths     int
	Spawned    int
	Lines      []string
}

// D29Lab is the instance as the DRILLS29 drill uses it.
type D29Lab interface {
	FieldLab
	ArmorHold(ctx context.Context, r *Run, simMS int) ArmorHold
}

// D29 returns the DRILLS29 group (letter W).
func D29(l D29Lab) []Drill {
	return []Drill{
		{ID: "D29-C", Primitive: "P10", Title: "Frozen Armor recast while a pack holds the character in melee at the due time (armorcut retry), binary named", Limit: 30 * time.Minute, Do: func(ctx context.Context, r *Run) Outcome {
			f := l.ArmorHold(ctx, r, 300000)
			o := Outcome{Intended: "a pack is hitting the player when the armor is due: the bot's Buff pass recasts, a hit that cuts the cast is retried at once, and the armor is never absent longer than one poll interval"}
			if f.Err != "" {
				return Blocks(f.Err)
			}
			var rc []string
			for _, c := range f.Recasts {
				rc = append(rc, fmt.Sprintf("[%.1f-%.1f s cast=%v cut=%v mode %d>%d adj=%d on=%v life=%d]", float64(c.AtMS)/1000, float64(c.EndMS)/1000, c.Cast, c.Cut, c.ModeIn, c.ModeOut, c.Adjacent, c.StateOn, c.Life))
			}
			o.Observed = fmt.Sprintf("binary %s | %d s; casts %d, cuts %d, cut-retries %d; %d samples (longest interval %d ms), armor absent %d ms in total, longest run %d ms; hit-recovery in %d of %d mode reads (by mode %v); life min %d, refills %d, deaths %d, spawned %d | recasts %s | log: %s",
				f.Binary, f.SimMS/1000, f.Casts, f.Cuts, f.CutRetries, f.Samples, f.PollMS, f.DownMS, f.MaxDownMS, f.Hits, f.ModeReads, f.Modes, f.MinLife, f.Refills, f.Deaths, f.Spawned, strings.Join(rc, " "), strings.Join(f.Lines[:min12(len(f.Lines))], " ; "))
			o.Sig = fmt.Sprintf("casts=%d cuts=%d", f.Casts, f.Cuts)
			switch {
			case f.Casts < 2:
				o.Verdict, o.Reason = Fail, "no recast came from the Buff pass"
			case f.MaxDownMS > f.PollMS:
				o.Verdict, o.Reason = Fail, fmt.Sprintf("armor absent for %d ms in a row, longer than the poll interval %d ms", f.MaxDownMS, f.PollMS)
			case f.Cuts == 0:
				o.Verdict, o.Reason = Fail, fmt.Sprintf("armor held (absent %d ms) but no cast was cut: the retry was not exercised (U)", f.DownMS)
			case f.CutRetries == 0:
				o.Verdict, o.Reason = Fail, "a cast was cut but no retry cast followed"
			default:
				o.Verdict, o.Reason = Pass, fmt.Sprintf("RECORD on %s: %d cuts, %d retried, armor absent %d ms", f.Binary, f.Cuts, f.CutRetries, f.DownMS)
			}
			return o
		}},
	}
}
