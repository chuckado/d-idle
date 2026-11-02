package bot

import (
	"testing"
	"time"
)

// Replay of the tpglitch dump pattern: full tome, stat 70 absent for 63 s (the longest drop-out), full again; one
// look every 2 s. The verdict must hold, so neither the kit gate nor the poverty check sees an empty tome.
func TestTPDebounceDropout(t *testing.T) {
	var d tpDebounce
	now := time.Unix(0, 0)
	for i := 0; i < 200; i++ {
		raw := !(i >= 20 && i < 52) // 32 looks x 2 s = 64 s of false
		if !lookV(&d, raw, now.Add(time.Duration(i)*2*time.Second)) {
			t.Fatalf("look %d: verdict flipped to empty during a drop-out", i)
		}
	}
}

func TestTPDebounceRealEmpty(t *testing.T) {
	var d tpDebounce
	now := time.Unix(0, 0)
	d.look(true, now)
	if !lookV(&d, false, now.Add(time.Second)) || !lookV(&d, false, now.Add(tpDebounceSpan-time.Second)) {
		t.Fatal("must hold inside the span")
	}
	if lookV(&d, false, now.Add(time.Second+tpDebounceSpan)) {
		t.Fatal("a tome empty for the whole span must read empty")
	}
	if !lookV(&d, true, now.Add(5*time.Minute)) {
		t.Fatal("a new scroll is taken at once")
	}
}

func lookV(d *tpDebounce, raw bool, now time.Time) bool { v, _ := d.look(raw, now); return v }
