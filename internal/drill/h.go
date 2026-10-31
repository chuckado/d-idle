package drill

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ApproachInfo is one d2dctl NPC approach (ApproachNPC, what the merc claim and revive code calls directly).
type ApproachInfo struct {
	Err    string
	Opened bool
	Ticks  int
	WallMs int
}

// ObjectClick is what the bot's object click (cmd/d2bot objectAdapter.click: Interact, then ClickWorld on error) did to a
// town waypoint, and what the client's own command did.
type ObjectClick struct {
	InteractErr string // the controller's answer to Interact on an object that is not an npcs row
	OpenedByBot bool   // waypoint panel after Interact failed and ClickWorld (cmd move) walked onto the object
	OpenedByCmd bool   // waypoint panel after `cmd interact <npcs index>` (the control for the same object)
	Ticks       int
}

// GroundClick is one try at taking a ground item: by pointer (hover, click) with the given frame mode, or by cmd pickup.
type GroundClick struct {
	Mode   string
	Picked bool
	Moved  bool // the player walked instead
	Ticks  int
}

// HoverLab is the instance as the hover audit drills use it.
type HoverLab interface {
	TownLab
	// ApproachKashya runs d2dctl.ApproachNPC to Kashya with d2dctl.Pause as the bot has it (wall time) or as the lab
	// has it (virtual ticks), after the logged cheat goto to the arrival cell.
	ApproachKashya(ctx context.Context, r *Run, wall bool) ApproachInfo
	// ClickObjectLikeBot replays objectAdapter.click on the town waypoint, then the control `cmd interact`.
	ClickObjectLikeBot(ctx context.Context, r *Run) ObjectClick
	// TakeGround gives a ring at the feet and takes it with mode "pointer", "pointer-render" or "pickup".
	TakeGround(ctx context.Context, r *Run, mode string) GroundClick
}

// H returns the hover blind-spot drills (docs/probes/drills/hover-audit.md): which bot verbs lean on the pointer pick
// that a lockstep d2d only computes while it draws a frame.
func H(l HoverLab) []Drill {
	lim := 6 * time.Minute
	return []Drill{
		{ID: "H.pause", Primitive: "AUDIT", Title: "NPC approach as the merc claim and revive code call it: d2dctl.Pause on wall time vs virtual ticks", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome { return hPause(ctx, r, l) }},
		{ID: "H.obj", Primitive: "AUDIT", Title: "object click as the bot does it (Interact, then ClickWorld): does a waypoint open", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome { return hObj(ctx, r, l) }},
		{ID: "H.ground", Primitive: "AUDIT", Title: "ground item by pointer (hover+click, with and without a rendered frame) vs cmd pickup", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome { return hGround(ctx, r, l) }},
		{ID: "H.monster", Primitive: "AUDIT", Title: "attack by pointer vs cmd skill with a unit id", Needs: []string{"a monster to fight: no debug verb spawns or places one (debug help lists none); the bot attacks with `cmd skill ... <unit> left`, which carries the unit id and needs no pointer pick (code), soaks kill under lockstep"}},
	}
}

func hPause(ctx context.Context, r *Run, l HoverLab) Outcome {
	virt := l.ApproachKashya(ctx, r, false)
	wall := l.ApproachKashya(ctx, r, true)
	obs := fmt.Sprintf("virtual-tick Pause: opened=%v err=%q ticks=%d wall=%dms | wall-time Pause (as cmd/d2bot): opened=%v err=%q ticks=%d wall=%dms", virt.Opened, virt.Err, virt.Ticks, virt.WallMs, wall.Opened, wall.Err, wall.Ticks, wall.WallMs)
	o := Outcome{Intended: "the approach opens Kashya's menu with either Pause", Observed: obs,
		Sig: fmt.Sprintf("virt=%v wall=%v", virt.Opened, wall.Opened)}
	switch {
	case !virt.Opened:
		o.Verdict, o.Reason = Fail, "control failed: the approach does not open Kashya even with the lab's stepping Pause"
	case !wall.Opened:
		o.Verdict, o.Reason = Fail, "the bot's own Pause (LateClock.Install, game time) does not open Kashya's menu under lockstep: fix-obj did not hold"
	default:
		o.Verdict, o.Reason = Pass, "the approach opens Kashya's menu with the bot's own Pause (a game-time deadline under lockstep), as with the lab's stepping Pause"
	}
	return o
}

func hObj(ctx context.Context, r *Run, l HoverLab) Outcome {
	c := l.ClickObjectLikeBot(ctx, r)
	obs := fmt.Sprintf("Interact(object) err %q; waypoint open after the ClickWorld fallback (cmd move onto it): %v; after `cmd interact <index>`: %v; %d ticks", c.InteractErr, c.OpenedByBot, c.OpenedByCmd, c.Ticks)
	o := Outcome{Intended: "the bot's object click opens the waypoint", Observed: obs, Sig: fmt.Sprintf("err=%v bot=%v cmd=%v", c.InteractErr != "", c.OpenedByBot, c.OpenedByCmd)}
	switch {
	case !c.OpenedByCmd:
		o.Verdict, o.Reason = Fail, "control failed: cmd interact did not open the waypoint either"
	case !c.OpenedByBot:
		o.Verdict, o.Reason = Fail, "RISK PROVEN: the bot's object click (Interact unsupported, then cmd move onto the object) does not operate it; only `cmd interact <index>` does"
	default:
		o.Verdict, o.Reason = Pass, "the bot's object click opens the waypoint"
	}
	return o
}

func hGround(ctx context.Context, r *Run, l HoverLab) Outcome {
	var parts, sig []string
	var ptr, render, pick GroundClick
	for _, m := range []string{"pointer", "pointer-render", "pickup"} {
		g := l.TakeGround(ctx, r, m)
		parts = append(parts, fmt.Sprintf("%s: picked=%v walked=%v (%d ticks)", g.Mode, g.Picked, g.Moved, g.Ticks))
		sig = append(sig, fmt.Sprintf("%s:%v", g.Mode, g.Picked))
		switch m {
		case "pointer":
			ptr = g
		case "pointer-render":
			render = g
		default:
			pick = g
		}
	}
	o := Outcome{Intended: "cmd pickup takes the item; a pointer click does only after a rendered frame", Observed: strings.Join(parts, " | "), Sig: strings.Join(sig, ";")}
	switch {
	case !pick.Picked:
		o.Verdict, o.Reason = Fail, "control failed: cmd pickup did not take the item"
	case ptr.Picked:
		o.Verdict, o.Reason = Pass, "a pointer click also picks up under lockstep (no blind spot for ground items)"
	default:
		o.Verdict, o.Reason = Pass, fmt.Sprintf("blind spot confirmed for pointer clicks (picked=%v without a frame, %v after a rendered frame); the bot uses cmd pickup, so it is not affected", ptr.Picked, render.Picked)
	}
	return o
}
