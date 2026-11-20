package leveler

import "testing"

// necro1: the Necromancer summoner table is selectable by name, lands its points as planned at clvl 30 and changes no
// default; a second class proves the selector registration changed nothing else.
func TestNecromancerSummonerBuildSelectableAndDefaultsUnchanged(t *testing.T) {
	lb, err := LoadBuild("necromancer_summoner_guide", "")
	if err != nil {
		t.Fatal(err)
	}
	if lb.Class != "necromancer" || lb.Combat != "summoner" {
		t.Fatalf("class %q combat %q", lb.Class, lb.Combat)
	}
	if TableCombat("necromancer_summoner_guide") != "summoner" || TableCombat("necromancer_summoner_normal30") != "" {
		t.Errorf("combat lines: guide %q, normal30 %q", TableCombat("necromancer_summoner_guide"), TableCombat("necromancer_summoner_normal30"))
	}
	got := map[string]int{}
	for _, r := range lb.Rows {
		if r.Clvl > 30 {
			break
		}
		for _, s := range r.Skills {
			got[s.Skill] += s.Points
		}
	}
	want := map[string]int{"Teeth": 1, "Raise Skeleton": 10, "Skeleton Mastery": 10, "Clay Golem": 1, "Amplify Damage": 1, "Bone Armor": 1, "Raise Skeletal Mage": 6, "Corpse Explosion": 3}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: %d points at clvl 30, want %d", k, got[k], v)
		}
	}
	if len(got) != len(want) {
		t.Errorf("skills %v, want exactly %v", got, want)
	}
	// Defaults: nothing selects the new table unless named.
	if n := DefaultBuildName("necromancer"); n != "necromancer_summoner_normal30" {
		t.Errorf("default necromancer build %q", n)
	}
	if n := EndgameBuildName("necromancer"); n != "necromancer_summoner_endgame90" {
		t.Errorf("endgame necromancer build %q", n)
	}
	if n := SelectBuild("necromancer", 1, nil); n != "necromancer_summoner_normal30" {
		t.Errorf("fresh necromancer build %q", n)
	}
	if n := ResolveBuild("necromancer_summoner_guide", "necromancer"); n != "necromancer_summoner_guide" {
		t.Errorf("explicit selector resolved to %q", n)
	}
	if n := DefaultBuildName("paladin"); n != "paladin-normal30" {
		t.Errorf("default paladin build %q", n)
	}
	if n := DefaultBuildName("sorceress"); n != "fallback" {
		t.Errorf("default sorceress build %q", n)
	}
	if _, err := LoadBuild("necromancer_summoner_nope", ""); err == nil {
		t.Error("an unknown table must be an error")
	}
}
