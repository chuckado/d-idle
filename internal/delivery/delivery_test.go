package delivery

import (
	"bytes"
	"os"
	"strings"
	"testing"

	capdata "d2lod-bot/data/capabilities"
)

func load(t *testing.T) *Table {
	t.Helper()
	tb, err := Load(capdata.JSON)
	if err != nil {
		t.Fatal(err)
	}
	return tb
}

// The generated views must equal the checked-in files: the CI check of the card (go run ./tools/capgen -check).
func TestGeneratedViewsAreCheckedIn(t *testing.T) {
	tb := load(t)
	want, err := SkillsJSON(tb)
	if err != nil {
		t.Fatal(err)
	}
	have, err := os.ReadFile("../combat/classdata/skills.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(have, want) {
		t.Error("classdata/skills.json differs from the table: run go run ./tools/capgen")
	}
	old, err := os.ReadFile("../power/profiles.json")
	if err != nil {
		t.Fatal(err)
	}
	prof, err := ProfilesJSON(old, tb)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(old, prof) {
		t.Error("power/profiles.json skillKinds differs from the table: run go run ./tools/capgen")
	}
}

// Rows of two classes keep the aiming kind the old hand-typed table gave them (it was the combat engine's input).
func TestCombatKindKeepsOldValues(t *testing.T) {
	tb := load(t)
	v, err := tb.CombatView()
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]CombatSkill{}
	for _, s := range v {
		got[s.Name] = s
	}
	for _, c := range []struct{ skill, kind, hand string }{
		{"Fire Bolt", "bolt", ""}, {"Blizzard", "ground", ""}, {"Frost Nova", "self", ""}, {"Frozen Armor", "buff", ""}, // Sorceress
		{"Zeal", "melee", "left"}, {"Might", "aura", ""}, {"Blessed Hammer", "ground", "left"}, // Paladin
	} {
		s, ok := got[c.skill]
		if !ok || s.Kind != c.kind || s.Hand != c.hand {
			t.Errorf("%s: %+v (ok %v), want kind %s hand %q", c.skill, s, ok, c.kind, c.hand)
		}
	}
	if !got["Zeal"].Hold {
		t.Error("Zeal holds the button")
	}
}

func TestTierRule(t *testing.T) {
	tb := load(t)
	none := map[string]bool{}
	tiers := tb.Tiers(none)
	// Sorceress row with a driver, one with none, a passive of the Paladin.
	for skill, want := range map[string]string{"Fire Bolt": Experimental, "Inferno": Unsupported, "Natural Resistance": Supported} {
		if tiers[skill] != want {
			t.Errorf("%s tier %s, want %s", skill, tiers[skill], want)
		}
	}
	if got := tb.Tiers(map[string]bool{"Fire Bolt": true})["Fire Bolt"]; got != Supported {
		t.Errorf("a proven row must be supported, got %s", got)
	}
	// a proven skill with no driver stays unsupported: the scenario cannot lift a missing loop.
	if got := tb.Tiers(map[string]bool{"Inferno": true})["Inferno"]; got != Unsupported {
		t.Errorf("proof without a driver must stay unsupported, got %s", got)
	}
}

func TestLoadNamesTheBadRow(t *testing.T) {
	good := string(capdata.JSON)
	for name, c := range map[string]struct{ from, to string }{
		"purpose": {`"purpose": "damage"`, `"purpose": "bash"`},
		"driver":  {`"driver": "attack"`, `"driver": "wizard"`},
		"target":  {`"target": "unit"`, `"target": "sky"`},
	} {
		bad := strings.Replace(good, c.from, c.to, 1)
		if bad == good {
			t.Fatalf("%s: marker not found", name)
		}
		if _, err := Load([]byte(bad)); err == nil || !strings.Contains(err.Error(), c.to[strings.LastIndex(c.to, " ")+2:len(c.to)-1]) {
			t.Errorf("%s: want an error that names the value, got %v", name, err)
		}
	}
	if _, err := Load([]byte(strings.Replace(good, "didle.cap/2", "didle.cap/1", 1))); err == nil {
		t.Error("an old schema must be refused")
	}
}

func TestProvenSkillsNeedAScenario(t *testing.T) {
	if _, err := ProvenSkills([]byte(`{"rows":[{"skill":"Fire Bolt"}]}`), "d2d"); err == nil {
		t.Error("a proven row without a scenario must be refused")
	}
	m, err := ProvenSkills([]byte(`{"rows":[{"skill":"A","scenario":"s","backend":"d2d"},{"skill":"B","scenario":"s","backend":"other"}]}`), "d2d")
	if err != nil || !m["A"] || m["B"] {
		t.Errorf("backend filter: %v %v", m, err)
	}
}
