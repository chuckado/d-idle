package main

import (
	"os"
	"path/filepath"
	"sync"

	"d2lod-bot/internal/vault"
	"d2lod-bot/internal/watch"
)

// The Ethereal Vault page's desk (docs/vault.md, slice V4): the active hero's vault under <state dir>/<hero>/vault, opened
// only when it already exists on disk (the page never creates one; the recorder of slice V3 does), and shared: V3 must take
// its Store from vaultHome.Store as well, because two Store values over one log would each hold a different index.
type vaultHome struct {
	root string
	mu   sync.Mutex
	hero string
	st   *vault.Store
	opt  vault.Options
}

func newVaultHome(root string, tiers *vault.Tiers) *vaultHome {
	return &vaultHome{root: root, opt: vault.Options{Tiers: tiers}}
}

// Store returns the hero's vault, opening it when its log exists; nil while there is none. A hero switch closes the old one.
func (h *vaultHome) Store(hero string) *vault.Store {
	if hero == "" || vault.ValidHero(hero) != nil {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.st != nil && h.hero == hero {
		return h.st
	}
	if _, err := os.Stat(filepath.Join(vault.HeroDir(h.root, hero), vault.LogFile)); err != nil {
		return nil
	}
	st, err := vault.OpenHero(h.root, hero, h.opt)
	if err != nil {
		return nil
	}
	if h.st != nil {
		h.st.Close()
	}
	h.hero, h.st = hero, st
	return st
}

// vaultAPI builds the page's routes. on reads the knob loot.vault; hero names the active hero.
func vaultAPI(stateDir string, on func() bool, hero func() string) (*watch.VaultAPI, error) {
	tiers, err := vault.DefaultTiers()
	if err != nil {
		return nil, err
	}
	home := newVaultHome(stateDir, tiers)
	api := watch.NewVaultAPI()
	api.Bind(watch.VaultDesk{On: on, Tiers: tiers, Store: func() *vault.Store { return home.Store(hero()) }})
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
