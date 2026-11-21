package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/quest"
	"d2lod-bot/internal/task"
)

// Act 1 object work: the cairn stones of Stony Field and the "click each of several identical objects once" rule.
// Tags: classes 17..21 are S (Kolbot StoneAlpha..StoneLambda); V on the d2d lab build (docs/probes/search-for-cain-l1-2026-11-18.md):
// the order is a random permutation drawn at the first click, a wrong click is ignored (nothing resets), the lit state
// is the state column of debug objects, the fifth stone opens a portal to Tristram (level 38). Game.exe: U.

// CairnStones is the Interact name task.cainNodes passes for the five stones.
const CairnStones = "cairn-stones"

const stonyFieldArea = 4

// cairnClasses are the five stones (S, Kolbot). quest.ObjCairnStone (17) is the first.
var cairnClasses = []int{quest.ObjCairnStone, quest.ObjCairnStone + 1, quest.ObjCairnStone + 2, quest.ObjCairnStone + 3, quest.ObjCairnStone + 4}

// onceSet remembers which object ids were already clicked, per area, so three calls for "cage" click three different
// cages and Cain's gibbet is not clicked twice. A new area forgets (ids are only unique within a game, and a visit of
// the area after a restart must be able to click again).
type onceSet struct {
	mu   sync.Mutex
	area int
	ids  map[uint32]bool
}

func (s *onceSet) add(area int, id uint32) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.area != area || s.ids == nil {
		s.area, s.ids = area, map[uint32]bool{}
	}
	s.ids[id] = true
}

func (s *onceSet) has(area int, id uint32) bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.area == area && s.ids[id]
}

func (o objectAdapter) findUnclicked(ctx context.Context, class, area int) (*d2.Unit, error) {
	return o.findNearSkip(ctx, class, nil, func(u *d2.Unit) bool { return o.clicked.has(area, u.ID) })
}

// stoneOps is what the stone search needs from the game, so the search can be tested without one. Stones are numbered
// 0..4 by class (17..21); an implementation resolves the stone by class on every call because object ids and indexes
// shift as rooms load (V lab drill L1).
type stoneOps interface {
	Click(ctx context.Context, i int) error // click stone i; returns after the click settled
	// Lit is the lit flag of every stone; known is false when the backend cannot read it (no mode source, U).
	Lit(ctx context.Context) (lit []bool, known bool, err error)
	Portal(ctx context.Context) bool // the way to Tristram exists (it may take a moment to appear)
}

// maxStoneClicks bounds the search with the lit state readable: the order is a random permutation drawn at the first
// click, a wrong click is ignored and a lit stone is not clicked again, so 5+4+3+2+1 clicks always suffice (V lab drill
// L1: nothing resets). maxBlindClicks is the bound without a lit state: five full cycles of five (a cycle lights at
// least one more stone, and clicking a lit stone is harmless, V).
const (
	maxStoneClicks = 15
	maxBlindClicks = 25
)

// solveCairn lights all n stones by cycling through them: click every stone that is not lit, read the state, repeat
// until the portal exists. Wrong clicks cost nothing but a click. With the lit state unreadable it clicks every stone
// in each cycle and watches the portal only. It fails with the lit set (and the likely cause) when a whole cycle lit
// nothing, and after the click bound.
func solveCairn(ctx context.Context, ops stoneOps, n int) error {
	clicks := 0
	for {
		if ops.Portal(ctx) {
			return nil
		}
		lit, known, err := ops.Lit(ctx)
		if err != nil {
			return err
		}
		if known && len(lit) != n {
			return fmt.Errorf("cairn stones: %d stones read, want %d", len(lit), n)
		}
		max := maxStoneClicks
		if !known {
			max = maxBlindClicks
		}
		before := litCount(lit)
		if known && before == n {
			// the fifth click is accepted at once, the portal object follows a moment later
			if ops.Portal(ctx) {
				return nil
			}
			return fmt.Errorf("cairn stones: all %d stones are lit but no portal appeared", n)
		}
		for i := 0; i < n; i++ {
			if known && lit[i] {
				continue
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if clicks >= max {
				return fmt.Errorf("cairn stones: %d clicks without a portal (lit %s)", max, litString(lit, known))
			}
			clicks++
			if err := ops.Click(ctx, i); err != nil {
				return err
			}
			if ops.Portal(ctx) {
				return nil
			}
			if known {
				if lit, known, err = ops.Lit(ctx); err != nil {
					return err
				} else if !known {
					return fmt.Errorf("cairn stones: the lit state became unreadable after click %d", clicks)
				}
			}
		}
		if known && litCount(lit) == before {
			return fmt.Errorf("cairn stones: a whole cycle lit no stone (lit %s): is the Key to the Cairn Stones in the pack?", litString(lit, known))
		}
	}
}

func litCount(l []bool) int {
	n := 0
	for _, v := range l {
		if v {
			n++
		}
	}
	return n
}

func litString(l []bool, known bool) string {
	if !known {
		return "unknown"
	}
	b := make([]byte, len(l))
	for i, v := range l {
		b[i] = '0'
		if v {
			b[i] = '1'
		}
	}
	return string(b)
}

// tristramArea is the level id the Tristram portal leads to (V lab drill L1: `debug level` 38 after `cmd interact -2002`).
const tristramArea = quest.AreaTristram

// cairnStones opens the Tristram portal and walks through it, then checks the level is Tristram. With the portal
// already open (a restart in the same game) it only enters. Needs the Key to the Cairn Stones in the pack (the step's
// Ready says so before this runs); without it every click is ignored and the search ends in an error that says so.
func (o objectAdapter) cairnStones(ctx context.Context) error {
	s0, err := o.e.Snapshot()
	if err != nil {
		return err
	}
	if s0.Player.Area == tristramArea {
		return nil
	}
	if s0.Player.Area != stonyFieldArea {
		return task.Errorf(task.KindPrereq, "the cairn stones are in Stony Field, the character is in area %d", s0.Player.Area)
	}
	if !hasClass(s0, quest.ObjRedPortal) {
		ops := &liveStones{o: o}
		if err := ops.load(ctx); err != nil {
			return task.Wrap(task.KindNotFound, err)
		}
		if err := solveCairn(ctx, ops, len(cairnClasses)); err != nil {
			return task.Wrap(task.KindTimeout, err)
		}
	}
	if err := o.Interact(ctx, "Tristram portal"); err != nil {
		return err
	}
	s, err := o.e.Snapshot()
	if err != nil {
		return err
	}
	if s.Player.Area != tristramArea {
		return task.Errorf(task.KindNotFound, "the Tristram portal led to area %d, not Tristram (%d)", s.Player.Area, tristramArea)
	}
	return nil
}

func hasClass(s *d2.Snapshot, class int) bool {
	for i := range s.Objects {
		if s.Objects[i].Class == class {
			return true
		}
	}
	return false
}

// liveStones is stoneOps over the running game. Each call finds the stone by class in a fresh snapshot.
type liveStones struct{ o objectAdapter }

// load walks to the stones until all five are in view, so that a missing one is an error before the first click.
func (l *liveStones) load(ctx context.Context) error {
	for _, c := range cairnClasses {
		if _, err := l.o.find(ctx, c); err != nil {
			return fmt.Errorf("stone class %d: %w", c, err)
		}
	}
	return nil
}

func (l *liveStones) Click(ctx context.Context, i int) error {
	u, err := l.o.find(ctx, cairnClasses[i])
	if err != nil {
		return err
	}
	if err := l.o.click(ctx, u); err != nil {
		return err
	}
	l.o.e.Clock.Sleep(900 * time.Millisecond)
	return nil
}

// Lit reads the Mode of the five stones: 0 untouched, above 0 lit, -1 unknown (the backend has no state source, S the
// d2d source fills it from debug objects and Game.exe reads the unit mode, U live). One unreadable stone makes the
// whole read unknown, so the solver falls back to the blind cycle instead of trusting a half state.
func (l *liveStones) Lit(ctx context.Context) ([]bool, bool, error) {
	s, err := l.o.e.Snapshot()
	if err != nil {
		return nil, false, err
	}
	lit := make([]bool, len(cairnClasses))
	for i, c := range cairnClasses {
		var u *d2.Unit
		for j := range s.Objects {
			if s.Objects[j].Class == c {
				u = &s.Objects[j]
				break
			}
		}
		switch {
		case u == nil:
			return nil, false, fmt.Errorf("cairn stones: stone class %d left the view", c)
		case u.Mode < 0:
			return nil, false, nil
		}
		lit[i] = u.Mode > 0
	}
	return lit, true, nil
}

// Portal waits briefly: the portal object follows the fifth click by a tick or two.
func (l *liveStones) Portal(ctx context.Context) bool {
	for try := 0; try < 3; try++ {
		if s, err := l.o.e.Snapshot(); err == nil && hasClass(s, quest.ObjRedPortal) {
			return true
		}
		if try < 2 {
			l.o.e.Clock.Sleep(300 * time.Millisecond)
		}
	}
	return false
}
