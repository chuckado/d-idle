package main

import (
	"path/filepath"

	"d2lod-bot/internal/goals"
	"d2lod-bot/internal/goldfarm"
	"d2lod-bot/internal/repair"
	"d2lod-bot/internal/watch"
)

// The Loop order's bot side (docs/loop-order.md): the config the Fun is built with and the repair desk the loop talks
// to. The loop itself is a sequencer in internal/watch; it places the same orders as POST /api/orders and asks for the
// repair the way the Repair now button does, so no new behaviour lives in internal/bot.

// loopConfig is FunConfig.OrderLoop: the state next to the other ladder state, the shared Pact / low-life halt, the
// journey status for the row, and the charter quote of the tally strip.
func loopConfig(stateDir string, halt func() string, status func() goals.Status, quote *watch.TallyStrip) *watch.OrderLoopConfig {
	c := &watch.OrderLoopConfig{Halt: halt}
	if stateDir != "" {
		c.Path = filepath.Join(stateDir, "orderloop.json")
	}
	if status != nil {
		c.Blocked = func() string { return status().JourneyBlocked }
	}
	if quote != nil {
		c.Quote = quote.Quote
	}
	return c
}

// bindLoop gives the loop the repair desk. No-op without a loop or a controller.
func bindLoop(f *watch.Fun, c *repair.Controller, gf *goldfarm.Controller) {
	if f == nil || f.Loop == nil || c == nil {
		return
	}
	f.Loop.Bind(loopDesk{c: c, a: repairAdapter{c, gf}})
}

// loopDesk is the repair controller as the loop sees it. Request goes through the same adapter as the button, so a
// request that has to farm first also switches the Farm gold order on for itself.
type loopDesk struct {
	c *repair.Controller
	a repairAdapter
}

func (d loopDesk) Worn() bool     { return repair.Worst(d.c.View().Gear) != "" }
func (d loopDesk) Request() error { _, err := d.a.Request(); return err }
func (d loopDesk) Cancel() error  { return d.a.Cancel() }

func (d loopDesk) Status() (string, bool, string) {
	q := d.c.View().Request
	if q == nil {
		return watch.LoopReqNone, false, ""
	}
	return q.State, q.Auto, q.Note
}

var _ watch.LoopRepairDesk = loopDesk{}
