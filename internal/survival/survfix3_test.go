package survival

import (
	"testing"

	"d2lod-bot/internal/power"
)

// Card survhold3 (holdout 2 seed 103): the risk route is closed while k sits at its limit, a caster trusts its realized
// yield sooner, and a starved zone is worth no more than it paid. Both builds.
func TestRiskRouteClosedAtKLimit(t *testing.T) {
	for _, b := range twoBuilds {
		t.Run(b.name, func(t *testing.T) {
			e := newEnv(t, b.arch)
			d, _ := e.m.P.Dial("balanced")
			cand := func(area int, name string, badBest, badLow, typLow, xp float64) Candidate {
				return Candidate{Key: ZoneKey{area, "normal"}, Name: name,
					ZM:  power.ZoneMargin{Bad: power.Result{Best: badBest, Low: badLow}, Typical: power.Result{Best: typLow * 1.2, Low: typLow}},
					Eco: Eco{XPPerMin: xp}, TravelMin: 2}
			}
			in := Input{Now: t0, Dial: d, Power: 150, Archetype: b.arch, Purse: Purse{Gold: 5000},
				Cands: []Candidate{cand(2, "Safe", 9, 6, 90, 60), cand(5, "Tight", 8, 1.1, 90, 440)}}
			l := NewLedger(e.p)
			if dec, _ := l.Choose(in); dec.Area != 5 || dec.Why.Code != WhyRiskPaid {
				t.Fatalf("k 1: the route should be taken: %s", dec.Line())
			}
			l = NewLedger(e.p)
			l.Observe(ZoneKey{5, "normal"}, t0, Event{Kind: EvFight, Seconds: 600, Kills: 100, PredTaken: 100, Taken: 100 * 10 * e.p.KMax}, 150)
			dec, _ := l.Choose(in)
			if dec.Area == 5 && dec.Why.Code == WhyRiskPaid {
				t.Fatalf("k at the limit must close the route: %s", dec.Line())
			}
		})
	}
}

func TestCasterTrustsRealizedYieldSooner(t *testing.T) {
	p := params(t)
	if p.yieldMin("caster") >= p.yieldMin("bruiser") || p.yieldMin("") != p.YieldMinMinutes {
		t.Fatalf("yieldMin caster %v bruiser %v global %v", p.yieldMin("caster"), p.yieldMin("bruiser"), p.YieldMinMinutes)
	}
	if p.starveMul("caster") >= p.starveMul("bruiser") {
		t.Fatalf("starveMul caster %v bruiser %v", p.starveMul("caster"), p.starveMul("bruiser"))
	}
	k := ZoneKey{3, "normal"}
	lc, lb := NewLedger(p), NewLedger(p)
	for _, l := range []*Ledger{lc, lb} {
		l.Observe(k, t0, Event{Kind: EvXP, Seconds: 120, XP: 12}, PowerOf(sorceress(14)))
	}
	if _, r := lc.BlendYield(k, 150, "caster"); r == 0 {
		t.Fatal("the caster should trust two minutes")
	}
	if _, r := lb.BlendYield(k, 150, "bruiser"); r != 0 {
		t.Fatal("the bruiser needs three")
	}
}
