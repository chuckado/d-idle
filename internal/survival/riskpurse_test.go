package survival

import (
	"testing"

	"d2lod-bot/internal/power"
)

// Card survhold: the pay-for-risk route is closed to a hero whose purse cannot rebuy a belt. Both builds: a poor hero is
// never sent into a risk-paid zone, at any level.
func TestRiskRouteClosedWhenPoor(t *testing.T) {
	for _, b := range twoBuilds {
		t.Run(b.name, func(t *testing.T) {
			e := newEnv(t, b.arch)
			for lvl := 4; lvl <= 40; lvl++ {
				poor := e.input(b.hero(lvl), ZoneKey{}, t0)
				poor.Purse.Gold = 20
				dp, err := NewLedger(e.p).Choose(poor)
				if err != nil {
					t.Fatal(err)
				}
				if dp.Why.Code == WhyRiskPaid {
					t.Errorf("lvl %d: a poor hero was sent to %s on the risk route", lvl, dp.Name)
				}
			}
		})
	}
}

// Hand-made candidates, both archetypes: the risky field is taken at the limit of the purse and refused one gold under
// it, and the dial rows carry the condition as data.
func TestRiskRoutePurseCandidates(t *testing.T) {
	cand := func(area int, name string, badBest, badLow, typLow, xp float64) Candidate {
		return Candidate{Key: ZoneKey{area, "normal"}, Name: name,
			ZM:  power.ZoneMargin{Bad: power.Result{Best: badBest, Low: badLow}, Typical: power.Result{Best: typLow * 1.2, Low: typLow}},
			Eco: Eco{XPPerMin: xp}, TravelMin: 2}
	}
	for _, arch := range []string{"bruiser", "caster"} {
		t.Run(arch, func(t *testing.T) {
			e := newEnv(t, arch)
			for _, dial := range []string{"balanced", "bold"} {
				row, err := e.p.stepDown(dial)
				if err != nil || row.RiskMinGold <= 0 {
					t.Fatalf("%s must carry a purse condition: %+v %v", dial, row, err)
				}
				d, _ := e.m.P.Dial(dial)
				pick := func(gold int) Decision {
					in := Input{Now: t0, Dial: d, Power: 150, Archetype: arch, Purse: Purse{Gold: gold},
						Cands: []Candidate{cand(2, "Safe", 9, 6, 90, 60), cand(5, "Tight", 2.4, 1.1, 50, 440)}}
					dec, err := NewLedger(e.p).Choose(in)
					if err != nil {
						t.Fatal(err)
					}
					return dec
				}
				if got := pick(row.RiskMinGold); got.Area != 5 || got.Why.Code != WhyRiskPaid {
					t.Errorf("%s: purse at the limit pays for the risk: %+v", dial, got)
				}
				got := pick(row.RiskMinGold - 1)
				if got.Area != 2 {
					t.Errorf("%s: a poor hero stays in the safe field: %+v", dial, got)
				}
				for _, a := range got.Alts {
					if a.Area == 5 && a.Why.Code != WhyPoor {
						t.Errorf("%s: the refused zone says %q, want %q", dial, a.Why.Code, WhyPoor)
					}
				}
			}
		})
	}
}
