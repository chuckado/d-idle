package combat

import (
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
)

func lvlCat() MapCatalog {
	return MapCatalog{Skills: map[int]*data.Skill{
		38: {ID: 38, Key: "Charged Bolt", Name: "Charged Bolt", Class: "sor", Page: 2, EType: "ltng", Mana: 24, LvlMana: 4, ManaShift: 5, MinMana: 1},
	}}
}

func lvlPlayer(st d2.Stats, base int) *d2.Player {
	return &d2.Player{Class: d2.Sorceress, Stats: st, Skills: []d2.Skill{{ID: 38, Level: base}}}
}

func chargedCost(t *testing.T, p *d2.Player) int {
	t.Helper()
	sp, ok := KnownSkills(p, lvlCat(), nil)["Charged Bolt"]
	if !ok {
		t.Fatal("Charged Bolt not known")
	}
	return sp.Cost
}

// No bonus stat: the level is the base points (regression of the old behaviour).
func TestKnownSkillsNoBonus(t *testing.T) {
	// (24 + 4*4) << 5 / 256 = 5
	if got := chargedCost(t, lvlPlayer(d2.Stats{}, 5)); got != 5 {
		t.Errorf("cost %d, want 5", got)
	}
}

// A +3 Lightning tab (stat 188, sub class*8+tab-1 = 9) raises the cast level, so the mana cost, while the base
// points stay the same.
func TestKnownSkillsTabBonus(t *testing.T) {
	p := lvlPlayer(d2.Stats{9<<16 | 188: 3}, 5)
	// level 8: (24 + 4*7) << 5 / 256 = 6.5 -> 7
	if got := chargedCost(t, p); got != 7 {
		t.Errorf("cost %d, want 7", got)
	}
	if p.Base(38) != 5 {
		t.Errorf("base %d, want 5", p.Base(38))
	}
}

// A skill with no points yet but +all skills is usable at that level.
func TestKnownSkillsAllSkillsUnlearned(t *testing.T) {
	if _, ok := KnownSkills(lvlPlayer(d2.Stats{}, 0), lvlCat(), nil)["Charged Bolt"]; ok {
		t.Error("level 0 skill without bonus must be unknown")
	}
	if _, ok := KnownSkills(lvlPlayer(d2.Stats{127: 1}, 0), lvlCat(), nil)["Charged Bolt"]; !ok {
		t.Error("+1 all skills must make it known")
	}
}

func TestSkillRefOf(t *testing.T) {
	p := &d2.Player{Class: d2.Sorceress}
	got := SkillRefOf(p, lvlCat().Skills[38])
	if got != (d2.SkillRef{Own: true, Tab: 2, Elem: d2.ElemLight}) {
		t.Errorf("ref %+v", got)
	}
	if SkillRefOf(p, &data.Skill{Class: ""}).Own || SkillRefOf(p, &data.Skill{Class: "pal"}).Own {
		t.Error("generic and other-class skills are not own-class")
	}
}
