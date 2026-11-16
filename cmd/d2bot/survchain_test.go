package main

import (
	"testing"

	"d2lod-bot/internal/combat"
)

// TestChainSurvFightWithoutSensesRig: the fight hook must reach the ledger when no senses rig chains OnSummary.
func TestChainSurvFightWithoutSensesRig(t *testing.T) {
	var got, prevSeen int
	fn := func(combat.FightSummary) { got++ }
	survFight.Store(&fn)
	defer survFight.Store(nil)
	f := &combat.Fighter{OnSummary: func(combat.FightSummary) { prevSeen++ }}
	chainSurvFight(f)
	f.OnSummary(combat.FightSummary{Kills: 3})
	if got != 1 || prevSeen != 1 {
		t.Fatalf("ledger hook %d, previous hook %d", got, prevSeen)
	}
	chainSurvFight(nil)
}
