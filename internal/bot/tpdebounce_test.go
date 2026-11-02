package bot

import (
	"testing"
	"time"
)

// Replay of the 17:50 cycle: HasTP false in area 3, true in area 2, one look per area entry about 35 s wall.
func TestTPDebounceReplay(t *testing.T) {
	var d tpDebounce
	now := time.Unix(0, 0)
	d.look(true, now)
	flips, last := 0, true
	for i := 0; i < 40; i++ { // 40 looks, 3 virtual minutes apart
		now = now.Add(3 * time.Minute)
		got := d.look(i%2 == 1, now)
		if got != last {
			flips++
			last = got
		}
	}
	if flips != 0 {
		t.Fatalf("alternating looks flipped the verdict %d times", flips)
	}
}

func TestTPDebounceHolds(t *testing.T) {
	var d tpDebounce
	now := time.Unix(0, 0)
	d.look(true, now)
	if !d.look(false, now) || !d.look(false, now.Add(time.Second)) {
		t.Fatal("two looks must not flip")
	}
	if d.look(false, now.Add(2*time.Second)) {
		t.Fatal("three consecutive looks flip")
	}
	d = tpDebounce{}
	d.look(true, now)
	d.look(false, now.Add(time.Second))
	if d.look(false, now.Add(3*time.Minute)) {
		t.Fatal("a verdict held for the span flips")
	}
}
