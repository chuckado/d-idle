package d2denv

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/d2dctl/fake"
)

func freezeRig(t *testing.T, fix fake.Fixtures) (*Freezer, *fake.Server) {
	t.Helper()
	srv, err := fake.New(filepath.Join(t.TempDir(), "d.sock"), fix, fake.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(srv.Close)
	c := d2dctl.New(srv.Path(), d2dctl.Options{Timeout: 2 * time.Second})
	t.Cleanup(func() { c.Close() })
	return NewFreezer(c, &LateClock{c: c, ctx: context.Background()}), srv
}

// A free-running fork can freeze and thaw; the clock verb shows each switch.
func TestFreezerSwitchesLockstep(t *testing.T) {
	f, _ := freezeRig(t, fake.Fixtures{Realtime: true})
	ctx := context.Background()
	if !f.CanFreeze(ctx) {
		t.Fatal("a free-running fork with the verb must be freezable")
	}
	if err := f.Freeze(ctx, true); err != nil {
		t.Fatal(err)
	}
	if ci, _ := f.C.Clock(ctx); !ci.Lockstep {
		t.Fatal("not frozen after Freeze(true)")
	}
	if f.CanFreeze(ctx) {
		t.Fatal("a frozen game reads as a lockstep run: the runner keeps its own state, not the probe")
	}
	if err := f.Freeze(ctx, false); err != nil {
		t.Fatal(err)
	}
	if ci, _ := f.C.Clock(ctx); ci.Lockstep {
		t.Fatal("still frozen after Freeze(false)")
	}
}

// No button on a lockstep run (the bot steps the game itself) and on a build without the runtime switch.
func TestFreezerRefusals(t *testing.T) {
	f, _ := freezeRig(t, fake.Fixtures{})
	if f.CanFreeze(context.Background()) {
		t.Fatal("a lockstep d2d must not be freezable")
	}
	f, _ = freezeRig(t, fake.Fixtures{Realtime: true, NoLockstepVerb: true})
	if f.CanFreeze(context.Background()) {
		t.Fatal("no switch verb, no freeze")
	}
	if err := f.Freeze(context.Background(), true); err == nil {
		t.Fatal("Freeze without the verb must fail")
	}
}

// The wall clock stands still while frozen and does not jump when it runs again.
func TestLateClockFrozenStandsStill(t *testing.T) {
	l := &LateClock{}
	a := l.Now()
	l.SetFrozen(true)
	f1 := l.Now()
	time.Sleep(30 * time.Millisecond)
	if f2 := l.Now(); !f2.Equal(f1) {
		t.Fatalf("clock moved while frozen: %v -> %v", f1, f2)
	}
	l.SetFrozen(true) // idempotent
	l.SetFrozen(false)
	b := l.Now()
	if b.Before(f1) || b.Sub(a) > 25*time.Millisecond {
		t.Fatalf("the frozen stretch leaked into the clock: a=%v b=%v", a, b)
	}
}
