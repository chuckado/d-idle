package bot

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
)

// The stall diagnostic (Questy death 11: 14 s with no log line and no survival action, seven hostiles): the guard logs one
// `stall: goroutine stack` after guardStallSec of silence with enough hostiles near, rate limited, counted, and fed to the soft
// invariant guard-stall. Quiet for less, fewer hostiles, in town or a recent line change nothing. Two classes, fake clock.
func TestGuardStallLogsOneStack(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		for _, tc := range []struct {
			name     string
			hostiles int
			quiet    time.Duration
			town     bool
			want     bool
		}{
			{"silent with a pack", 3, 15 * time.Second, false, true},
			{"silent but one hostile", 1, 15 * time.Second, false, false},
			{"a recent log line", 3, 4 * time.Second, false, false},
			{"in town", 3, 15 * time.Second, true, false},
		} {
			g := newRig(t)
			buf := &bytes.Buffer{}
			g.r.e.Log = slog.New(slog.NewTextHandler(buf, nil))
			g.w.set(func(w *world) { w.class, w.level, w.area = class, 12, 2 })
			for i := 0; i < tc.hostiles; i++ {
				g.w.addMonster(uint32(20+i), 10+i, 8, 0)
			}
			if tc.town {
				g.w.set(func(w *world) { w.area = 1 })
			}
			g.r.poll()
			now := g.e.Clock.Now()
			g.r.lastLogAt.Store(now.Add(-tc.quiet).UnixNano())
			got := g.r.stallCheck(g.r.latest(), now)
			if got != tc.want || strings.Contains(buf.String(), "stall: goroutine stack") != tc.want {
				t.Errorf("%v %s: stall %v (log %q), want %v", class, tc.name, got, buf.String(), tc.want)
			}
			if tc.want {
				if g.r.stalls.Load() != 1 || len(g.r.inv.pending) != 1 || g.r.inv.pending[0].Name != "guard-stall" {
					t.Errorf("%v: stalls %d pending %+v", class, g.r.stalls.Load(), g.r.inv.pending)
				}
				if again := g.r.stallCheck(g.r.latest(), now.Add(time.Minute)); again {
					t.Errorf("%v: a second stall line within the gap", class)
				}
				if later := g.r.stallCheck(g.r.latest(), now.Add(guardStallGap+time.Second)); !later {
					t.Errorf("%v: no stall line after the gap", class)
				}
			}
		}
	}
}
