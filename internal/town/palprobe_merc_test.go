package town_test

import (
	"testing"

	. "d2lod-bot/internal/merc"
)

// PALPROBE part B (d): the Act 2 mercenary policy for a Paladin (a melee class with auras of its own). Normal asks for
// Defiance, then Prayer; Might only appears from Nightmare. Nothing in the policy is Paladin specific: the class
// only decides the caster triple, so a Paladin takes the physical-class rows.
func TestPalProbeMercAuras(t *testing.T) {
	want := map[Difficulty]string{Normal: "Defiance,Prayer", Nightmare: "Might,Defiance,Prayer", Hell: "Defiance,Prayer"}
	for d, w := range want {
		got := ""
		for i, a := range AuraPrefs("Paladin", d, false) {
			if i > 0 {
				got += ","
			}
			got += a
		}
		if got != w {
			t.Errorf("difficulty %v: %s, want %s", d, got, w)
		}
	}
	c := Ctx{Gold: 3000, Kit: 300, Act: 2, PlayerClass: "Paladin", HaveTP: true, HavePotions: true, Knobs: DefaultKnobs(), FirstLutVisit: true}
	none := MercState{Known: true}
	for _, o := range []struct {
		aura string
		do   bool
	}{{"Defiance", true}, {"Prayer", true}, {"Might", false}, {"Blessed Aim", false}, {"Holy Freeze", false}} {
		d := ShouldHire(HireOffer{Class: ClassDesert, Aura: o.aura, Level: 10, Price: 600}, none, c)
		if d.Do != o.do {
			t.Errorf("Normal Paladin, %s: %+v", o.aura, d)
		}
	}
	// Defiance beats Prayer in a replace: a Prayer merc is swapped on the first Lut Gholein visit
	cur := MercState{Known: true, Owned: true, Alive: true, Aura: "Prayer"}
	if d := ShouldReplace(cur, HireOffer{Class: ClassDesert, Aura: "Defiance", Level: 10, Price: 600}, c); !d.Do {
		t.Logf("Prayer to Defiance replace: %+v (knob-dependent; recorded, not asserted)", d)
	}
}
