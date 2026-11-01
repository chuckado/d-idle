package combat

import (
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
)

func teleCatalog() MapCatalog {
	c := testCatalog()
	c.Skills[54] = &data.Skill{ID: 54, Key: "Teleport", Name: "Teleport", Mana: 24, LvlMana: 1, ManaShift: 0, MinMana: 1}
	return c
}

func TestTeleportKeyFollowsKnownSkillAndLayout(t *testing.T) {
	cat := teleCatalog()
	sorc := func(skills ...d2.Skill) *d2.Player {
		p := baseSnap().Player
		p.Skills = skills
		return p
	}
	teleKey := func(p *d2.Player, b Build) string {
		for _, w := range WantedBindings(p, cat, b) {
			if w.Skill.Name == "Teleport" {
				return w.Key
			}
		}
		return ""
	}
	if got := teleKey(sorc(d2.Skill{ID: 54, Level: 1}, d2.Skill{ID: 38, Level: 1}), BuildBlizzard); got != "f2" {
		t.Fatalf("sorceress with Teleport: key %q, want the profile layout key f2", got)
	}
	if got := teleKey(sorc(d2.Skill{ID: 38, Level: 1}), BuildBlizzard); got != "" {
		t.Fatalf("sorceress without Teleport: key %q, want none", got)
	}
	if got := teleKey(sorc(d2.Skill{ID: 54}), BuildBlizzard); got != "" {
		t.Fatalf("Teleport at level 0 counts as not learned: key %q", got)
	}
	// a Paladin cannot learn Teleport, so its list never has it and its layout never names it
	cat.Skills[115] = &data.Skill{ID: 115, Key: "Holy Fire", Name: "Holy Fire", Mana: 0}
	pal := sorc(d2.Skill{ID: 115, Level: 1})
	pal.Class = d2.Paladin
	if got := teleKey(pal, BuildBlizzard); got != "" {
		t.Fatalf("paladin: key %q, want none", got)
	}
}

func TestFighterTeleportKeyReadsVerifiedBind(t *testing.T) {
	cat := teleCatalog()
	f := &Fighter{cat: cat, keys: map[int]string{}}
	if k := f.TeleportKey(); k != "" {
		t.Fatalf("nothing bound: %q", k)
	}
	f.SetHotkeys(map[int]string{38: "f4", 54: "f7"}) // a build that binds Teleport to another key
	if k := f.TeleportKey(); k != "f7" {
		t.Fatalf("key %q, want f7", k)
	}
	f.SetHotkeys(map[int]string{38: "f4"})
	if k := f.TeleportKey(); k != "" {
		t.Fatalf("Teleport unbound again: %q", k)
	}
	if k := (&Fighter{keys: map[int]string{54: "f2"}}).TeleportKey(); k != "" {
		t.Fatalf("no catalog: %q", k)
	}
	if k := f.SkillKey("No Such Skill"); k != "" {
		t.Fatalf("unknown skill name: %q", k)
	}
}
