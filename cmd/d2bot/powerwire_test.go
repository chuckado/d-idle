package main

import (
	"testing"

	"d2lod-bot/internal/danger"
	"d2lod-bot/internal/leveler"
	"d2lod-bot/internal/plandata"
	"d2lod-bot/internal/power"
)

// TestShadowBuildResolves: for every class the leveler selects a build at several levels and the shadow hook can turn it
// into an archetype (the build name maps to a plandata file, the combat profile exists, every skill has a kind).
func TestShadowBuildResolves(t *testing.T) {
	m, err := power.NewModel(danger.DefaultParams())
	if err != nil {
		t.Fatal(err)
	}
	for _, class := range []string{"amazon", "sorceress", "necromancer", "paladin", "barbarian", "druid", "assassin"} {
		for _, lvl := range []int{1, 12, 30} {
			name := leveler.SelectBuild(class, lvl, nil)
			spec, err := power.SpecFromBuild(plandata.FS, "build-"+name+".json", lvl, classAttacks)
			if err != nil {
				t.Errorf("%s@%d (%s): %v", class, lvl, name, err)
				continue
			}
			if _, err := m.P.Resolve(spec); err != nil {
				t.Errorf("%s@%d (%s): %v", class, lvl, name, err)
			}
		}
	}
}

func TestWirePowerWithoutTables(t *testing.T) {
	if wirePower(nil) != nil {
		t.Error("no tables must give no hook")
	}
}

func TestWireSurvivalWithoutTables(t *testing.T) {
	c, o := wireSurvival(nil, t.TempDir()+"/ledger.json", nil)
	if c != nil || o != nil {
		t.Error("no tables must give no hooks")
	}
}

func TestWireSurvivalModelWithoutTables(t *testing.T) {
	c, o, d, b := wireSurvivalModel(nil, t.TempDir()+"/ledger.json", nil)
	if c != nil || o != nil || d != nil || b != nil {
		t.Error("no tables must give no hooks: the planner then keeps the plan tables")
	}
}
