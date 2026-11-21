package goals

import (
	"testing"
	"time"
)

// A manager on an injected (virtual) clock decides only by that clock: the same scripted virtual timeline gives the same
// credit and the same end whatever the wall time between ticks was (card pgclock: the d2d path now passes the env clock).
func TestManagerRunsOnTheInjectedClockNotTheWall(t *testing.T) {
	for _, tc := range []struct {
		name string
		g    Goal
	}{
		{"sorceress farm", Goal{Type: FarmArea, Params: Params{Area: "Mephisto", UntilMins: 2}}},
		{"paladin farm", Goal{Type: FarmArea, Params: Params{Area: "Andariel", UntilMins: 2}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run := func(wallPause time.Duration) (int64, State) {
				m, _, c := newMgr(t, "")
				mustAdd(t, m, tc.g)
				m.Start()
				for i := 0; i < 5; i++ { // 5 x 30 virtual seconds
					time.Sleep(wallPause)
					c.advance(30 * time.Second)
					m.Tick(c.now())
				}
				g, _ := m.Goal("g1")
				return g.Progress.ActiveSeconds, g.State
			}
			fastSecs, fastState := run(0)
			slowSecs, slowState := run(25 * time.Millisecond)
			if fastSecs != slowSecs || fastState != slowState {
				t.Fatalf("wall time changed the outcome: %d s %s against %d s %s", fastSecs, fastState, slowSecs, slowState)
			}
			if !fastState.Finished() && fastSecs < 120 {
				t.Fatalf("2 virtual minutes of a 2-minute goal not credited: %d s, %s", fastSecs, fastState)
			}
		})
	}
}
