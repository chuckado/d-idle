package main

import (
	"d2lod-bot/internal/leveler"
	"strings"
	"testing"

	"d2lod-bot/internal/agency"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/points"
	"d2lod-bot/internal/power"
)

func testAllocWire(t *testing.T) *allocWire {
	caps, err := agency.LoadCaps()
	if err != nil {
		t.Fatal(err)
	}
	w := &allocWire{db: &data.DB{}, caps: caps}
	if p, err := power.Load(); err == nil {
		w.kinds = p.SkillKinds
	}
	return w
}

func TestAllocBuildListTwoClasses(t *testing.T) {
	w := testAllocWire(t)
	by := map[string][]points.BuildInfo{}
	for _, b := range w.builds() {
		by[strings.ToLower(b.Class)] = append(by[strings.ToLower(b.Class)], b)
		if b.Tier != points.BuildSupported && b.Tier != points.BuildExperimental {
			t.Errorf("%s: tier %q", b.Name, b.Tier)
		}
		if len(b.Preview) == 0 || b.Preview[len(b.Preview)-1].Clvl > previewLevels {
			t.Errorf("%s: preview rows %d", b.Name, len(b.Preview))
		}
		if b.Tier == points.BuildExperimental && len(b.Needs) == 0 {
			t.Errorf("%s: experimental without a stated need", b.Name)
		}
	}
	for _, class := range []string{"paladin", "sorceress", "necromancer"} {
		if len(by[class]) < 2 {
			t.Fatalf("%s: builds %v", class, by[class])
		}
		names := map[string]bool{}
		for _, b := range by[class] {
			names[b.Name] = true
			t.Logf("%s %s tier=%s needs=%d", class, b.Name, b.Tier, len(b.Needs))
		}
	}
	for _, want := range []string{"necromancer_summoner_sme", "necromancer_poison_nova_sme", "paladin-normal30-v2", "sorceress_normal30", "sorceress_fire_normal30"} {
		found := false
		for _, bs := range by {
			for _, b := range bs {
				found = found || b.Name == want
			}
		}
		if !found {
			t.Errorf("build %s is not offered", want)
		}
	}
}

// necropick: the picker offers each class only its own builds and takes the note from the build catalogue.
func TestAllocPickerNecromancerNotes(t *testing.T) {
	w := testAllocWire(t)
	for _, b := range w.builds() {
		switch b.Name {
		case "necromancer_summoner_sme", "necromancer_poison_nova_sme":
			if !strings.EqualFold(b.Class, "necromancer") || b.Note == "" {
				t.Errorf("%s: class %q note %q", b.Name, b.Class, b.Note)
			}
		case "sorceress_normal30":
			if !strings.EqualFold(b.Class, "sorceress") {
				t.Errorf("%s: class %q", b.Name, b.Class)
			}
		}
	}
}

// The summoner opens with Vitality, 3 Energy at clvl 2 only (SME answers 2026-11-22); the table once put 10 Energy first
// and the hero sat at clvl 4 with its mana pool grown and no Vitality (user report 2026-11-23).
func TestSummonerTableStatsAreVitalityFirst(t *testing.T) {
	b, err := leveler.LoadBuild("necromancer_summoner_normal30", "")
	if err != nil {
		t.Fatal(err)
	}
	ene, vit := 0, 0
	for _, r := range b.Rows {
		if r.Clvl > 10 {
			break
		}
		ene += r.Attr[3]
		vit += r.Attr[2]
	}
	if ene != 3 || vit != 5*9-3 {
		t.Errorf("clvl 2-10: energy %d vitality %d", ene, vit)
	}
}
