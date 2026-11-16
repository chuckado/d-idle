package bot

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
)

type fakeFreezer struct {
	mu    sync.Mutex
	can   bool
	err   error
	calls []bool
}

func (f *fakeFreezer) CanFreeze(context.Context) bool { return f.can }
func (f *fakeFreezer) Freeze(_ context.Context, on bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	f.calls = append(f.calls, on)
	return nil
}
func (f *fakeFreezer) log() []bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]bool(nil), f.calls...)
}

// fclock is a virtual clock that stands still between SetFrozen(true) and (false) like d2denv.LateClock; pass moves the
// underlying time (the wall) even while frozen.
type fclock struct {
	vclock
	frozen bool
	at     time.Time
	off    time.Duration
}

func newFClock() *fclock { return &fclock{vclock: *newClock()} }

func (c *fclock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.frozen {
		return c.at.Add(-c.off)
	}
	return c.now.Add(-c.off)
}
func (c *fclock) Sleep(d time.Duration) { c.pass(d); time.Sleep(50 * time.Microsecond) }
func (c *fclock) pass(d time.Duration)  { c.mu.Lock(); c.now = c.now.Add(d); c.mu.Unlock() }
func (c *fclock) SetFrozen(on bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch {
	case on && !c.frozen:
		c.frozen, c.at = true, c.now
	case !on && c.frozen:
		c.frozen, c.off = false, c.off+c.now.Sub(c.at)
	}
}

type wallClock struct {
	mu sync.Mutex
	t  time.Time
}

func (w *wallClock) now() time.Time { w.mu.Lock(); defer w.mu.Unlock(); return w.t }
func (w *wallClock) add(d time.Duration) {
	w.mu.Lock()
	w.t = w.t.Add(d)
	w.mu.Unlock()
}

// freezeRig is a pause rig with a freezer, a freezable env clock and a scripted wall clock.
func freezeRig(t *testing.T, class d2.Class, area int, fz *fakeFreezer) (*rig, *fclock, *wallClock) {
	t.Helper()
	fc, wc := newFClock(), &wallClock{t: time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)}
	g := newRig(t, func(c *Config, d *Deps, tu *env.Tuning) { d.Freezer = fz })
	g.r.e.Clock = fc
	g.r.wall = wc.now
	g.w.set(func(w *world) { w.class, w.area = class, area })
	g.mover.exploreFn = blockExplore
	return g, fc, wc
}

// The frozen pause stops the game, parks the loop mid-explore with no further game call, and Resume goes on with the plan.
func TestFreezeStopsTheLoopAndResumes(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		fz := &fakeFreezer{can: true}
		g, fc, _ := freezeRig(t, class, 2, fz)
		g.run()
		g.r.Start(reach(10))
		g.waitFor("exploring", func() bool { return g.w.count("explore") > 0 })
		v := g.r.PlayerPause(true)
		if !v.Frozen || v.State != PausePaused || !v.CanFreeze || v.Error != "" {
			t.Fatalf("%v: view %+v", class, v)
		}
		g.waitFor("parked", func() bool { return g.r.State() == StatePaused })
		time.Sleep(20 * time.Millisecond)
		calls, snaps := len(g.w.calls), g.w.snaps.Load()
		time.Sleep(60 * time.Millisecond)
		if len(g.w.calls) != calls || g.w.snaps.Load() != snaps {
			t.Fatalf("%v: the loop touched the frozen game (calls %d->%d, snaps %d->%d)", class, calls, len(g.w.calls), snaps, g.w.snaps.Load())
		}
		if !fc.frozen || len(fz.log()) != 1 || !fz.log()[0] {
			t.Fatalf("%v: clock frozen %v, freezer %v", class, fc.frozen, fz.log())
		}
		before := g.w.count("explore")
		v = g.r.PlayerPause(false)
		if v.Frozen || v.State != PauseRunning || v.Ended == nil || v.Ended.Auto {
			t.Fatalf("%v: after resume %+v", class, v)
		}
		g.waitFor("plan continues", func() bool { return g.r.State() != StatePaused && g.w.count("explore") > before })
		if fc.frozen || fz.log()[1] {
			t.Fatalf("%v: not thawed: %v", class, fz.log())
		}
		g.finish()
	}
}

// Hours of wall time pass while frozen: the stall ladder, the stuck watchdog and the drought clock stay quiet on resume,
// on a clock that stands still AND on one that does not (the resume re-arms the progress clocks itself).
func TestFreezeTimersDoNotFire(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		for _, still := range []bool{true, false} {
			fz := &fakeFreezer{can: true}
			g, fc, _ := freezeRig(t, class, 2, fz)
			if !still {
				g.r.e.Clock = newClock() // no SetFrozen: the clock runs through the pause
			}
			g.r.d.Cycler = fakeCycler{g.w}
			g.r.e.Tune.Bot.StallWarnMin, g.r.e.Tune.Bot.StallExitMin = 30, 30
			g.r.cfg.StuckAfterSec, g.r.cfg.GlobalStuckSec = 120, 300
			g.r.poll()
			g.r.campaignStall(context.Background(), g.r.latest(), true)
			g.r.Start(reach(10))
			g.r.setState(StateClearing, "fight") // a state the stuck watchdog counts
			g.r.obMu.Lock()
			g.r.ob.lastProgress = g.r.e.Clock.Now()
			g.r.obMu.Unlock()
			if v := g.r.PlayerPause(true); !v.Frozen {
				t.Fatalf("%v still=%v: %+v", class, still, v)
			}
			if g.r.StuckFor() != 0 {
				t.Fatal("StuckFor must be 0 while frozen")
			}
			if still {
				fc.pass(5 * time.Hour)
			} else {
				g.r.e.Clock.Sleep(5 * time.Hour)
			}
			if !g.r.frozenStep() {
				t.Fatal("frozenStep must keep the loop parked")
			}
			g.r.PlayerPause(false)
			g.r.setState(StateClearing, "fight") // the loop's next step: a state the watchdog counts
			if g.r.campaignStall(context.Background(), g.r.latest(), true) || g.r.loop.stallStage != 0 {
				t.Fatalf("%v still=%v: stall ladder fired after the pause", class, still)
			}
			if why := g.r.stuckCheck(g.r.e.Clock.Now()); why != "" {
				t.Fatalf("%v still=%v: stuck watchdog fired after the pause: %s", class, still, why)
			}
			if g.stalled() != nil {
				t.Fatal("the stall ladder exited the bot")
			}
		}
	}
}

// The cap ends a frozen pause by itself, on the wall clock, and leaves a note for the Needs you row.
func TestFreezeCapResumes(t *testing.T) {
	fz := &fakeFreezer{can: true}
	g, _, wc := freezeRig(t, d2.Sorceress, 1, fz)
	g.r.e.Tune.Bot.PauseCapMin = 60
	v := g.r.PlayerPause(true)
	if v.CapMin != 60 || v.EndsAt == nil || !v.EndsAt.Equal(wc.now().Add(time.Hour)) {
		t.Fatalf("cap fields: %+v", v)
	}
	wc.add(59 * time.Minute)
	if !g.r.frozenStep() || !g.r.Frozen() {
		t.Fatal("must stay frozen before the cap")
	}
	wc.add(2 * time.Minute)
	g.r.frozenStep()
	v = g.r.PlayerPauseView()
	if g.r.Frozen() || v.Ended == nil || !v.Ended.Auto || v.Ended.Minutes != 61 {
		t.Fatalf("after the cap: %+v ended %+v", v, v.Ended)
	}
	if got := v.Ended.PauseNeedText(); got == "" {
		t.Fatal("no row text")
	}
	wc.add(40 * time.Minute) // the note fades
	if g.r.PlayerPauseView().Ended != nil {
		t.Fatal("the note must fade")
	}
	// cap 0: no cap
	g.r.e.Tune.Bot.PauseCapMin = 0
	g.r.PlayerPause(true)
	wc.add(48 * time.Hour)
	if !g.r.frozenStep() || !g.r.Frozen() {
		t.Fatal("cap 0 must never resume by itself")
	}
}

// A backend that refuses the freeze leaves the hero running and says why; a thaw that fails keeps the pause.
func TestFreezeFailures(t *testing.T) {
	fz := &fakeFreezer{can: true, err: errors.New("no lockstep verb")}
	g, fc, _ := freezeRig(t, d2.Paladin, 1, fz)
	v := g.r.PlayerPause(true)
	if v.Frozen || v.Error == "" || v.State != PauseRunning || fc.frozen {
		t.Fatalf("failed freeze: %+v clock frozen %v", v, fc.frozen)
	}
	fz.err = nil
	g.r.PlayerPause(true)
	fz.err = errors.New("socket gone")
	v = g.r.PlayerPause(false)
	if !v.Frozen || v.Error == "" {
		t.Fatalf("failed thaw must stay paused and say so: %+v", v)
	}
	fz.err = nil
	if v = g.r.PlayerPause(false); v.Frozen {
		t.Fatalf("second press must resume: %+v", v)
	}
}

// Without a freezer (or one that cannot) Pause keeps its older meaning; freeze:true is refused with a reason.
func TestFreezeUnavailableKeepsWalkHome(t *testing.T) {
	for _, fz := range []*fakeFreezer{nil, {can: false}} {
		g := pauseRig(t, d2.Sorceress, 1)
		if fz != nil {
			g.r.d.Freezer = fz
		}
		yes := true
		v := g.r.PlayerPauseMode(true, &yes)
		if v.Error == "" || v.Frozen || v.CanFreeze {
			t.Fatalf("forced freeze must be refused: %+v", v)
		}
		if v = g.r.PlayerPause(true); v.State != PauseGoingHome || v.Frozen {
			t.Fatalf("plain pause must walk home: %+v", v)
		}
	}
}

// freeze:false asks for the walk-home pause even where the game can be frozen.
func TestFreezeOptOutWalksHome(t *testing.T) {
	fz := &fakeFreezer{can: true}
	g, _, _ := freezeRig(t, d2.Sorceress, 2, fz)
	no := false
	if v := g.r.PlayerPauseMode(true, &no); v.Frozen || v.State != PauseGoingHome || len(fz.log()) != 0 {
		t.Fatalf("%+v %v", v, fz.log())
	}
}

// Reset is break glass: it thaws a frozen game before it starts, and the frozen state is gone.
func TestFreezeResetThaws(t *testing.T) {
	fz := &fakeFreezer{can: true}
	g, fc, _ := freezeRig(t, d2.Sorceress, 1, fz)
	g.r.d.Cycler = fakeCycler{g.w}
	g.r.PlayerPause(true)
	v := g.r.PlayerReset(true)
	if g.r.Frozen() || fc.frozen || v.State != ResetResetting {
		t.Fatalf("frozen %v clock %v reset %+v", g.r.Frozen(), fc.frozen, v)
	}
	if l := fz.log(); len(l) != 2 || l[1] {
		t.Fatalf("freezer calls %v", l)
	}
	// even when the game refuses to thaw the state clears (the loop runs its own recovery)
	g.r.PlayerPause(true)
	fz.err = errors.New("stuck")
	g.r.ForceThaw()
	if g.r.Frozen() {
		t.Fatal("ForceThaw must clear the pause even on error")
	}
}

// A Resume from the goal manager ends a frozen pause too.
func TestFreezeManagerResumeThaws(t *testing.T) {
	fz := &fakeFreezer{can: true}
	g, _, _ := freezeRig(t, d2.Paladin, 1, fz)
	g.r.PlayerPause(true)
	g.r.Resume()
	if g.r.Frozen() {
		t.Fatal("Resume left the game frozen")
	}
}

// A go-home pause that is already held turns into the freeze (one pause, never two).
func TestFreezeReplacesWalkHome(t *testing.T) {
	fz := &fakeFreezer{can: true}
	g, _, _ := freezeRig(t, d2.Sorceress, 1, fz)
	no := false
	g.r.PlayerPauseMode(true, &no)
	v := g.r.PlayerPause(true)
	if !v.Frozen || g.r.plp.want {
		t.Fatalf("%+v want=%v", v, g.r.plp.want)
	}
}

// The frozen view keeps the schema and adds only optional fields (the older goldens stay valid).
func TestFreezeViewShape(t *testing.T) {
	fz := &fakeFreezer{can: true}
	g, _, wc := freezeRig(t, d2.Sorceress, 1, fz)
	g.r.e.Tune.Bot.PauseCapMin = 60
	if got := mustJSON(g.r.PlayerPauseView()); got != `{"schema":"watch.pause/1","state":"running","can_freeze":true}` {
		t.Fatalf("running: %s", got)
	}
	v := g.r.PlayerPause(true)
	want := `{"schema":"watch.pause/1","state":"paused","since":` + mustJSON(wc.now()) + `,"reason":"the game is frozen","frozen":true,"can_freeze":true,"cap_min":60,"ends_at":` + mustJSON(wc.now().Add(time.Hour)) + `}`
	if got := mustJSON(v); got != want {
		t.Fatalf("frozen:\n got %s\nwant %s", got, want)
	}
}
