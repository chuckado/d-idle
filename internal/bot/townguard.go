package bot

import (
	"sync/atomic"
	"time"

	"d2lod-bot/internal/d2"
)

// choreGrace is how long the hero may read as outside a town while a town chore runs before the guard aborts the chore:
// a zone change reads an odd area for a moment. The 02:51 death of 2026-11-26 (Fable review 3.1, P3 d) ran the shop chore
// for 21 s in Cold Plains with 12 hostiles around; the town package now checks before each NPC step, and this is the
// check for the walk inside one step.
const choreGrace = 2 * time.Second

// choreWatch tracks one running town chore and when the hero was first seen outside a town during it.
type choreWatch struct {
	running  atomic.Bool
	outSince atomic.Int64 // virtual clock unix nanos of the first outside reading, 0 = inside
}

func (c *choreWatch) begin() { c.outSince.Store(0); c.running.Store(true) }
func (c *choreWatch) end()   { c.running.Store(false); c.outSince.Store(0) }

// outside reports that the chore has run outside town for longer than choreGrace.
func (c *choreWatch) outside(snap *d2.Snapshot, now time.Time) bool {
	if !c.running.Load() || snap == nil || snap.Player == nil || snap.State != d2.StateInGame || snap.Player.InTown() {
		c.outSince.Store(0)
		return false
	}
	since := c.outSince.Load()
	if since == 0 {
		c.outSince.Store(now.UnixNano())
		return false
	}
	return now.Sub(time.Unix(0, since)) >= choreGrace
}
