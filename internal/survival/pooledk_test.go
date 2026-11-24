package survival

import (
	"testing"

	"d2lod-bot/internal/power"
)

// Card toolstrade39 (live 2026-11-24, dial bold): the hero-wide k sat at its clamp 3 and every zone without kills carried it
// whole, while the zone visited had k 1.5 (Dark Wood), so Tower Cellar rows (typical low 3.0, k 3.00, low margin 1.0 under
// the 1.2 bold needs) could never win. With poolUnvisitedK the unvisited zone takes the kill-weighted k of the visited
// ones. Both builds; the flag is on for bold only.
func TestPooledKLetsAnUnvisitedZoneEnter(t *testing.T) {
	for _, b := range twoBuilds {
		t.Run(b.name, func(t *testing.T) {
			e := newEnv(t, b.arch)
			cand := func(area int, name string, badBest, badLow, typLow, xp float64) Candidate {
				return Candidate{Key: ZoneKey{area, "normal"}, Name: name,
					ZM:  power.ZoneMargin{Bad: power.Result{Best: badBest, Low: badLow}, Typical: power.Result{Best: typLow * 1.2, Low: typLow}},
					Eco: Eco{XPPerMin: xp}, TravelMin: 2}
			}
			mk := func(p *Params) *Ledger {
				l := NewLedger(p)
				l.HeroKill, l.HeroPred, l.HeroSec = 400, 100, 300 // hero k at the clamp (3)
				// the visited field took half the damage the hero-wide ratio predicts: k_zone 0.5, k 1.5
				l.Observe(ZoneKey{2, "normal"}, t0, Event{Kind: EvFight, Seconds: 600, Kills: 200, PredTaken: 1000, Taken: 1000 * 1.5 / 3}, 150)
				return l
			}
			in := func(dial string) Input {
				d, _ := e.m.P.Dial(dial)
				return Input{Now: t0, Dial: d, Power: 150, Archetype: b.arch, Purse: Purse{Gold: 5000},
					Cands: []Candidate{cand(2, "Field", 20, 14, 90, 30), cand(5, "Cellar", 4.2, 3.0, 90, 100)}}
			}
			if k := mk(e.p).KHeroFor(false); k < 2.9 {
				t.Fatalf("setup: hero k %.2f is not at the clamp", k)
			}
			if d, _ := mk(e.p).Choose(in("bold")); d.Area != 5 {
				t.Errorf("bold with the pooled k should enter the unvisited cellar: %s", d.Line())
			}
			off := *e.p
			off.Dials = map[string]StepDown{}
			for k, v := range e.p.Dials {
				v.PoolUnvisitedK = false
				off.Dials[k] = v
			}
			if d, _ := mk(&off).Choose(in("bold")); d.Area == 5 {
				t.Errorf("flag off: the hero k keeps the cellar out: %s", d.Line())
			}
			if d, _ := mk(e.p).Choose(in("balanced")); d.Area == 5 {
				t.Errorf("balanced does not pool: %s", d.Line())
			}
		})
	}
}
