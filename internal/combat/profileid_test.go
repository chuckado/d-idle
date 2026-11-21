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

// necro1: the Necromancer summoner line is a second profile selected by name; the class default stays the Teeth line,
// and an unknown Necromancer profile is an error that names it, never another class or the default.
func TestNecromancerSummonerProfile(t *testing.T) {
	def, err := LookupProfile(d2.Necromancer, "")
	if err != nil || def.Build != "" || def.Bands[0].Attack[0] != "Bone Spear" {
		t.Fatalf("necromancer default changed: %v %+v", err, def)
	}
	sm, err := LookupProfile(d2.Necromancer, "summoner")
	if err != nil || sm.ProfileID() != "necromancer/summoner" {
		t.Fatalf("summoner profile: %v %+v", err, sm)
	}
	if _, ok := sm.Layout["Raise Skeleton"]; !ok {
		t.Errorf("summoner layout has no Raise Skeleton key: %v", sm.Layout)
	}
	// the summoner never ranks the Bone line as an attack
	for _, b := range sm.Bands {
		for _, a := range b.Attack {
			if a == "Bone Spear" {
				t.Errorf("summoner band attacks with %s", a)
			}
		}
	}
	for _, c := range []struct {
		class d2.Class
		id    Build
	}{{d2.Necromancer, "bone"}, {d2.Paladin, "summoner"}, {d2.Sorceress, "summoner"}, {d2.Class(42), "summoner"}} {
		if p, err := LookupProfile(c.class, c.id); err == nil || p != nil {
			t.Errorf("%v %q: got %v, want an error", c.class, c.id, p)
		}
	}
}

// hammer1: the Paladin Hammerdin line is a second profile selected by name; the class default stays the Zeal and
// Sacrifice line, and an unknown Paladin profile (or another class asking for "hammerdin") is an error that names it.
func TestPaladinHammerdinProfile(t *testing.T) {
	def, err := LookupProfile(d2.Paladin, "")
	if err != nil || def.Build != "" || def.Bands[1].Attack[0] != "Zeal" {
		t.Fatalf("paladin default changed: %v %+v", err, def)
	}
	hd, err := LookupProfile(d2.Paladin, "hammerdin")
	if err != nil || hd.ProfileID() != "paladin/hammerdin" {
		t.Fatalf("hammerdin profile: %v %+v", err, hd)
	}
	if hd.Bands[0].Attack[0] != "Blessed Hammer" {
		t.Errorf("hammerdin does not lead with Blessed Hammer: %v", hd.Bands[0].Attack)
	}
	for _, n := range []string{"Blessed Hammer", "Holy Shield", "Concentration", "Vigor"} {
		if _, ok := hd.Layout[n]; !ok {
			t.Errorf("hammerdin layout has no key for %s: %v", n, hd.Layout)
		}
	}
	// never a Zeal or Sacrifice attack, and no life price
	for _, b := range hd.Bands {
		for _, a := range b.Attack {
			if a == "Zeal" || a == "Sacrifice" {
				t.Errorf("hammerdin band attacks with %s", a)
			}
		}
	}
	if len(hd.LifePrice) != 0 {
		t.Errorf("hammerdin pays life for %v", hd.LifePrice)
	}
	for _, c := range []struct {
		class d2.Class
		id    Build
	}{{d2.Paladin, "zealot"}, {d2.Necromancer, "hammerdin"}, {d2.Sorceress, "hammerdin"}, {d2.Class(42), "hammerdin"}} {
		if p, err := LookupProfile(c.class, c.id); err == nil || p != nil {
			t.Errorf("%v %q: got %v, want an error", c.class, c.id, p)
		}
	}
}
