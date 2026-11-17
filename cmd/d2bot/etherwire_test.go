package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"d2lod-bot/internal/ether"
	"d2lod-bot/internal/goals"
	"d2lod-bot/internal/vault"
)

type etherRig struct {
	dir, root string
	on        atomic.Bool
	hero      atomic.Value
	eh        *etherHome
}

func newEtherRig(t *testing.T) *etherRig {
	t.Helper()
	r := &etherRig{dir: t.TempDir(), root: t.TempDir()}
	r.hero.Store("Sorc")
	r.eh = etherFor(r.dir, r.root, r.on.Load, func() string { return r.hero.Load().(string) })
	return r
}

func TestEtherWalletOffUntilKnob(t *testing.T) {
	r := newEtherRig(t)
	if r.eh.Wallet() != nil {
		t.Fatal("wallet with loot.vault 0")
	}
	if _, err := os.Stat(filepath.Join(r.dir, ether.FileName)); err == nil {
		t.Fatal("the ledger was created with the knob off")
	}
	if got := vaultHomeFor(r.root).storeCapacity(); got != 400 {
		t.Fatalf("base capacity %d", got)
	}
	r.on.Store(true)
	w := r.eh.Wallet()
	if w == nil || r.eh.Wallet() != w {
		t.Fatal("one wallet once the knob is on")
	}
}

func TestEtherNoAccountLedgerNoWallet(t *testing.T) {
	if etherFor("", t.TempDir(), nil, nil) != nil {
		t.Fatal("a wallet without a currency directory")
	}
	var nilHome *etherHome
	if nilHome.Wallet() != nil {
		t.Fatal("nil home")
	}
}

// Owned tiers raise the capacity of every hero's store of the account; a hero switch keeps the wallet and the tiers; the refund
// rule reads the ACTIVE hero's count.
func TestEtherCapacityAndHeroSwitch(t *testing.T) {
	r := newEtherRig(t)
	r.on.Store(true)
	w := r.eh.Wallet()
	if _, err := w.Earn("quest", "q", "normal", 50); err != nil {
		t.Fatal(err)
	}
	home := vaultHomeFor(r.root)
	a, err := home.Open("Sorc")
	if err != nil {
		t.Fatal(err)
	}
	if c := a.Status().Capacity; c != 400 {
		t.Fatalf("base %d", c)
	}
	if _, err := w.Upgrade(""); err != nil {
		t.Fatal(err)
	}
	if c := a.Status().Capacity; c != 600 {
		t.Fatalf("after upgrade %d (an open store must see it)", c)
	}
	// the switch: another hero, same account
	r.hero.Store("Pala")
	b, err := home.Open("Pala")
	if err != nil {
		t.Fatal(err)
	}
	if c := b.Status().Capacity; c != 600 {
		t.Fatalf("second hero %d", c)
	}
	if r.eh.Wallet() != w {
		t.Fatal("the wallet was replaced by the switch")
	}
	if owned := r.eh.owned(); len(owned) != 1 || owned[0] != "v1" {
		t.Fatalf("owned %v", owned)
	}
	// the active hero's vault decides the refund: 450 records do not fit 400
	for i := 0; i < 450; i++ {
		if _, err := b.Add(vault.Item{Hero: "Pala", Raw: fmt.Sprintf("%08x", i+1), Code: "cap", Name: "Cap", Quality: 4, Ilvl: 5}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := w.Downgrade(""); err == nil {
		t.Fatal("downgrade with 450 records")
	}
	r.hero.Store("Sorc") // the other hero is empty: the refund fits
	if _, err := w.Downgrade(""); err != nil {
		t.Fatal(err)
	}
	if c := a.Status().Capacity; c != 400 {
		t.Fatalf("after refund %d", c)
	}
}

// The store takes the wallet's capacity under its own lock and the wallet reads the store's count under its own: a store
// writer and a refund at once must not deadlock (the lock free Capacity).
func TestEtherNoDeadlockStoreAndWallet(t *testing.T) {
	r := newEtherRig(t)
	r.on.Store(true)
	w := r.eh.Wallet()
	w.Earn("quest", "q", "normal", 50)
	w.Upgrade("")
	home := vaultHomeFor(r.root)
	st, _ := home.Open("Sorc")
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 300; i++ {
			st.Add(vault.Item{Hero: "Sorc", Raw: fmt.Sprintf("%08x", i+1), Code: "cap", Name: "Cap", Quality: 4, Ilvl: 5})
			st.Status()
			w.Summary()
			w.Downgrade("")
			w.Upgrade("")
		}
	}()
	select {
	case <-done:
	case <-time.After(20 * time.Second):
		t.Fatal("deadlock between the store and the wallet")
	}
}

func fakeHero(name, diff string) func() ether.Hero {
	return func() ether.Hero { return ether.Hero{Name: name, Difficulty: diff} }
}

func ev(seq uint64, t goals.EventType) goals.Event { return goals.Event{Seq: seq, Type: t} }

// The feed with fakes: kill, boss, first area, level; the knob off earns nothing; a restart (a new feed over the reopened ledger)
// books nothing twice; a second account has its own wallet.
func TestEtherFeedEventsRestartAndAccounts(t *testing.T) {
	r := newEtherRig(t)
	town := func(a string) bool { return a == "Rogue Encampment" }
	events := func() []goals.Event {
		k := ev(1, goals.EvKill)
		k.Monster = "Andariel"
		k.Boss = true
		f := ev(2, goals.EvKill)
		f.Monster = "fallen"
		ar := ev(3, goals.EvAreaChange)
		ar.Area = "Blood Moor"
		tw := ev(4, goals.EvAreaChange)
		tw.Area = "Rogue Encampment"
		lv := ev(5, goals.EvLevelUp)
		lv.Level = 10
		return []goals.Event{k, f, ar, tw, lv}
	}
	feed := etherFeed(r.eh.Wallet, fakeHero("Sorc", "normal"), town, nil)
	for _, e := range events() {
		feed(e) // knob off
	}
	if r.eh.Wallet() != nil {
		t.Fatal("wallet exists with the knob off")
	}
	r.on.Store(true)
	for _, e := range events() {
		feed(e)
	}
	want := (5 + 10) + 5 + 5 + 5 // Andariel first+boss, fallen first, Blood Moor first, clvl10
	if b, _ := r.eh.Wallet().Balance(); b != want {
		t.Fatalf("balance %d, want %d", b, want)
	}
	// restart: a fresh home over the same directory, replaying the same events
	r2 := &etherHome{dir: r.dir, log: func(string) {}, on: r.on.Load, hero: func() string { return "Sorc" }}
	feed2 := etherFeed(r2.Wallet, fakeHero("Sorc", "normal"), town, nil)
	for _, e := range events() {
		feed2(e)
	}
	if b, _ := r2.Wallet().Balance(); b != want {
		t.Fatalf("restart double credit: %d", b)
	}
	// another account: its own directory, its own wallet
	other := etherFor(t.TempDir(), "", r.on.Load, nil)
	if b, _ := other.Wallet().Balance(); b != 0 {
		t.Fatalf("second account balance %d", b)
	}
	// a hero switch (the second hero of the account) keeps the wallet and earns only its own firsts
	feed3 := etherFeed(r2.Wallet, fakeHero("Pala", "normal"), town, nil)
	for _, e := range events() {
		feed3(e)
	}
	if b, _ := r2.Wallet().Balance(); b != want+5+5+5 { // fallen, Andariel (first, not boss), Blood Moor
		t.Fatalf("second hero %d", b)
	}
}

func TestEtherDifficulty(t *testing.T) {
	for in, want := range map[string]string{"": "normal", "Normal": "normal", "Nightmare": "nm", "hell": "hell", "weird": "normal"} {
		if got := etherDifficulty(in); got != want {
			t.Errorf("%q -> %q", in, got)
		}
	}
}
