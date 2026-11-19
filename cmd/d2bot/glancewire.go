package main

import (
	"time"

	"fmt"

	"d2lod-bot/internal/bot"
	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/goals"
	"d2lod-bot/internal/leveler"
	"d2lod-bot/internal/merc"
	"d2lod-bot/internal/points"
	"d2lod-bot/internal/questboard"
	"d2lod-bot/internal/resource"
	"d2lod-bot/internal/watch"
)

// The Hub sources behind /api/glance that only the bot can fill (docs/ui/redesign-2026-11-14, B5): the Mercenary light
// and the Supplies light. Both are read-only views of what the runner already holds; nothing is polled from the game.
// Other cards add their own source (the Needs rows of B6) through glanceExtra, so one SetGlanceSources call serves all.

// glanceExtra lets a card extend the sources without editing this file (B6 appends in an init func).
var glanceExtra []func(*watch.GlanceSources, glanceDeps)

// glanceDeps is what an extension may read: the runner status, the pause and reset views and the quest board desk.
type glanceDeps struct {
	Status func() goals.Status
	Pause  func() string // the player pause state (bot.PauseRunning, PauseGoingHome, PausePaused)
	PauseV func() bot.PlayerPauseView
	Reset  func() bot.ResetView
	Stuck  func() time.Duration    // how long the bot made no progress (the stall detector), 0 = not stuck
	Beat   func() time.Time        // the loop's last finished iteration
	Quests func() *questboard.Desk // the quest board desk (questwire.go); nil result: none
	Points func() *points.Desk     // the points desk (pointswire.go); nil result: none
}

// useGlance binds the bot's sources to the Hub. No rig, no Hub or no runner: nothing (the lights stay "unknown").
func useGlance(rig *ladderRig, r *bot.Runner, db *data.DB, levelerBuild string, ref *runnerRef) {
	if rig == nil || rig.Hub == nil || r == nil {
		return
	}
	var cat combat.Catalog // without game tables the Supplies light stays unknown
	if db != nil {
		cat = combat.NewCatalog(db)
	}
	src := watch.GlanceSources{
		Merc:     glanceMerc(mercAPIAdapter{}.view),
		Supplies: glanceSupplies(r.LastSnapshot, cat, db, levelerBuild),
		Intake: func() (float64, int, bool) {
			in, ok := r.Intake()
			return in.LifeLostPerMin, in.WindowMinutes, ok
		},
	}
	deps := glanceDeps{Status: r.Status,
		Pause: func() string { return r.PlayerPauseView().State }, PauseV: r.PlayerPauseView,
		Reset: r.ResetView, Stuck: r.StuckFor, Beat: r.LastStep, Quests: func() *questboard.Desk {
			if ref == nil {
				return nil
			}
			return ref.qd
		}, Points: func() *points.Desk {
			if ref == nil {
				return nil
			}
			return ref.pd
		}}
	for _, f := range glanceExtra {
		f(&src, deps)
	}
	rig.Hub.SetGlanceSources(src)
}

// glanceMerc reads the merc revive desk's own verdict (alive, dead, none; the same state /api/merc reports). While the
// desk is not running or has not looked yet the light is unknown, never green.
func glanceMerc(view func() (merc.PlayerView, *mercWire)) func() (watch.GlanceLight, bool) {
	return func() (watch.GlanceLight, bool) {
		v, w := view()
		if w == nil {
			return watch.GlanceLight{}, false
		}
		switch v.State {
		case "alive":
			return watch.GlanceLight{Light: "ok", Word: "Alive"}, true
		case "dead":
			return watch.GlanceLight{Light: "bad", Word: "Dead", Detail: "The mercenary is dead and the hero fights without him"}, true
		case "none":
			return watch.GlanceLight{Light: "warn", Word: "None", Detail: "No mercenary is hired"}, true
		}
		return watch.GlanceLight{}, false
	}
}

// glanceKit is what the Supplies light reads: the belt as the runner sees it and the plan of the hero's build.
type glanceKit struct {
	Class      d2.Class
	Level      int
	Diff       d2.Difficulty
	Build      string // resolved leveler table, "" when the class has none
	Slots      int    // potion slots of the worn belt
	Heal, Mana int    // belt potions: healing plus rejuvenation, mana
	Portal     bool   // a town portal scroll or a non-empty tome in the pack
}

// supplyLight compares the kit with the belt plan of the hero's build: green at plan, amber below, red at none (no
// healing potion where the plan has a healing column, or no portal source). The plan numbers are the plan's columns
// times the belt rows (resource.PlanForBuild), so a class or build with another plan reads its own numbers.
func supplyLight(k glanceKit) watch.GlanceLight {
	plan := resource.PlanForBuild(k.Class, k.Level, k.Diff, k.Build)
	rows := max(k.Slots, 4) / 4
	var wantHeal, wantMana int
	for _, c := range plan.Cols {
		switch c {
		case resource.PotionLife, resource.PotionRejuv:
			wantHeal += rows
		case resource.PotionMana:
			wantMana += rows
		}
	}
	detail := fmt.Sprintf("healing %d of %d, mana %d of %d, portal %s", k.Heal, wantHeal, k.Mana, wantMana, yesNo(k.Portal))
	switch {
	case !k.Portal:
		return watch.GlanceLight{Light: "bad", Word: "No portal", Detail: detail}
	case wantHeal > 0 && k.Heal == 0:
		return watch.GlanceLight{Light: "bad", Word: "No potions", Detail: detail}
	case k.Heal < wantHeal || k.Mana < wantMana:
		return watch.GlanceLight{Light: "warn", Word: "Low", Detail: detail}
	}
	return watch.GlanceLight{Light: "ok", Word: "Stocked", Detail: detail}
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// glanceSupplies reads the newest snapshot; a nil snapshot, no catalog or a snapshot outside the game is "unknown".
func glanceSupplies(snap func() *d2.Snapshot, cat combat.Catalog, db *data.DB, levelerBuild string) func() (watch.GlanceLight, bool) {
	return func() (watch.GlanceLight, bool) {
		s := snap()
		if s == nil || s.Player == nil || s.Inventory == nil || cat == nil {
			return watch.GlanceLight{}, false
		}
		k := glanceKit{Class: s.Player.Class, Level: s.Player.Level, Diff: s.Player.Difficulty,
			Build: leveler.ResolveBuild(levelerBuild, s.Player.Class.String()),
			Slots: combat.BeltCapacity(s.Inventory, db), Portal: combat.PortalSource(s, cat, nil) != nil}
		b := combat.BeltOf(s.Inventory, cat)
		for _, c := range b.Cols {
			switch c.Kind {
			case combat.PotionHealing, combat.PotionRejuv:
				k.Heal += c.Count
			case combat.PotionMana:
				k.Mana += c.Count
			}
		}
		return supplyLight(k), true
	}
}
