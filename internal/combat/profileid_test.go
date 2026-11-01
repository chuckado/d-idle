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
		{d2.Sorceress, "fire"},      // unknown profile of a known class
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
	if _, err := SelectProfileChecked(&d2.Player{Class: d2.Sorceress}, cat, "fire"); err == nil {
		t.Fatal("unknown profile must be an error")
	}
	// the lenient form still returns something the fight loop can use
	if pr := SelectProfile(&d2.Player{Class: d2.Sorceress}, cat, "fire"); pr == nil || pr.Class != "sorceress" {
		t.Fatalf("%v", pr)
	}
}
