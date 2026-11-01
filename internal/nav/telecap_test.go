package nav

import (
	"testing"

	"d2lod-bot/internal/input"
)

func hopRun(t *testing.T, set func(n *Navigator)) *sim {
	t.Helper()
	g := openField(2000, 2000, 400, 100)
	s := newSim(g, input.Point{X: 2010, Y: 2050})
	n := s.navigator()
	set(n)
	oldHopRules(n)
	tuningOf(n).Nav.ExploreTimeoutSec = 2000
	done := func() bool { return s.pos().X > 2200 }
	if err := n.ExploreToward(ctxT(t), input.Point{X: 2390, Y: 2050}, done); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestTeleportKeyFnDrivesHops(t *testing.T) {
	s := hopRun(t, func(n *Navigator) { n.TeleportKeyFn = func() string { return "f7" } })
	if len(s.casts) == 0 {
		t.Fatal("no hops with a bound key")
	}
	if s.lastKey != "f7" {
		t.Fatalf("cast key %q, want f7 from the function", s.lastKey)
	}
}

func TestNoTeleportKeyWalksWithoutError(t *testing.T) {
	for name, set := range map[string]func(*Navigator){
		"fn empty": func(n *Navigator) { n.TeleportKeyFn = func() string { return "" } },
		"no fn":    func(n *Navigator) {},
	} {
		t.Run(name, func(t *testing.T) {
			s := hopRun(t, set)
			if len(s.casts) != 0 {
				t.Fatalf("%d casts without a Teleport key", len(s.casts))
			}
		})
	}
}

func TestTeleportKeyIsReadWhenLearnedLater(t *testing.T) {
	key := ""
	n := &Navigator{TeleportKeyFn: func() string { return key }}
	if n.HasTeleport() {
		t.Fatal("capability before the skill is bound")
	}
	key = "f2"
	if !n.HasTeleport() {
		t.Fatal("capability missing after the bind")
	}
	n.TeleportKey = "f3" // a fixed key (rigs) wins over the function
	if n.teleKey() != "f3" {
		t.Fatalf("teleKey %q", n.teleKey())
	}
}

func TestTeleportRefusedLevelNeverCasts(t *testing.T) {
	g := openField(2000, 2000, 400, 100)
	s := newSim(g, input.Point{X: 2010, Y: 2050})
	n := s.navigator()
	n.TeleportKey = "f2"
	n.TeleportOK = func(int) bool { return false }
	tuningOf(n).Nav.ExploreTimeoutSec = 2000
	done := func() bool { return s.pos().X > 2100 }
	if err := n.ExploreToward(ctxT(t), input.Point{X: 2390, Y: 2050}, done); err != nil {
		t.Fatal(err)
	}
	if len(s.casts) != 0 {
		t.Fatalf("%d casts on a level that refuses Teleport", len(s.casts))
	}
}

func TestRefusedHopsAreBounded(t *testing.T) {
	g := openField(2000, 2000, 400, 100)
	s := newSim(g, input.Point{X: 2010, Y: 2050})
	s.noTeleport = true // every cast is refused (d2d still spends mana, T3)
	n := s.navigator()
	n.TeleportKey = "f2"
	tuningOf(n).Nav.ExploreTimeoutSec = 2000
	done := func() bool { return s.pos().X > 2150 }
	if err := n.ExploreToward(ctxT(t), input.Point{X: 2390, Y: 2050}, done); err != nil {
		t.Fatal(err)
	}
	if len(s.casts) > 3*teleportBadHops+3 {
		t.Fatalf("%d refused casts, want a bounded number", len(s.casts))
	}
}
