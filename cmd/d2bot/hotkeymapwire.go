package main

import (
	"context"
	"fmt"
	"log/slog"

	"d2lod-bot/internal/agency"
	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/hotkeymap"
)

// The generated hotkey map (knob input.hotkeyMap, default 0; docs/design/fable-review-2026-11-19/07-input-layer.md I7):
// once a game has started the hero's hotkey table is filled from the planner's slots and read back from the game state.
// No save file is read or written here; the only route is the game's own skill popup (hover, F key), which Fighter.BindSkill
// already drives and verifies. U: not run on a live game; the unit tests use a fake Binder.

// townPortalSkill is the Skills.txt row of the skill the scroll or tome of town portal casts (U: row name from memory of
// the table, 1.14d id 220; skipped with a log line when the loaded table has no such row).
const townPortalSkill = "Book of Townportal"

// popupSlots is how many slots the popup route can reach: F1 to F8 (S: the classic controls page; slots 9 to 16 have no key we drive).
const popupSlots = 8

// stateReader asks the game for its state line; d2dctl.Client satisfies it.
type stateReader interface {
	Do(ctx context.Context, line string) (*d2dctl.Reply, error)
}

// popupBinder writes through the popup route and reads the table from `state`.
type popupBinder struct {
	state stateReader
	bind  func(ctx context.Context, skillID int, key string) error
}

func (b popupBinder) Writable() int { return popupSlots }

func (b popupBinder) Bind(ctx context.Context, slot, skillID int) error {
	return b.bind(ctx, skillID, fmt.Sprintf("f%d", slot+1))
}

func (b popupBinder) Read(ctx context.Context) ([hotkeymap.Slots]int, error) {
	r, err := b.state.Do(ctx, "state")
	if err != nil {
		return [hotkeymap.Slots]int{}, err
	}
	return hotkeymap.ParseState(r.Text())
}

// applyHotkeyMap is the knob-gated step: knob off (the default) does nothing and asks nothing. The plan comes from plan
// (the agency planner on the current snapshot); a failure is logged, never fatal (the hotkeys are a convenience).
func applyHotkeyMap(ctx context.Context, e *env.Env, plan func() (agency.Plan, error), resolve func(string) (int, bool), tp string, b hotkeymap.Binder, log *slog.Logger) bool {
	if e == nil || e.Tune == nil || e.Tune.Input.HotkeyMap != 1 {
		return false
	}
	pl, err := plan()
	if err != nil {
		log.Warn("hotkey map: no plan", "err", err)
		return false
	}
	if _, ok := resolve(tp); tp != "" && !ok {
		log.Info("hotkey map: no town portal skill row, left out", "skill", tp)
		tp = ""
	}
	m, err := hotkeymap.Generate(pl, hotkeymap.Options{Resolve: resolve, TownPortal: tp})
	if err != nil {
		log.Warn("hotkey map: not generated", "err", err)
		return false
	}
	res, err := hotkeymap.Apply(ctx, b, m)
	if err != nil {
		log.Warn("hotkey map: not applied", "err", err, "slots", len(m.Slots))
		return false
	}
	log.Info("hotkey map applied", "slots", len(m.Slots), "written", res.Written, "already", res.Already, "beyond", res.Beyond)
	return true
}

// d2dHotkeyStep builds the prepare-time step for the d2d stack; nil when the pieces are missing.
func d2dHotkeyStep(e *env.Env, aw *agencyWire, c *d2dctl.Client, f *combat.Fighter, db *data.DB, log *slog.Logger) func(ctx context.Context) {
	if e == nil || aw == nil || c == nil || f == nil || db == nil {
		return nil
	}
	resolve := func(name string) (int, bool) {
		if s := db.SkillByName(name); s != nil {
			return s.ID, true
		}
		return 0, false
	}
	name := func(id int) string {
		for _, s := range db.Skills {
			if s.ID == id {
				return s.Key
			}
		}
		return ""
	}
	b := popupBinder{state: c, bind: func(ctx context.Context, id int, key string) error { return f.BindSkill(ctx, name(id), key) }}
	return func(ctx context.Context) {
		plan := func() (agency.Plan, error) {
			snap, err := e.Snapshot()
			if err != nil {
				return agency.Plan{}, err
			}
			if snap.Player == nil {
				return agency.Plan{}, fmt.Errorf("no player in the snapshot")
			}
			_, pl, err := aw.eval(snap)
			return pl, err
		}
		applyHotkeyMap(ctx, e, plan, resolve, townPortalSkill, b, log)
	}
}
