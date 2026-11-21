package task

import "testing"

// Sandbox hero 2026-11-21: record 4 read 0x0018 (bit 3 left town, bit 4 stones done), the pack held no scroll and no key,
// and the tree gave nothing (the game marks the tree used for a hero that joined with bit 4). Every field step before the
// cage must read done from the record word, for a melee hero and a caster alike.
func TestCainFieldStepsFollowStonesBit(t *testing.T) {
	before := []string{"wp-dark-wood", "to-wood", "tree", "read", "to-stony", "stones", "to-tristram"}
	for _, caster := range []bool{false, true} { // Paladin, Sorceress
		mk := func(raw uint16) World {
			return World{Area: DarkWood, Caster: caster, Items: map[string]int{},
				Quests: map[string]QuestState{QCain: {Stage: StageStarted, Ev: EvMemory, Raw: raw}}}
		}
		for _, n := range cainNodes() {
			base := n.Step.ID[len(id(QCain, "")):]
			isBefore := false
			for _, b := range before {
				isBefore = isBefore || b == base
			}
			if !isBefore {
				continue
			}
			if need, _ := n.Step.Need(mk(0x0018)); need {
				t.Errorf("caster=%v: step %s still needed with the word 0x0018", caster, base)
			}
			if base == "tree" {
				for _, raw := range []uint16{0x0000, 0x0008, 0x000c} { // bit 4 clear: the scroll is still to take
					if need, _ := n.Step.Need(mk(raw)); !need {
						t.Errorf("caster=%v: the tree step reads done with the word 0x%04x", caster, raw)
					}
				}
			}
		}
	}
}
