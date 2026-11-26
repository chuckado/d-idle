package combat

import (
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
)

func TestDecideArrivalLeavesAtOnce(t *testing.T) {
	in := SafetyInput{Life: 1400, Max: 1446, Tune: env.DefaultCombatTuning(), Arrival: true, HasPortal: true}
	if a := Decide(in); a != CrowdAbandon {
		t.Fatalf("arrival with a portal: %v, want crowd-abandon", a)
	}
	in.HasPortal, in.CanRetreat = false, true
	if a := Decide(in); a != CrowdRetreat {
		t.Fatalf("arrival without a portal: %v, want crowd-retreat", a)
	}
	in.Arrival = false
	if a := Decide(in); a != SafetyNone {
		t.Fatalf("no arrival crowd at full life: %v, want none", a)
	}
}

type stepClock struct{ t time.Time }

func (c *stepClock) Now() time.Time        { return c.t }
func (c *stepClock) Sleep(d time.Duration) { c.t = c.t.Add(d) }

func crowdAt(n int, dist float64) []Target {
	hs := make([]Target, n)
	for i := range hs {
		hs[i].Dist = dist
	}
	return hs
}

// The rule is data (admission.json arrival): a listed area, inside the window, enough hostiles, Normal only, any class.
func TestArrivalCrowdRule(t *testing.T) {
	const cvt, field = 58, 41
	clk := &stepClock{t: time.Unix(1000, 0)}
	f := &Fighter{e: &env.Env{Clock: clk}}
	pl := func(area int, d d2.Difficulty) *d2.Player {
		return &d2.Player{Area: area, Life: 1400, MaxLife: 1446, Difficulty: d}
	}
	if f.arrivalCrowd(pl(field, d2.Normal), crowdAt(30, 2)) {
		t.Fatal("an area with no row fired")
	}
	if !f.arrivalCrowd(pl(cvt, d2.Normal), crowdAt(25, 2)) {
		t.Fatal("25 hostiles at 2 subtiles on arrival did not fire")
	}
	if f.arrivalCrowd(pl(cvt, d2.Normal), crowdAt(5, 2)) {
		t.Fatal("5 hostiles fired")
	}
	if f.arrivalCrowd(pl(cvt, d2.Nightmare), crowdAt(25, 2)) {
		t.Fatal("rows apply in Normal only")
	}
	clk.Sleep(10 * time.Second)
	if f.arrivalCrowd(pl(cvt, d2.Normal), crowdAt(25, 2)) {
		t.Fatal("fired outside the arrival window")
	}
	if f.arr.bounce.Load() {
		t.Fatal("bounce flag left set")
	}
}
