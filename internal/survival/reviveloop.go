package survival

import (
	"sync"
	"time"
)

// ReviveLoop is the breaker of a merc revive loop: he is revived, dies within Within in the same zone, is revived again.
// After N such quick deaths in a zone the zone is closed to reviving for Hold (card survfix2, seed 103 of the screen:
// four revives of 175 to 531 gold, each worth about a minute). A death after a longer life, or the Hold running out,
// resets the zone. N 0 turns it off.
type ReviveLoop struct {
	N      int
	Within time.Duration
	Hold   time.Duration

	mu        sync.Mutex
	revivedAt time.Time
	rows      map[ZoneKey]*loopRow
}

type loopRow struct {
	fast  int
	until time.Time
}

// Revived records that he is alive again (the chore paid, or the lab revived him).
func (l *ReviveLoop) Revived(now time.Time) {
	l.mu.Lock()
	l.revivedAt = now
	l.mu.Unlock()
}

// Died records his death in k. life is the time since the last revive (0 and known false when he died without one in
// this run); fast is true when it was within the limit; tripped is true when this death closed the zone.
func (l *ReviveLoop) Died(k ZoneKey, now time.Time) (life time.Duration, known, fast, tripped bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.revivedAt.IsZero() {
		return 0, false, false, false
	}
	life = now.Sub(l.revivedAt)
	l.revivedAt = time.Time{}
	if l.rows == nil {
		l.rows = map[ZoneKey]*loopRow{}
	}
	r := l.rows[k]
	if r == nil {
		r = &loopRow{}
		l.rows[k] = r
	}
	if life > l.Within {
		r.fast = 0
		return life, true, false, false
	}
	r.fast++
	if l.N > 0 && r.fast >= l.N && !now.Before(r.until) {
		r.until = now.Add(l.Hold)
		tripped = true
	}
	return life, true, true, tripped
}

// Blocked is true while reviving in k is closed.
func (l *ReviveLoop) Blocked(k ZoneKey, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.blocked(l.rows[k], now)
}

// BlockedArea is Blocked for any difficulty of the area; area 0 asks about every zone.
func (l *ReviveLoop) BlockedArea(area int, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for k, r := range l.rows {
		if (area == 0 || k.Area == area) && l.blocked(r, now) {
			return true
		}
	}
	return false
}

func (l *ReviveLoop) blocked(r *loopRow, now time.Time) bool {
	return l.N > 0 && r != nil && now.Before(r.until)
}

// Reset forgets zone k (the hero got stronger, or the user asked).
func (l *ReviveLoop) Reset(k ZoneKey) {
	l.mu.Lock()
	delete(l.rows, k)
	l.mu.Unlock()
}
