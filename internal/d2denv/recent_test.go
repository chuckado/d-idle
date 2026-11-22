package d2denv

import (
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2dsrc"
)

func TestReaderRecent(t *testing.T) {
	r := &Reader{}
	if _, ok := r.Recent(time.Second); ok {
		t.Fatal("no snapshot read yet: Recent must say no")
	}
	snap := &d2.Snapshot{State: d2.StateInGame}
	r.last, r.lastAt = &d2dsrc.Result{Snap: snap}, time.Now()
	if got, ok := r.Recent(time.Second); !ok || got != snap {
		t.Fatalf("a snapshot just read must be returned: %v %v", got, ok)
	}
	r.lastAt = time.Now().Add(-2 * time.Second)
	if _, ok := r.Recent(time.Second); ok {
		t.Fatal("a snapshot older than maxAge must not be returned")
	}
}

// Recent measures a snapshot's age on the Reader's clock (card pgclock): a guard poll reuses the last snapshot or queues a
// fresh read by VIRTUAL age, so it does not depend on how long the machine took in wall time. Before, Recent used the wall
// clock and a loaded builder re-read where a quiet one reused, which moved the command stream (journal pair, tick 2242).
func TestReaderRecentUsesTheInjectedClock(t *testing.T) {
	for _, cls := range []d2.Class{d2.Sorceress, d2.Paladin} {
		virt := time.Unix(0, 0)
		r := &Reader{Now: func() time.Time { return virt }}
		snap := &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Class: cls}}
		r.last, r.lastAt = &d2dsrc.Result{Snap: snap}, virt
		time.Sleep(20 * time.Millisecond) // the wall moves, the env clock does not
		if got, ok := r.Recent(10 * time.Millisecond); !ok || got != snap {
			t.Fatalf("class %v: wall time aged a snapshot the env clock calls fresh", cls)
		}
		virt = virt.Add(350 * time.Millisecond)
		if _, ok := r.Recent(300 * time.Millisecond); ok {
			t.Fatalf("class %v: a snapshot older than maxAge on the env clock was reused", cls)
		}
	}
}
