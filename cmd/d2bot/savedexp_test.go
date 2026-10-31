package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSavedExpReaderCachesByMtime(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "Soaky.d2s")
	reads, val := 0, uint32(100)
	var fail error
	f := savedExpReader(func() string { return dir }, func() string { return "Soaky" }, func(string) (int, uint32, error) {
		reads++
		return 5, val, fail
	})
	if _, ok := f(); ok {
		t.Fatal("missing file must be unknown")
	}
	os.WriteFile(p, []byte("a"), 0o644)
	if v, ok := f(); !ok || v != 100 || reads != 1 {
		t.Fatalf("first read: %d %v %d", v, ok, reads)
	}
	f()
	if reads != 1 {
		t.Fatalf("unchanged file re-read: %d", reads)
	}
	val = 250
	os.WriteFile(p, []byte("bb"), 0o644)
	if v, _ := f(); v != 250 || reads != 2 {
		t.Fatalf("changed file: %d %d", v, reads)
	}
	fail = errors.New("torn")
	os.WriteFile(p, []byte("ccc"), 0o644)
	if v, ok := f(); !ok || v != 250 {
		t.Fatalf("torn read must keep last good: %d %v", v, ok)
	}
}

// d2d: the save file sits somewhere under the slot's d2d home; the dir is found once, kept, and a miss is retried only
// after the retry delay.
func TestD2DSavedExpFindsTheFileUnderHome(t *testing.T) {
	home := t.TempDir()
	now := time.Unix(0, 0)
	exp := uint32(0)
	dir := d2dSaveDir(home, func() string { return "Dbot" }, 10*time.Second, func() time.Time { return now })
	if dir() != "" {
		t.Fatal("no file yet must give no dir")
	}
	sub := filepath.Join(home, "data", "d2", "save")
	os.MkdirAll(sub, 0o755)
	os.WriteFile(filepath.Join(sub, "Dbot.d2s"), []byte("a"), 0o644)
	if dir() != "" {
		t.Fatal("a miss must not be re-walked inside the retry delay")
	}
	now = now.Add(11 * time.Second)
	if dir() != sub {
		t.Fatalf("dir %q, want %q", dir(), sub)
	}
	exp = 777
	f := savedExpReader(dir, func() string { return "Dbot" }, func(string) (int, uint32, error) { return 5, exp, nil })
	if v, ok := f(); !ok || v != 777 {
		t.Fatalf("saved exp %d %v, want 777", v, ok)
	}
	if d2dSavedExp("", func() string { return "Dbot" }) != nil {
		t.Fatal("no home (attached d2d) must leave SavedExp nil")
	}
}
