package ether

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

func open(t *testing.T, dir string, content func() int, c *clock) *Wallet {
	t.Helper()
	w, err := Open(dir, Options{Content: content, Now: c.now})
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func newClock() *clock { return &clock{time.Date(2026, 11, 17, 10, 0, 0, 0, time.UTC)} }

func mustEarn(t *testing.T, w *Wallet, src, sub, diff string, hook int) Result {
	t.Helper()
	r, err := w.Earn(src, sub, diff, hook)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestEarnIdempotentAndReplay(t *testing.T) {
	dir, c := t.TempDir(), newClock()
	w := open(t, dir, nil, c)
	if r := mustEarn(t, w, "quest", "den-of-evil", "normal", 5); r.Credited != 5 {
		t.Fatal(r)
	}
	if r := mustEarn(t, w, "quest", "den-of-evil", "normal", 5); !r.Already || r.Credited != 0 {
		t.Fatalf("double credit: %+v", r)
	}
	// another difficulty is another key
	mustEarn(t, w, "quest", "den-of-evil", "nm", 8)
	mustEarn(t, w, "boss", "andariel", "normal", 0)
	mustEarn(t, w, "first", "monster/fallen", "", 0)
	mustEarn(t, w, "milestone", "clvl10", "", 0)
	if b, _ := w.Balance(); b != 5+8+10+5+5 {
		t.Fatalf("balance %d", b)
	}
	// replay: reopen, and the same journal books nothing
	w2 := open(t, dir, nil, c)
	b1, _ := w.Balance()
	if b2, _ := w2.Balance(); b2 != b1 {
		t.Fatalf("replay %d != %d", b2, b1)
	}
	if r := mustEarn(t, w2, "boss", "andariel", "normal", 0); !r.Already {
		t.Fatal("replayed boss paid again")
	}
	if n := len(w2.Entries()); n != 5 {
		t.Fatalf("%d entries", n)
	}
}

func TestEarnErrors(t *testing.T) {
	w := open(t, t.TempDir(), nil, newClock())
	for _, c := range []struct {
		src, sub, diff string
		hook           int
	}{
		{"nope", "x", "", 0},
		{"quest", "q", "", 5},         // difficulty missing
		{"quest", "q", "easy", 5},     // unknown difficulty
		{"quest", "q", "normal", 999}, // over maxHook
		{"boss", "b", "nightmare", 0},
		{"milestone", "clvl99", "", 0}, // not a row
		{"first", "", "", 0},
	} {
		if _, err := w.Earn(c.src, c.sub, c.diff, c.hook); err == nil {
			t.Errorf("%+v accepted", c)
		}
	}
	if b, _ := w.Balance(); b != 0 {
		t.Fatal(b)
	}
}

func TestKillTrickleDailyCap(t *testing.T) {
	c := newClock()
	w := open(t, t.TempDir(), nil, c)
	paid := 0
	for i := 0; i < 25; i++ {
		r := mustEarn(t, w, "kills", fmt.Sprintf("run1/%d", i), "", 0)
		paid += r.Credited
		if i >= 10 && !r.Capped {
			t.Fatalf("kill %d not capped", i)
		}
	}
	if paid != 10 {
		t.Fatalf("paid %d", paid)
	}
	c.t = c.t.Add(24 * time.Hour)
	if r := mustEarn(t, w, "kills", "run1/99", "", 0); r.Credited != 1 {
		t.Fatalf("next day: %+v", r)
	}
}

func TestUpgradeSpendRefundRules(t *testing.T) {
	dir, c := t.TempDir(), newClock()
	count := 0
	w := open(t, dir, func() int { return count }, c)

	if _, err := w.Upgrade(""); !errors.Is(err, ErrInsufficient) {
		t.Fatalf("empty wallet: %v", err)
	}
	var se *ShortError
	_, err := w.Upgrade("")
	if !errors.As(err, &se) || se.Need != 20 {
		t.Fatalf("short error: %v", err)
	}
	mustEarn(t, w, "quest", "a", "hell", 100)
	mustEarn(t, w, "quest", "b", "hell", 100)
	if _, err := w.Upgrade("v2"); !errors.Is(err, ErrOutOfOrder) {
		t.Fatalf("skip a step: %v", err)
	}
	r, err := w.Upgrade("v1")
	if err != nil || r.Capacity != 600 || r.Balance != 180 || r.Already {
		t.Fatalf("v1: %+v %v", r, err)
	}
	if r, _ := w.Upgrade("v1"); !r.Already || r.Balance != 180 {
		t.Fatalf("double click: %+v", r)
	}
	if r, err = w.Upgrade(""); err != nil || r.Step != "v2" || r.Capacity != 900 || r.Balance != 140 {
		t.Fatalf("v2: %+v %v", r, err)
	}
	if owned, _ := w.OwnedSteps(); fmt.Sprint(owned) != "[v1 v2]" {
		t.Fatal(owned)
	}

	// refund: only the last step, only while the content fits the lower tier
	count = 700
	if _, err := w.Downgrade("v1"); !errors.Is(err, ErrOutOfOrder) {
		t.Fatalf("refund not last: %v", err)
	}
	var fe *FitError
	if _, err := w.Downgrade(""); !errors.As(err, &fe) || fe.Count != 700 || fe.Capacity != 600 {
		t.Fatalf("in use: %v", err)
	}
	if s := w.Summary(); s.Downgrade.Available || s.Downgrade.Reason == "" {
		t.Fatalf("summary %+v", s.Downgrade)
	}
	if b, _ := w.Balance(); b != 140 {
		t.Fatalf("refused refund changed balance %d", b)
	}
	count = 600
	r, err = w.Downgrade("")
	if err != nil || r.Step != "v2" || r.Balance != 180 || r.Capacity != 600 {
		t.Fatalf("refund: %+v %v", r, err)
	}
	if r, _ := w.Downgrade("v2"); !r.Already {
		t.Fatal("second refund of the same step must be a no-op")
	}
	// buy again after a refund (key carries the refund count), then replay
	if _, err := w.Upgrade("v2"); err != nil {
		t.Fatal(err)
	}
	w2 := open(t, dir, nil, c)
	b1, _ := w.Balance()
	b2, _ := w2.Balance()
	o1, _ := w.OwnedSteps()
	o2, _ := w2.OwnedSteps()
	if b1 != b2 || fmt.Sprint(o1) != fmt.Sprint(o2) || b1 != 140 {
		t.Fatalf("replay %d %v vs %d %v", b2, o2, b1, o1)
	}
	if err := (&Wallet{}).SpendKey("", 1); err == nil {
		t.Fatal("empty key")
	}
}

func TestAllStepsThenNothingToBuy(t *testing.T) {
	w := open(t, t.TempDir(), nil, newClock())
	for i := 0; i < 30; i++ {
		mustEarn(t, w, "quest", fmt.Sprint("q", i), "hell", 100)
	}
	for i := 0; i < 6; i++ {
		if _, err := w.Upgrade(""); err != nil {
			t.Fatalf("step %d: %v", i, err)
		}
	}
	if _, err := w.Upgrade(""); !errors.Is(err, ErrNothingToBuy) {
		t.Fatal(err)
	}
	if c := w.Capacity(0); c != 2400 {
		t.Fatalf("capacity %d", c)
	}
	if s := w.Summary(); s.Next != nil || s.Capacity != 2400 {
		t.Fatalf("%+v", s)
	}
}

// A crash after the ledger line and before any tier change cannot exist (the tier is derived), but a crash in the
// middle of the line can: the torn tail is dropped and the earlier state is intact.
func TestCrashTornTail(t *testing.T) {
	dir, c := t.TempDir(), newClock()
	w := open(t, dir, nil, c)
	mustEarn(t, w, "quest", "a", "hell", 100)
	if _, err := w.Upgrade(""); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, FileName)
	f, _ := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0)
	f.WriteString(`{"seq":3,"ts":"2026-11-17T10:00:00Z","kind":"spend","amount":40,"key":"vault/v2#0","step":"v`)
	f.Close()
	w2 := open(t, dir, nil, c)
	if w2.Repaired() == 0 {
		t.Fatal("tail not reported")
	}
	if o, _ := w2.OwnedSteps(); fmt.Sprint(o) != "[v1]" {
		t.Fatalf("owned %v", o)
	}
	if b, _ := w2.Balance(); b != 80 {
		t.Fatal(b)
	}
	// the file is whole again and the next append is seq 3
	e, err := w2.Upgrade("")
	if err != nil || e.Step != "v2" {
		t.Fatal(e, err)
	}
	if es := open(t, dir, nil, c).Entries(); len(es) != 3 || es[2].Seq != 3 {
		t.Fatalf("%+v", es)
	}
}

func TestCorruptMiddleRefused(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, FileName), []byte("garbage\n{\"seq\":1}\n"), 0o644)
	if _, err := Open(dir, Options{}); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("%v", err)
	}
	// a forged spend over the balance is corrupt too (a hand-edited wallet is a cheat)
	os.WriteFile(filepath.Join(dir, FileName), []byte(`{"seq":1,"ts":"x","kind":"spend","amount":5,"key":"k"}`+"\n"), 0o644)
	if _, err := Open(dir, Options{}); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("%v", err)
	}
}

func TestTwoProfilesAreSeparate(t *testing.T) {
	c := newClock()
	a, b := open(t, t.TempDir(), nil, c), open(t, t.TempDir(), nil, c)
	mustEarn(t, a, "quest", "q", "normal", 5)
	mustEarn(t, b, "quest", "q", "normal", 7) // same key in another account pays there
	ba, _ := a.Balance()
	bb, _ := b.Balance()
	if ba != 5 || bb != 7 {
		t.Fatal(ba, bb)
	}
}

// Two heroes share the wallet: the capacity is account-wide, the refund check reads the active hero's content.
func TestHeroesShareWalletRefundReadsActiveHero(t *testing.T) {
	hero := map[string]int{"a": 650, "b": 10}
	cur := "a"
	w := open(t, t.TempDir(), func() int { return hero[cur] }, newClock())
	mustEarn(t, w, "quest", "q", "hell", 100)
	if _, err := w.Upgrade(""); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Downgrade(""); !errors.Is(err, ErrInUse) {
		t.Fatal(err)
	}
	cur = "b"
	if _, err := w.Downgrade(""); err != nil {
		t.Fatalf("hero b fits: %v", err)
	}
}

func TestPropertyBalanceNeverNegativeAndReplays(t *testing.T) {
	for seed := int64(1); seed <= 20; seed++ {
		rng := rand.New(rand.NewSource(seed))
		dir, c := t.TempDir(), newClock()
		count := 0
		w := open(t, dir, func() int { return count }, c)
		for i := 0; i < 300; i++ {
			switch rng.Intn(6) {
			case 0, 1:
				w.Earn("quest", fmt.Sprint("q", rng.Intn(15)), []string{"normal", "nm", "hell"}[rng.Intn(3)], rng.Intn(60))
			case 2:
				w.Earn("kills", fmt.Sprint("k", i), "", 0)
			case 3:
				w.Upgrade("")
			case 4:
				count = rng.Intn(2500)
				w.Downgrade("")
			case 5:
				w.SpendKey(fmt.Sprint("misc", rng.Intn(20)), 1+rng.Intn(30))
			}
			if b, _ := w.Balance(); b < 0 {
				t.Fatalf("seed %d step %d: balance %d", seed, i, b)
			}
			if rng.Intn(50) == 0 {
				c.t = c.t.Add(12 * time.Hour)
			}
		}
		w2 := open(t, dir, nil, c)
		b1, _ := w.Balance()
		b2, _ := w2.Balance()
		o1, _ := w.OwnedSteps()
		o2, _ := w2.OwnedSteps()
		if b1 != b2 || fmt.Sprint(o1) != fmt.Sprint(o2) {
			t.Fatalf("seed %d replay %d %v vs %d %v", seed, b2, o2, b1, o1)
		}
		sum := 0
		for _, e := range w.Entries() {
			if e.Kind == Spend {
				sum -= e.Amount
			} else {
				sum += e.Amount
			}
		}
		if sum != b1 {
			t.Fatalf("seed %d entries sum %d != balance %d", seed, sum, b1)
		}
	}
}

func TestSourcesParse(t *testing.T) {
	if _, err := ParseSources([]byte(`{"currency":"E","sources":[{"id":"a","amount":1},{"id":"a","amount":2}]}`)); err == nil {
		t.Fatal("duplicate id")
	}
	if _, err := ParseSources([]byte(`{"currency":"E","sources":[{"id":"a"}]}`)); err == nil {
		t.Fatal("no amount style")
	}
	if _, err := DefaultSources(); err != nil {
		t.Fatal(err)
	}
}
