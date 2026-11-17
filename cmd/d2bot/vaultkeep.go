package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"

	"d2lod-bot/internal/bot"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/vault"
	"d2lod-bot/internal/vaultkeep"
	"d2lod-bot/internal/watch"
)

// The Keep queue of the Ethereal Vault in the bot (docs/vault.md, Keep; slice V5): the executor of internal/vaultkeep over the
// d2d client, the bot step (bot.VaultKeep), the page's routes (watch.VaultKeeper) and the Needs-you row.

// activeKeeper is the process's keeper for the Needs-you row (one hero per process).
var activeKeeper atomic.Pointer[vaultkeep.Keeper]

// keepOn is the kill switch: loot.vault and loot.vaultKeep on, and never in a measured run (D2BOT_RUN).
func keepOn(e *env.Env) func() bool {
	return func() bool {
		if e == nil || e.Tune == nil || os.Getenv("D2BOT_RUN") != "" {
			return false
		}
		return e.Tune.Loot.Vault != 0 && e.Tune.Loot.VaultKeep != 0
	}
}

// keepReady is the safe-moment test of the snapshot: alive, in town, no blocking panel, the cursor free. The bot step has
// checked fight and trip state already. Proximity to the stash chest is not required: `item create` is a memory verb of
// the fork (U: the fork's own proximity test exists only for the move verbs).
func keepReady(snap func() *d2.Snapshot) func() string {
	return func() string {
		s := snap()
		switch {
		case s == nil || s.Player == nil:
			return vaultkeep.ReasonNoGame
		case s.Player.MaxLife > 0 && s.Player.Life <= 0:
			return vaultkeep.ReasonDead
		case !s.Player.InTown():
			return "not in town"
		case s.UI.BlockingPanelOpen():
			return "a panel is open"
		case s.Inventory != nil && s.Inventory.Cursor != nil:
			return "an item is on the cursor"
		}
		return ""
	}
}

// wireVaultKeep builds the keeper for the hero and hangs it on the shared home and the runner ref.
func wireVaultKeep(ref *runnerRef, e *env.Env, c *d2dctl.Client, dataDir string, hero func() string, log *slog.Logger) {
	if ref == nil || c == nil || dataDir == "" || e == nil {
		return
	}
	home := vaultHomeFor(filepath.Join(dataDir, "vaults"))
	snap := func() *d2.Snapshot {
		if r := ref.p.Load(); r != nil {
			return r.LastSnapshot()
		}
		return nil
	}
	// Backup and Save are the fork's save verb: it writes through the character store, reads the bytes back and keeps its
	// .bak of the previous file (V design doc section 2: character_store.save). The bot has no save directory on the d2d
	// path and saves.Snapshot refuses in a game, so this is the existing mechanism; the vault record itself stays after a
	// Keep, so a bad create can never lose the item data.
	save := func(ctx context.Context) error { return c.SaveCharacter(ctx) }
	k := &vaultkeep.Keeper{
		Store:    func() *vault.Store { return home.Existing(hero()) },
		API:      c,
		Enabled:  keepOn(e),
		Ready:    keepReady(snap),
		Backup:   save,
		Save:     save,
		Notice:   func(text string) { log.Info("vault keep: " + text) },
		PerVisit: func() int { return e.Tune.Loot.VaultKeepPerVisit },
	}
	home.mu.Lock()
	home.keep = k
	home.mu.Unlock()
	ref.vk = k
	activeKeeper.Store(k)
}

// vaultKeepDep is bot.Deps.VaultKeep (nil interface when there is no keeper).
func vaultKeepDep(ref *runnerRef) bot.VaultKeep {
	if ref == nil || ref.vk == nil {
		return nil
	}
	return ref.vk
}

// keepFace adapts the keeper of a vaultHome to watch.VaultKeeper; late bound, so the page works before and without one.
type keepFace struct{ h *vaultHome }

func (f keepFace) k() *vaultkeep.Keeper {
	f.h.mu.Lock()
	defer f.h.mu.Unlock()
	return f.h.keep
}

func (f keepFace) Gate(ctx context.Context) (bool, string, string) {
	k := f.k()
	if k == nil {
		return false, "off", "this bot has no Keep queue"
	}
	ok, why := k.Gate(ctx)
	if ok {
		return true, "", ""
	}
	code := "stopped"
	switch why {
	case vaultkeep.ReasonOff:
		code = "off"
	case vaultkeep.ReasonDead:
		code = "dead"
	case vaultkeep.ReasonVerbs:
		code = "verbs"
	}
	return false, code, why
}

func (f keepFace) Enqueue(ctx context.Context, ids []string, target string, now bool) []watch.VaultKeepResult {
	k := f.k()
	if k == nil {
		return nil
	}
	var out []watch.VaultKeepResult
	for _, r := range k.Enqueue(ctx, ids, target, now) {
		out = append(out, watch.VaultKeepResult{ID: r.ID, OK: r.OK, Already: r.Already, Code: r.Code, Why: r.Why})
	}
	return out
}

func (f keepFace) Progress() watch.VaultKeepView {
	k := f.k()
	if k == nil {
		return watch.VaultKeepView{Why: "this bot has no Keep queue"}
	}
	in := k.Info()
	ok, why := k.Peek()
	return watch.VaultKeepView{Enabled: ok, Why: why, Queued: in.Queued, Creating: in.Creating, Done: in.Done, Full: in.Full, Line: k.ProgressLine()}
}

func init() {
	glanceExtra = append(glanceExtra, func(src *watch.GlanceSources, _ glanceDeps) {
		prev := src.Needs
		src.Needs = func() []watch.GlanceNeed {
			var out []watch.GlanceNeed
			if prev != nil {
				out = prev()
			}
			if k := activeKeeper.Load(); k != nil {
				if text := k.NeedsText(); text != "" {
					out = append(out, watch.GlanceNeed{ID: "vault:keep", Kind: "vault_keep", Sev: "warn", Text: text,
						Sub: "The items stay queued and arrive at the next town visit once there is room.", Actions: []watch.GlanceVerb{}})
				}
			}
			return out
		}
	})
}
