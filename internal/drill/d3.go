package drill

import (
	"cmp"
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

// WPLab is the instance as the DW drills use it. The verb is shared code (internal/nav), called as the bot calls it.
type WPLab interface {
	WView(ctx context.Context) (WPView, error)
	Tick(ctx context.Context, n int) error
	Goto(ctx context.Context, r *Run, level int) error
	ActivateAll(ctx context.Context, r *Run) error
	// UnlockAct runs `debug act N` (ledgered): the waypoint panel then lists the tabs 1..N.
	UnlockAct(ctx context.Context, r *Run, act int) error
	// Tabs is the panel's `wp_tabs` rects as `dump ui` showed them at the last tab click of TakeWaypoint ("" none).
	Tabs() string
	TakeWaypoint(ctx context.Context, level int) WPCall
}

// actOf is the act of a level id (Levels.txt: 1-39, 40-74, 75-102, 103-108, 109 and up).
func actOf(level int) int {
	switch {
	case level < 40:
		return 1
	case level < 75:
		return 2
	case level < 103:
		return 3
	case level < 109:
		return 4
	}
	return 5
}

// Arrival is "x+3, y+3" of the waypoint (trace), about 4 subtiles away; 10 leaves room for the walk-off.
const wpArrivalMax = 10

// DW returns the waypoint group (P23). Order matters: DW.b runs before anything activates a waypoint (the character
// is fresh per process).
func DW(l WPLab) []Drill {
	lim := 10 * time.Minute
	t := func(from, to int, id, title string) Drill {
		return Drill{ID: id, Primitive: "P23", Title: title, Limit: lim,
			Do: func(ctx context.Context, r *Run) Outcome { return d3travel(ctx, r, l, from, to) }}
	}
	return []Drill{
		{ID: "DW.b", Primitive: "P23", Title: "waypoint not activated: the verb fails gracefully (error, area unchanged, panel closed)", Limit: lim,
			Do: func(ctx context.Context, r *Run) Outcome { return d3inactive(ctx, r, l) }},
		t(1, 3, "DW.a1", "Rogue Encampment to Cold Plains (act 1 row 1)"),
		t(1, 35, "DW.a2", "Rogue Encampment to Catacombs level 2 (act 1 row 8)"),
		t(1, 40, "DW.a3", "Rogue Encampment to Lut Gholein (cross act: tab 2 row 0, act 2 unlocked with `debug act 2`)"),
		t(40, 3, "DW.a3b", "Lut Gholein to Cold Plains (act 2 area to an act 1 waypoint: tab 1 row 1)"),
		t(40, 46, "DW.a4", "Lut Gholein to Canyon of the Magi (act 2 row 8)"),
		t(40, 52, "DW.a5", "Lut Gholein to Palace Cellar level 1 (act 2 row 6)"),
		t(1, 75, "DW.a6", "Rogue Encampment to Kurast Docks (act 3 tab 3 row 0, `debug act 3`)"),
		t(1, 103, "DW.a7", "Rogue Encampment to the Pandemonium Fortress (act 4 tab 4 row 0, `debug act 4`)"),
		t(1, 109, "DW.a8", "Rogue Encampment to Harrogath (act 5 tab 5 row 0, `debug act 5`)"),
		t(75, 1, "DW.a9", "Kurast Docks to the Rogue Encampment (act 3 area to act 1 row 0)"),
		t(103, 1, "DW.a10", "Pandemonium Fortress to the Rogue Encampment (act 4 area to act 1 row 0)"),
		t(109, 1, "DW.a11", "Harrogath to the Rogue Encampment (act 5 area to act 1 row 0)"),
		{ID: "DW.c", Primitive: "P23", Title: "second use right after an arrival: d2d refuses inside 10 s, the verb waits and succeeds", Limit: lim,
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

func d3start(ctx context.Context, r *Run, l WPLab, from int, activate bool, act int) (WPView, error) {
	if err := l.Goto(ctx, r, from); err != nil {
		return WPView{}, err
	}
	if activate {
		if err := l.ActivateAll(ctx, r); err != nil {
			return WPView{}, err
		}
	}
	if act > 1 {
		if err := l.UnlockAct(ctx, r, act); err != nil {
			return WPView{}, err
		}
	}
	if err := l.Tick(ctx, 50); err != nil {
		return WPView{}, err
	}
	return l.WView(ctx)
}

func d3travel(ctx context.Context, r *Run, l WPLab, from, to int) Outcome {
	v0, err := d3start(ctx, r, l, from, true, max(actOf(from), actOf(to)))
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
	observed := fmt.Sprintf("from area %d: area %d, %d ticks, %d clicks, verb %s, nearest waypoint object %d subtiles (found %v; player %d,%d, waypoint objects %v), panel open %v, wp_tabs at the tab click %s; lines: %s",
		from, v.Area, c.Ticks, c.Clicks, errOrOK(c.Err), d, found, v.X, v.Y, v.Waypoints, v.PanelOpen, cmp.Or(l.Tabs(), "none"), strings.Join(c.Log, " / "))
	sig := fmt.Sprintf("from %d area %d verb %s near %v panel %v", from, v.Area, errOrOK(c.Err), found && d <= wpArrivalMax, v.PanelOpen)
	o := Outcome{Intended: intended, Observed: observed, Sig: sig}
	var bad []string
	if c.Err != "" {
		bad = append(bad, "verb failed: "+c.Err)
	}
	if v.Area != to {
		bad = append(bad, fmt.Sprintf("area %d, want %d", v.Area, to))
	}
	// Arrival in the Rogue Encampment is at the town's start point (141,103), 66 subtiles from its waypoint (V live DW.a9
	// to a11, identical from acts 3, 4 and 5); whether Game.exe arrives beside the waypoint there is UNVERIFIED, so the
	// area id is the pass condition for target 1 and the distance is only recorded.
	if (!found || d > wpArrivalMax) && to != 1 {
		bad = append(bad, fmt.Sprintf("arrival %d subtiles from the nearest waypoint (found %v)", d, found))
	}
	if len(bad) > 0 {
		o.Verdict, o.Reason = Fail, strings.Join(bad, "; ")
		return o
	}
	o.Verdict, o.Reason = Pass, fmt.Sprintf("verb travelled %d -> %d in %d ticks", from, to, c.Ticks)
	if to == 1 && (!found || d > wpArrivalMax) {
		o.Reason += fmt.Sprintf("; arrival at the town start, %d subtiles from the waypoint (not required for target 1)", d)
	}
	return o
}

func d3inactive(ctx context.Context, r *Run, l WPLab) Outcome {
	v0, err := d3start(ctx, r, l, 1, false, 1)
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
	if _, err := d3start(ctx, r, l, 1, true, 1); err != nil {
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
