package leveler

import "testing"

// SORC2: the second Sorceress build is selectable by name, lands its skill points as planned at clvl 30 and changes no default.
func TestSorceressFireBuildSelectableAndDefaultsUnchanged(t *testing.T) {
	lb, err := LoadBuild("sorceress_fire_normal30", "")
	if err != nil {
		t.Fatal(err)
	}
	if lb.Class != "sorceress" || lb.Combat != "fire" {
		t.Fatalf("class %q combat %q", lb.Class, lb.Combat)
	}
	got := map[string]int{}
	for _, r := range lb.Rows {
		if r.Clvl > 30 {
			t.Fatalf("row clvl %d past 30", r.Clvl)
		}
		for _, s := range r.Skills {
			got[s.Skill] += s.Points
		}
	}
	want := map[string]int{"Fire Bolt": 9, "Frozen Armor": 1, "Warmth": 1, "Telekinesis": 1, "Teleport": 1, "Fire Ball": 16, "Fire Mastery": 1}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: %d points at clvl 30, want %d", k, got[k], v)
		}
	}
	if len(got) != len(want) {
		t.Errorf("skills %v, want exactly %v", got, want)
	}
	// Defaults: nothing selects the new table unless named.
	if n := DefaultBuildName("sorceress"); n != "fallback" {
		t.Errorf("default sorceress build %q, want fallback", n)
	}
	if n := SelectBuild("sorceress", 1, nil); n != SorceressNormal30 {
		t.Errorf("fresh sorceress build %q, want %s", n, SorceressNormal30)
	}
	// Paladin: the selector registration changed nothing for a second class.
	if n := DefaultBuildName("paladin"); n != "paladin-normal30" {
		t.Errorf("default paladin build %q", n)
	}
	if n := ResolveBuild("sorceress_fire_normal30", "sorceress"); n != "sorceress_fire_normal30" {
		t.Errorf("explicit selector resolved to %q", n)
	}
}
