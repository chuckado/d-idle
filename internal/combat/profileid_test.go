package combat

import (
	"testing"

	"d2lod-bot/internal/d2"
)

func TestLookupProfileByID(t *testing.T) {
	for _, c := range []struct {
		class d2.Class
		id    Build
		want  string
	}{
		{d2.Sorceress, "", "sorceress/blizzard"}, // the zero value is the old default
		{d2.Sorceress, BuildBlizzard, "sorceress/blizzard"},
		{d2.Sorceress, BuildLightning, "sorceress/lightning"},
		{d2.Sorceress, "sorceress/lightning", "sorceress/lightning"},
		{d2.Paladin, "", "paladin"},
		{d2.Paladin, BuildBlizzard, "paladin"}, // legacy ids still reach a single-profile class
		{d2.Paladin, "paladin", "paladin"},
	} {
		p, err := LookupProfile(c.class, c.id)
		if err != nil || p.ProfileID() != c.want {
			t.Errorf("%v %q: %v %v, want %s", c.class, c.id, p, err, c.want)
		}
	}
}

func TestLookupProfileUnknownIsAnError(t *testing.T) {
	for _, c := range []struct {
		class d2.Class
		id    Build
	}{
		{d2.Class(42), ""},          // unknown class: never the Sorceress
		{d2.Sorceress, "poison"},    // unknown profile of a known class
		{d2.Paladin, "zealot-lite"}, // single-profile classes only know their own id
	} {
		if p, err := LookupProfile(c.class, c.id); err == nil || p != nil {
			t.Errorf("%v %q: got %v, want an error", c.class, c.id, p)
		}
	}
}

func TestSelectProfileCheckedErrorsAndFallbackLogs(t *testing.T) {
	cat := testCatalog()
	if _, err := SelectProfileChecked(&d2.Player{Class: d2.Class(42)}, cat, ""); err == nil {
		t.Fatal("unknown class must be an error")
	}
	if _, err := SelectProfileChecked(&d2.Player{Class: d2.Sorceress}, cat, "poison"); err == nil {
		t.Fatal("unknown profile must be an error")
	}
	// the lenient form still returns something the fight loop can use
	if pr := SelectProfile(&d2.Player{Class: d2.Sorceress}, cat, "poison"); pr == nil || pr.Class != "sorceress" {
		t.Fatalf("%v", pr)
	}
}

// The sorceress/fire line (sorc2): the second Sorceress build selects it by name, Fire Ball leads once owned.
func TestSorceressFireProfile(t *testing.T) {
	pr, err := LookupProfile(d2.Sorceress, "fire")
	if err != nil || pr == nil || pr.Bands[0].Requires != "Fire Ball" || pr.Bands[0].Attack[0] != "Fire Ball" {
		t.Fatalf("fire profile: %v %+v", err, pr)
	}
}
