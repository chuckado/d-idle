package d2denv

import (
	"context"
	"testing"
	"time"

	"d2lod-bot/internal/input"
)

// B1 (change fixrun): the restart path. Build wires Proc.OnInGame to the controller's ResetRun; when a (re)started
// game reports ingame, the next move must press the run key again, once. The fake has no process to restart, so
// WaitInGame (the seam every Cycler.EnterGame, restart and chicken re-entry goes through) is called directly.
func TestFreshGameResetsRunBelief(t *testing.T) {
	t.Setenv("D2D_RUN", "")
	srv, p := startFake(t)
	be, err := Build(context.Background(), Config{Sock: p, Log: quiet(), ID: "t"})
	if err != nil {
		t.Fatal(err)
	}
	defer be.Close()
	if be.Ctl.Run == nil {
		t.Fatal("run policy not wired")
	}
	keys := func() int {
		n := 0
		for _, l := range srv.Lines() {
			if l == "key r" {
				n++
			}
		}
		return n
	}
	w := input.Point{X: 175, Y: 70}
	for i := 0; i < 2; i++ {
		if err := be.Ctl.Move(w); err != nil {
			t.Fatal(err)
		}
	}
	if keys() != 1 {
		t.Fatalf("first game: %d run keys, want 1", keys())
	}
	if err := be.Proc.WaitInGame(context.Background(), 5*time.Second); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := be.Ctl.Move(w); err != nil {
			t.Fatal(err)
		}
	}
	if keys() != 2 {
		t.Fatalf("second game: %d run keys, want 2", keys())
	}
}
