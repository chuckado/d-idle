package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJournalWrites(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.jnl")
	body := "# journal\n>\t1\tdump player\tabc\tf:1\n>\t2\tdebug gold 5\tabc\tf:2\n>\t3\tdebug hit 9\terr:lab verbs off\tf:3\nC\t4\tdead\n>\t5\titems\tabc\tf:4\n>\t6\tdebug xp 1\tabc\tf:5\n"
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if n, ok := journalWrites(p); !ok || n != 2 {
		t.Fatalf("two accepted writes, got %d %v", n, ok)
	}
	if _, ok := journalWrites(filepath.Join(t.TempDir(), "none")); ok {
		t.Fatal("a missing journal is unobserved")
	}
}
