package town

import (
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
)

// No game tables: the automatic identify purchases are off by default and the need checks say so.
func TestAutoIdentifyBuysAreOffByDefault(t *testing.T) {
	tu := env.DefaultTuning()
	if tu.Town.IdentifyTome != 0 {
		t.Fatalf("town.identifyTome defaults to %d, want 0", tu.Town.IdentifyTome)
	}
	tn := &Town{e: &env.Env{Tune: &tu}}
	s := &d2.Snapshot{Player: &d2.Player{Gold: 9999}, Inventory: &d2.Inventory{}}
	if tn.needIDTome(s) {
		t.Fatal("the identify tome chore wants to run with the knob off")
	}
}
