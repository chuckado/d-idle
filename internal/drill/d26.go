package drill

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// DRILLS26 (card drills26): the Frozen Armor recast proof in a real fight against a pack that is replaced as it dies,
// on a named fork binary. Harness code only: the bot's own combat loop (Fighter.Clear) does every cast but the first.

// ArmorFight is one bounded fight of a Frozen Armor Sorceress against a plain pack.
type ArmorFight struct {
	Err         string
	Binary      string
	SimMS       int
	Rounds      int
	CastsMS     []int // game ms from the start of every Frozen Armor cast in the log (the first is the opening Buff)
	FightCastMS []int // those made inside a Clear round
	Samples     int
	UpSamples   int
	DownMS      int // game ms of the samples with the armor absent after it was first on
	UpMS        int // game ms with the armor up (state read at each sample, interval-weighted)
	MaxGapMS    int
	GapBoundMS  int // game ms the armor must have been off: cast-to-cast time beyond the real 120.3 s duration
	MinLife     int
	MaxLife     int
	Kills       int // monsters that the Clear rounds report killed
	Spawned     int
	Potions     int
	Refills     int // harness life refills (logged cheats) after a read below 25 percent
	Deaths      int
	Lines       []string
}

// D26Lab is the instance as the DRILLS26 drill uses it.
type D26Lab interface {
	FieldLab
	ArmorFight(ctx context.Context, r *Run, simMS int) ArmorFight
}

// D26 returns the DRILLS26 group (letter U).
func D26(l D26Lab) []Drill {
	return []Drill{
		{ID: "D26-A", Primitive: "P10", Title: "Frozen Armor recast in a long real fight against a pack, binary named", Limit: 30 * time.Minute, Do: func(ctx context.Context, r *Run) Outcome {
			f := l.ArmorFight(ctx, r, 300000)
			o := Outcome{Intended: "the bot recasts Frozen Armor from the fight loop before each expiry (limit 114 s at level 1) while a pack keeps attacking; record recasts, minimum life, armor-up time, kills"}
			if f.Err != "" {
				return Blocks(f.Err)
			}
			o.Observed = fmt.Sprintf("binary %s | %d s in %d Clear rounds; casts at game s %v (inside a round: %v); %d samples (longest gap %d ms), armor up %d ms of %d, absent %d ms (sampled), at least %d ms off by the cast spacing; life min %d of max %d, refills %d, deaths %d; kills %d of %d spawned, potions %d | log: %s",
				f.Binary, f.SimMS/1000, f.Rounds, secs(f.CastsMS), secs(f.FightCastMS), f.Samples, f.MaxGapMS, f.UpMS, f.SimMS, f.DownMS, f.GapBoundMS, f.MinLife, f.MaxLife, f.Refills, f.Deaths, f.Kills, f.Spawned, f.Potions, strings.Join(f.Lines[:min12(len(f.Lines))], " ; "))
			o.Sig = fmt.Sprintf("casts=%d", len(f.CastsMS))
			switch {
			case len(f.FightCastMS) == 0:
				o.Verdict, o.Reason = Fail, "no armor cast came from inside the fight"
			case f.DownMS > 0 || f.GapBoundMS > 0:
				o.Verdict, o.Reason = Fail, fmt.Sprintf("armor absent: %d ms sampled, %d ms by cast spacing", f.DownMS, f.GapBoundMS)
			default:
				o.Verdict, o.Reason = Pass, fmt.Sprintf("RECORD on %s: %d recasts from inside the fight at game s %v, armor up %d of %d ms, life min %d, kills %d", f.Binary, len(f.FightCastMS), secs(f.FightCastMS), f.UpMS, f.SimMS, f.MinLife, f.Kills)
			}
			return o
		}},
	}
}

func min12(n int) int {
	if n > 12 {
		return 12
	}
	return n
}
