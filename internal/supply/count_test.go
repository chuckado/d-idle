package supply

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func pots(belt, pack int, k Kind) []Potion {
	var ps []Potion
	for i := 0; i < belt; i++ {
		ps = append(ps, Potion{Kind: k, Tier: 1, InBelt: true})
	}
	for i := 0; i < pack; i++ {
		ps = append(ps, Potion{Kind: k, Tier: 1})
	}
	return ps
}

func TestCountBeltPlusPack(t *testing.T) {
	c := CountOf(append(pots(3, 2, HP), pots(1, 1, MP)...), Floors{})
	if c.Healing() != 5 || c.Mana() != 2 {
		t.Fatalf("healing %d mana %d, want 5 and 2 (belt plus pack for both)", c.Healing(), c.Mana())
	}
}

func TestBelowFloorSeparate(t *testing.T) {
	ps := []Potion{{Kind: HP, Tier: 1, InBelt: true}, {Kind: HP, Tier: 3, InBelt: true}}
	c := CountOf(ps, Floors{HP: 2})
	if c.Healing() != 1 || c.BelowFloor[HP] != 1 {
		t.Fatalf("healing %d belowfloor %d", c.Healing(), c.BelowFloor[HP])
	}
}

func TestUnreadableStashCountsZero(t *testing.T) {
	c := Count{Purse: 10, Stash: 6900}
	if c.Wealth() != 10 {
		t.Fatalf("wealth %d, want 10 with an unreadable stash", c.Wealth())
	}
}

// T9: the planner names no class or skill.
func TestNoClassLiterals(t *testing.T) {
	re := regexp.MustCompile(`(?i)\b(sorceress|paladin|barbarian|amazon|necromancer|druid|assassin|frost nova|static field)\b`)
	ents, _ := os.ReadDir(".")
	for _, e := range ents {
		if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		b, _ := os.ReadFile(e.Name())
		if m := re.Find(b); m != nil {
			t.Errorf("%s names %q", e.Name(), m)
		}
	}
}

// The old kit and guard counters counted rejuvenation as healing, and the resource policy read the best belt tier.
func TestHealingIncludesRejuvAndTier(t *testing.T) {
	ps := []Potion{{Kind: HP, Tier: 2, InBelt: true}, {Kind: RV, Tier: 1, InBelt: true}, {Kind: RV, Tier: 2}, {Kind: HP, Tier: 4}}
	c := CountOf(ps, Floors{})
	if c.Healing() != 4 || c.BeltTier[HP] != 2 || c.Belt[RV] != 1 || c.Pack[RV] != 1 {
		t.Fatalf("%+v healing %d", c, c.Healing())
	}
}
