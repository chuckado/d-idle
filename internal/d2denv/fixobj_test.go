package d2denv

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/d2dctl/fake"
	"d2lod-bot/internal/input"
)

// fixobj (drill H.obj): a chest the level placed near the player is listed by the snapshot and the d2d input
// operates it with "cmd interact <preset index>" (before: no object was listed and Interact answered ErrUnsupported).
func TestFieldChestListedAndOperated(t *testing.T) {
	p := sockPath(t)
	fix := fake.Sample()
	fix.Presets = []string{"0\t51\t30.0\t10.0\t0\tWaypoint", "6\t580\t33.0\t12.0\t0\tChest"}
	srv, err := fake.New(p, fix, fake.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	be, err := Build(context.Background(), Config{Sock: p, Log: quiet(), ID: "t"})
	if err != nil {
		t.Fatal(err)
	}
	defer be.Close()
	old := d2dctl.Pause
	t.Cleanup(func() { d2dctl.Pause = old })
	if err := be.Start(context.Background(), 0); err != nil {
		t.Fatal(err)
	}
	s, err := be.Env.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	var chest *d2.Unit
	for i := range s.Objects {
		if s.Objects[i].Class == 580 {
			chest = &s.Objects[i]
		}
	}
	if chest == nil {
		t.Fatalf("chest not listed: %+v", s.Objects)
	}
	err = be.Env.In.Interact(input.Target{Type: 2, ID: chest.ID, Pos: input.Point{X: chest.X, Y: chest.Y}})
	if err != nil {
		t.Fatalf("interact: %v", err)
	}
	ls := srv.Lines()
	if ls[len(ls)-1] != "cmd interact 6" {
		t.Fatalf("last command %q", ls[len(ls)-1])
	}
}

// fixobj (drill H.pause): under lockstep the pause between hover and click steps the game; on a free-running d2d it is
// a wall sleep that sends nothing (before: always a wall sleep, 0 ticks under lockstep).
func TestPauseStepsUnderLockstep(t *testing.T) {
	srv, p := startFake(t)
	be, err := Build(context.Background(), Config{Sock: p, Log: quiet(), ID: "t"})
	if err != nil {
		t.Fatal(err)
	}
	defer be.Close()
	old := d2dctl.Pause
	t.Cleanup(func() { d2dctl.Pause = old })
	if err := be.Start(context.Background(), 0); err != nil {
		t.Fatal(err)
	}
	steps := func() (n int) {
		for _, l := range srv.Lines() {
			if strings.HasPrefix(l, "step ") {
				n++
			}
		}
		return
	}
	before := steps()
	w := time.Now()
	if err := d2dctl.Pause(context.Background(), 4*time.Second); err != nil {
		t.Fatal(err)
	}
	if steps() == before || time.Since(w) > time.Second {
		t.Fatalf("lockstep pause: %d new step commands, wall %v", steps()-before, time.Since(w))
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := d2dctl.Pause(ctx, time.Second); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled pause: %v", err)
	}

	// real time: no Pacer, a wall sleep, no step command
	wall := &LateClock{c: be.Client, ctx: context.Background()}
	before = steps()
	w = time.Now()
	if err := wall.Pause(context.Background(), 120*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if steps() != before || time.Since(w) < 100*time.Millisecond {
		t.Fatalf("wall pause: %d new steps, took %v", steps()-before, time.Since(w))
	}
}
