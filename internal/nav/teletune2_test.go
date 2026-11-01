package nav

import (
	"testing"
	"time"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/input"
)

// bayMaze is the Halls of the Dead 1 shape of DRILLS24: a wall between the character and the exit whose only gap is
// far away, so a hop straight toward the exit lands against the wall and the walk from there is no shorter.
func bayMaze() *d2map.Grid {
	g := openField(2000, 2000, 300, 300)
	blockRect(g, d2map.Rect{X: 2150, Y: 2030, W: 6, H: 270}) // gap at the top (y 2000..2029)
	return g
}

func mapKnown(s *sim) {
	for y := 2010; y < 2300; y += 30 { // the whole level is known, as on d2d
		for x := 2010; x < 2300; x += 30 {
			s.px, s.py = float64(x), float64(y)
			_, _ = s.Update()
		}
	}
}

func TestPathHopNeverLengthensThePath(t *testing.T) {
	g := bayMaze()
	start := input.Point{X: 2132, Y: 2100}
	target := input.Point{X: 2290, Y: 2100}
	s := newSim(g, start)
	mapKnown(s)
	s.px, s.py = float64(start.X), float64(start.Y)
	n := s.navigator()
	n.TeleportKey = "f2"
	o, err := n.observeWait(ctxT(t))
	if err != nil {
		t.Fatal(err)
	}
	cell, saved, found, planned := n.pathHop(o, target)
	if !planned || !found {
		t.Fatalf("planned %v found %v", planned, found)
	}
	before, after := n.walkLen(o, target), n.walkLen(&obs{lm: o.lm, pos: cell, snap: o.snap}, target)
	if after > before-saved+1 || saved <= 13 {
		t.Fatalf("hop to %v: path %.0f before, %.0f after, saved %.0f", cell, before, after, saved)
	}
	// the old straight hop (nearest to the exit) lands against the wall: its remaining path is longer than the path hop's
	gc, ok := n.hopCell(o, target)
	if !ok {
		t.Fatal("no greedy cell")
	}
	if ga := n.walkLen(&obs{lm: o.lm, pos: gc, snap: o.snap}, target); ga <= after {
		t.Fatalf("greedy cell %v leaves %.0f, path cell %v leaves %.0f: the case does not discriminate", gc, ga, cell, after)
	}
}

func TestStonyTombHopsThenManaReserveStopsThem(t *testing.T) {
	g := openField(2000, 2000, 400, 100)
	s := newSim(g, input.Point{X: 2010, Y: 2050})
	s.maxMana, s.mana, s.castCost, s.manaRegen = 150, 150, 24, 0
	mapKnown(s)
	s.px, s.py = 2010, 2050
	n := s.navigator()
	n.TeleportKey = "f2"
	start := s.now
	if err := n.teleportToward(ctxT(t), input.Point{X: 2390, Y: 2050}, 10); err != nil {
		t.Fatal(err)
	}
	// 150 -> 126 -> 102 -> 78; the 4th would leave 54, below the 60 reserve
	if len(s.casts) != 3 || s.mana != 78 {
		t.Fatalf("%d casts, mana %d, want 3 and 78", len(s.casts), s.mana)
	}
	if s.now.Sub(start) > 10*time.Second {
		t.Fatalf("waited %v", s.now.Sub(start))
	}
}

func TestShortLegsWalk(t *testing.T) {
	g := openField(2000, 2000, 400, 100)
	s := newSim(g, input.Point{X: 2010, Y: 2050})
	n := s.navigator()
	n.TeleportKey = "f2"
	if err := n.teleportToward(ctxT(t), input.Point{X: 2021, Y: 2050}, 1); err != nil {
		t.Fatal(err)
	}
	if len(s.casts) != 0 {
		t.Fatalf("%d hops on a 11 subtile leg", len(s.casts))
	}
}

// After a stall the leg rule is off and the smaller escape reserve applies.
func TestStallAllowsHopBelowNormalFloor(t *testing.T) {
	for _, c := range []struct {
		name  string
		mana  int
		stall bool
		casts int
	}{
		{"no stall, 60 mana: 36 left is under the 60 reserve", 60, false, 0},
		{"stall, 60 mana: 36 left is above the 23 escape floor", 60, true, 1},
		{"stall, 40 mana: 16 left is under the escape floor", 40, true, 0},
	} {
		g := openField(2000, 2000, 400, 100)
		s := newSim(g, input.Point{X: 2010, Y: 2050})
		s.maxMana, s.mana, s.castCost, s.manaRegen = 150, c.mana, 24, 0
		n := s.navigator()
		n.TeleportKey = "f2"
		if c.stall {
			n.spots = append(n.spots, &stuckSpot{level: s.cur().no, at: input.Point{X: 2012, Y: 2050}, last: n.clk.Now(), hits: 1})
		}
		// a leg that is 12 subtiles: too short for the leg rule, so only the stall lets it hop
		o, err := n.observeWait(ctxT(t))
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := n.hop(ctxT(t), o, input.Point{X: 2022, Y: 2050}); err != nil {
			t.Fatal(err)
		}
		if len(s.casts) != c.casts {
			t.Errorf("%s: %d casts, want %d", c.name, len(s.casts), c.casts)
		}
	}
}

// The refused-hop bound: casts that move nothing stop after teleportBadHops, exactly, even with the leg rule off.
func TestRefusedHopBoundIsExercised(t *testing.T) {
	g := openField(2000, 2000, 400, 100)
	s := newSim(g, input.Point{X: 2010, Y: 2050})
	s.noTeleport = true
	n := s.navigator()
	n.TeleportKey = "f2"
	oldHopRules(n)
	if err := n.teleportToward(ctxT(t), input.Point{X: 2390, Y: 2050}, 10); err != nil {
		t.Fatal(err)
	}
	if len(s.casts) != teleportBadHops {
		t.Fatalf("%d casts, want exactly %d", len(s.casts), teleportBadHops)
	}
}
