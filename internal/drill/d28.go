package drill

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// DRILLS28 (card drills28): live proof of card armorcut. Harness code only.

// ArmorIdle is the no-monster Frozen Armor refresh test (step 0): cast at about 60 s and 114 s, read the state around 121-125 s.
type ArmorIdle struct {
	Err      string
	Binary   string
	CastsMS  []int    // game ms of each cast command
	Timeline []string // "t=...ms state=... mode=..." lines
	AtMS     []int    // sample times
	OnAt     []bool   // armor on at each sample
	Window   bool     // armor on at every sample between 121 s and 125 s
	Seen     int      // samples in that window
}

// D28Lab is the instance as the DRILLS28 drills use it.
type D28Lab interface {
	FieldLab
	ArmorIdle(ctx context.Context, r *Run) ArmorIdle
}

// D28 returns the DRILLS28 group (letter V); the D26-A rerun is D26 (group U).
func D28(l D28Lab) []Drill {
	return []Drill{
		{ID: "D28-0", Primitive: "P10", Title: "Frozen Armor refresh with no monsters: cast at 60 s and 114 s, state still present at 121-125 s", Limit: 10 * time.Minute, Do: func(ctx context.Context, r *Run) Outcome {
			f := l.ArmorIdle(ctx, r)
			o := Outcome{Intended: "a second cast 54 s into a 120 s armor extends it: the state is present at 121-125 s (so d2d refreshes a set state)"}
			if f.Err != "" {
				return Blocks(f.Err)
			}
			o.Observed = fmt.Sprintf("binary %s | casts at game s %v | window samples %d, all on: %v | %s", f.Binary, secs(f.CastsMS), f.Seen, f.Window, strings.Join(f.Timeline, " ; "))
			o.Sig = fmt.Sprintf("casts=%d", len(f.CastsMS))
			switch {
			case f.Seen == 0:
				o.Verdict, o.Reason = Fail, "no sample in the 121-125 s window"
			case !f.Window:
				o.Verdict, o.Reason = Fail, "state gone inside 121-125 s after a second cast: d2d did not refresh a set state"
			default:
				o.Verdict, o.Reason = Pass, "RECORD: the state is present through 121-125 s after the second cast"
			}
			return o
		}},
	}
}
