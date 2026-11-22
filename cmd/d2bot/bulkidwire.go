package main

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"d2lod-bot/internal/bulkid"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/ether"
	"d2lod-bot/internal/task"
	"d2lod-bot/internal/town"
)

// wireBulkID installs the bulk identify chore on the town (docs/design/bulk-identify-2026-11-18.md). The journal sits in
// the data directory so a restart in the middle of a Cain trip finds its items again. A damaged journal leaves the chore
// off and says so: guessing would lose items.
func wireBulkID(ref *runnerRef, tw *town.Town, e *env.Env, dataDir string, hero func() string, log *slog.Logger) {
	if ref == nil || tw == nil || e == nil || dataDir == "" {
		return
	}
	f, err := bulkid.DefaultFilter()
	if err != nil {
		log.Warn("bulkid: filter table unusable, bulk identify off", "err", err)
		return
	}
	name := "bulkid.json"
	if hero != nil && hero() != "" {
		name = "bulkid-" + filepath.Base(hero()) + ".json"
	}
	j, err := bulkid.FileJournal(filepath.Join(dataDir, name))
	if err != nil {
		log.Warn("bulkid: journal unreadable, bulk identify off", "err", err)
		return
	}
	tw.SetBulkID(&town.BulkID{
		Filter:   f,
		Tuning:   func() env.IdentifyTuning { return e.Tune.Identify },
		Measured: func() bool { return os.Getenv("D2BOT_RUN") != "" },
		Journal:  j,
		Rescued:  cainQuest(ref, task.StageCleared),
	})
}

// lazyWallet is the Ether wallet as the vault keeper sees it: opened on use, an error while loot.vault is off.
type lazyWallet struct{ get func() *ether.Wallet }

var errNoWallet = errors.New("no Ether wallet (loot.vault is off)")

func (l lazyWallet) Balance() (int, error) {
	if w := l.get(); w != nil {
		return w.Balance()
	}
	return 0, errNoWallet
}

func (l lazyWallet) Spend(key string, n int) error {
	if w := l.get(); w != nil {
		return w.Pay().Spend(key, n)
	}
	return errNoWallet
}
