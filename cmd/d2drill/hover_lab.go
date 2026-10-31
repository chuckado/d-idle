package main

import (
	"context"
	"strings"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/d2dsrc"
	"d2lod-bot/internal/drill"
	"d2lod-bot/internal/input"
)

// The hover audit drills (docs/probes/drills/hover-audit.md). Each replays what the bot's code does and reads the end
// state; the control is the verb that is known to work.

func (l *lab) ApproachKashya(ctx context.Context, r *drill.Run, wall bool) drill.ApproachInfo {
	var ai drill.ApproachInfo
	if err := l.Goto(ctx, r, 1); err != nil {
		ai.Err = "goto: " + err.Error()
		return ai
	}
	_ = l.be.Ctl.CloseAll()
	npcs, err := l.be.Client.NPCs(ctx)
	if err != nil {
		ai.Err = err.Error()
		return ai
	}
	idx := -1
	for _, n := range npcs.Items {
		if strings.EqualFold(n.Name, "Kashya") {
			idx = n.Index
		}
	}
	if idx < 0 {
		ai.Err = "no Kashya in npcs"
		return ai
	}
	savedP, savedN := d2dctl.Pause, d2dctl.Now
	if wall { // the bot's own wiring (fix-obj): Backend.Start / LateClock.Install put the game-time Pause and Now into d2dctl
		l.be.Clock.Install(ctx)
	}
	defer func() { d2dctl.Pause, d2dctl.Now = savedP, savedN }()
	t0, w0 := l.tick(), time.Now()
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	err = l.be.Client.ApproachNPC(cctx, idx, 6*time.Second)
	ai.Ticks, ai.WallMs = l.tick()-t0, int(time.Since(w0).Milliseconds())
	if err != nil {
		ai.Err = err.Error()
	}
	if u, uerr := l.UI(ctx); uerr == nil {
		for _, p := range u.Panels {
			if p == "npc-dialog" || p == "vendor" {
				ai.Opened = true
			}
		}
	}
	d2dctl.Pause, d2dctl.Now = savedP, savedN
	_ = l.CloseAll(ctx)
	return ai
}

func (l *lab) ClickObjectLikeBot(ctx context.Context, r *drill.Run) drill.ObjectClick {
	var oc drill.ObjectClick
	if err := l.Goto(ctx, r, 1); err != nil {
		oc.InteractErr = "goto: " + err.Error()
		return oc
	}
	_ = l.be.Ctl.CloseAll()
	s, err := l.be.Env.Snapshot()
	if err != nil {
		oc.InteractErr = err.Error()
		return oc
	}
	npcs, err := l.be.Client.NPCs(ctx)
	if err != nil {
		oc.InteractErr = err.Error()
		return oc
	}
	wpID := uint32(0)
	for _, n := range npcs.Items {
		if strings.EqualFold(n.Name, "Waypoint") {
			wpID = uint32(d2dsrc.NPCIDBase + n.Index)
		}
	}
	var wp *d2.Unit
	for i := range s.Objects {
		if wpID != 0 && s.Objects[i].ID == wpID {
			wp = &s.Objects[i]
		}
	}
	if wp == nil {
		oc.InteractErr = "no waypoint object in the snapshot"
		return oc
	}
	t0 := l.tick()
	// objectAdapter.click: an object that is not an npcs row has an id below d2dsrc.NPCIDBase (and type object)
	tg := input.Target{Type: 2, ID: wp.ID, Pos: input.Point{X: wp.X, Y: wp.Y}}
	if err := l.be.Ctl.Interact(tg); err != nil {
		oc.InteractErr = err.Error()
		if err := l.be.Ctl.ClickWorld(tg.Pos); err != nil {
			oc.InteractErr += "; ClickWorld: " + err.Error()
		}
	}
	_, _ = l.be.Client.Step(ctx, 150)
	if f, err := l.be.Client.UI(ctx); err == nil {
		oc.OpenedByBot = f.Waypoint
	}
	_ = l.CloseAll(ctx)
	// control: the npcs row of the same object through cmd interact (what d2dinput does for NPCIDBase ids)
	if id := wp.ID; id >= d2dsrc.NPCIDBase {
		_ = l.be.Ctl.InteractNPC(int(id - d2dsrc.NPCIDBase))
		_, _ = l.be.Client.Step(ctx, 200)
		if f, err := l.be.Client.UI(ctx); err == nil {
			oc.OpenedByCmd = f.Waypoint
		}
	}
	oc.Ticks = l.tick() - t0
	_ = l.CloseAll(ctx)
	return oc
}

func (l *lab) TakeGround(ctx context.Context, r *drill.Run, mode string) drill.GroundClick {
	gc := drill.GroundClick{Mode: mode}
	c := l.be.Client
	if err := l.Goto(ctx, r, 1); err != nil {
		return gc
	}
	_, _ = r.Cheat(ctx, "debug clearinv", "empty the pack so a picked ring is countable")
	if err := l.GiveGround(ctx, r, "rin"); err != nil {
		return gc
	}
	_, _ = c.Step(ctx, 30)
	g, err := c.Ground(ctx)
	if err != nil || len(g.Items) == 0 {
		return gc
	}
	it := g.Items[0]
	l0, _ := c.Level(ctx)
	t0 := l.tick()
	step := func(n int) { _, _ = c.Step(ctx, n) }
	switch mode {
	case "pointer", "pointer-render":
		render := mode == "pointer-render"
		_ = c.Hover(ctx, it.SX-3, it.SY)
		if render {
			_, _ = c.Do(ctx, "step 2 render")
		} else {
			step(3)
		}
		_ = c.Hover(ctx, it.SX, it.SY)
		if render {
			_, _ = c.Do(ctx, "step 2 render")
		} else {
			step(15)
		}
		_ = c.Click(ctx, it.SX, it.SY)
		step(120)
	case "pickup":
		_ = c.CmdPickup(ctx, it.Unit)
		step(120)
	}
	gc.Ticks = l.tick() - t0
	l1, _ := c.Level(ctx)
	gc.Moved = (l1.X-l0.X)*(l1.X-l0.X)+(l1.Y-l0.Y)*(l1.Y-l0.Y) > 0.25
	if v, err := l.TView(ctx); err == nil {
		gc.Picked = v.Pack["rin"] > 0
	}
	return gc
}

var _ drill.HoverLab = (*lab)(nil)
