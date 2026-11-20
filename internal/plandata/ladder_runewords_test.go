package plandata

import (
	"slices"
	"testing"
)

func TestLadderRunewordsTwoClasses(t *testing.T) {
	for _, class := range []string{"paladin", "sorceress"} {
		w, err := LadderRunewords(class)
		if err != nil || !slices.Contains(w, "Spirit") || !slices.Contains(w, "Insight") {
			t.Errorf("%s: %v %v", class, w, err)
		}
	}
	if _, err := LadderRunewords("barbarian"); err == nil {
		t.Error("an unlisted class must be an error, not a fallback")
	}
}
