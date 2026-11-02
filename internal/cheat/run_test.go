package cheat

import (
	"path/filepath"
	"testing"
	"time"
)

func TestWritesAfterStart(t *testing.T) {
	l := NewLedger(filepath.Join(t.TempDir(), "l.jsonl"))
	t0 := time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)
	add := func(k Kind, inst string, ok bool) {
		t.Helper()
		if err := l.Append(Entry{Time: t0, Actor: "a", Instance: inst, Kind: k, Reason: "r", Success: ok}); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := l.WritesAfterStart("r1"); err == nil {
		t.Fatal("unknown run must be an error, not 0 writes")
	}
	add(KindLevel, "i5", true) // setup, before the marker
	if err := l.MarkRun("r1", "a", "i5", "gate", t0); err != nil {
		t.Fatal(err)
	}
	if n, _, err := l.WritesAfterStart("r1"); err != nil || n != 0 {
		t.Fatalf("clean run: %d %v", n, err)
	}
	add(KindHeal, "i5", true)
	add(KindGold, "i5", false) // refused: not a write
	add(KindHeal, "i4", true)  // other instance
	n, es, err := l.WritesAfterStart("r1")
	if err != nil || n != 1 || es[0].Kind != KindHeal {
		t.Fatalf("got %d %v %v", n, es, err)
	}
	if err := l.MarkRun(" ", "a", "i5", "x", t0); err == nil {
		t.Fatal("empty run id accepted")
	}
}
