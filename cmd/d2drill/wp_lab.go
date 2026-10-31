package main

import (
	"context"
	"errors"
	"fmt"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/drill"
	"d2lod-bot/internal/nav"
)

// wpNavigator is the bot's navigator (d2denv.Backend.Navigator, what cmd/d2bot hands the goals as Mover) with the spy in
// front of the input. It lives as long as the lab: the arrival clock of the waypoint cooldown is part of its state.
// TakeWaypoint is shared code (internal/nav), called here exactly as internal/bot and internal/town call it, so no
// copy of bot code is needed for the DW drills.
func (l *lab) wpNavigator() (*nav.Navigator, *spy) {
	if l.wpNav == nil {
		e := *l.be.Env
		l.wpSpy = &spy{Controller: e.In, l: l, last: -1}
		e.In = l.wpSpy
		n := nav.New(&e)
		n.UseWorld(l.be.World)
		n.Escape = l.tw.ToTown
		n.ClearCursor = l.tw.ClearCursor
		l.wpNav = n
	}
	return l.wpNav, l.wpSpy
}

func (l *lab) WView(ctx context.Context) (drill.WPView, error) {
	s, err := l.be.Env.Snapshot()
	if err != nil {
		return drill.WPView{}, err
	}
	if s.Player == nil {
		return drill.WPView{}, errors.New("no player in the snapshot")
	}
	v := drill.WPView{Area: s.Player.Area, X: s.Player.X, Y: s.Player.Y, Tick: l.tick()}
	for _, o := range s.Objects {
		if d2map.IsWaypointObject(o.Class) {
			v.Waypoints = append(v.Waypoints, [2]int{o.X, o.Y})
		}
	}
	if u, err := l.UI(ctx); err == nil {
		for _, p := range u.Panels {
			if p == "waypoint" {
				v.PanelOpen = true
			}
		}
	}
	return v, nil
}

func (l *lab) Goto(ctx context.Context, r *drill.Run, level int) error {
	// `debug goto` is an arrival for d2d (its 10 s window restarts) even into the same level, which a navigator that
	// watches the area id cannot see: a fresh one counts the window from its first look, as after a real level change.
	l.wpNav = nil
	if _, err := r.Cheat(ctx, fmt.Sprintf("debug goto %d", level), fmt.Sprintf("arrive on level %d at a free cell: the start of the portal and waypoint drills", level)); err != nil {
		return err
	}
	return l.Tick(ctx, 25)
}

func (l *lab) ActivateAll(ctx context.Context, r *drill.Run) error {
	_, err := r.Cheat(ctx, "debug wp all", "activate every waypoint of the throwaway character: the drill checks the bot's waypoint verb, not the activation")
	return err
}

func (l *lab) TakeWaypoint(ctx context.Context, level int) drill.WPCall {
	n, sp := l.wpNavigator()
	c0, mark, t0 := sp.clicks, l.rec.Mark(), l.tick()
	err := n.TakeWaypoint(ctx, level)
	c := drill.WPCall{Ticks: l.tick() - t0, Clicks: sp.clicks - c0, Log: l.rec.Since(mark)}
	if err != nil {
		c.Err = err.Error()
	}
	return c
}
