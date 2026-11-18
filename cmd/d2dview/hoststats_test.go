package main

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestReadCgroupAndSample(t *testing.T) {
	p := filepath.Join(t.TempDir(), "cpu.stat")
	write := func(per, thr, us int) {
		os.WriteFile(p, []byte("usage_usec 5\nnr_periods "+itoa(per)+"\nnr_throttled "+itoa(thr)+"\nthrottled_usec "+itoa(us)+"\nnr_bursts 0\n"), 0o600)
	}
	write(100, 10, 1000)
	if per, thr, us, ok := readCgroup(p); !ok || per != 100 || thr != 10 || us != 1000 {
		t.Fatalf("%d %d %d %v", per, thr, us, ok)
	}
	if _, _, _, ok := readCgroup(filepath.Join(t.TempDir(), "none")); ok {
		t.Fatal("missing file reported ok")
	}
	h := newHostStats(nil)
	h.path = p
	if s := h.Sample(); s.ThrottlePct != 0 || s.AtMS == 0 {
		t.Fatalf("first sample %+v", s)
	}
	if s := h.Sample(); s.AtMS == 0 { // within 900 ms: the same answer
		t.Fatal("cached sample lost")
	}
	h.at = h.at.Add(-time.Second)
	write(110, 15, 3_000_000)
	if s := h.Sample(); s.ThrottlePct < 45 || s.ThrottlePct > 55 {
		t.Fatalf("5 of 10 periods throttled gave %.1f%%", s.ThrottlePct)
	}
}

func itoa(n int) string { return strconv.Itoa(n) }
