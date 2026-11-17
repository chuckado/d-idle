package survival

import (
	"testing"

	"d2lod-bot/internal/power"
)

func withoutRiskRoute(p *Params) *Params {
	c := *p
	c.Dials = map[string]StepDown{}
	for k, v := range p.Dials {
		v.TypicalEnter = 0
		c.Dials[k] = v
	}
	return &c
}

// The pay-for-risk route (card survtune) widens the balanced dial's choice toward zones that pay more, never to a zone
// whose bad pack is lost, and the cautious dial does not use it. Both builds.
func TestRiskRoutePaysForRiskWithinTheFloor(t *testing.T) {
	for _, tc := range []struct {
		name, arch string
		hero       func(int) power.Hero
	}{{"paladin", "bruiser", paladin}, {"sorceress", "caster", sorceress}} {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv(t, tc.arch)
			row := e.p.Dials["balanced"]
			if row.TypicalEnter <= 0 {
				t.Fatal("balanced has no risk route in params.json")
			}
			widened := 0
			for lvl := 2; lvl <= 40; lvl++ {
				in := e.input(tc.hero(lvl), ZoneKey{}, t0)
				on, err := NewLedger(e.p).Choose(in)
				if err != nil {
					t.Fatal(err)
				}
				off, err := NewLedger(withoutRiskRoute(e.p)).Choose(in)
				if err != nil {
					t.Fatal(err)
				}
				if off.Why.Code == WhyRiskPaid {
					t.Fatalf("lvl %d: route off but %q", lvl, off.Why.Code)
				}
				if on.Why.Code != WhyRiskPaid {
					if on.Area != off.Area && !on.Hole && !off.Hole {
						t.Errorf("lvl %d: picks differ (%d against %d) without the risk route", lvl, on.Area, off.Area)
					}
					continue
				}
				widened++
				for _, c := range in.Cands {
					if c.Key.Area == on.Area && c.ZM.Bad.Best < row.BadFloor {
						t.Errorf("lvl %d: risk-paid zone %d has bad best %.2f under the floor %.2f", lvl, on.Area, c.ZM.Bad.Best, row.BadFloor)
					}
				}
				in.Dial, _ = e.m.P.Dial("cautious")
				if d, _ := NewLedger(e.p).Choose(in); d.Why.Code == WhyRiskPaid {
					t.Errorf("lvl %d: the cautious dial took the risk route", lvl)
				}
			}
			t.Logf("%s: the route changed the pick at %d of 39 levels", tc.name, widened)
		})
	}
}

// Hand-made candidates: the safe field pays little, the riskier one pays a lot. The route takes the riskier one only
// within its two numbers, with the value discount, and the dial decides. Both archetypes (the controller is class
// generic, the archetype only selects the yield constants).
func TestRiskRouteCandidates(t *testing.T) {
	cand := func(area int, name string, badBest, badLow, typLow, xp float64) Candidate {
		return Candidate{Key: ZoneKey{area, "normal"}, Name: name,
			ZM:  power.ZoneMargin{Bad: power.Result{Best: badBest, Low: badLow}, Typical: power.Result{Best: typLow * 1.2, Low: typLow}},
			Eco: Eco{XPPerMin: xp}, TravelMin: 2}
	}
	for _, arch := range []string{"bruiser", "caster"} {
		t.Run(arch, func(t *testing.T) {
			e := newEnv(t, arch)
			pick := func(dial string, risky Candidate, p *Params) Decision {
				d, err := e.m.P.Dial(dial)
				if err != nil {
					t.Fatal(err)
				}
				in := Input{Now: t0, Dial: d, Power: 150, Archetype: arch, Cands: []Candidate{cand(2, "Safe", 9, 6, 90, 60), risky}}
				dec, err := NewLedger(p).Choose(in)
				if err != nil {
					t.Fatal(err)
				}
				return dec
			}
			tight := cand(5, "Tight", 2.4, 1.1, 50, 440) // bad pack calls it too tight (low 1.1 under 1.5), typical is comfortable
			if d := pick("balanced", tight, e.p); d.Area != 5 || d.Why.Code != WhyRiskPaid || d.Margin != "risky" {
				t.Errorf("balanced should pay for the risk: %+v", d)
			}
			if d := pick("balanced", tight, withoutRiskRoute(e.p)); d.Area != 2 {
				t.Errorf("route off: the safe field stays: %+v", d)
			}
			if d := pick("cautious", tight, e.p); d.Area != 2 {
				t.Errorf("cautious never pays for risk: %+v", d)
			}
			lost := cand(5, "Lost", 0.5, 0.3, 50, 440) // the bad pack is lost: under the floor
			if d := pick("balanced", lost, e.p); d.Area != 2 {
				t.Errorf("a lost bad pack stays out: %+v", d)
			}
			thin := cand(5, "Thin", 2.4, 1.1, 4, 440) // the typical pack is not comfortable
			if d := pick("balanced", thin, e.p); d.Area != 2 {
				t.Errorf("an uncomfortable typical pack stays out: %+v", d)
			}
			dull := cand(5, "Dull", 2.4, 1.1, 50, 62) // pays about the same: the value discount keeps the safe field
			if d := pick("balanced", dull, e.p); d.Area != 2 {
				t.Errorf("no clear pay: the safe field wins: %+v", d)
			}
		})
	}
}
