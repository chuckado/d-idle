package questboard

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func openStore(t *testing.T, dir, hero string) *Store {
	t.Helper()
	s, err := Open(mustBoard(t), dir, hero)
	if err != nil {
		t.Fatal(err)
	}
	s.now = func() time.Time { return time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC) }
	return s
}

func TestStoreAcceptRestartCancel(t *testing.T) {
	dir := t.TempDir()
	s := openStore(t, dir, "Pali")
	if err := s.Accept("den_of_evil", "clear_den", 0, "train"); err != nil {
		t.Fatal(err)
	}
	if err := s.Accept("den_of_evil", "clear_den", 0, "train"); err != nil {
		t.Errorf("repeating the same tap: %v", err)
	}
	err := s.Accept("sisters_burial_grounds", "blood_raven", 0, "train")
	if err == nil || !strings.Contains(err.Error(), "den_of_evil/clear_den") {
		t.Errorf("a second accepted chunk = %v", err)
	}
	// Another difficulty is a fresh board.
	if err := s.Accept("sisters_burial_grounds", "blood_raven", 1, "train"); err != nil {
		t.Errorf("nm accept: %v", err)
	}

	// A restart: a new Store on the same directory sees the same records, and the board derives the same state.
	s2 := openStore(t, dir, "Pali")
	p, info, err := s2.Load()
	if err != nil || !info.Exists || len(p.Accepted) != 2 {
		t.Fatalf("reload: %+v %+v %v", p, info, err)
	}
	b := mustBoard(t)
	v := view(t, b, Input{Facts: facts{}, Plan: p})["den_of_evil"]
	if v.State != Accepted {
		t.Errorf("after restart = %s", v.State)
	}
	if err := s2.Cancel("den_of_evil", "clear_den", 0); err != nil {
		t.Fatal(err)
	}
	p, _, _ = s2.Load()
	if len(p.Accepted) != 1 || p.Accepted[0].Diff != "nm" {
		t.Errorf("after cancel = %+v", p.Accepted)
	}
}

func TestStoreHeroesAreSeparate(t *testing.T) {
	dir := t.TempDir()
	a, c := openStore(t, dir, "Pali"), openStore(t, dir, "Sorci")
	if err := a.Accept("den_of_evil", "clear_den", 0, ""); err != nil {
		t.Fatal(err)
	}
	if p, _, _ := c.Load(); len(p.Accepted) != 0 {
		t.Errorf("Sorci sees Pali's record: %+v", p)
	}
	for _, bad := range []string{"", "a", "../x", "a/b", "Pali.json", strings.Repeat("x", 30), "1abc"} {
		if _, err := Open(mustBoard(t), dir, bad); err == nil {
			t.Errorf("hero %q accepted", bad)
		}
	}
}

func TestStoreUnknownIDsAreErrors(t *testing.T) {
	s := openStore(t, t.TempDir(), "Pali")
	for name, err := range map[string]error{
		"accept quest": s.Accept("ghost", "c", 0, ""),
		"accept chunk": s.Accept("den_of_evil", "zz", 0, ""),
		"accept diff":  s.Accept("den_of_evil", "clear_den", 5, ""),
		"cancel":       s.Cancel("ghost", "c", 0),
		"skip":         s.Skip("ghost", "", 0),
		"skip chunk":   s.Skip("den_of_evil", "zz", 0),
		"unskip":       s.Unskip("ghost", "", 0),
		"auto":         s.SetAutoStart("ghost", 0, true),
	} {
		if err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestStoreSkipAndAutoStart(t *testing.T) {
	s := openStore(t, t.TempDir(), "Pali")
	if err := s.Skip("tools_of_the_trade", "", 0); err != nil {
		t.Fatal(err)
	}
	if err := s.Skip("tools_of_the_trade", "", 0); err != nil {
		t.Fatal(err)
	}
	p, _, _ := s.Load()
	if len(p.Skipped) != 1 {
		t.Errorf("skip not idempotent: %+v", p.Skipped)
	}
	if err := s.Unskip("tools_of_the_trade", "", 0); err != nil {
		t.Fatal(err)
	}
	if p, _, _ = s.Load(); len(p.Skipped) != 0 {
		t.Errorf("unskip: %+v", p.Skipped)
	}
	// Auto-start is per difficulty and refused where the data says never (Andariel is a hazard).
	if err := s.SetAutoStart("den_of_evil", 0, true); err != nil {
		t.Fatal(err)
	}
	if err := s.SetAutoStart("sisters_to_the_slaughter", 0, true); err == nil || !strings.Contains(err.Error(), "does not allow") {
		t.Errorf("hazard auto-start = %v", err)
	}
	p, _, _ = s.Load()
	if !p.AutoStart(0, "den_of_evil") || p.AutoStart(1, "den_of_evil") {
		t.Errorf("auto flags wrong: %+v", p.Auto)
	}
	if err := s.SetAutoStart("den_of_evil", 0, false); err != nil {
		t.Fatal(err)
	}
	if p, _, _ = s.Load(); p.AutoStart(0, "den_of_evil") {
		t.Error("auto-start not cleared")
	}
	if err := s.MarkSeen("quest/den_of_evil/-/normal"); err != nil {
		t.Fatal(err)
	}
	if err := s.MarkSeen("quest/den_of_evil/-/normal"); err != nil {
		t.Fatal(err)
	}
	if p, _, _ = s.Load(); !reflect.DeepEqual(p.Seen, []string{"quest/den_of_evil/-/normal"}) {
		t.Errorf("seen = %v", p.Seen)
	}
}

func TestStoreAtomicBackupAndRecovery(t *testing.T) {
	dir := t.TempDir()
	s := openStore(t, dir, "Pali")
	if err := s.Accept("den_of_evil", "clear_den", 0, ""); err != nil {
		t.Fatal(err)
	}
	if err := s.Skip("tools_of_the_trade", "", 0); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "Pali.quests.json")
	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Fatalf("no .bak after the second write: %v", err)
	}
	if _, err := os.Stat(path + ".tmp"); err == nil {
		t.Error("tmp file left behind")
	}
	// A torn main file: the load falls back to the backup (the state before the last write) and says so.
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	p, info, err := openStore(t, dir, "Pali").Load()
	if err != nil || !info.FromBackup || info.Warning == "" || len(p.Accepted) != 1 {
		t.Errorf("recovery: %+v %+v %v", p, info, err)
	}
}

func TestStoreRefusesNewerVersionAndStaleIDs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Pali.quests.json")
	if err := os.WriteFile(path, []byte(`{"version":9,"data":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	s := openStore(t, dir, "Pali")
	if _, _, err := s.Load(); err == nil {
		t.Error("future version loaded")
	}
	if err := s.Accept("den_of_evil", "clear_den", 0, ""); err == nil {
		t.Error("future version overwritten")
	}
	// A record from an older data file names a quest that no longer exists: an error naming it.
	if err := os.WriteFile(path, []byte(`{"version":1,"data":{"accepted":[{"quest":"old_quest","chunk":"c","difficulty":"normal","at":"2026-11-16T12:00:00Z"}]}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Load(); err == nil || !strings.Contains(err.Error(), `"old_quest"`) {
		t.Errorf("stale id = %v", err)
	}
}

func TestStoreEmptyLoad(t *testing.T) {
	p, info, err := openStore(t, t.TempDir(), "Pali").Load()
	if err != nil || info.Exists || len(p.Accepted)+len(p.Skipped)+len(p.Auto) != 0 {
		t.Errorf("empty: %+v %+v %v", p, info, err)
	}
}
