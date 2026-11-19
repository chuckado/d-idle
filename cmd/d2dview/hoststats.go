package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"d2lod-bot/internal/watch/palfeed"
)

// hostStats gives the stream monitor the numbers only this process can see: how much of the last second the container's
// CPU quota throttled everything (cgroup v2 cpu.stat; the sandbox has one core, and a throttled 100 ms period freezes d2d,
// d2dview and the bot together, which shows as a hitch in the picture), this process' own CPU, and the ring rate. Sampled
// lazily (at most once per 900 ms, when the page polls), so an unwatched run pays nothing. Linux only; elsewhere zeros.
type hostStats struct {
	mu       sync.Mutex
	path     string
	last     palfeed.ServerStats
	at       time.Time
	periods  int64
	throttle int64
	usec     int64
	cpu      time.Duration
	pubs     int64
	pubsFn   func() int64
}

func newHostStats(pubs func() int64) *hostStats {
	return &hostStats{path: "/sys/fs/cgroup/cpu.stat", pubsFn: pubs}
}

// readCgroup parses cpu.stat text: nr_periods, nr_throttled, throttled_usec.
func readCgroup(path string) (periods, throttled, usec int64, ok bool) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, 0, false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		k, v, found := strings.Cut(sc.Text(), " ")
		if !found {
			continue
		}
		n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			continue
		}
		switch k {
		case "nr_periods":
			periods, ok = n, true
		case "nr_throttled":
			throttled = n
		case "throttled_usec":
			usec = n
		}
	}
	return
}

// Sample returns the numbers over the time since the previous sample (the previous answer again when asked within 900 ms).
func (h *hostStats) Sample() palfeed.ServerStats {
	h.mu.Lock()
	defer h.mu.Unlock()
	now := time.Now()
	dt := now.Sub(h.at)
	if !h.at.IsZero() && dt < 900*time.Millisecond {
		return h.last
	}
	var s palfeed.ServerStats
	if p, t, u, ok := readCgroup(h.path); ok {
		if !h.at.IsZero() && p > h.periods && dt < 10*time.Second {
			s.ThrottlePct = 100 * float64(t-h.throttle) / float64(p-h.periods)
			s.ThrottleMS = float64(u-h.usec) / 1000 * float64(time.Second) / float64(dt)
		}
		h.periods, h.throttle, h.usec = p, t, u
	}
	cpu := procCPU()
	if !h.at.IsZero() && dt < 10*time.Second {
		s.CPUPct = 100 * float64(cpu-h.cpu) / float64(dt)
		if h.pubsFn != nil {
			s.RingFPS = float64(h.pubsFn()-h.pubs) * float64(time.Second) / float64(dt)
		}
	}
	h.cpu, h.at = cpu, now
	if h.pubsFn != nil {
		h.pubs = h.pubsFn()
	}
	s.AtMS = now.UnixMilli()
	h.last = s
	return s
}
