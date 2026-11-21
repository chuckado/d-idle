package main

import (
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
	for _, class := range []string{"paladin", "sorceress"} {
		if len(by[class]) < 2 {
			t.Fatalf("%s: builds %v", class, by[class])
		}
		names := map[string]bool{}
		for _, b := range by[class] {
			names[b.Name] = true
			t.Logf("%s %s tier=%s needs=%d", class, b.Name, b.Tier, len(b.Needs))
		}
	}
	for _, want := range []string{"paladin-normal30-v2", "sorceress_normal30", "sorceress_fire_normal30"} {
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
