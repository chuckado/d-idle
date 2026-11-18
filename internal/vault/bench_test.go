package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func fixture(b *testing.B, n int) string {
	b.Helper()
	dir := b.TempDir()
	s := open(b, dir, "sorc", n+10)
	for i := 0; i < n; i++ {
		it := mk(i, []string{"cap", "wsp", "ltp", "rin"}[i%4], 1+i%7, i%90)
		it.Sockets = i % 4
		if _, err := s.Add(it); err != nil {
			b.Fatal(err)
		}
	}
	if err := s.Close(); err != nil {
		b.Fatal(err)
	}
	return dir
}

func BenchmarkInsert(b *testing.B) {
	for _, n := range []int{10_000, 50_000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			s := open(b, b.TempDir(), "sorc", n+b.N+10)
			for i := 0; i < n; i++ {
				s.Add(mk(i, "cap", 4, i%90))
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := s.Add(mk(n+i, "cap", 4, i%90)); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkInsertTrim is the full-vault insert: a scan for the victim, an archive line, the add.
func BenchmarkInsertTrim(b *testing.B) {
	s := open(b, b.TempDir(), "sorc", 10_000)
	for i := 0; i < 10_000; i++ {
		s.Add(mk(i, "cap", 4, i%90))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Add(mk(10_000+i, "cap", 4, 90))
	}
}

// BenchmarkIndexRebuild is Open from the log alone (the snapshot removed), the worst case.
func BenchmarkIndexRebuild(b *testing.B) {
	for _, n := range []int{10_000, 50_000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			dir := fixture(b, n)
			os.Remove(filepath.Join(dir, IndexFile))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s := open(b, dir, "sorc", n+10)
				s.f.Close()
			}
		})
	}
}

// BenchmarkOpenSnapshot is Open with a matching snapshot.
func BenchmarkOpenSnapshot(b *testing.B) {
	dir := fixture(b, 50_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := open(b, dir, "sorc", 50_010)
		s.f.Close()
	}
}

func BenchmarkQuery(b *testing.B) {
	for _, n := range []int{10_000, 50_000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			s := open(b, b.TempDir(), "sorc", n+10)
			for i := 0; i < n; i++ {
				it := mk(i, []string{"cap", "wsp", "ltp", "rin"}[i%4], 1+i%7, i%90)
				s.Add(it)
			}
			two := 2
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s.Query(Filter{Code: "wsp", Quality: []int{6}, MinIlvl: 30, Sockets: &two, Limit: 50})
				s.Query(Filter{Name: "ltp #4", Limit: 50})
			}
		})
	}
}
