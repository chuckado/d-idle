package currency

import (
	"testing"
	"time"
)

// BenchmarkTallyPoll is what one viewer poll of the tally strip asks of the ledger: three summaries (all, trip, last
// trip) and the upgrade levels, on a ledger of about 312 KB (1500 entries) and 5 MB (24000).
func BenchmarkTallyPoll(b *testing.B) {
	for _, n := range []struct {
		name string
		n    int
	}{{"312KB", 1500}, {"5MB", 24000}} {
		b.Run(n.name, func(b *testing.B) {
			c := &clock{time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)}
			l, _ := Open(b.TempDir(), Options{Mode: didle, Now: c.now})
			for i := 0; i < n.n; i++ {
				if _, err := l.Append(Entry{Kind: KindCredit, Currency: "tally", Source: "grant", Amount: 1, Key: "k" + itoa(i),
					Character: "Booti", Run: "r", Note: "synthetic ledger line to give the file a realistic size"}); err != nil {
					b.Fatal(err)
				}
			}
			ta, err := NewTally(l, DefaultTallyRows(), "Booti", "r")
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for j := 0; j < 3; j++ {
					if _, err := ta.SummaryFor(Window{}); err != nil {
						b.Fatal(err)
					}
				}
				if _, err := l.Levels(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func itoa(i int) string {
	var b [12]byte
	p := len(b)
	for ; i > 0 || p == len(b); i /= 10 {
		p--
		b[p] = byte('0' + i%10)
	}
	return string(b[p:])
}
