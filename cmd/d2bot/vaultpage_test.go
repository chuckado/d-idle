package main

import (
	"os"
	"path/filepath"
	"testing"

	"d2lod-bot/internal/vault"
)

func TestVaultHomeOpensOnlyExistingVaults(t *testing.T) {
	root := t.TempDir()
	h := vaultHomeFor(root)
	if h.Existing("Sorc") != nil {
		t.Fatal("a vault that was never recorded must not be opened (the page does not create one)")
	}
	if _, err := os.Stat(filepath.Join(root, "Sorc")); err == nil {
		t.Fatal("the page created a directory")
	}
	// the recorder (V3) creates it; the page then sees the same store on every call and per hero
	rec, err := vault.OpenHero(root, "Sorc", vault.Options{})
	if err != nil {
		t.Fatal(err)
	}
	rec.Add(vault.Item{Hero: "Sorc", Raw: "0102", Code: "cap", Name: "Cap", Quality: 4, Ilvl: 5})
	rec.Close()
	a := h.Existing("Sorc")
	if a == nil || h.Existing("Sorc") != a {
		t.Fatal("the existing vault is not opened once and shared")
	}
	if _, total := a.Query(vault.Filter{}); total != 1 {
		t.Errorf("total %d", total)
	}
	if h.Existing("Pala") != nil || h.Existing("../x") != nil || h.Existing("") != nil {
		t.Error("another hero or a bad key opened a vault")
	}
	// the recorder's Open returns the very store the page reads
	if rs, err := h.Open("Sorc"); err != nil || rs != a {
		t.Errorf("Open gave another store than Existing: %v", err)
	}
}

func TestVaultKnobOn(t *testing.T) {
	val := 0.0
	knob := func(k string) (float64, string, error) {
		if k != "loot.vault" {
			t.Errorf("read %q", k)
		}
		return val, "", nil
	}
	on := vaultKnobOn(knob)
	if on() {
		t.Error("default 0 must be off")
	}
	val = 1
	if !on() || vaultKnobOn(nil)() {
		t.Error("knob 1 on, no knob off")
	}
}
