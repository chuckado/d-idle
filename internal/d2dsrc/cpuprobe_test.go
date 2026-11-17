package d2dsrc

import (
	"context"
	"testing"
)

// BenchmarkReadLate is one full in-game snapshot (8 pipelined verbs, head and item dumps parsed) of a late-game
// character: what every nav poll costs the bot besides d2d's own frame (docs/probes/bot-cpu-2026-11-16.md).
func BenchmarkReadLate(b *testing.B) {
	_, c := startScript(b, lateReplies(b))
	src := New(c, Options{Tables: fakeTables{}, Logf: quiet})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := src.Read(context.Background()); err != nil {
			b.Fatal(err)
		}
	}
}
