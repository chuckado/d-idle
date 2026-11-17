package main

import (
	"os"
	"path/filepath"
	"sync"

	"d2lod-bot/internal/vault"
	"d2lod-bot/internal/watch"
)

// The Ethereal Vault of the process (docs/vault.md): one Store per root and hero, shared by the recorder (wireVault, slice V3) and
// the page (slice V4). Two Store values over one log would each hold their own index and miss the other's records, so both
// go through vaultHome. The recorder Opens (creating the vault on its first record); the page only reads an Existing one and
// never creates a directory.
type vaultHome struct {
	root string
	mu   sync.Mutex
	hero string
	st   *vault.Store
}

var vaultHomes sync.Map // root -> *vaultHome

func vaultHomeFor(root string) *vaultHome {
	h, _ := vaultHomes.LoadOrStore(root, &vaultHome{root: root})
	return h.(*vaultHome)
}

// Open returns the hero's vault, creating it if needed. A hero switch closes the old Store.
func (h *vaultHome) Open(hero string) (*vault.Store, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.st != nil && h.hero == hero {
		return h.st, nil
	}
	st, err := vault.OpenHero(h.root, hero, vault.Options{})
	if err != nil {
		return nil, err
	}
	if h.st != nil {
		_ = h.st.Close()
	}
	h.hero, h.st = hero, st
	return st, nil
}

// Existing returns the hero's vault when it is open or its log exists on disk; nil while there is none.
func (h *vaultHome) Existing(hero string) *vault.Store {
	if hero == "" || vault.ValidHero(hero) != nil {
		return nil
	}
	h.mu.Lock()
	open := h.st != nil && h.hero == hero
	h.mu.Unlock()
	if open {
		st, _ := h.Open(hero)
		return st
	}
	if _, err := os.Stat(filepath.Join(vault.HeroDir(h.root, hero), vault.LogFile)); err != nil {
		return nil
	}
	st, err := h.Open(hero)
	if err != nil {
		return nil
	}
	return st
}

// vaultAPI builds the page's routes over <root>/<hero>/vault (root is <data>/vaults, as in wireVault). on reads the knob loot.vault; hero names the active hero.
func vaultAPI(root string, on func() bool, hero func() string) (*watch.VaultAPI, error) {
	tiers, err := vault.DefaultTiers()
	if err != nil {
		return nil, err
	}
	home := vaultHomeFor(root)
	api := watch.NewVaultAPI()
	api.Bind(watch.VaultDesk{On: on, Tiers: tiers, Store: func() *vault.Store { return home.Existing(hero()) }})
	return api, nil
}

// vaultKnobOn is the loot.vault knob as a bool; an unreadable knob counts as off.
func vaultKnobOn(knob func(string) (float64, string, error)) func() bool {
	return func() bool {
		if knob == nil {
			return false
		}
		v, _, err := knob("loot.vault")
		return err == nil && v != 0
	}
}
