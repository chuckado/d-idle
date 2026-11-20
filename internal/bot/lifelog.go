package bot

import (
	"time"

	"d2lod-bot/internal/input"
	"fmt"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/power"
)

// Life log fields (docs/lifelog.md): logging only, nothing here feeds a decision.

// mercWord is alive, dead or none for the zone entry line.
func mercWord(s *d2.Snapshot) string {
	mc := d2.FindMerc(s)
	switch {
	case !mc.Present:
		return "none"
	case mc.Alive:
		return "alive"
	}
	return "dead"
}

// vitalsKV is the life, max life and potion counts added to the retreat and death lines.
func (r *Runner) vitalsKV(s *d2.Snapshot, p *d2.Player) []any {
	k := r.kitOf(s)
	return []any{"life", p.Life, "maxLife", p.MaxLife, "heal", k.BeltHP + k.PackHP, "beltMP", k.BeltMP, "rejuv", k.BeltRV, "tome", k.TownPortals}
}

// zoneEntry writes the one line per area change: `zone entry area=<id> clvl=<n> life=<n>/<n> heal=<n> mana=<n>/<n>
// tome=<n> merc=<alive|dead|none>`. heal is belt plus pack healing potions, tome the town portals in the kit.
func (r *Runner) zoneEntry(s *d2.Snapshot, p *d2.Player) {
	k := r.kitOf(s)
	kv := []any{"area", p.Area, "clvl", p.Level, "life", fmt.Sprintf("%d/%d", p.Life, p.MaxLife),
		"heal", k.BeltHP + k.PackHP, "mana", fmt.Sprintf("%d/%d", p.Mana, p.MaxMana), "tome", k.TownPortals, "merc", mercWord(s)}
	r.e.Log.Info("zone entry", append(kv, heroTermsKV(p)...)...)
}

// heroTermsKV is the power estimator's own inputs, appended to the zone entry line so a replay needs no default for them
// (card survcal): `def=<armour class> block=<gear block percent> dmg=<min>-<max> dmgpct=<n> ar=<n>`. A stat the player read
// does not carry is left out (never a zero standing for unknown). The weapon is the min-max damage with enhanced damage
// of the stat vector. Logging only: no decision reads it.
func heroTermsKV(p *d2.Player) []any {
	h := power.HeroFromPlayer(p, nil, -1)
	var kv []any
	if h.Quality[power.TermDefence] == power.Measured {
		kv = append(kv, "def", h.Defence)
	}
	if h.Quality[power.TermBlock] >= power.Derived {
		kv = append(kv, "block", h.BlockPct)
	}
	if h.DmgMax > 0 {
		kv = append(kv, "dmg", fmt.Sprintf("%d-%d", h.DmgMin, h.DmgMax), "dmgpct", h.DmgPct, "ar", h.AR)
	}
	return kv
}

// arrival is how the character came into the current area.
type arrival struct {
	from int
	at   time.Time
	pos  input.Point
}

// noteArrival runs at an area change, before o.area moves on. It logs where the character stands on arrival and, when
// the stay in the area it is leaving was a bounce (straight back where it came from within bounceMaxStay, neither side a
// town), says so with the position it left from and counts it (noteBounceLocked).
func (r *Runner) noteArrival(o *observer, p *d2.Player, now time.Time) {
	prev := o.arrived
	if o.area != 0 && !prev.at.IsZero() && p.Area == prev.from && !r.d.Plan.IsTown(o.area) && !r.d.Plan.IsTown(p.Area) {
		if stay := now.Sub(prev.at); stay < bounceMaxStay {
			r.e.Log.Info("arrival bounce", "area", o.area, "name", r.areaName(o.area), "back", p.Area,
				"stayMs", stay.Milliseconds(), "arrivedAt", prev.pos, "leftFrom", o.pos)
			o.bounces++
			r.noteBounceLocked(o.area, p.Area, now)
		}
	}
	o.arrivals++
	r.loop.flail.noteMove(now, p.Area, o.real)
	o.arrived = arrival{from: o.area, at: now, pos: input.Point{X: p.X, Y: p.Y}}
	r.e.Log.Info("arrival", "area", p.Area, "from", o.area, "x", p.X, "y", p.Y,
		"arrivals", o.arrivals, "bounces", o.bounces) // cumulative since start: bounces/arrivals is the rate to compare per build
}
