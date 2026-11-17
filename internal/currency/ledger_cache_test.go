package currency

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"
)

// fill books n credits spread over two characters and two runs (two accounts/heroes), plus spends.
func fill(t testing.TB, l *Ledger, from, n int) {
	t.Helper()
	for i := from; i < from+n; i++ {
		ch, run := "Booti", "runA"
		if i%2 == 1 {
			ch, run = "Palli", "runB"
		}
		e := Entry{Kind: KindCredit, Currency: "tally", Source: "grant", Amount: 3, Key: "k" + strconv.Itoa(i), Character: ch, Run: run, Note: "synthetic ledger line to give the file a realistic size"}
		if _, err := l.Append(e); err != nil {
			t.Fatal(err)
		}
	}
}

func sameAsFresh(t *testing.T, l *Ledger) {
	t.Helper()
	fresh, err := Open(l.Dir(), Options{Mode: l.Mode(), Now: l.opt.Now})
	if err != nil {
		t.Fatal(err)
	}
	want, werr := fresh.Entries()
	got, gerr := l.Entries()
	if (werr == nil) != (gerr == nil) {
		t.Fatalf("errors differ: fresh %v cached %v", werr, gerr)
	}
	if werr == nil && !reflect.DeepEqual(want, got) {
		t.Fatalf("cached entries differ from a full load (%d vs %d)", len(got), len(want))
	}
	wb, _ := fresh.Balance("tally")
	gb, _ := l.Balance("tally")
	if wb != gb {
		t.Fatalf("balance %+v vs %+v", gb, wb)
	}
}

func TestLedgerCacheEqualsFullLoad(t *testing.T) {
	a, c := newLedger(t, didle)
	b, err := Open(a.Dir(), Options{Mode: didle, Instance: "i5", Now: c.now}) // a second writer on the same dir
	if err != nil {
		t.Fatal(err)
	}
	fill(t, a, 0, 20)
	sameAsFresh(t, a)
	fill(t, b, 20, 15) // appended behind a's back: a reads only the tail
	sameAsFresh(t, a)
	sameAsFresh(t, b)
	fill(t, a, 35, 5)
	sameAsFresh(t, b)
	if _, err := a.Spend(SpendRequest{Upgrade: "fade_faster", Key: "s1", Character: "Booti"}); err != nil {
		t.Fatal(err)
	}
	sameAsFresh(t, b)
	if err := a.Verify(); err != nil {
		t.Fatal(err)
	}
	// a duplicate key is still refused from the cached state
	if _, err := a.Append(Entry{Kind: KindCredit, Currency: "tally", Source: "grant", Amount: 3, Key: "k3"}); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("want duplicate, got %v", err)
	}
}

func TestLedgerCacheDetectsTruncationAndTamper(t *testing.T) {
	l, _ := newLedger(t, didle)
	fill(t, l, 0, 10)
	if _, err := l.Entries(); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(l.Dir(), "ledger.jsonl")
	raw, _ := os.ReadFile(p)
	// truncated to a shorter whole-line prefix: the head names a later seq
	cut := 0
	for n := 0; n < 5; n++ {
		cut += indexNL(raw[cut:]) + 1
	}
	if err := os.WriteFile(p, raw[:cut], 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Entries(); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("truncation not detected: %v", err)
	}
	// same size, different content, new mtime: the chain check catches it
	tam := append([]byte(nil), raw...)
	i := indexOf(tam, "k1\"")
	tam[i+1] = 'x'
	if err := os.WriteFile(p, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Entries(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(p, tam, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := l.Verify(); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("tamper not detected by Verify: %v", err)
	}
	// replaced file (other inode) with a good copy is re-read
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	sameAsFresh(t, l)
}

func TestLedgerCacheTornTail(t *testing.T) {
	l, _ := newLedger(t, didle)
	fill(t, l, 0, 4)
	if _, err := l.Entries(); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(l.Dir(), "ledger.jsonl")
	f, _ := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o600)
	f.WriteString(`{"seq":5,"ts"`)
	f.Close()
	if _, err := l.Entries(); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("torn tail not reported: %v", err)
	}
	fill(t, l, 4, 1) // Append repairs it
	if l.Repaired() == 0 {
		t.Fatal("repair not reported")
	}
	sameAsFresh(t, l)
	if es, _ := l.Entries(); len(es) != 5 {
		t.Fatalf("%d entries", len(es))
	}
}

func indexNL(b []byte) int {
	for i, c := range b {
		if c == '\n' {
			return i
		}
	}
	return -1
}

func indexOf(b []byte, s string) int {
	for i := 0; i+len(s) <= len(b); i++ {
		if string(b[i:i+len(s)]) == s {
			return i
		}
	}
	return -1
}

func benchLedger(b *testing.B, entries int, cached bool) {
	c := &clock{time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)}
	dir := b.TempDir()
	l, _ := Open(dir, Options{Mode: didle, Now: c.now})
	fill(b, l, 0, entries)
	fi, _ := os.Stat(filepath.Join(dir, "ledger.jsonl"))
	b.ReportMetric(float64(fi.Size())/1024, "KiB")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !cached {
			l.cache = nil
		}
		if _, err := l.Entries(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLedgerRead(b *testing.B) {
	for _, n := range []int{1500, 24000} { // about 312 KB and 5 MB
		for _, cached := range []bool{false, true} {
			b.Run(fmt.Sprintf("n%d/cached=%v", n, cached), func(b *testing.B) { benchLedger(b, n, cached) })
		}
	}
}
