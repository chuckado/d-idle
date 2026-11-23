package main

import (
	"path/filepath"
	"testing"

	"d2lod-bot/internal/loot"
)

func TestLootRulesWireActivatesStoreAndListsClasses(t *testing.T) {
	cl := heroClasses()
	if len(cl) < 7 || cl[0] != "Amazon" {
		t.Fatalf("classes %v", cl)
	}
	defer loot.SetActiveRules(loot.DefaultEngine.Rules())
	api := lootRulesAPI(filepath.Join(t.TempDir(), "state"), "p1", nil)
	if api == nil || loot.DefaultEngine.Rules().Name != "Starter" {
		t.Fatal("not wired")
	}
}
