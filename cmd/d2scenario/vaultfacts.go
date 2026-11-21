package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"d2lod-bot/internal/scenario"
	"d2lod-bot/internal/vault"
)

// vaultSampleFile is the side file that holds the last GET /api/vault answer of the run.
func vaultSampleFile(trace string) string {
	return strings.TrimSuffix(trace, ".bot.log") + ".vault.json"
}

// sampleVault reads GET /api/vault every `every` until ctx ends and keeps the last good answer in path (the viewer
// listener is gone when the bot has exited, so the end state is the last sample before it).
func sampleVault(ctx context.Context, view, token string, every time.Duration, path string) {
	client := &http.Client{Timeout: 10 * time.Second}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(every):
		}
		req, _ := http.NewRequestWithContext(ctx, "GET", view+"/api/vault", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
		resp.Body.Close()
		if resp.StatusCode == 200 && len(body) > 0 {
			_ = os.WriteFile(path, body, 0o644)
		}
	}
}

// addVaultFacts reads the hero's vault store (the bot has exited and closed it), the journal of the run and the last API
// sample into the state. Nothing is added for a scenario that asserts nothing on the vault.
func addVaultFacts(st *scenario.State, s scenario.Scenario, dataDir, hero, trace string, so seedOut) {
	if !scenario.NeedsVault(s) {
		return
	}
	dir := vault.HeroDir(filepath.Join(dataDir, "vaults"), hero)
	vf := &scenario.VaultFacts{}
	if store, err := vault.Open(dir, hero, vault.Options{NoSync: true}); err == nil {
		seeded := map[string]bool{}
		for _, id := range so.VaultIDs {
			seeded[id] = true
		}
		for _, r := range store.Dump() {
			vf.Records = append(vf.Records, scenario.VaultRecFact{ID: r.ID, Code: r.Code, State: string(r.State), Pin: r.Pin, Identified: r.Identified, Seeded: seeded[r.ID]})
		}
		_ = store.Close()
		log, _ := os.ReadFile(filepath.Join(dir, vault.LogFile))
		bot, _ := os.ReadFile(trace)
		vf.Journal = scenario.BuildJournal(log, so.Seq, bot)
		st.Has = append(st.Has, "vault")
	}
	if b, err := os.ReadFile(vaultSampleFile(trace)); err == nil {
		vf.API = b
	}
	st.Vault = vf
}
