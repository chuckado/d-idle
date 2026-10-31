package drill

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

// WPView is the waypoint-relevant state (StateReader: player, objects, UI flag).
type WPView struct {
	Area, X, Y int
	Tick       int
	Waypoints  [][2]int // waypoint objects in the snapshot (subtiles)
	PanelOpen  bool
}

// WPCall is one call of the bot's waypoint verb (nav.Navigator.TakeWaypoint).
type WPCall struct {
	Err    string
	Ticks  int // virtual ticks from the call to its return
	Clicks int // clicks the verb sent
	Log    []string
}

// WPLab is the instance as the D3 drills use it. The verb is shared code (internal/nav), called as the bot calls it.
type WPLab interface {
	WView(ctx context.Context) (WPView, error)
	Tick(ctx context.Context, n int) error
	Goto(ctx context.Context, r *Run, level int) error
	ActivateAll(ctx context.Context, r *Run) error
	TakeWaypoint(ctx context.Context, level int) WPCall
}

const missingAct = "act 2 unlocked on the throwaway character (cheat `debug act 2` or the Andariel quest word): d2d lists only the current act's tab until then (V live: `dump ui` wp_tabs has one rect after `debug wp all`, clicks at the tab 2 and 3 positions change nothing), so a cross-act row cannot be chosen"

// Arrival is "x+3, y+3" of the waypoint (trace), about 4 subtiles away; 10 leaves room for the walk-off.
const wpArrivalMax = 10

// D3 returns the waypoint group (P23). Order matters: D3.b runs before anything activates a waypoint (the character
// is fresh per process).
func D3(l WPLab) []Drill {
	lim := 10 * time.Minute
	t := func(from, to int, id, title string) Drill {
		return Drill{ID: id, Primitive: "P23", Title: title, Limit: lim,
			Do: func(ctx context.Context, r *Run) Outcome { return d3travel(ctx, r, l, from, to) }}
	}
	return []Drill{
		{ID: "D3.b", Primitive: "P23", Title: "waypoint not activated: the verb fails gracefully (error, area unchanged, panel closed)", Limit: lim,
			Do: func(ctx context.Context, r *Run) Outcome { return d3inactive(ctx, r, l) }},
		t(1, 3, "D3.a1", "Rogue Encampment to Cold Plains (act 1 row 1)"),
		t(1, 35, "D3.a2", "Rogue Encampment to Catacombs level 2 (act 1 row 8)"),
		{ID: "D3.a3", Primitive: "P23", Title: "Rogue Encampment to Lut Gholein (cross act: tab 2 row 0)", Needs: []string{missingAct}},
		t(40, 46, "D3.a4", "Lut Gholein to Canyon of the Magi (act 2 row 8)"),
		t(40, 52, "D3.a5", "Lut Gholein to Palace Cellar level 1 (act 2 row 6)"),
		{ID: "D3.c", Primitive: "P23", Title: "second use right after an arrival: d2d refuses inside 10 s, the verb waits and succeeds", Limit: lim,
			Do: func(ctx context.Context, r *Run) Outcome { return d3second(ctx, r, l) }},
	}
}

func nearest(v WPView) (int, bool) {
	best, ok := 1<<30, false
	for _, w := range v.Waypoints {
		if d := int(math.Hypot(float64(w[0]-v.X), float64(w[1]-v.Y))); d < best {
			best, ok = d, true
		}
	}
	return best, ok
}

func d3start(ctx context.Context, r *Run, l WPLab, from int, activate bool) (WPView, error) {
	if err := l.Goto(ctx, r, from); err != nil {
		return WPView{}, err
	}
	if activate {
		if err := l.ActivateAll(ctx, r); err != nil {
			return WPView{}, err
		}
	}
	if err := l.Tick(ctx, 50); err != nil {
		return WPView{}, err
	}
	return l.WView(ctx)
}

func d3travel(ctx context.Context, r *Run, l WPLab, from, to int) Outcome {
	v0, err := d3start(ctx, r, l, from, true)
	if err != nil {
		return setupFail(err)
	}
	if v0.Area != from {
		return Blocks(fmt.Sprintf("`debug goto %d` did not arrive (area %d): the character cannot be put in that act", from, v0.Area))
	}
	c := l.TakeWaypoint(ctx, to)
	if err := l.Tick(ctx, 30); err != nil {
		return setupFail(err)
	}
	v, err := l.WView(ctx)
	if err != nil {
		return setupFail(err)
	}
	d, found := nearest(v)
	intended := fmt.Sprintf("area %d within the verb's call, the character beside the target waypoint (<= %d subtiles), the verb returns nil, no panel left open", to, wpArrivalMax)
	observed := fmt.Sprintf("from area %d: area %d, %d ticks, %d clicks, verb %s, nearest waypoint object %d subtiles (found %v), panel open %v; lines: %s",
		from, v.Area, c.Ticks, c.Clicks, errOrOK(c.Err), d, found, v.PanelOpen, strings.Join(c.Log, " / "))
	sig := fmt.Sprintf("from %d area %d verb %s near %v panel %v", from, v.Area, errOrOK(c.Err), found && d <= wpArrivalMax, v.PanelOpen)
	o := Outcome{Intended: intended, Observed: observed, Sig: sig}
	var bad []string
	if c.Err != "" {
		bad = append(bad, "verb failed: "+c.Err)
	}
	if v.Area != to {
		bad = append(bad, fmt.Sprintf("area %d, want %d", v.Area, to))
	}
	if !found || d > wpArrivalMax {
		bad = append(bad, fmt.Sprintf("arrival %d subtiles from the nearest waypoint (found %v)", d, found))
	}
	if len(bad) > 0 {
		o.Verdict, o.Reason = Fail, strings.Join(bad, "; ")
		return o
	}
	o.Verdict, o.Reason = Pass, fmt.Sprintf("verb travelled %d -> %d in %d ticks", from, to, c.Ticks)
	return o
}

func d3inactive(ctx context.Context, r *Run, l WPLab) Outcome {
	v0, err := d3start(ctx, r, l, 1, false)
	if err != nil {
		return setupFail(err)
	}
	c := l.TakeWaypoint(ctx, 3)
	if err := l.Tick(ctx, 30); err != nil {
		return setupFail(err)
	}
	v, err := l.WView(ctx)
	if err != nil {
		return setupFail(err)
	}
	intended := "Cold Plains never activated: the verb returns an error, the area stays 1, no panel is left open, bounded ticks"
	observed := fmt.Sprintf("area %d -> %d, verb %s, %d ticks, %d clicks, panel open %v; lines: %s", v0.Area, v.Area, errOrOK(c.Err), c.Ticks, c.Clicks, v.PanelOpen, strings.Join(c.Log, " / "))
	sig := fmt.Sprintf("area %d verb failed %v panel %v", v.Area, c.Err != "", v.PanelOpen)
	o := Outcome{Intended: intended, Observed: observed, Sig: sig}
	switch {
	case c.Err == "":
		o.Verdict, o.Reason = Fail, "the verb reported success for an inactive waypoint"
	case v.Area != 1 || v.PanelOpen:
		o.Verdict, o.Reason = Fail, "the verb left the area changed or the panel open"
	default:
		o.Verdict, o.Reason = Pass, "inactive waypoint: error, nothing changed"
	}
	return o
}

func d3second(ctx context.Context, r *Run, l WPLab) Outcome {
	if _, err := d3start(ctx, r, l, 1, true); err != nil {
		return setupFail(err)
	}
	c1 := l.TakeWaypoint(ctx, 3)
	if c1.Err != "" {
		return Failed("first travel failed: "+c1.Err, "", strings.Join(c1.Log, " / "))
	}
	c2 := l.TakeWaypoint(ctx, 4) // at once: d2d refuses inside 10 s of the arrival
	if err := l.Tick(ctx, 30); err != nil {
		return setupFail(err)
	}
	v, err := l.WView(ctx)
	if err != nil {
		return setupFail(err)
	}
	intended := "the immediate second use (Cold Plains to Stony Field) waits out d2d's 10 s window (>= 250 ticks) and arrives in 4"
	observed := fmt.Sprintf("second verb %s after %d ticks, %d clicks, area %d; lines: %s", errOrOK(c2.Err), c2.Ticks, c2.Clicks, v.Area, strings.Join(c2.Log, " / "))
	sig := fmt.Sprintf("second %s area %d waited %v", errOrOK(c2.Err), v.Area, c2.Ticks >= 250)
	o := Outcome{Intended: intended, Observed: observed, Sig: sig}
	switch {
	case c2.Err != "" || v.Area != 4:
		o.Verdict, o.Reason = Fail, "the second use did not arrive: "+c2.Err
	default:
		o.Verdict, o.Reason = Pass, fmt.Sprintf("second use arrived after %d ticks", c2.Ticks)
	}
	return o
}

func errOrOK(s string) string {
	if s == "" {
		return "ok"
	}
	return "error: " + s
}
