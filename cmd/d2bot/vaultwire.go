package main

import (
	"log/slog"
	"path/filepath"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/loot"
)

// wireVault gives the looter the vault verbs of the d2d client and a per-hero store under <data>/vaults (docs/vault.md).
// Inert until the knob loot.vault is 1; the looter forces it off in measured runs and when the fork lacks the verbs.
// The store opens on the first record, for the hero the process plays at that moment.
func wireVault(lt *loot.Looter, c *d2dctl.Client, dataDir string, hero func() string, log *slog.Logger) {
	if lt == nil || c == nil || dataDir == "" {
		return
	}
	home := vaultHomeFor(filepath.Join(dataDir, "vaults")) // shared with the Vault page (vaultpage.go): one Store per hero
	lt.UseVault(&loot.VaultWiring{
		API: c,
		Store: func() (loot.VaultRecorder, error) {
			s, err := home.Open(hero())
			if err != nil {
				return nil, err
			}
			return s, nil
		},
		Notice: func(text string) { log.Info("vault: " + text) },
	})
}
