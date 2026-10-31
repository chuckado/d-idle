package drill

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

// PortalView is the portal-relevant end state (StateReader: player, objects, pack).
type PortalView struct {
	Area, X, Y, Mode int
	InTown           bool
	Near, Any        int      // town portal objects (class 59) within 25 subtiles of the player / anywhere in the level
	PortalPos        [][2]int // their positions
	PortalIDs        []int    // their unit ids (a new cast makes a new object)
	Tome             int      // charges in the pack's tome (bot rule: no stat 70 = 0), -1 no tome
	TomeID           int
	Scrolls          int
	ScrollIDs        []int
	Tick             int
}

// BotCall is one call of a bot function (the portal verb, ToTown, BackToField) with what it sent to the game.
type BotCall struct {
	Err       string
	NoPortal  bool // errors.Is(err, combat.ErrNoPortal)
	Rights    int  // right clicks sent
	Clicks    int  // all clicks sent
	Keys      int  // key presses sent
	Ticks     int  // virtual ticks from the call to its return
	ClickTick int  // ticks from the call to the right click, -1 none
	Log       []string
}

// PortalLab is the instance as the D2 drills use it. The Run* methods call the bot's own code (a verbatim copy of
// cmd/d2bot's d2dPortalWith, d2dTown.ToTown and townAdapter.BackToField, kept in step by a test); the rest is setup.
type PortalLab interface {
	PView(ctx context.Context) (PortalView, error)
	Tick(ctx context.Context, n int) error
	ClearPack(ctx context.Context, r *Run) error
	GiveGround(ctx context.Context, r *Run, code string) error
	PickupRaw(ctx context.Context) (int, error)
	Goto(ctx context.Context, r *Run, level int) error
	PortalList(ctx context.Context) (string, error)

	RunPortal(ctx context.Context) BotCall
	// RunPortalHit runs the portal verb and sends `debug hit 5` (a logged cheat) delay ticks after its first right click.
	RunPortalHit(ctx context.Context, r *Run, delay int) BotCall
	RunToTown(ctx context.Context) BotCall
	RunBack(ctx context.Context, level int) BotCall
	// EnterViaVerb takes the open portal next to the player with d2dctl's EnterPortal (the verified walk and click on
	// the sprite). It is not what the bot calls; it separates "the portal exists" from "the bot's walk-in works".
	EnterViaVerb(ctx context.Context) error

	// RawCast is the harness's own cast of item id (not the bot's): ticks from the click to the portal object, -1 none.
	RawCast(ctx context.Context, id int) (int, error)
	CastSkill(ctx context.Context, skill int, dx, dy float64) error
	Walk(ctx context.Context, dx, dy float64) error
}

const (
	field      = 2  // Blood Moor
	townArea   = 1  // Rogue Encampment
	fireBolt   = 36 // a sorceress skill that puts the player in cast mode (V live: mode SC for ~10 ticks)
	modeSC     = 10 // PlrMode.txt order: SC (cast)
	traceTicks = 7  // docs/research/town-portal-trace.md: d2d casts about 7 ticks (sorceress, FCR 0)
)

// D2 returns the portal and town entry group: P10 (cast and enter), P11 (return), with the failure modes of fix-portal F4 and fix-portal2.
func D2(l PortalLab) []Drill {
	lim := 8 * time.Minute
	do := func(f func(context.Context, *Run, PortalLab) Outcome) func(context.Context, *Run) Outcome {
		return func(ctx context.Context, r *Run) Outcome { return f(ctx, r, l) }
	}
	return []Drill{
		{ID: "D2.a", Primitive: "P10", Title: "tome with charges and a scroll: portal verb from the field, arrive in town, one source spent, one portal", Limit: lim, Do: do(d2a)},
		{ID: "D2.t", Primitive: "P10", Title: "bot ToTown with a portal open next to the player enters it (arrive in town, no second scroll)", Limit: lim, Do: do(d2t)},
		{ID: "D2.a2", Primitive: "P10", Title: "ticks from the right click to portal object 59 (raw cast, 3 casts)", Limit: lim, Do: do(d2a2)},
		{ID: "D2.b", Primitive: "P10", Title: "empty tome, no scroll: no portal source, nothing clicked, no-portal error at once", Limit: lim, Do: do(d2b)},
		{ID: "D2.c", Primitive: "P10", Title: "empty tome and a loose scroll: the scroll is used", Limit: lim, Do: do(d2c)},
		{ID: "D2.g", Primitive: "P10", Title: "control: d2dctl EnterPortal takes the open portal (separates a bad bot walk-in from a bad portal)", Limit: lim, Do: do(d2g)},
		{ID: "D2.d", Primitive: "P10", Title: "hit in the same tick as the click: the click is dropped, source unchanged, the retry takes it", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome { return d2d(ctx, r, l, 0) }},
		{ID: "D2.d3", Primitive: "P10", Title: "hit 3 ticks into the cast: the cast is cut with the scroll spent, the retry takes the next source", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome { return d2d(ctx, r, l, 3) }},
		{ID: "D2.e", Primitive: "P10", Title: "busy player: the verb waits for idle, then clicks", Limit: lim, Do: do(d2e)},
		{ID: "D2.h", Primitive: "P10", Title: "one-pair rule: a second cast closes the first portal", Limit: lim, Do: do(d2h)},
		{ID: "D2.f", Primitive: "P11", Title: "round trip town and back, arrival near the origin, one portal pair", Limit: 12 * time.Minute, Do: do(d2f)},
	}
}

// goField arrives in the field and walks off any portal an earlier drill left near the arrival cell: a stale portal
// within 25 subtiles would make the verb's own success test (a portal object near the player) pass without a cast.
func goField(ctx context.Context, r *Run, l PortalLab) error {
	if err := l.Goto(ctx, r, field); err != nil {
		return err
	}
	return clearSpot(ctx, l)
}

// clearSpot walks away from a portal an earlier case left (the town portal object stays on the level): each direction is
// walked up to 4 times while a portal is within 25 subtiles, so a stale portal at the arrival cell is left behind
// whatever the previous case did (drills7: the full D2 group failed D2.a2 on one 8-cell step per direction).
func clearSpot(ctx context.Context, l PortalLab) error {
	for _, d := range [][2]float64{{8, 0}, {-8, 0}, {0, 8}, {0, -8}} {
		for i := 0; i < 4; i++ {
			v, err := l.PView(ctx)
			if err != nil {
				return err
			}
			if v.Near == 0 {
				return nil
			}
			if err := l.Walk(ctx, d[0], d[1]); err != nil {
				break
			}
		}
	}
	v, err := l.PView(ctx)
	if err != nil {
		return err
	}
	if v.Near != 0 {
		return fmt.Errorf("a stale portal stays within 25 subtiles after walking away")
	}
	return nil
}

func setupFail(err error) Outcome { return Failed("setup: "+err.Error(), "", "") }

// stage empties the pack and puts a tome and/or scrolls in it with the engine's own pickup.
func stage(ctx context.Context, r *Run, l PortalLab, tome bool, scrolls int) (PortalView, error) {
	if err := l.ClearPack(ctx, r); err != nil {
		return PortalView{}, err
	}
	// scrolls first and picked up alone: a scroll picked up while a tome is in the pack goes into the tome (V live,
	// D2.a first run: tome 1 + scroll became tome 2), so a loose scroll next to a tome needs this order
	for i := 0; i < scrolls; i++ {
		if err := l.GiveGround(ctx, r, "tsc"); err != nil {
			return PortalView{}, err
		}
	}
	if err := l.Tick(ctx, 30); err != nil {
		return PortalView{}, err
	}
	if _, err := l.PickupRaw(ctx); err != nil {
		return PortalView{}, err
	}
	if tome {
		if err := l.GiveGround(ctx, r, "tbk"); err != nil {
			return PortalView{}, err
		}
	}
	if err := l.Tick(ctx, 30); err != nil {
		return PortalView{}, err
	}
	if _, err := l.PickupRaw(ctx); err != nil {
		return PortalView{}, err
	}
	v, err := l.PView(ctx)
	if err != nil {
		return v, err
	}
	if (tome && v.Tome < 1) || (!tome && v.Tome >= 0) || v.Scrolls != scrolls {
		return v, fmt.Errorf("stage failed: tome %d (wanted %v), scrolls %d (wanted %d)", v.Tome, tome, v.Scrolls, scrolls)
	}
	return v, nil
}

// burn casts the tome with the harness's raw cast until it holds no charge.
func burn(ctx context.Context, l PortalLab, v PortalView) (PortalView, error) {
	for i := 0; i < 25 && v.Tome != 0; i++ {
		if v.TomeID == 0 {
			return v, fmt.Errorf("no tome to burn")
		}
		if err := clearSpot(ctx, l); err != nil {
			return v, err
		}
		if v, err := l.PView(ctx); err != nil || v.TomeID == 0 {
			return v, fmt.Errorf("tome gone while walking off a portal: %v", err)
		}
		if t, err := l.RawCast(ctx, v.TomeID); err != nil || t < 0 {
			return v, fmt.Errorf("burning the tome: cast gave ticks %d err %v", t, err)
		}
		if err := l.Tick(ctx, 10); err != nil {
			return v, err
		}
		var err error
		if v, err = l.PView(ctx); err != nil {
			return v, err
		}
	}
	if v.Tome != 0 {
		return v, fmt.Errorf("tome still holds %d charges", v.Tome)
	}
	return v, nil
}

func portalLines(c BotCall) string {
	var out []string
	for _, ln := range c.Log {
		s := strings.ToLower(ln)
		if strings.Contains(s, "portal") || strings.Contains(s, "back to field") || strings.Contains(s, "d2d town") {
			out = append(out, ln)
		}
	}
	if len(out) == 0 {
		return "none"
	}
	return strings.Join(out, " // ")
}

func errText(c BotCall) string {
	if c.Err == "" {
		return "ok"
	}
	return c.Err
}

func (v PortalView) pos() string { return fmt.Sprintf("(%d,%d)", v.X, v.Y) }

func dist(a, b PortalView) float64 { return math.Hypot(float64(a.X-b.X), float64(a.Y-b.Y)) }

func d2a(ctx context.Context, r *Run, l PortalLab) Outcome {
	v0, err := stage(ctx, r, l, true, 1)
	if err != nil {
		return setupFail(err)
	}
	if err := goField(ctx, r, l); err != nil {
		return setupFail(err)
	}
	v1, err := l.PView(ctx)
	if err != nil {
		return setupFail(err)
	}
	c := l.RunPortal(ctx)
	if err := l.Tick(ctx, 5); err != nil {
		return setupFail(err)
	}
	v2, err := l.PView(ctx)
	if err != nil {
		return setupFail(err)
	}
	list, _ := l.PortalList(ctx)
	spent := fmt.Sprintf("tome %d->%d, scrolls %d->%d", v0.Tome, v2.Tome, v0.Scrolls, v2.Scrolls)
	which := (v0.Tome - v2.Tome) + (v0.Scrolls - v2.Scrolls)
	intended := "from the field (area 2) the verb opens one portal object next to the player, sends one right click and spends exactly one source"
	observed := fmt.Sprintf("area %d; verb %s: %d right click, %d ticks (click at tick +%d), field portal objects near %d any %d (a stale one from the previous drill stands elsewhere: %d before), %s, debug portals [%s]; verb lines: %s",
		v2.Area, errText(c), c.Rights, c.Ticks, c.ClickTick, v2.Near, v2.Any, v1.Any, spent, list, portalLines(c))
	sig := fmt.Sprintf("area %d verb %s rights %d ticks %d click+%d near %d any %d spent %d", v2.Area, errText(c), c.Rights, c.Ticks, c.ClickTick, v2.Near, v2.Any, which)
	var bad []string
	if c.Err != "" {
		bad = append(bad, "portal verb failed: "+c.Err)
	}
	if v2.Near != 1 || v2.Any != 1 {
		bad = append(bad, fmt.Sprintf("portal objects after the cast: near %d any %d (want 1 and 1)", v2.Near, v2.Any))
	}
	if which != 1 {
		bad = append(bad, fmt.Sprintf("%d sources spent (want exactly 1)", which))
	}
	if c.Rights != 1 {
		bad = append(bad, fmt.Sprintf("%d right clicks (want 1)", c.Rights))
	}
	o := Outcome{Intended: intended, Observed: observed, Sig: sig}
	if len(bad) > 0 {
		o.Verdict, o.Reason = Fail, strings.Join(bad, "; ")
		return o
	}
	o.Verdict, o.Reason = Pass, fmt.Sprintf("verb returned after %d ticks with one portal pair and one source spent (%s)", c.Ticks, spent)
	return o
}

// d2t is the bot's ToTown with a portal open next to the player: it must arrive in town without a second cast.
func d2t(ctx context.Context, r *Run, l PortalLab) Outcome {
	v0, err := stage(ctx, r, l, false, 3)
	if err != nil {
		return setupFail(err)
	}
	if err := goField(ctx, r, l); err != nil {
		return setupFail(err)
	}
	c := l.RunPortal(ctx)
	if c.Err != "" {
		return Failed("setup: the portal verb failed: "+c.Err, "", "")
	}
	_ = l.Tick(ctx, 5)
	t := l.RunToTown(ctx)
	v, err := l.PView(ctx)
	if err != nil {
		return setupFail(err)
	}
	list, _ := l.PortalList(ctx)
	intended := "d2dTown.ToTown with the portal open next to the player enters it: area 1 with no further scroll spent"
	observed := fmt.Sprintf("scrolls %d -> %d; ToTown %s after %d ticks, %d right clicks through the spy (the town package's own casts are not spied), area %d, debug portals [%s]; lines: %s",
		v0.Scrolls, v.Scrolls, errText(t), t.Ticks, t.Rights, v.Area, list, portalLines(t))
	o := Outcome{Intended: intended, Observed: observed, Sig: fmt.Sprintf("%s|%d|%d", errText(t), v.Area, v.Scrolls)}
	switch {
	case v.Area != townArea || t.Err != "":
		o.Verdict, o.Reason = Fail, fmt.Sprintf("ToTown did not arrive in town: err %q, area %d, scrolls %d of 3 left (the verb used 1)", t.Err, v.Area, v.Scrolls)
	case v.Scrolls != v0.Scrolls-1:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("arrived but spent %d scrolls (want 1)", v0.Scrolls-v.Scrolls)
	default:
		o.Verdict, o.Reason = Pass, "entered the open portal, arrived in town, no second scroll"
	}
	return o
}

func d2a2(ctx context.Context, r *Run, l PortalLab) Outcome {
	v0, err := stage(ctx, r, l, false, 3)
	if err != nil {
		return setupFail(err)
	}
	if err := goField(ctx, r, l); err != nil {
		return setupFail(err)
	}
	var ts []int
	v := v0
	for i := 0; i < 3; i++ {
		if err := clearSpot(ctx, l); err != nil {
			return setupFail(err)
		}
		if v, err = l.PView(ctx); err != nil {
			return setupFail(err)
		}
		if len(v.ScrollIDs) == 0 {
			return Failed("setup: no scroll left for cast "+fmt.Sprint(i+1), "", "")
		}
		t, err := l.RawCast(ctx, v.ScrollIDs[0])
		if err != nil {
			return setupFail(err)
		}
		ts = append(ts, t)
		if err := l.Tick(ctx, 10); err != nil {
			return setupFail(err)
		}
		if v, err = l.PView(ctx); err != nil {
			return setupFail(err)
		}
	}
	intended := fmt.Sprintf("d2d casts: object 59 about %d ticks (280 ms, sorceress FCR 0) after the right click (docs/research/town-portal-trace.md); game.exe would be 0", traceTicks)
	observed := fmt.Sprintf("ticks from click to object, 3 casts: %v; scrolls left %d; objects near %d any %d (one pair)", ts, v.Scrolls, v.Near, v.Any)
	o := Outcome{Intended: intended, Observed: observed}
	for _, t := range ts {
		if t < 1 || t > 12 {
			o.Verdict, o.Reason = Fail, fmt.Sprintf("a cast took %d ticks (want about %d, at most 12), or no object appeared (-1)", t, traceTicks)
			return o
		}
	}
	if v.Any != 1 {
		o.Verdict, o.Reason = Fail, fmt.Sprintf("%d portal objects after 3 casts (one pair rule wants 1)", v.Any)
		return o
	}
	o.Verdict, o.Reason = Pass, fmt.Sprintf("observed %v ticks vs the trace's about %d", ts, traceTicks)
	return o
}

func d2b(ctx context.Context, r *Run, l PortalLab) Outcome {
	v, err := stage(ctx, r, l, true, 0)
	if err != nil {
		return setupFail(err)
	}
	if err := goField(ctx, r, l); err != nil {
		return setupFail(err)
	}
	if v, err = burn(ctx, l, v); err != nil {
		return setupFail(err)
	}
	c := l.RunPortal(ctx)
	v2, err := l.PView(ctx)
	if err != nil {
		return setupFail(err)
	}
	intended := "tome at 0 charges and no scroll: the verb finds no source, sends no click and no key, returns the no-portal error with no ticks spent (fix-portal F4)"
	observed := fmt.Sprintf("tome %d, scrolls %d; verb error %q (no-portal %v), clicks %d, keys %d, ticks %d; tome after %d; lines: %s", v.Tome, v.Scrolls, errText(c), c.NoPortal, c.Clicks, c.Keys, c.Ticks, v2.Tome, portalLines(c))
	o := Outcome{Intended: intended, Observed: observed}
	switch {
	case !c.NoPortal:
		o.Verdict, o.Reason = Fail, "the verb did not return the no-portal error: "+errText(c)
	case c.Clicks != 0 || c.Keys != 0:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("the verb sent %d clicks and %d keys with nothing to cast", c.Clicks, c.Keys)
	case c.Ticks != 0:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("the no-portal answer took %d ticks, not immediate", c.Ticks)
	default:
		o.Verdict, o.Reason = Pass, "no portal source: no click, no key, error at tick 0"
	}
	return o
}

func d2c(ctx context.Context, r *Run, l PortalLab) Outcome {
	v, err := stage(ctx, r, l, true, 1)
	if err != nil {
		return setupFail(err)
	}
	if err := goField(ctx, r, l); err != nil {
		return setupFail(err)
	}
	if v, err = burn(ctx, l, v); err != nil {
		return setupFail(err)
	}
	if v.Scrolls != 1 {
		return Failed(fmt.Sprintf("setup: burning the tome used the scroll (scrolls %d)", v.Scrolls), "", "")
	}
	c := l.RunPortal(ctx)
	if err := l.Tick(ctx, 5); err != nil {
		return setupFail(err)
	}
	v2, err := l.PView(ctx)
	if err != nil {
		return setupFail(err)
	}
	intended := "tome at 0 plus one scroll: the verb skips the tome, casts the scroll once, portal object near the player, tome untouched"
	observed := fmt.Sprintf("before: tome %d, scrolls %d; verb %s, %d right click, %d ticks; after: tome %d, scrolls %d, portal objects near %d any %d; lines: %s",
		v.Tome, v.Scrolls, errText(c), c.Rights, c.Ticks, v2.Tome, v2.Scrolls, v2.Near, v2.Any, portalLines(c))
	o := Outcome{Intended: intended, Observed: observed}
	switch {
	case c.Err != "":
		o.Verdict, o.Reason = Fail, "the verb failed with an empty tome and a scroll in the pack: "+c.Err
	case v2.Scrolls != 0 || v2.Tome != 0:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("wrong source: scrolls %d, tome %d after (want 0 and 0)", v2.Scrolls, v2.Tome)
	case v2.Near != 1 || v2.Any != 1:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("portal objects near %d any %d (want 1, 1)", v2.Near, v2.Any)
	case c.Rights != 1:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("%d right clicks (want 1)", c.Rights)
	default:
		o.Verdict, o.Reason = Pass, "scroll used, empty tome never clicked"
	}
	return o
}

func d2e(ctx context.Context, r *Run, l PortalLab) Outcome {
	if _, err := stage(ctx, r, l, false, 2); err != nil {
		return setupFail(err)
	}
	if err := goField(ctx, r, l); err != nil {
		return setupFail(err)
	}
	idle := l.RunPortal(ctx) // baseline: the click tick of an idle player
	if err := l.Tick(ctx, 30); err != nil {
		return setupFail(err)
	}
	if err := l.CastSkill(ctx, fireBolt, 3, 0); err != nil {
		return setupFail(err)
	}
	if err := l.Tick(ctx, 1); err != nil {
		return setupFail(err)
	}
	vb, err := l.PView(ctx)
	if err != nil {
		return setupFail(err)
	}
	busy := l.RunPortal(ctx)
	if err := l.Tick(ctx, 5); err != nil {
		return setupFail(err)
	}
	v2, err := l.PView(ctx)
	if err != nil {
		return setupFail(err)
	}
	intended := "player in cast mode when the verb is called: it waits until idle, then clicks once and the portal opens (fix-portal2); the click tick is later than the idle baseline"
	observed := fmt.Sprintf("player mode at call %d (SC=%d); idle baseline: %s, click at tick +%d; busy call: %s, %d right clicks, click at tick +%d, %d ticks total; scrolls left %d, portal objects near %d any %d; lines: %s",
		vb.Mode, modeSC, errText(idle), idle.ClickTick, errText(busy), busy.Rights, busy.ClickTick, busy.Ticks, v2.Scrolls, v2.Near, v2.Any, portalLines(busy))
	o := Outcome{Intended: intended, Observed: observed}
	switch {
	case idle.Err != "":
		o.Verdict, o.Reason = Fail, "setup: the idle baseline cast failed: "+idle.Err
	case vb.Mode != modeSC:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("setup: the player was not in cast mode at the call (mode %d)", vb.Mode)
	case busy.Err != "":
		o.Verdict, o.Reason = Fail, "the verb failed on a busy player: "+busy.Err
	case busy.Rights != 1 || v2.Near != 1 || v2.Scrolls != 0:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("%d right clicks, portal near %d, scrolls left %d (want 1, 1, 0)", busy.Rights, v2.Near, v2.Scrolls)
	case busy.ClickTick <= idle.ClickTick:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("click at +%d not later than the idle baseline +%d: the verb did not wait for idle", busy.ClickTick, idle.ClickTick)
	default:
		o.Verdict, o.Reason = Pass, fmt.Sprintf("waited: click at +%d vs idle +%d ticks, one click, portal open", busy.ClickTick, idle.ClickTick)
	}
	return o
}

func d2g(ctx context.Context, r *Run, l PortalLab) Outcome {
	if _, err := stage(ctx, r, l, false, 1); err != nil {
		return setupFail(err)
	}
	if err := goField(ctx, r, l); err != nil {
		return setupFail(err)
	}
	c := l.RunPortal(ctx)
	if c.Err != "" {
		return Failed("setup: the portal verb failed: "+c.Err, "", "")
	}
	t0, _ := l.PView(ctx)
	verr := l.EnterViaVerb(ctx)
	_ = l.Tick(ctx, 10)
	v, _ := l.PView(ctx)
	intended := "the d2dctl EnterPortal verb (walk to the portal, click its sprite) takes the open portal: area 1"
	observed := fmt.Sprintf("area %d -> %d, verb error %v, portal objects in town %d", t0.Area, v.Area, verr, v.Any)
	if verr != nil || v.Area != townArea {
		return Failed(fmt.Sprintf("the verified walk-in did not arrive: err %v, area %d", verr, v.Area), intended, observed)
	}
	return Passed("EnterPortal arrives in town", intended, observed)
}

// d2h is the one-pair rule: a second cast while the first portal is open closes the first (trace: one pair per player).
func d2h(ctx context.Context, r *Run, l PortalLab) Outcome {
	v0, err := stage(ctx, r, l, false, 3)
	if err != nil {
		return setupFail(err)
	}
	if err := goField(ctx, r, l); err != nil {
		return setupFail(err)
	}
	c1 := l.RunPortal(ctx)
	_ = l.Tick(ctx, 5)
	v1, _ := l.PView(ctx)
	walkErr := l.Walk(ctx, 12, 0)
	if walkErr != nil {
		walkErr = l.Walk(ctx, -12, 0)
	}
	vw, _ := l.PView(ctx)
	c2 := l.RunPortal(ctx)
	_ = l.Tick(ctx, 5)
	v2, _ := l.PView(ctx)
	list, _ := l.PortalList(ctx)
	moved := 0.0
	if len(v1.PortalPos) > 0 && len(v2.PortalPos) > 0 {
		moved = math.Hypot(float64(v1.PortalPos[0][0]-v2.PortalPos[0][0]), float64(v1.PortalPos[0][1]-v2.PortalPos[0][1]))
	}
	intended := "cast, walk 12 cells, cast again with the first portal still open: exactly one portal object in the level, at the new place, both casts spent a scroll"
	observed := fmt.Sprintf("cast1 %s, objects near %d any %d at %v; walked to %s (%v), %.0f subtiles; cast2 %s, objects near %d any %d at %v, %.0f subtiles from the first; scrolls %d -> %d; debug portals [%s]; lines: %s / %s",
		errText(c1), v1.Near, v1.Any, v1.PortalPos, vw.pos(), walkErr, dist(vw, v1), errText(c2), v2.Near, v2.Any, v2.PortalPos, moved, v0.Scrolls, v2.Scrolls, list, portalLines(c1), portalLines(c2))
	o := Outcome{Intended: intended, Observed: observed, Sig: fmt.Sprintf("%v %v %d %d %d %d %.0f %d %s", c1.Err, c2.Err, v1.Near, v1.Any, v2.Near, v2.Any, moved, v2.Scrolls, list)}
	switch {
	case c1.Err != "" || c2.Err != "" || walkErr != nil:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("cast1 %q, walk %v, cast2 %q", c1.Err, walkErr, c2.Err)
	case v2.Any != 1 || v2.Near != 1 || moved < 20:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("objects near %d any %d, new portal %.0f subtiles from the old (want 1, 1, >= 20)", v2.Near, v2.Any, moved)
	case v0.Scrolls-v2.Scrolls != 2:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("%d scrolls spent for 2 casts", v0.Scrolls-v2.Scrolls)
	default:
		o.Verdict, o.Reason = Pass, fmt.Sprintf("second cast closed the first: one object, %.0f subtiles away", moved)
	}
	return o
}

const missingWalkIn = "a portal walk-in that works in the lab: the bot's ToTown (D2.t) and d2dctl EnterPortal (D2.g) both fail to take an open portal under lockstep; needs the proposed `cmd use <portal>` verb (docs/d2d-town.md) or a verified click point"

// d2f is the round trip: cast, ToTown, BackToField through the town portal, arrival near the origin.
func d2f(ctx context.Context, r *Run, l PortalLab) Outcome {
	if _, err := stage(ctx, r, l, false, 6); err != nil {
		return setupFail(err)
	}
	if err := goField(ctx, r, l); err != nil {
		return setupFail(err)
	}
	origin, err := l.PView(ctx)
	if err != nil {
		return setupFail(err)
	}
	if c := l.RunPortal(ctx); c.Err != "" {
		return Failed("setup: the portal verb failed: "+c.Err, "", "")
	}
	_ = l.Tick(ctx, 5)
	c2 := l.RunToTown(ctx)
	v2, _ := l.PView(ctx)
	if v2.Area != townArea {
		if err := l.EnterViaVerb(ctx); err != nil {
			return Blocks("missing: " + missingWalkIn + fmt.Sprintf(" (bot ToTown: %s; EnterPortal: %v)", errText(c2), err))
		}
		_ = l.Tick(ctx, 10)
		if v2, _ = l.PView(ctx); v2.Area != townArea {
			return Blocks("missing: " + missingWalkIn)
		}
	}
	c3 := l.RunBack(ctx, field)
	_ = l.Tick(ctx, 10)
	v3, _ := l.PView(ctx)
	list3, _ := l.PortalList(ctx)
	arrival := dist(v3, origin)
	viaPortal := strings.Contains(portalLines(c3), "back to field: portal in town") && !strings.Contains(strings.Join(c3.Log, " // "), "did not lead")
	intended := "after ToTown, the bot's BackToField returns by the town portal into area 2 within 25 subtiles of the origin"
	observed := fmt.Sprintf("origin %s; ToTown %s -> area %d; BackToField %s -> area %d at %s, %.0f subtiles from the origin, via portal %v, debug portals [%s]; lines: %s",
		origin.pos(), errText(c2), v2.Area, errText(c3), v3.Area, v3.pos(), arrival, viaPortal, list3, portalLines(c3))
	o := Outcome{Intended: intended, Observed: observed, Sig: fmt.Sprintf("%s|%s|%d|%.0f|%v", errText(c2), errText(c3), v3.Area, arrival, viaPortal)}
	var bad []string
	if c2.Err != "" {
		bad = append(bad, "the bot's ToTown did not enter the portal ("+errText(c2)+")")
	}
	if v3.Area != field || !viaPortal {
		bad = append(bad, fmt.Sprintf("BackToField did not return by the portal (area %d, err %s)", v3.Area, errText(c3)))
	}
	if arrival > 25 {
		bad = append(bad, fmt.Sprintf("arrival %.0f subtiles from the origin (want <= 25)", arrival))
	}
	if len(bad) > 0 {
		o.Verdict, o.Reason = Fail, strings.Join(bad, "; ")
		return o
	}
	o.Verdict, o.Reason = Pass, fmt.Sprintf("round trip through the portals, arrival %.0f subtiles from the origin", arrival)
	return o
}

// d2d: two scrolls and a hit delay ticks after the verb's click. A click without a portal is a dropped or cut cast: with
// delay 0 the hit is handled before the click (dropped, source unchanged), with delay 3 the cast has begun (cut, source
// spent, fix-portal F1/F4); either way d2dPortalWith tries again with the next source.
func d2d(ctx context.Context, r *Run, l PortalLab, delay int) Outcome {
	if _, err := stage(ctx, r, l, false, 2); err != nil {
		return setupFail(err)
	}
	if err := goField(ctx, r, l); err != nil {
		return setupFail(err)
	}
	c := l.RunPortalHit(ctx, r, delay)
	if err := l.Tick(ctx, 40); err != nil {
		return setupFail(err)
	}
	v, err := l.PView(ctx)
	if err != nil {
		return setupFail(err)
	}
	wantLeft := 1 // dropped click: the first scroll was not spent
	what := "the click is dropped (source unchanged), the retry opens the portal with the same scroll"
	if delay > 0 {
		wantLeft, what = 0, "the hit cuts the cast (scroll spent, no portal), the retry opens the portal with the next scroll"
	}
	observed := fmt.Sprintf("verb: %s, %d right clicks, %d ticks; scrolls left %d, portal objects near %d any %d; lines: %s", errText(c), c.Rights, c.Ticks, v.Scrolls, v.Near, v.Any, portalLines(c))
	o := Outcome{Intended: what, Observed: observed, Sig: fmt.Sprintf("err=%v rights=%d scrolls=%d near=%d", c.Err != "", c.Rights, v.Scrolls, v.Near)}
	switch {
	case c.Rights < 2 && v.Near == 0:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("one click, no portal and the verb stopped (%s): the failed cast was not retried", errText(c))
	case c.Rights < 2:
		o.Verdict, o.Reason = Blocked, "the hit did not disturb the cast (one click, a portal opened)"
	case c.Err != "" || v.Near != 1:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("retried (%d clicks) but ended with err %q and %d portals near", c.Rights, c.Err, v.Near)
	case v.Scrolls != wantLeft:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("%d scrolls left, want %d: the source accounting differs from the %s case", v.Scrolls, wantLeft, map[bool]string{true: "cut cast", false: "dropped click"}[delay > 0])
	default:
		o.Verdict, o.Reason = Pass, fmt.Sprintf("%d clicks, %d scrolls left, one portal", c.Rights, v.Scrolls)
	}
	return o
}
