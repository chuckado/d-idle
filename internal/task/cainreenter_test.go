package task

import "testing"

func cainCage(t *testing.T) Step {
	for _, n := range cainNodes() {
		if n.Step.ID == id(QCain, "cage") {
			return n.Step
		}
	}
	t.Fatal("no cage node")
	return Step{}
}

// Proving Ground cain.chain (2026-11-19): the hero left Tristram on a potion trip and the way back led to the Dark Wood.
// The cage step is needed again outside Tristram and splices in the walk to Stony Field and the red portal; it asks for
// nothing in Tristram, and not once Cain is free.
func TestCainCageReentersThroughStonyField(t *testing.T) {
	cage := cainCage(t)
	for _, area := range []Area{RogueEncampment, DarkWood, StonyField} {
		w := World{Area: area, InTown: area == RogueEncampment, Items: map[string]int{ItemPortal: 1}}
		if need, _ := cage.Need(w); !need {
			t.Fatalf("area %d: the cage step is not needed", area)
		}
		f := cage.Ready(w)
		if f == nil || f.Kind != KindPrereq || len(f.Insert) != 2 {
			t.Fatalf("area %d: fault %+v, want a prereq with two inserted steps", area, f)
		}
		travel, portal := f.Insert[0], f.Insert[1]
		if need, _ := travel.Need(w); need != (area != StonyField) {
			t.Errorf("area %d: travel to Stony Field needed = %v", area, need)
		}
		if need, _ := portal.Need(w); need != (area == StonyField) {
			t.Errorf("area %d: portal step needed = %v", area, need)
		}
	}
	if f := cage.Ready(World{Area: Tristram}); f != nil {
		t.Fatalf("in Tristram: %+v", f)
	}
}
