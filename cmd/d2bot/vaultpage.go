package main

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"d2lod-bot/internal/bulkid"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/loot"
	"d2lod-bot/internal/vault"
	"d2lod-bot/internal/vaultkeep"
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
	capf atomic.Value      // func() int: the capacity the stores read at every insert (the ETHER wallet's owned tiers, etherwire.go)
	keep *vaultkeep.Keeper // the Keep queue (vaultkeep.go); nil until wireVaultKeep
}

var vaultHomes sync.Map // root -> *vaultHome

func vaultHomeFor(root string) *vaultHome {
	h, _ := vaultHomes.LoadOrStore(root, &vaultHome{root: root})
	return h.(*vaultHome)
}

// setCapacity late-binds the capacity of every store of this root, those already open included.
func (h *vaultHome) setCapacity(f func() int) { h.capf.Store(f) }

// Open returns the hero's vault, creating it if needed. A hero switch closes the old Store.
func (h *vaultHome) Open(hero string) (*vault.Store, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.st != nil && h.hero == hero {
		return h.st, nil
	}
	st, err := vault.OpenHero(h.root, hero, vault.Options{Capacity: h.storeCapacity})
	if err != nil {
		return nil, err
	}
	if h.st != nil {
		_ = h.st.Close()
	}
	h.hero, h.st = hero, st
	return st, nil
}

// storeCapacity is nil-safe for the vault: a bound wallet answers, otherwise the table base.
func (h *vaultHome) storeCapacity() int {
	if f, _ := h.capf.Load().(func() int); f != nil {
		if n := f(); n > 0 {
			return n
		}
	}
	t, err := vault.DefaultTiers()
	if err != nil {
		return 400
	}
	return t.Capacity(nil, 0)
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
	return vaultAPIWith(root, on, hero, nil, nil)
}

// vaultAPIWith is vaultAPI with the ETHER wallet of the account: the capacity line and the tiers owned come from it.
func vaultAPIWith(root string, on func() bool, hero func() string, eh *etherHome, db *data.DB) (*watch.VaultAPI, error) {
	tiers, err := vault.DefaultTiers()
	if err != nil {
		return nil, err
	}
	home := vaultHomeFor(root)
	api := watch.NewVaultAPI()
	desk := watch.VaultDesk{On: on, Tiers: tiers, Store: func() *vault.Store { return home.Existing(hero()) }, Keep: keepFace{home}, Act: actFace{home}, Tip: vaultTipFor(db), Quality: loot.QualityName, ScrollGold: scrollGold(db)}
	if f, ferr := bulkid.DefaultFilter(); ferr == nil {
		desk.MassFilter = f
	}
	if eh != nil {
		desk.Owned = eh.owned
	}
	api.Bind(desk)
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

// scrollGold is the base price of a Scroll of Identify from the item tables (information in the mass reveal line; 0 when
// the tables lack it).
func scrollGold(db *data.DB) int {
	if db == nil {
		return 0
	}
	if b := db.ItemByCode("isc"); b != nil {
		return b.Cost
	}
	return 0
}
