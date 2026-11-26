package combat

import (
	"sync"
	"sync/atomic"
	"time"

	"d2lod-bot/internal/admit"
	"d2lod-bot/internal/d2"
)

// arrivalState tracks the last area the guard saw and when the hero came into it.
type arrivalState struct {
	mu     sync.Mutex
	area   int
	at     time.Time
	bounce atomic.Bool // the last Safety verdict was the arrival crowd rule (crowdAct skips the retreat walk)
}

// arrivalCrowd is the crowd-on-arrival rule (card act2temple, data: admission.json "arrival"): within the row's window after
// an area change into a listed area, CrowdN hostiles within Radius subtiles and life at or below LifePct. The quest crowd
// override fights up to 40 hostiles, which killed the Claw Viper Temple stand-in 1 s after arrival (V lab runs, CHEATED setup,
// mechanics only). Normal only, like every admission row. U: whether the live game gives that density.
func (f *Fighter) arrivalCrowd(p *d2.Player, hs []Target) bool {
	now := f.now()
	f.arr.mu.Lock()
	if p.Area != f.arr.area {
		f.arr.area, f.arr.at = p.Area, now
	}
	at := f.arr.at
	f.arr.mu.Unlock()
	f.arr.bounce.Store(false)
	if p.Difficulty != d2.Normal || p.InTown() {
		return false
	}
	rows, err := admit.Load()
	if err != nil {
		return false
	}
	row, ok := rows.ArrivalFor(p.Area)
	if !ok || now.Sub(at) > time.Duration(row.WithinSec)*time.Second || pct(p.Life, p.MaxLife) > row.LifePct {
		return false
	}
	if CrowdWithin(hs, float64(row.Radius)) < row.CrowdN {
		return false
	}
	f.arr.bounce.Store(true)
	return true
}
