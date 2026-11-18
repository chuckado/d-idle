package bot

import (
	"testing"
	"time"

	"d2lod-bot/internal/d2"
)

// The merc death is reported once, at the alive-to-dead edge, never on the first reading and never twice.
func TestZoneMercEdge(t *testing.T) {
	var got []ZoneEvent
	r := &Runner{d: Deps{ZoneObserve: func(ev ZoneEvent) { got = append(got, ev) }}}
	p := &d2.Player{Area: 7}
	now := time.Unix(1000, 0)
	for _, alive := range []bool{false, true, true, false, false, true, false} {
		r.zoneMercEdge(p, alive, now)
	}
	if len(got) != 2 || got[0].Kind != ZoneEvMercDied || got[0].Area != 7 {
		t.Fatalf("events %+v, want two merc deaths", got)
	}
	r.d.ZoneObserve = nil
	r.zoneObserve(ZoneEvent{Kind: ZoneEvDeath}) // no hook: no panic
}
