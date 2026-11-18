package ether

import (
	"testing"
	"time"
)

func bal(t *testing.T, w *Wallet) int {
	t.Helper()
	b, err := w.Balance()
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// The earn rules as a table: what each event pays, from data/vault/ether.json.
func TestEarnerRulesTable(t *testing.T) {
	hero := Hero{Name: "Sorc", Difficulty: "normal"}
	for _, c := range []struct {
		name string
		do   func(e *Earner)
		want int
	}{
		{"first monster", func(e *Earner) { e.First(hero, "monster", "fallen") }, 5},
		{"first area", func(e *Earner) { e.First(hero, "area", "Blood Moor") }, 5},
		{"kill of a boss pays first and boss", func(e *Earner) { e.Kill(hero, "Andariel", true) }, 5 + 10},
		{"boss in nightmare", func(e *Earner) { e.Kill(Hero{Name: "Sorc", Difficulty: "nm"}, "Andariel", true) }, 5 + 15},
		{"boss in hell", func(e *Earner) { e.Kill(Hero{Name: "Sorc", Difficulty: "hell"}, "Andariel", true) }, 5 + 25},
		{"plain kill pays the first only", func(e *Earner) { e.Kill(hero, "fallen", false) }, 5},
		{"level 9 pays nothing", func(e *Earner) { e.Level(hero, 9) }, 0},
		{"level 20 pays both rows", func(e *Earner) { e.Level(hero, 20) }, 5 + 10},
		{"reaching nightmare", func(e *Earner) { e.Level(Hero{Name: "Sorc", Difficulty: "nm"}, 1) }, 20},
		{"reaching hell pays both rows", func(e *Earner) { e.Reached(Hero{Name: "Sorc", Difficulty: "hell"}) }, 20 + 40},
		{"no hero name, no first", func(e *Earner) { e.First(Hero{}, "monster", "fallen") }, 0},
	} {
		w := open(t, t.TempDir(), nil, newClock())
		c.do(&Earner{W: w})
		if got := bal(t, w); got != c.want {
			t.Errorf("%s: balance %d, want %d", c.name, got, c.want)
		}
	}
}

// A restart (a second Earner over the reopened ledger) and a repeated event pay nothing again; a second hero earns its own firsts
// but not the account-wide boss and milestone rows.
func TestEarnerNoDoubleCreditAcrossRestartAndHeroes(t *testing.T) {
	dir, c := t.TempDir(), newClock()
	a := Hero{Name: "Sorc", Difficulty: "normal"}
	w := open(t, dir, nil, c)
	e := &Earner{W: w}
	e.Kill(a, "Andariel", true)
	e.Level(a, 10)
	before := bal(t, w)
	w2 := open(t, dir, nil, c)
	e2 := &Earner{W: w2}
	e2.Kill(a, "Andariel", true)
	e2.Level(a, 10)
	if bal(t, w2) != before {
		t.Fatalf("restart paid again: %d -> %d", before, bal(t, w2))
	}
	b := Hero{Name: "Pally", Difficulty: "normal"}
	e2.Kill(b, "Andariel", true) // first for the new hero only
	e2.Level(b, 10)
	if got := bal(t, w2); got != before+5 {
		t.Fatalf("second hero: %d, want %d", got, before+5)
	}
}

// The trickle: one credit per `per` kills, at most dailyCap a day, and the day rolls at local midnight; a restart in the same
// day continues at the next free key and never rewrites an old one.
func TestEarnerKillTrickleCapAndDay(t *testing.T) {
	dir := t.TempDir()
	c := &clock{time.Date(2026, 11, 17, 23, 30, 0, 0, time.Local)}
	w := open(t, dir, nil, c)
	e := &Earner{W: w}
	h := Hero{Name: "Sorc", Difficulty: "normal"}
	kill := func(n int) {
		for i := 0; i < n; i++ {
			e.Kill(h, "", false) // an empty name is ignored
		}
	}
	kill(1000)
	if bal(t, w) != 0 {
		t.Fatal("unnamed kills paid")
	}
	per, capd := 25, 10
	n := 0
	for i := 0; i < per*(capd+5); i++ {
		e.Kill(h, "Z", false)
		n++
	}
	if got := bal(t, w); got != 5+capd { // the first of Z, then the cap
		t.Fatalf("balance %d, want %d", got, 5+capd)
	}
	// restart on the same day: the cap still holds
	w2 := open(t, dir, nil, c)
	e2 := &Earner{W: w2}
	for i := 0; i < per*3; i++ {
		e2.Kill(h, "Z", false)
	}
	if got := bal(t, w2); got != 5+capd {
		t.Fatalf("after restart %d", got)
	}
	c.t = c.t.Add(time.Hour) // past local midnight
	for i := 0; i < per; i++ {
		e2.Kill(h, "Z", false)
	}
	if got := bal(t, w2); got != 5+capd+1 {
		t.Fatalf("next day %d", got)
	}
}

func TestTrickleErrors(t *testing.T) {
	w := open(t, t.TempDir(), nil, newClock())
	if _, err := w.Trickle("boss"); err == nil {
		t.Error("an uncapped source is no trickle")
	}
	if _, err := w.Trickle("nope"); err == nil {
		t.Error("unknown source")
	}
}
