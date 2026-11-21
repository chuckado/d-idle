package agency

import (
	"strings"
	"testing"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2"
)

// hammer1: the shadow planner gives a sane plan for the Hammerdin table at every level from 1 to 30 and the same
// invariants hold for the default Paladin table, so the check is not Hammerdin-only. No unsupported skill is ever chosen.
func TestPaladinHammerdinShadowPlan(t *testing.T) {
	env := testEnv(t)
	pl := &Planner{Env: env}
	for _, table := range []string{"paladin_hammerdin_guide", "paladin-normal30-v2"} {
		for clvl := 1; clvl <= 30; clvl++ {
			a, comb := allocFromTable(t, table, clvl)
			v, p, err := pl.Evaluate(Input{Hero: "Testy", Alloc: a, Combat: comb, Weapon: Weapon{Class: "1HS", Avg: 20}})
			if err != nil {
				t.Fatalf("%s clvl %d: %v", table, clvl, err)
			}
			if line := ShadowLine(v, p); !strings.Contains(line, "class=paladin") || !strings.Contains(v.Line, "your build:") {
				t.Errorf("%s clvl %d: line %q / view %q", table, clvl, line, v.Line)
			}
			for _, ch := range append([]*Choice{p.Main, p.Attack, p.Secondary}, ptrs(p.Buffs)...) {
				if ch != nil && env.Caps.Skills[ch.Skill].Tier == Unsupported {
					t.Errorf("%s clvl %d: planner chose the unsupported %s", table, clvl, ch.Skill)
				}
			}
			if table == "paladin_hammerdin_guide" {
				if clvl >= 18 && (p.Attack == nil || p.Attack.Skill != "Blessed Hammer") {
					t.Errorf("clvl %d: planner attack %v, want Blessed Hammer", clvl, p.Attack)
				}
				if clvl == 30 {
					// Charge is a Holy Shield prerequisite point the engine never casts (unsupported); a Hammerdin
					// at 30 burns more mana than an unequipped hero sustains: both are flagged, never hidden.
					if p.Unplayed.Skills["Charge"] != 1 || !p.Has(WUnsupportedPts) || !p.Has(WManaStarved) {
						t.Errorf("clvl 30 warnings %s, unplayed %v", warnCodes(p), p.Unplayed.Skills)
					}
				}
			}
		}
	}
}

// Agreement on the attack and the aura slot at 6, 12, 18, 24 and 30: where the plan and the profile disagree the
// fixture (testdata/agreement.json) carries the reason; here the profile's own invariants are pinned.
func TestPaladinHammerdinPlanAndProfile(t *testing.T) {
	env := testEnv(t)
	prof, err := combat.LookupProfile(d2.Paladin, "hammerdin")
	if err != nil {
		t.Fatal(err)
	}
	for _, clvl := range []int{6, 12, 18, 24, 30} {
		a, comb := allocFromTable(t, "paladin_hammerdin_guide", clvl)
		env.Combat = combat.Build(comb)
		if comb != "hammerdin" {
			t.Fatalf("combat line %q", comb)
		}
		p := mustDerive(t, a, env)
		pick := botPick(t, "paladin", comb, clvl, a.Skills)
		want := "Smite"
		if clvl >= 18 {
			want = "Blessed Hammer"
		}
		if pick != want {
			t.Errorf("clvl %d: the profile attacks with %s, want %s", clvl, pick, want)
		}
		if _, ok := prof.Layout[pick]; !ok {
			t.Errorf("clvl %d: profile attack %s has no key", clvl, pick)
		}
		if clvl >= 18 && a.Skills["Concentration"] > 0 && p.Archetype == "" {
			t.Errorf("clvl %d: no archetype", clvl)
		}
	}
}
