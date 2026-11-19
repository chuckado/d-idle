package bot

import (
	"context"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/goals"
)

// The live bug of 2026-11-17: the hero farmed Dark Wood under a farm_area goal, the player tapped Start, the accept was
// recorded and nothing ran because boardStep was only reached under a campaign goal. The accept now preempts a farm,
// level or boss goal (two classes; the knob at 0 changes nothing).
func TestQuestBoardStartRunsUnderAFarmGoal(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		for _, goal := range []goals.Goal{farm("Blood Moor"), {ID: "g2", Type: goals.ReachLevel, Params: goals.Params{Level: 30}}} {
			t.Run(class.String()+"/"+string(goal.Type), func(t *testing.T) {
				g, x := boardRig(t, true)
				// The farm must still be running when the accept lands: with an instant explore the farm clears its
				// area and ends in the "no fresh game" error before a loaded test goroutine reaches Accept (flake of
				// 2026-11-18). Explore holds until the accept is recorded (bounded, so a bug fails instead of hanging).
				accepted := make(chan struct{})
				g.mover.exploreFn = func(ctx context.Context, done func() bool) error {
					select {
					case <-accepted:
					case <-ctx.Done():
					case <-time.After(30 * time.Second):
					}
					return nil
				}
				g.w.set(func(w *world) { w.class = class })
				g.run()
				defer g.finish()
				if err := g.r.Start(goal); err != nil {
					t.Fatal(err)
				}
				g.waitFor("the goal runs", func() bool { return g.w.snaps.Load() > 3 })
				if err := x.Accept("den_of_evil", "clear_den"); err != nil {
					t.Fatal(err)
				}
				close(accepted)
				// the log, not x.Next: the test goroutine must not read the board's world while the loop writes it (-race)
				g.waitFor("awaiting hand-in", func() bool {
					l := g.w.log()
					i := strings.Index(l, "exit 8")
					return i >= 0 && strings.Contains(l[i:], "talk Akara")
				})
				if !strings.Contains(g.w.log(), "exit 8") {
					t.Fatalf("the chunk never went to the Den: %s", g.w.log())
				}
				if strings.Contains(g.w.log(), "talk Akara/reward") {
					t.Fatalf("handed in by itself: %s", g.w.log())
				}
			})
		}
	}
}

func TestQuestBoardKnobOffIgnoresAcceptUnderFarm(t *testing.T) {
	g, _ := boardRig(t, false)
	if g.r.boardPreempts(&d2.Snapshot{}, control{spec: goalSpec{kind: goals.FarmArea}}) {
		t.Fatal("preempted with the knob at 0")
	}
}

// A fight in the clear radius delays the start of a chunk and the reason is shown; a campaign goal and a Runs goal are
// not touched; once calm the chunk starts.
func TestQuestBoardQueuedReasonWhileFighting(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		t.Run(class.String(), func(t *testing.T) {
			g, x := boardRig(t, true)
			g.w.set(func(w *world) { w.class = class; w.area = 2 })
			snap, _ := g.w.snapshot()
			g.r.obMu.Lock()
			g.r.ob.lastSnap = snap
			g.r.obMu.Unlock()
			if err := x.Accept("den_of_evil", "clear_den"); err != nil {
				t.Fatal(err)
			}
			farmCtl := control{spec: goalSpec{kind: goals.FarmArea}}
			if !g.r.boardPreempts(snap, farmCtl) {
				t.Fatal("calm field: the chunk should start")
			}
			if got := g.r.QuestQueued(); got != "" {
				t.Fatalf("queued reason with nothing in the way: %q", got)
			}
			if g.r.boardPreempts(snap, control{spec: goalSpec{kind: goals.Runs}}) {
				t.Fatal("a Runs goal was preempted")
			}
			g.w.addMonster(500, 1, 3, 3)
			snap, _ = g.w.snapshot()
			g.r.obMu.Lock()
			g.r.ob.lastSnap = snap
			g.r.obMu.Unlock()
			if g.r.hostileNear(snap, g.r.e.Tune.Bot.ClearRadius) == 0 {
				t.Skip("the fake monster is not hostile here")
			}
			if g.r.boardPreempts(snap, farmCtl) {
				t.Fatal("started in the middle of a fight")
			}
			if q := g.r.QuestQueued(); !strings.Contains(q, "fight") {
				t.Fatalf("queued reason %q", q)
			}
		})
	}
}
