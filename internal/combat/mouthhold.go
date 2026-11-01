package combat

import (
	"errors"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/input"
	"d2lod-bot/internal/nav"
)

// FIXMOUTHLOOP. After the mouth gate's rewrite cap a walk fails with nav.ErrMouthGate; every engage tick then started
// a fresh walk (and step-away) from the same spot: 117 times in one Den cycle (V DRILLS19: monsters blocked the away
// route, the cycle lasted the 120 s phase limit, so about one restart per second). The Fighter now remembers the
// hold and does not walk from that spot again for a while; the tick costs a poll and the fight goes on in place.
const (
	mouthHoldBackoff = 3 * time.Second  // between restarts from one spot, for the first mouthHoldCap holds
	mouthHoldCap     = 3                // holds at one spot, then no restart until mouthHoldReset passes
	mouthHoldReset   = 30 * time.Second // long hold after the cap, and the window that links holds to one spot
)

// mouthHoldState is the back-off memory; holds counts consecutive holds at pos.
type mouthHoldState struct {
	pos   input.Point
	at    time.Time
	holds int
	skips int // walks not started because of the back-off (reported with the next hold)
}

func (f *Fighter) playerPos() (input.Point, bool) {
	if f.snapFn == nil {
		return input.Point{}, false
	}
	var s *d2.Snapshot
	s, err := f.snapFn()
	if err != nil || s == nil || s.Player == nil {
		return input.Point{}, false
	}
	return input.Point{X: s.Player.X, Y: s.Player.Y}, true
}

// mouthHoldActive: a walk from here would repeat a held one.
func (f *Fighter) mouthHoldActive() bool {
	h := &f.mouthHold
	if h.holds == 0 {
		return false
	}
	wait := mouthHoldBackoff
	if h.holds >= mouthHoldCap {
		wait = mouthHoldReset
	}
	if f.now().Sub(h.at) >= wait {
		return false
	}
	pos, ok := f.playerPos()
	if !ok || pdist(pos, h.pos) > pinRadius {
		h.holds = 0
		return false
	}
	return true
}

// noteMouthHold records a walk that failed with the mouth gate and logs it once per episode (debug).
func (f *Fighter) noteMouthHold(to input.Point, err error) {
	pos, _ := f.playerPos()
	h := &f.mouthHold
	now := f.now()
	if h.holds > 0 && now.Sub(h.at) < mouthHoldReset && pdist(pos, h.pos) <= pinRadius {
		h.holds++
	} else {
		*h = mouthHoldState{holds: 1}
	}
	h.pos, h.at = pos, now
	if h.holds == 1 || h.holds == mouthHoldCap {
		f.log().Debug("combat move held by the mouth gate, backing off", "at", pos, "to", to, "holds", h.holds,
			"skipped", h.skips, "backoff", mouthHoldBackoff, "long", mouthHoldReset, "err", err)
	}
}

func isMouthHold(err error) bool { return errors.Is(err, nav.ErrMouthGate) }
