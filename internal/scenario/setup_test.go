package scenario

import (
	"strings"
	"testing"
)

// pgfix4: the ring of loot.rare-ring-classified was dropped in the town because "debug give" ran before "debug goto":
// a ground item is placed beside the hero, so it must come after the position verb. Both classes get the same order.
func TestGroundItemAfterPosition(t *testing.T) {
	for _, class := range []string{"Paladin", "Sorceress"} {
		s := Scenario{ID: "x.ring", Kind: KindBehaviour, Pre: Pre{Level: 10, Start: StartField, Area: 2,
			Items: []Item{{Code: "rin", Where: "ground", Quality: 6, Unid: true}}}}
		su, err := BuildSetup(s, class, 3)
		if err != nil {
			t.Fatal(err)
		}
		goto_, give := -1, -1
		for i, v := range su.Verbs {
			switch {
			case strings.HasPrefix(v.Line, "debug goto"):
				goto_ = i
			case strings.HasPrefix(v.Line, "debug give rin"):
				give = i
			}
		}
		if goto_ < 0 || give < 0 || give < goto_ {
			t.Errorf("%s: goto at %d, give at %d: the ring must be dropped after the hero is in the field", class, goto_, give)
		}
	}
}
