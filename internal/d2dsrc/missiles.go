package d2dsrc

import (
	"context"
	"sync"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2dctl"
)

// missileOff remembers the sources whose d2d answered err to "debug missiles" (no listing): they are not asked again.
var missileOff sync.Map // *Source -> struct{}

// FillMissiles adds the missiles in flight to a snapshot read by Read, from the requested "debug missiles" listing
// (docs/fixmissile.md; the lab fork has the verb, docs/forkcheats.md). Rows with speed 0 are hit effects (a monster's
// explosion is listed with owner -1) and are dropped here, not in the parser, so no one dodges an explosion. It fills what Game.exe's reader fills: Class =
// missiles.txt id, OwnerType/OwnerID = the shooter, X/Y in subtiles (same origin rule as every other unit). A d2d
// without the verb is asked once and then never: the snapshot keeps no missiles, so no dodge fires. Only transport
// errors are returned; they are not fatal to the snapshot, the caller may ignore them.
func (s *Source) FillMissiles(ctx context.Context, res *Result) error {
	if res == nil || res.Snap == nil || res.Snap.State != d2.StateInGame {
		return nil
	}
	if _, off := missileOff.Load(s); off {
		return nil
	}
	l, err := s.c.Missiles(ctx)
	if err != nil {
		if isCmdErr(err) {
			missileOff.Store(s, struct{}{})
			s.warnOnce("nomissiles", "d2dsrc: debug missiles verb missing: no missile dodge on this d2d")
			return nil
		}
		return err
	}
	ox, oy := 0, 0
	if s.o.AddOrigin && res.Extras.Level != nil {
		ox, oy = 5*res.Extras.Level.OriginX, 5*res.Extras.Level.OriginY
	}
	ms := make([]d2.Unit, 0, len(l.Items))
	for _, m := range l.Items {
		if m.Speed <= 0 {
			continue
		}
		ms = append(ms, d2.Unit{Type: d2.TypeMissile, ID: uint32(m.Unit), Class: m.Class,
			OwnerType: uint32(m.OwnerType), OwnerID: uint32(m.OwnerID),
			X: d2dctl.SubTile(m.X) + ox, Y: d2dctl.SubTile(m.Y) + oy})
	}
	res.Snap.Missiles = ms
	return nil
}
