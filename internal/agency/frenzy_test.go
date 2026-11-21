package agency

import (
	"strings"
	"testing"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2"
)

// barb1: the shadow planner gives a sane plan for the Frenzy table at every level from 1 to 30 and the same invariants
// hold for the default Barbarian table, so the check is not Frenzy-only. No unsupported skill is ever chosen.
func TestBarbarianFrenzyShadowPlan(t *testing.T) {
	env := testEnv(t)
	pl := &Planner{Env: env}
	for _, table := range []string{"barbarian_frenzy_guide", "barbarian_melee_normal30"} {
		for clvl := 1; clvl <= 30; clvl++ {
			a, comb := allocFromTable(t, table, clvl)
			v, p, err := pl.Evaluate(Input{Hero: "Testy", Alloc: a, Combat: comb, Weapon: Weapon{Class: "1HS", Avg: 20}})
			if err != nil {
				t.Fatalf("%s clvl %d: %v", table, clvl, err)
			}
			if line := ShadowLine(v, p); !strings.Contains(line, "class=barbarian") || !strings.Contains(v.Line, "your build:") {
				t.Errorf("%s clvl %d: line %q / view %q", table, clvl, line, v.Line)
			}
			for _, ch := range append([]*Choice{p.Main, p.Attack, p.Secondary}, ptrs(p.Buffs)...) {
				if ch != nil && env.Caps.Skills[ch.Skill].Tier == Unsupported {
					t.Errorf("%s clvl %d: planner chose the unsupported %s", table, clvl, ch.Skill)
				}
			}
			if table == "barbarian_frenzy_guide" && clvl >= 26 && (p.Attack == nil || p.Attack.Skill != "Frenzy") {
				t.Errorf("clvl %d: planner attack %v, want Frenzy", clvl, p.Attack)
			}
		}
	}
}

// Agreement on the attack at 6, 12, 18, 24 and 30: where the plan and the profile disagree the fixture
// (testdata/agreement.json) carries the reason; here the profile's own invariants are pinned.
func TestBarbarianFrenzyPlanAndProfile(t *testing.T) {
	env := testEnv(t)
	prof, err := combat.LookupProfile(d2.Barbarian, "frenzy")
	if err != nil {
		t.Fatal(err)
	}
	for _, clvl := range []int{6, 12, 18, 24, 30} {
		a, comb := allocFromTable(t, "barbarian_frenzy_guide", clvl)
		env.Combat = combat.Build(comb)
		if comb != "frenzy" {
			t.Fatalf("combat line %q", comb)
		}
		p := mustDerive(t, a, env)
		pick := botPick(t, "barbarian", comb, clvl, a.Skills)
		want := "Attack"
		switch {
		case clvl >= 24 && a.Skills["Frenzy"] > 0: // the first Frenzy point lands after clvl 24 (the Double Swing target fills first)
			want = "Frenzy"
		case clvl >= 6:
			want = "Double Swing"
		}
		if pick != want {
			t.Errorf("clvl %d: the profile attacks with %s, want %s", clvl, pick, want)
		}
		if pick != "Attack" {
			if _, ok := prof.Layout[pick]; !ok {
				t.Errorf("clvl %d: profile attack %s has no key", clvl, pick)
			}
		}
		if p.Archetype == "" {
			t.Errorf("clvl %d: no archetype", clvl)
		}
	}
}
