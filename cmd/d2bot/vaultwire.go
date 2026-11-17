package main

import (
	"log/slog"
	"path/filepath"
	"sync"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/loot"
	"d2lod-bot/internal/vault"
)

// wireVault gives the looter the vault verbs of the d2d client and a per-hero store under <data>/vaults (docs/vault.md).
// Inert until the knob loot.vault is 1; the looter forces it off in measured runs and when the fork lacks the verbs.
// The store opens on the first record, for the hero the process plays at that moment.
func wireVault(lt *loot.Looter, c *d2dctl.Client, dataDir string, hero func() string, log *slog.Logger) {
	if lt == nil || c == nil || dataDir == "" {
		return
	}
	var mu sync.Mutex
	var cur *vault.Store
	var curHero string
	lt.UseVault(&loot.VaultWiring{
		API: c,
		Store: func() (loot.VaultRecorder, error) {
			mu.Lock()
			defer mu.Unlock()
			h := hero()
			if cur != nil && curHero == h {
				return cur, nil
			}
			s, err := vault.OpenHero(filepath.Join(dataDir, "vaults"), h, vault.Options{})
			if err != nil {
				return nil, err
			}
			if cur != nil {
				_ = cur.Close()
			}
			cur, curHero = s, h
			return s, nil
		},
		Notice: func(text string) { log.Info("vault: " + text) },
	})
}
