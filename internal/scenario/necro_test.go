package scenario

import (
	"strings"
	"testing"
)

func TestSpawnSetupLines(t *testing.T) {
	s := Scenario{ID: "necro.x", Pre: Pre{Level: 20, Start: StartField, Area: 2, Spawn: []Spawn{{Monster: "fallen1", N: 6}, {Monster: "fallen1", Mod: "unique"}}}}
	su, err := BuildSetup(s, "Necromancer", 11)
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, v := range su.Verbs {
		lines = append(lines, v.Line)
	}
	all := strings.Join(lines, "|")
	if !strings.Contains(all, "debug goto 2|debug spawn fallen1 6|debug spawn fallen1 1 unique") {
		t.Fatalf("spawn comes after the warp, in order: %s", all)
	}
	s.Pre.Start = ""
	if _, err := BuildSetup(s, "Necromancer", 11); err == nil {
		t.Fatal("spawn without a field start must be an error")
	}
	s.Pre.Start, s.Pre.Spawn = StartField, []Spawn{{Monster: "fallen1", Mod: "boss"}}
	if _, err := BuildSetup(s, "Paladin", 11); err == nil {
		t.Fatal("unknown mod must be an error")
	}
}

func TestMinionAndCastAssertions(t *testing.T) {
	two, zero := 2, 0
	st := State{Minions: &two, HeroCasts: map[string]int{"Raise Skeleton": 3}}
	as := []Assert{{Kind: AMinionsAtLeast, Min: 2}, {Kind: AHeroCasts, Code: "Raise Skeleton", Min: 3}}
	for _, r := range Evaluate(as, st) {
		if !r.OK {
			t.Errorf("%+v", r)
		}
	}
	st.Minions = &zero // must fail when the upkeep did nothing
	st.HeroCasts = map[string]int{}
	for _, r := range Evaluate(as, st) {
		if r.OK || r.Unobserved {
			t.Errorf("must be red, not blocked: %+v", r)
		}
	}
	for _, r := range Evaluate(as, State{}) {
		if !r.Unobserved {
			t.Errorf("no read must be UNOBSERVED: %+v", r)
		}
	}
	if err := (Assert{Kind: AHeroCasts}).Validate(); err == nil {
		t.Error("hero_casts_at_least needs a skill name")
	}
}
