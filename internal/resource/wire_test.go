package resource

import (
	"fmt"
	"testing"
)

type wireCase struct {
	name string
	mod  func(*Scenario)
}

// The early-game cases of docs/skmeleewire.md: the stock Blood Moor pack, an undead pack (Den, graveyard), a pack
// the character is weak against by level but cannot hit (defense 100), and the same with Weak set.
var wireCases = []wireCase{
	{"stock living", func(*Scenario) {}},
	{"undead", func(s *Scenario) { s.Foe.Undead = true }},
	{"undead weak", func(s *Scenario) { s.Foe.Undead, s.Weak = true, true }},
	{"defense 100 weak", func(s *Scenario) { s.Foe = Foe{Defense: 100, Level: 8}; s.Weak = true }},
}

// Fed versus not fed: only the policy's view changes; the world rolls the same blows either way. The model never
// makes a death where the blind policy had none, and where it changes the choice it does so in the stated direction.
func TestWireFedVersusBlind(t *testing.T) {
	tu := DefaultTuning()
	for _, c := range wireCases {
		sc := BloodMoorSorc()
		c.mod(&sc)
		blind := sc
		blind.Blind = true
		fed, bl := Replay(sc, ResourcePolicy(), tu), Replay(blind, ResourcePolicy(), tu)
		t.Logf("%-18s blind: kills %2d melee %3d casts %3d retreats %d deaths %d | fed: kills %2d melee %3d casts %3d retreats %d deaths %d",
			c.name, bl.Kills, bl.MeleeHits, bl.Casts, bl.Retreats, bl.Deaths, fed.Kills, fed.MeleeHits, fed.Casts, fed.Retreats, fed.Deaths)
		if fed.Deaths > bl.Deaths {
			t.Errorf("%s: the model costs a death", c.name)
		}
		if again := Replay(sc, ResourcePolicy(), tu); again != fed {
			t.Errorf("%s: replay is not deterministic", c.name)
		}
	}
}

// Rung choice with no mana (the melee rung is all that is left) by pack size, fed and blind.
func TestWireRungTable(t *testing.T) {
	tu := DefaultTuning()
	for _, c := range wireCases {
		sc := BloodMoorSorc()
		c.mod(&sc)
		row := c.name + ":"
		changed := false
		for n := 1; n <= 6; n++ {
			in := Input{MaxMana: 45, Life: 60, MaxLife: 60, Hostiles: n, WeakPack: sc.Weak, Skills: []Skill{{Name: "x", Melee: true, Ready: true}}}
			blind := AttackChoice(in, tu).Kind
			in.Blow, in.Foe = sc.Blow, &sc.Foe
			fed := AttackChoice(in, tu).Kind
			cell := fmt.Sprintf(" %d=%v", n, fed)
			if fed != blind {
				cell = fmt.Sprintf(" %d=%v(was %v)", n, fed, blind)
				changed = true
			}
			row += cell
		}
		t.Logf("%s changed=%v", row, changed)
	}
}
