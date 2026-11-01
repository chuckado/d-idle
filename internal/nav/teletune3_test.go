package nav

import (
	"testing"
	"time"

	"d2lod-bot/internal/input"
)

// hopHere asks for one hop toward a point 12 subtiles ahead (a leg the leg rule declines) and returns the casts.
func hopHere(t *testing.T, s *sim, n *Navigator) int {
	t.Helper()
	o, err := n.observeWait(ctxT(t))
	if err != nil {
		t.Fatal(err)
	}
	p := o.pos
	if _, _, err := n.hop(ctxT(t), o, input.Point{X: p.X + 12, Y: p.Y}); err != nil {
		t.Fatal(err)
	}
	return len(s.casts)
}

// The stall rule through the real stall memory (noteStuck, as walk.go records it): a stall at the spot turns the leg
// rule off and lowers the mana floor to the escape reserve (15 percent of 150 = 23); a stall elsewhere, on another
// level or older than the memory window leaves the leg rule in force.
func TestStalledHereThroughNoteStuck(t *testing.T) {
	for _, c := range []struct {
		name  string
		mana  int
		stall func(n *Navigator, s *sim)
		casts int
	}{
		{"no stall", 150, func(*Navigator, *sim) {}, 0},
		{"stall at the spot", 150, func(n *Navigator, s *sim) { n.noteStuck(s.cur().no, s.pos()) }, 1},
		{"stall 9 subtiles away is the same spot", 150, func(n *Navigator, s *sim) {
			n.noteStuck(s.cur().no, input.Point{X: s.pos().X + 9, Y: s.pos().Y})
		}, 1},
		{"stall 30 subtiles away", 150, func(n *Navigator, s *sim) {
			n.noteStuck(s.cur().no, input.Point{X: s.pos().X + 30, Y: s.pos().Y})
		}, 0},
		{"stall on another level", 150, func(n *Navigator, s *sim) { n.noteStuck(s.cur().no+1, s.pos()) }, 0},
		{"stall older than the window", 150, func(n *Navigator, s *sim) {
			n.noteStuck(s.cur().no, s.pos())
			s.now = s.now.Add(stuckMemory + time.Second)
		}, 0},
		{"stall inside the window", 150, func(n *Navigator, s *sim) {
			n.noteStuck(s.cur().no, s.pos())
			s.now = s.now.Add(stuckMemory - time.Second)
		}, 1},
		{"stall, mana 60: 36 left is above the 15 percent floor (23)", 60, func(n *Navigator, s *sim) { n.noteStuck(s.cur().no, s.pos()) }, 1},
		{"stall, mana 45: 21 left is under the 15 percent floor", 45, func(n *Navigator, s *sim) { n.noteStuck(s.cur().no, s.pos()) }, 0},
	} {
		for _, key := range []string{"f2", "sk54"} { // two builds: the rule is the same whatever the key
			s := newSim(openField(2000, 2000, 400, 100), input.Point{X: 2010, Y: 2050})
			s.maxMana, s.mana, s.castCost, s.manaRegen = 150, c.mana, 24, 0
			n := s.navigator()
			n.TeleportKey = key
			c.stall(n, s)
			if got := hopHere(t, s, n); got != c.casts {
				t.Errorf("%s (key %s): %d casts, want %d", c.name, key, got, c.casts)
			}
			if c.casts > 0 && s.lastKey != key {
				t.Errorf("%s: cast with %q, want %q", c.name, s.lastKey, key)
			}
		}
	}
}

// ResetStallMemory forgets the stall: the leg rule is back, and the recorded spots, hard blocks, stalled gates, refused
// cells and decline reason are empty (DRILLS25: one process running D22-PC then D22-P).
func TestResetStallMemory(t *testing.T) {
	s := newSim(openField(2000, 2000, 400, 100), input.Point{X: 2010, Y: 2050})
	s.maxMana, s.mana, s.castCost = 150, 150, 24
	n := s.navigator()
	n.TeleportKey = "f2"
	n.noteStuck(s.cur().no, s.pos())
	n.hardBlock(input.Point{X: 2030, Y: 2050}, 3)
	n.markStalledGate(s.cur().no, s.cur().no+1, input.Point{X: 2100, Y: 2050})
	n.teleBad[[2]int{1, 1}] = 2
	n.teleFails, n.teleWhy = 2, "leg"
	n.markNoExit(1, 2, 10)
	n.ResetStallMemory()
	if len(n.spots)+len(n.hard)+len(n.badGates)+len(n.noExit)+len(n.teleBad) != 0 || n.teleFails != 0 || n.teleWhy != "" {
		t.Fatalf("memory left: spots %d hard %d gates %d noExit %d teleBad %d fails %d why %q",
			len(n.spots), len(n.hard), len(n.badGates), len(n.noExit), len(n.teleBad), n.teleFails, n.teleWhy)
	}
	if got := hopHere(t, s, n); got != 0 {
		t.Fatalf("%d casts after the reset: the leg rule should decline a 12 subtile leg", got)
	}
	// and without the reset the same stall allows it (the control of the test above)
	s2 := newSim(openField(2000, 2000, 400, 100), input.Point{X: 2010, Y: 2050})
	s2.maxMana, s2.mana, s2.castCost = 150, 150, 24
	n2 := s2.navigator()
	n2.TeleportKey = "f2"
	n2.noteStuck(s2.cur().no, s2.pos())
	if got := hopHere(t, s2, n2); got != 1 {
		t.Fatalf("%d casts with the stall kept, want 1", got)
	}
}

// TELETUNE3 border leg: a border exit never reaches hop code, so a bound Teleport key must not change the walk (DRILLS25:
// 717 ticks with the key against 708 without, 0 hops; the logs show the walks split at a replan, not at a hop). Same
// clicks, same elapsed time, same number of world polls, 0 casts; the key is looked up once, for the ExitTo gate check.
func TestBorderLegIsTheSameWithAKeyBound(t *testing.T) {
	type run struct {
		clicks  int
		elapsed time.Duration
		updates int
		casts   int
		keyFn   int
	}
	do := func(key string) run {
		s, _ := gateWorld()
		start := s.pos()
		for y := 3010; y < 3120; y += 30 { // both levels are known on arrival, as on d2d: no exploring
			for x := 2810; x < 3200; x += 30 {
				s.px, s.py = float64(x), float64(y)
				_, _ = s.Update()
			}
		}
		s.px, s.py = float64(start.X), float64(start.Y)
		s.maxMana, s.mana, s.castCost = 150, 150, 24
		n := s.navigator()
		calls := 0
		n.TeleportKeyFn = func() string { calls++; return key }
		t0 := s.now
		if err := n.ExitTo(ctxT(t), 2); err != nil {
			t.Fatal(err)
		}
		return run{len(s.clicks), s.now.Sub(t0), s.updates, len(s.casts), calls}
	}
	for _, key := range []string{"f2", "sk54"} {
		with, without := do(key), do("")
		if with.clicks != without.clicks || with.elapsed != without.elapsed || with.updates != without.updates || with.casts != 0 {
			t.Errorf("key %s: with %+v, without %+v", key, with, without)
		}
		if with.keyFn > 2 {
			t.Errorf("key %s: the key was looked up %d times on a border leg", key, with.keyFn)
		}
	}
}
