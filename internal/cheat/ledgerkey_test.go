package cheat

import (
	"path/filepath"
	"testing"
	"time"
)

func TestLedgerKey(t *testing.T) {
	l := NewLedger(filepath.Join(t.TempDir(), "l.jsonl"))
	t0 := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	add := func(inst, char, id string, k Kind, ok bool) {
		t.Helper()
		e := Entry{Time: t0, Actor: "a", Instance: inst, Character: char, CharID: id, Kind: k, Reason: "r", Success: ok}
		if err := l.Append(e); err != nil {
			t.Fatal(err)
		}
	}
	add("i5", "Fresha", "", KindLevel, true)     // a cheat on the i5 character
	add("i4", "Fresha", "", KindSetupCopy, true) // a copy: setup, not a cheat
	add("i3", "Fresha", "", KindGold, false)     // refused
	add("", "Old", "", KindXP, true)             // legacy entry: no instance
	add("i1", "X", "uid-7", KindHeal, true)      // explicit id

	for _, c := range []struct {
		key  string
		want bool
	}{
		{"i5/Fresha", true}, {"i5/fresha", true},
		{"i4/Fresha", false},             // only a setup-copy
		{"i3/Fresha", false},             // only a refused attempt
		{"i2/Old", true},                 // legacy entry counts against the name in any prefix
		{"uid-7", true}, {"i1/X", false}, // CharID replaces the instance/name key
	} {
		if got := l.CheatedKey(t0.Add(-time.Hour), c.key); got != c.want {
			t.Errorf("CheatedKey(%q) = %v, want %v", c.key, got, c.want)
		}
	}
	if l.CheatedKey(t0.Add(time.Hour), "i5/Fresha") {
		t.Error("cheat before since counted")
	}
	if !l.Cheated(t0.Add(-time.Hour), "Fresha") || l.Cheated(t0.Add(-time.Hour), "Nobody") {
		t.Error("name-keyed Cheated changed")
	}
	es, _ := l.List(Filter{Key: "i5/Fresha"})
	if len(es) != 1 {
		t.Fatalf("filter by key: %d entries", len(es))
	}
	s := Summarize(mustList(t, l))
	if s.ByKey["i5/Fresha"] != 1 || s.ByKey["i4/Fresha"] != 1 || s.ByKey["Old"] != 1 || s.ByKey["uid-7"] != 1 || s.ByCharacter["Fresha"] != 3 {
		t.Fatalf("summary %+v", s)
	}
}

func mustList(t *testing.T, l *Ledger) []Entry {
	es, err := l.List(Filter{})
	if err != nil {
		t.Fatal(err)
	}
	return es
}
