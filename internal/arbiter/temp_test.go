package arbiter

import (
	"strings"
	"testing"
)

type lostSrc struct {
	FuncSource
	lost []string
	star int
}

func (l *lostSrc) Lost(in, by Intent, why string, starved bool) {
	l.lost = append(l.lost, in.Key()+" by "+by.Key())
	if starved {
		l.star++
	}
}

func tempFacts(potions int) map[string]int {
	return map[string]int{FactPotions: potions, FactPotionsMin: 4, FactObjective: 1}
}

func chosenKey(d Decision) string {
	if d.Chosen == nil {
		return "nothing"
	}
	return d.Chosen.Key()
}

// A refusal with the need "needs_potions" becomes a refill that outranks the quest step, and goes away when its wake
// condition holds. Run for a second need (too_weak) so the table is not one row deep.
func TestRefusalBecomesTemporaryIntent(t *testing.T) {
	quest := FuncSource{N: SrcQuest, F: func(View) *Intent {
		return &Intent{Kind: "step", Class: Objective, Urgency: 0.5, NeedsArea: Any}
	}}
	a := New(DefaultConfig(), quest)
	f := tempFacts(0)
	if got := chosenKey(a.Decide(View{Now: 0, Area: "dark_wood", Facts: f})); got != "quest/step" {
		t.Fatalf("before the refusal: %s", got)
	}
	if !a.Refused("needs_potions", "potions or portal scroll in the pack", 1) {
		t.Fatal("needs_potions has a row")
	}
	d := a.Decide(View{Now: 200, Area: "dark_wood", Facts: f})
	if chosenKey(d) != "resupply/refill" || !strings.Contains(d.Chosen.Reason, "wakes on") {
		t.Fatalf("temporary refill expected: %s", d)
	}
	if got := chosenKey(a.Decide(View{Now: 210, Area: "town", Facts: tempFacts(4)})); got != "quest/step" {
		t.Fatalf("Until holds, the temporary intent ends: %s", got)
	}
	// too_weak: filler level_up, never above the quest (objective) but it exists while the quest source is silent
	b := New(DefaultConfig())
	b.Refused("too_weak", "character level rises", 0)
	if got := chosenKey(b.Decide(View{Now: 5, Area: "x", Facts: map[string]int{}})); got != "level/level_up" {
		t.Fatalf("too_weak: %s", got)
	}
	if got := chosenKey(b.Decide(View{Now: 1000, Area: "x", Facts: map[string]int{}})); got != "nothing" {
		t.Fatalf("TTL expired: %s", got)
	}
	if b.Refused("not_ready", "", 0) {
		t.Fatal("not_ready only pauses, no intent")
	}
}

func TestTemporaryIntentEndsOnReportDone(t *testing.T) {
	a := New(DefaultConfig())
	a.Refused("needs_gold_farm", "", 0)
	d := a.Decide(View{Now: 1, Facts: map[string]int{}})
	if chosenKey(d) != "recovery/farm_gold" {
		t.Fatal(d)
	}
	a.Report(*d.Chosen, 2, true)
	if got := chosenKey(a.Decide(View{Now: 3, Facts: map[string]int{}})); got != "nothing" {
		t.Fatalf("done intent proposed again: %s", got)
	}
}

func TestLostControlAndStarvation(t *testing.T) {
	low := &lostSrc{FuncSource: FuncSource{N: "low", F: func(View) *Intent {
		return &Intent{Kind: "x", Class: Objective, NeedsArea: Any}
	}}}
	high := FuncSource{N: "high", F: func(View) *Intent { return &Intent{Kind: "y", Class: Upkeep, NeedsArea: Any} }}
	a := New(DefaultConfig(), low, high)
	var starved int
	for now := 0.0; now <= 1300; now += 60 {
		starved += len(a.Decide(View{Now: now, Area: "a", Facts: map[string]int{}}).Starved)
	}
	if len(low.lost) != 2 || low.lost[0] != "low/x by high/y" {
		t.Fatalf("lost: once for the winner, once more when starved, got %v", low.lost)
	}
	if low.star != 1 || starved != 1 {
		t.Fatalf("starved once: hook %d decisions %d", low.star, starved)
	}
	// must-fail half: an outranked source that stops proposing is not starved
	idle := false
	s2 := &lostSrc{FuncSource: FuncSource{N: "low", F: func(View) *Intent {
		if idle {
			return nil
		}
		return &Intent{Kind: "x", Class: Objective, NeedsArea: Any}
	}}}
	b := New(DefaultConfig(), s2, high)
	for now := 0.0; now <= 1300; now += 60 {
		idle = now > 300
		b.Decide(View{Now: now, Area: "a", Facts: map[string]int{}})
	}
	if s2.star != 0 {
		t.Fatalf("a source that stopped asking is not starved: %d", s2.star)
	}
}

func TestControlledMask(t *testing.T) {
	if !Controlled(3, SrcSurvive) || !Controlled(3, SrcPlayer) || Controlled(3, "chores") || Controlled(0, SrcSurvive) || Controlled(255, "nope") {
		t.Fatal("bit order or unknown family")
	}
	sh := NewShadow(DefaultConfig(), ShadowConfig{LogGapSec: 1}, StdSources()...)
	d, ok := sh.Decide(View{Area: "a", Facts: map[string]int{FactHostile: 3, FactLifePct: 20, FactSurviveBelow: 50}})
	if !ok || d.Chosen == nil || d.Chosen.Source != SrcSurvive {
		t.Fatalf("%v", d)
	}
	if line := sh.Compare(d, 0, "chores"); line == "" {
		t.Fatal("a divergence must be reported by Compare")
	}
}
