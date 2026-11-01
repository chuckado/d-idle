package drill

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// MeterCheck is what the live Meter and the harness's own sampler saw during one Fighter.Clear (docs/fixmeter.md).
// The sampler is independent of the Meter's derivation: it reads the raw `monsters` and `state` verbs at the same
// sleeps, per unit id.
type MeterCheck struct {
	Totals              string // Totals.String() of the rig's Meter
	Kills, Dealt, Taken int    // the Meter's
	DealtOK             bool
	GTDead              int // distinct spawned units seen dead (mode DD or DT, or hp 0) by the sampler
	GTDealtFull         int // sum over units: first alive hp minus last alive hp, plus the last alive hp of those seen dead
	GTDealtDrops        int // sum over units of first alive hp minus last alive hp (the drops only)
	GTTaken             int // sum of drops of the player's life between samples
	SpawnHP             int // sum of the first alive hp reads
	Units               int
	PerUnit             []string
	Samples, ClearKills int
}

// DMeter is the fixmeter live proof (P28 recipe of docs/fixmeter.md).
func DMeter(l FieldLab) Drill {
	return Drill{ID: "D-meter", Primitive: "P28", Title: "fixmeter: the Meter's totals of a Fighter.Clear against 20 spawned zombies equal what the per-id monster hp and the player's life reads say", Limit: 10 * time.Minute, Do: func(ctx context.Context, r *Run) Outcome {
		f := l.Fight(ctx, r, FightSpec{Level: 2, Kill: true, Meter: true, Spawns: []string{"zombie1 20 near"}, Life: 1000, Clear: 150 * time.Second})
		m := f.Meter
		o := Outcome{Intended: "Meter.Kills equals the units seen dead, DamageDealt equals the per-id hp drops (first alive minus last alive, plus the last alive hp of the dead) and is within the spawn hp sum, Damage taken equals the player's life drops; hits and misses print n/a"}
		if m == nil {
			o.Verdict, o.Reason, o.Observed = Blocked, "no meter result: "+f.Err, f.Err
			return o
		}
		o.Observed = fmt.Sprintf("Meter %s | sampler: %d units, spawn hp sum %d, dead %d, dealt full %d (drops only %d), taken %d | Fighter.Clear kills %d, elapsed %v, err %q | per unit: %s", m.Totals, m.Units, m.SpawnHP, m.GTDead, m.GTDealtFull, m.GTDealtDrops, m.GTTaken, f.Kills, f.Elapsed.Round(time.Second), f.Err, strings.Join(firstN(m.PerUnit, 8), "; "))
		o.Sig = fmt.Sprintf("kills=%v dealt=%v taken=%v na=%v", m.Kills == m.GTDead, m.Dealt == m.GTDealtFull, m.Taken == m.GTTaken, strings.Contains(m.Totals, "hits=n/a misses=n/a"))
		switch {
		case f.Spawned == 0 || m.Units == 0:
			o.Verdict, o.Reason = Blocked, "no hostile monster was listed after the spawn"
		case !m.DealtOK:
			o.Verdict, o.Reason = Fail, "the Meter has no monster totals (DealtOK false: dealt and kills print n/a) in a fight with listed monsters"
		case !strings.Contains(m.Totals, "hits=n/a misses=n/a"):
			o.Verdict, o.Reason = Fail, "hits or misses are not printed n/a: "+m.Totals
		case m.Kills != m.GTDead:
			o.Verdict, o.Reason = Fail, fmt.Sprintf("Meter kills %d, units seen dead %d (Fighter.Clear counted %d)", m.Kills, m.GTDead, f.Kills)
		case m.Dealt <= 0 || m.Dealt > m.SpawnHP:
			o.Verdict, o.Reason = Fail, fmt.Sprintf("dealt %d is not in (0, %d]", m.Dealt, m.SpawnHP)
		case m.Dealt != m.GTDealtFull:
			o.Verdict, o.Reason = Fail, fmt.Sprintf("dealt %d, per-id hp %d", m.Dealt, m.GTDealtFull)
		case m.Taken != m.GTTaken:
			o.Verdict, o.Reason = Fail, fmt.Sprintf("taken %d, life drops %d", m.Taken, m.GTTaken)
		case m.Taken == 0:
			o.Verdict, o.Reason = Pass, fmt.Sprintf("PARTIAL: kills %d, dealt %d agree; no damage was taken, so the taken total (0 = 0) was not exercised", m.Kills, m.Dealt)
		default:
			o.Verdict, o.Reason = Pass, fmt.Sprintf("kills %d = %d, dealt %d = %d (spawn hp sum %d), taken %d = %d, hits and misses n/a", m.Kills, m.GTDead, m.Dealt, m.GTDealtFull, m.SpawnHP, m.Taken, m.GTTaken)
		}
		return o
	}}
}

// M returns the meter group.
func M(l FieldLab) []Drill { return []Drill{DMeter(l)} }
