package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/cheat"
	"d2lod-bot/internal/cheat/d2s"
)

// synthetic character, never a real save (offsets as in internal/cheat/d2s)
func d2sImage(name string) []byte {
	h := make([]byte, 0x2FD)
	binary.LittleEndian.PutUint32(h[0:], d2s.Signature)
	binary.LittleEndian.PutUint32(h[4:], d2s.Version)
	copy(h[0x14:], name)
	h[0x24] = d2s.FlagExpansion
	h[0x28], h[0x2B], h[0xA8] = 1, 7, 0x80
	copy(h[0x14F:], "Woo!")
	copy(h[0x279:], "WS")
	copy(h[0x2CA:], "w4")
	b := append(h, 'g', 'f', 0xFF, 0x01, 'i', 'f')
	b = append(b, make([]byte, 30)...)
	b = append(b, 'J', 'M', 0, 0)
	binary.LittleEndian.PutUint32(b[8:], uint32(len(b)))
	binary.LittleEndian.PutUint32(b[0xC:], d2s.Checksum(b))
	return b
}

type rig struct {
	env    *Env
	dirs   map[string]string
	locked map[string]bool
	pre    []string
}

func newRig(t *testing.T) *rig {
	r := &rig{dirs: map[string]string{"i5": t.TempDir(), "i4": t.TempDir()}, locked: map[string]bool{}}
	r.env = &Env{
		Lib:     filepath.Join(t.TempDir(), "lib"),
		Ledger:  cheat.NewLedger(filepath.Join(t.TempDir(), "cheats.jsonl")),
		Now:     func() time.Time { return time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC) },
		SaveDir: func(i string) string { return r.dirs[i] },
		Preflight: func(i string) (func(), error) {
			r.pre = append(r.pre, i)
			if r.locked[i] {
				return nil, errors.New("instance lock held")
			}
			return func() {}, nil
		},
	}
	os.WriteFile(filepath.Join(r.dirs["i5"], "Fresha.d2s"), d2sImage("Fresha"), 0o644)
	os.WriteFile(filepath.Join(r.dirs["i5"], "Fresha.map"), []byte("map"), 0o644)
	return r
}

func (r *rig) do(args ...string) (string, error) {
	var b bytes.Buffer
	err := run(args, r.env, &b)
	return b.String(), err
}

func TestSnapshotPlaceLedger(t *testing.T) {
	r := newRig(t)
	if _, err := r.do("snapshot", "--name", "Fresha", "--inst", "i5"); err == nil {
		t.Fatal("snapshot without --at-menu accepted")
	}
	out, err := r.do("snapshot", "--name", "Fresha", "--inst", "i5", "--at-menu", "--id", "fresha-c7")
	if err != nil || !strings.Contains(out, "ledger honest") {
		t.Fatalf("%q %v", out, err)
	}
	if out, err = r.do("verify", "fresha-c7"); err != nil {
		t.Fatalf("%q %v", out, err)
	}
	if out, _ = r.do("list"); !strings.Contains(out, "fresha-c7") {
		t.Fatalf("list %q", out)
	}
	if _, err = r.do("place", "--snapshot", "fresha-c7", "--as", "Freshb", "--inst", "i4"); err == nil {
		t.Fatal("place without a reason accepted")
	}
	out, err = r.do("place", "--snapshot", "fresha-c7", "--as", "Freshb", "--inst", "i4", "--reason", "baseline late run 3", "--run", "run-3", "--actor", "t")
	if err != nil || !strings.Contains(out, "setup: copy-of-honest fresha-c7") {
		t.Fatalf("%q %v", out, err)
	}
	if _, err := os.Stat(filepath.Join(r.dirs["i4"], "Freshb.d2s")); err != nil {
		t.Fatal(err)
	}
	es, _ := r.env.Ledger.List(cheat.Filter{Kind: cheat.KindSetupCopy})
	if len(es) != 1 || !es[0].Success || es[0].Instance != "i4" || es[0].Character != "Freshb" || es[0].Key() != "i4/Freshb" ||
		es[0].Params["provenance"] != "copy-of-honest" || es[0].Params["run"] != "run-3" || es[0].Reason == "" {
		t.Fatalf("ledger %+v", es)
	}
	// a setup copy is not a cheat: the copy stays honest, and the original is untouched
	if r.env.Ledger.CheatedKey(time.Time{}, "i4/Freshb") || r.env.Ledger.CheatedKey(time.Time{}, "i5/Fresha") {
		t.Fatal("setup-copy counted as cheated")
	}
}

func TestPlaceRefusalsAreLogged(t *testing.T) {
	r := newRig(t)
	r.do("snapshot", "--name", "Fresha", "--inst", "i5", "--at-menu", "--id", "s1")
	args := func(as, inst string) []string {
		return []string{"place", "--snapshot", "s1", "--as", as, "--inst", inst, "--reason", "t"}
	}
	r.locked["i4"] = true
	if _, err := r.do(args("Freshb", "i4")...); err == nil || !strings.Contains(err.Error(), "lock") {
		t.Fatalf("locked prefix: %v", err)
	}
	if ents, _ := os.ReadDir(r.dirs["i4"]); len(ents) != 0 {
		t.Fatal("placed into a locked prefix")
	}
	r.locked["i4"] = false
	if _, err := r.do(args("Freshb", "i0")...); err == nil {
		t.Fatal("i0 accepted")
	}
	if _, err := r.do(args("Freshb", "i4")...); err != nil {
		t.Fatal(err)
	}
	if _, err := r.do(args("Freshb", "i4")...); err == nil { // never over an existing character
		t.Fatal("overwrote an existing character")
	}
	es, _ := r.env.Ledger.List(cheat.Filter{Kind: cheat.KindSetupCopy})
	var ok, fail int
	for _, e := range es {
		if e.Success {
			ok++
		} else if e.Error != "" {
			fail++
		}
	}
	if ok != 1 || fail != 2 {
		t.Fatalf("ledger ok=%d fail=%d: %+v", ok, fail, es)
	}
}

func TestSnapshotOfCheatedCharacter(t *testing.T) {
	r := newRig(t)
	r.env.Ledger.Append(cheat.Entry{Time: r.env.Now(), Actor: "t", Instance: "i5", Character: "Fresha", Kind: cheat.KindGold, Reason: "x", Success: true})
	out, err := r.do("snapshot", "--name", "Fresha", "--inst", "i5", "--at-menu", "--id", "s2")
	if err != nil || !strings.Contains(out, "ledger cheated") {
		t.Fatalf("%q %v", out, err)
	}
	if _, err := r.do("place", "--snapshot", "s2", "--as", "Freshb", "--inst", "i4", "--reason", "t"); err == nil {
		t.Fatal("cheated source placed")
	}
	// the same name cheated in ANOTHER prefix does not taint this one
	r2 := newRig(t)
	r2.env.Ledger.Append(cheat.Entry{Time: r2.env.Now(), Actor: "t", Instance: "i3", Character: "Fresha", Kind: cheat.KindGold, Reason: "x", Success: true})
	if out, _ := r2.do("snapshot", "--name", "Fresha", "--inst", "i5", "--at-menu"); !strings.Contains(out, "ledger honest") {
		t.Fatalf("%q", out)
	}
}

func TestLockAll(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.lock")
	rel, err := lockAll([]string{p})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lockAll([]string{p}); err == nil {
		t.Fatal("second lock succeeded")
	}
	rel()
	rel2, err := lockAll([]string{p})
	if err != nil {
		t.Fatal(err)
	}
	rel2()
}
