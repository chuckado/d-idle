package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/drill"
	"d2lod-bot/internal/input"
)

// spy counts what the bot's code sends to the game, so a drill can say "nothing was clicked" with evidence.
type spy struct {
	input.Controller
	l      *lab
	rights int // right clicks
	clicks int // all clicks
	keys   int
	last   int    // tick of the last right click, -1 none
	after  func() // runs once, right after the first right click returns (D2.d: a hit that cuts the cast)
}

func (s *spy) Click(b input.Button, x, y int, mods ...input.Mod) error {
	s.clicks++
	if b == input.Right {
		s.rights++
		s.last = s.l.tick()
	}
	err := s.Controller.Click(b, x, y, mods...)
	if b == input.Right && s.after != nil {
		f := s.after
		s.after = nil
		f()
	}
	return err
}

func (s *spy) Key(name string) error { s.keys++; return s.Controller.Key(name) }

// tick is the world tick counter (the virtual clock of the lockstep d2d).
func (l *lab) tick() int {
	c, err := l.be.Client.Clock(context.Background())
	if err != nil {
		return -1
	}
	return c.Tick
}

func (l *lab) catalog() combat.Catalog { return combat.NewCatalog(l.db) }

// bot returns the env with the spy in front of the input, the portal verb and the town-entry code the bot calls.
func (l *lab) bot() (*spy, func(context.Context) error, d2dTown) {
	e := *l.be.Env
	sp := &spy{Controller: e.In, l: l, last: -1}
	e.In = sp
	itemAt := func(ctx context.Context, id int) (int, int, error) { return l.be.Client.ItemAt(ctx, id) }
	portal := d2dPortalWith(&e, itemAt, l.catalog())
	return sp, portal, d2dTown{townAdapter: townAdapter{t: l.tw, e: &e, mv: l.nv}, portal: portal}
}

func (l *lab) stats() combat.Stats { return l.be.Env.Read.UnitStats }

// PView reads the portal-relevant state through the StateReader.
func (l *lab) PView(ctx context.Context) (drill.PortalView, error) {
	s, err := l.be.Env.Snapshot()
	if err != nil {
		return drill.PortalView{}, err
	}
	if s.Player == nil || s.Inventory == nil {
		return drill.PortalView{}, errors.New("no player or inventory in the snapshot")
	}
	v := drill.PortalView{Area: s.Player.Area, InTown: s.Player.InTown(), X: s.Player.X, Y: s.Player.Y, Mode: s.Player.Mode, Tome: -1, Tick: l.tick()}
	for _, o := range s.Objects {
		if o.Class == townPortalObject {
			v.Any++
			v.PortalPos = append(v.PortalPos, [2]int{o.X, o.Y})
			v.PortalIDs = append(v.PortalIDs, int(o.ID))
			if math.Hypot(float64(o.X-s.Player.X), float64(o.Y-s.Player.Y)) < 25 {
				v.Near++
			}
		}
	}
	cat := l.catalog()
	for i := range s.Inventory.Items {
		u := &s.Inventory.Items[i]
		if u.Item == nil || u.Item.Where != d2.ContInventory || !cat.Item(u.Class).Portal {
			continue
		}
		if cat.Item(u.Class).Tome {
			v.TomeID, v.Tome = int(u.ID), combat.PortalQty(u, cat, l.stats())
		} else {
			v.Scrolls++
			v.ScrollIDs = append(v.ScrollIDs, int(u.ID))
		}
	}
	return v, nil
}

func (l *lab) PortalList(ctx context.Context) (string, error) {
	r, err := l.be.Client.Debug(ctx, "portals")
	if err != nil {
		return "", err
	}
	var s []string
	for _, ln := range append(append([]string{}, r.Lines...), r.Rest) {
		if ln = strings.TrimSpace(ln); ln != "" && ln != "ok" {
			s = append(s, strings.ReplaceAll(ln, "\t", ","))
		}
	}
	return strings.Join(s, " / "), nil
}

// RunPortal calls the bot's portal verb once and reports what it sent and said.
func (l *lab) RunPortal(ctx context.Context) drill.BotCall {
	sp, portal, _ := l.bot()
	return l.call(ctx, sp, portal)
}

// RunPortalHit is RunPortal with a `debug hit` cheat sent delay ticks after the first right click.
func (l *lab) RunPortalHit(ctx context.Context, r *drill.Run, delay int) drill.BotCall {
	sp, portal, _ := l.bot()
	sp.after = func() {
		if delay > 0 {
			_, _ = l.be.Client.Step(ctx, delay)
		}
		_, _ = r.Cheat(ctx, "debug hit 5", "hit the player right after the portal right click: a monster blow cuts the cast (the cut-cast drill D2.d)")
	}
	return l.call(ctx, sp, portal)
}

// RunToTown calls the bot's d2dTown.ToTown (the portal verb when no portal is open next to the player, then the walk in).
func (l *lab) RunToTown(ctx context.Context) drill.BotCall {
	sp, _, t := l.bot()
	return l.call(ctx, sp, t.ToTown)
}

// RunBack calls the bot's townAdapter.BackToField.
func (l *lab) RunBack(ctx context.Context, level int) drill.BotCall {
	sp, _, t := l.bot()
	return l.call(ctx, sp, func(ctx context.Context) error { return t.BackToField(ctx, level) })
}

func (l *lab) call(ctx context.Context, sp *spy, f func(context.Context) error) drill.BotCall {
	mark, t0 := l.rec.Mark(), l.tick()
	err := f(ctx)
	c := drill.BotCall{Rights: sp.rights, Clicks: sp.clicks, Keys: sp.keys, Ticks: l.tick() - t0, ClickTick: -1, Log: l.rec.Since(mark)}
	if sp.last >= 0 {
		c.ClickTick = sp.last - t0
	}
	if err != nil {
		c.Err = err.Error()
		c.NoPortal = errors.Is(err, combat.ErrNoPortal)
	}
	return c
}

// RawCast is the harness's own cast, not the bot's: open the pack, right-click item id's box, then step ONE tick at a
// time until portal a portal object 59 appears at a position that had none before the click. It returns the ticks from the click to the object (-1: none in 40).
func (l *lab) RawCast(ctx context.Context, id int) (int, error) {
	c := l.be.Client
	if err := c.Key(ctx, "i"); err != nil {
		return -1, err
	}
	defer func() { _ = c.Key(ctx, "i"); _ = l.Tick(ctx, 5) }()
	if err := l.Tick(ctx, 18); err != nil {
		return -1, err
	}
	x, y, err := c.ItemAt(ctx, id)
	if err != nil {
		return -1, err
	}
	if err := c.Hover(ctx, x, y); err != nil {
		return -1, err
	}
	if err := l.Tick(ctx, 4); err != nil {
		return -1, err
	}
	old := map[[2]int]bool{}
	if v, err := l.PView(ctx); err == nil {
		for _, p := range v.PortalPos {
			old[p] = true
		}
	}
	t0 := l.tick()
	if err := c.RClick(ctx, x, y); err != nil {
		return -1, err
	}
	for i := 1; i <= 40; i++ {
		if err := l.Tick(ctx, 1); err != nil {
			return -1, err
		}
		v, err := l.PView(ctx)
		if err != nil {
			return -1, err
		}
		for _, p := range v.PortalPos {
			if !old[p] { // a portal at a place with none before: this cast's (the caller stands away from older ones; a recast at the same spot cannot be told apart)
				return l.tick() - t0, nil
			}
		}
	}
	return -1, nil
}

// CastSkill starts a skill cast at a cell near the player (a busy player: SC mode) and returns at once.
func (l *lab) CastSkill(ctx context.Context, skill int, dx, dy float64) error {
	s, err := l.be.Env.Snapshot()
	if err != nil || s.Player == nil {
		return fmt.Errorf("no player: %v", err)
	}
	return l.be.Client.CastCell(ctx, skill, float64(s.Player.X)/5+dx, float64(s.Player.Y)/5+dy)
}

// Walk walks to a cell offset from the player and steps (at most 600 ticks) until it is within 2 cells of the target or
// stopped for 3 reads in a row; an error when it ended farther than 4 cells away (a blocked way).
func (l *lab) Walk(ctx context.Context, dx, dy float64) error {
	s, err := l.be.Env.Snapshot()
	if err != nil || s.Player == nil {
		return fmt.Errorf("no player: %v", err)
	}
	tx, ty := float64(s.Player.X)/5+dx, float64(s.Player.Y)/5+dy
	if err := l.be.Client.CmdMove(ctx, tx, ty); err != nil {
		return err
	}
	still, lastX, lastY := 0, -1, -1
	for i := 0; i < 60; i++ {
		if err := l.Tick(ctx, 10); err != nil {
			return err
		}
		n, err := l.be.Env.Snapshot()
		if err != nil || n.Player == nil {
			continue
		}
		cx, cy := float64(n.Player.X)/5, float64(n.Player.Y)/5
		if math.Hypot(cx-tx, cy-ty) <= 2 {
			return nil
		}
		if n.Player.X == lastX && n.Player.Y == lastY {
			if still++; still >= 3 {
				break
			}
		} else {
			still = 0
		}
		lastX, lastY = n.Player.X, n.Player.Y
	}
	n, err := l.be.Env.Snapshot()
	if err != nil || n.Player == nil {
		return fmt.Errorf("no player after the walk: %v", err)
	}
	if d := math.Hypot(float64(n.Player.X)/5-tx, float64(n.Player.Y)/5-ty); d > 4 {
		return fmt.Errorf("walk stopped %.1f cells short at (%d,%d)", d, n.Player.X, n.Player.Y)
	}
	return nil
}

func (l *lab) EnterViaVerb(ctx context.Context) error { return l.be.Client.EnterPortal(ctx) }
