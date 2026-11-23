package agency

import (
	"strings"
	"testing"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2"
)

// druid1: the shadow planner gives a sane plan for the Druid fire table at every level from 1 to 30 and the same invariants
// hold for the default Druid table, so the check is not Frenzy-only. No unsupported skill is ever chosen.
func TestDruidFireShadowPlan(t *testing.T) {
	env := testEnv(t)
	pl := &Planner{Env: env}
	for _, table := range []string{"druid_fire_guide", "druid_fire_normal30"} {
		for clvl := 1; clvl <= 30; clvl++ {
			a, comb := allocFromTable(t, table, clvl)
			v, p, err := pl.Evaluate(Input{Hero: "Testy", Alloc: a, Combat: comb, Weapon: Weapon{Class: "1HS", Avg: 8}})
			if err != nil {
				t.Fatalf("%s clvl %d: %v", table, clvl, err)
			}
			if line := ShadowLine(v, p); !strings.Contains(line, "class=druid") || !strings.Contains(v.Line, "your build:") {
				t.Errorf("%s clvl %d: line %q / view %q", table, clvl, line, v.Line)
			}
			for _, ch := range append([]*Choice{p.Main, p.Attack, p.Secondary}, ptrs(p.Buffs)...) {
				if ch != nil && env.Caps.Skills[ch.Skill].Tier == Unsupported {
					t.Errorf("%s clvl %d: planner chose the unsupported %s", table, clvl, ch.Skill)
				}
			}
		}
	}
}

// Agreement on the attack at 6, 12, 18, 24 and 30: where the plan and the profile disagree the fixture
// (testdata/agreement.json) carries the reason; here the profile's own invariants are pinned.
func TestDruidFirePlanAndProfile(t *testing.T) {
	env := testEnv(t)
	prof, err := combat.LookupProfile(d2.Druid, "fire")
	if err != nil {
		t.Fatal(err)
	}
	for _, clvl := range []int{6, 12, 18, 24, 30} {
		a, comb := allocFromTable(t, "druid_fire_guide", clvl)
		env.Combat = combat.Build(comb)
		if comb != "fire" {
			t.Fatalf("combat line %q", comb)
		}
		p := mustDerive(t, a, env)
		pick := botPick(t, "druid", comb, clvl, a.Skills)
		want := "Firestorm"
		if clvl >= 12 {
			want = "Eruption" // Fissure in game; the first point lands at clvl 12
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
