package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"d2lod-bot/internal/ether"
	"d2lod-bot/internal/goals"
)

// The ETHER wallet of the process (docs/ether.md "Wiring"): one wallet per account, that is per currency directory of the
// profile, opened once and shared by the Vault recorder, the Vault page, the API routes and the earn hook. It lives on the
// account, not on a hero: a Character Select switch restarts the child and the next one opens the same ledger. Everything is
// behind loot.vault: with the knob at 0 the wallet is not opened for earning or spending (the capacity of a wallet that
// exists on disk is still read, so a read only page does not show a false over capacity).
type etherHome struct {
	dir string
	log func(string)

	mu   sync.Mutex
	on   func() bool
	hero func() string
	home *vaultHome
	w    *ether.Wallet
	bad  bool // an open failed: said once, not retried per request
}

var etherHomes sync.Map // dir -> *etherHome

// etherFor returns the wallet home of the account. root is the vault root (<data>/vaults) whose stores take their capacity
// from it. on is the knob loot.vault, hero the active hero's name; later calls replace them. nil when there is no
// currency directory (no account ledger, no wallet).
func etherFor(dir, root string, on func() bool, hero func() string) *etherHome {
	if dir == "" {
		return nil
	}
	v, _ := etherHomes.LoadOrStore(dir, &etherHome{dir: dir, log: func(s string) { slog.Warn(s) }})
	h := v.(*etherHome)
	h.mu.Lock()
	if on != nil {
		h.on = on
	}
	if hero != nil {
		h.hero = hero
	}
	if root != "" && h.home == nil {
		h.home = vaultHomeFor(root)
		h.home.setCapacity(h.capacity)
	}
	h.mu.Unlock()
	return h
}

func (h *etherHome) knobOn() bool {
	h.mu.Lock()
	on := h.on
	h.mu.Unlock()
	return on != nil && on()
}

// content is the record count of the active hero's vault, the number the refund rule reads. It must not call back into the
// wallet's locked paths: Store.Status reads the capacity, which is lock free (ether.Wallet.Capacity).
func (h *etherHome) content() int {
	h.mu.Lock()
	home, hero := h.home, h.hero
	h.mu.Unlock()
	if home == nil || hero == nil {
		return 0
	}
	if st := home.Existing(hero()); st != nil {
		return st.Status().Count
	}
	return 0
}

// open returns the wallet, opening it first when asked to (create) or when its ledger exists already.
func (h *etherHome) open(create bool) *ether.Wallet {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.w != nil || h.bad {
		return h.w
	}
	if !create {
		if _, err := os.Stat(filepath.Join(h.dir, ether.FileName)); err != nil {
			return nil
		}
	}
	w, err := ether.Open(h.dir, ether.Options{Content: h.content})
	if err != nil {
		h.bad = true
		h.log("ether wallet not available: " + err.Error())
		return nil
	}
	if n := w.Repaired(); n > 0 {
		h.log("ether ledger: dropped a torn last line")
	}
	h.w = w
	return w
}

// Wallet is the wallet for earning, spending and the API: nil while loot.vault is 0.
func (h *etherHome) Wallet() *ether.Wallet {
	if h == nil || !h.knobOn() {
		return nil
	}
	return h.open(true)
}

// capacity is the stores' capacity: the owned tiers when the account has a wallet, the table base otherwise.
func (h *etherHome) capacity() int {
	if w := h.open(h.knobOn()); w != nil {
		return w.Capacity(0)
	}
	return 0
}

// owned is VaultDesk.Owned.
func (h *etherHome) owned() []string {
	if w := h.open(false); w != nil {
		s, _ := w.OwnedSteps()
		return s
	}
	return nil
}

// etherDifficulty is the wallet's word for the hero's difficulty text.
func etherDifficulty(s string) string {
	return [...]string{"normal", "nm", "hell"}[difficultyOf(s)]
}

// etherFeed turns the goals events into earn calls (docs/ether.md "Earn call sites"). hero names who is playing now, wallet is nil
// while loot.vault is 0 (nothing is earned then and the events are not queued for later: the keys make a late start harmless).
// isTown names the areas that are not a first visit.
func etherFeed(wallet func() *ether.Wallet, hero func() ether.Hero, isTown func(string) bool, log func(string)) func(goals.Event) {
	var e *ether.Earner // one per wallet: it keeps the unpaid kill count
	return func(ev goals.Event) {
		w := wallet()
		if w == nil {
			return
		}
		if e == nil || e.W != w {
			e = &ether.Earner{W: w, Log: log}
		}
		h := hero()
		switch ev.Type {
		case goals.EvKill:
			e.Kill(h, ev.Monster, ev.Boss)
		case goals.EvAreaChange:
			if ev.Area != "" && !isTown(ev.Area) {
				e.First(h, "area", ev.Area)
			}
		case goals.EvLevelUp:
			e.Level(h, ev.Level)
		}
	}
}

// startEther subscribes the earn hook to the goal manager.
func startEther(ctx context.Context, eh *etherHome, m *goals.Manager, hero func() ether.Hero, log func(string)) {
	if eh == nil || m == nil || hero == nil {
		return
	}
	ch, unsub := m.Subscribe()
	feed := etherFeed(eh.Wallet, hero, townSet(), log)
	go func() {
		defer unsub()
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-ch:
				if !ok {
					return
				}
				feed(ev)
			}
		}
	}()
}

func townSet() func(string) bool {
	t := townNames()
	return func(a string) bool { return t[a] || strings.TrimSpace(a) == "" }
}
