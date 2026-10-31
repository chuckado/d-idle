package env

import "testing"

// earlyCastMs is the cast of the early Sorceress bolts (V-d2d, docs/d2d-casts-real.md).
const earlyCastMs = 560

// The cast gap must not be shorter than the cast animation, or the fighter re-issues casts that d2d drops mid-cast.
func TestCastGapCoversEarlyCast(t *testing.T) {
	if g := DefaultCombatTuning().CastGapMs; g < earlyCastMs {
		t.Fatalf("CastGapMs %d is below the %d ms early cast", g, earlyCastMs)
	}
}
