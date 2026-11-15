package main

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/danger"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/leveler"
	"d2lod-bot/internal/plandata"
	"d2lod-bot/internal/power"
)

var (
	dangerOnce sync.Once
	dangerTab  *danger.Table
)

// dangerTable builds the zone profile once per process (the API and the shadow line share it); nil without tables.
func dangerTable(db *data.DB) *danger.Table {
	if db == nil {
		return nil
	}
	dangerOnce.Do(func() {
		ov, err := danger.ShippedOverrides()
		if err != nil {
			slog.Warn("danger: bad overrides", "err", err)
			return
		}
		if dangerTab, err = danger.Build(db, danger.DefaultParams(), ov); err != nil {
			slog.Warn("danger: no profile", "err", err)
		}
	})
	return dangerTab
}

// classAttacks adapts the combat profiles (the attack bands per class) to the archetype resolver.
func classAttacks(class string) ([]power.Band, bool, error) {
	var pick *combat.Profile
	for _, p := range combat.Profiles() {
		if strings.EqualFold(p.Class, class) && (pick == nil || (pick.Build != "" && p.Build == "")) {
			pf := p
			pick = &pf
		}
	}
	if pick == nil {
		return nil, false, fmt.Errorf("no combat profile for class %q", class)
	}
	bs := make([]power.Band, 0, len(pick.Bands))
	for _, b := range pick.Bands {
		bs = append(bs, power.Band{MinLevel: b.MinLevel, Requires: b.Requires, Attack: b.Attack})
	}
	return bs, pick.Melee != nil, nil
}

// wirePower returns the ZoneShadow hook (docs/survival/power.md): the margin of the hero against the zone the planner
// chose, as one log line. Shadow only; without tables or on any error it logs a single warning per distinct reason and
// returns no line. The build is the one the leveler picks when the config names none (the operator's explicit table is
// not visible here: U).
func wirePower(db *data.DB) func(s *d2.Snapshot, area int, diff string, potions int) string {
	tab := dangerTable(db)
	if tab == nil {
		return nil
	}
	m, err := power.NewModel(tab.Params)
	if err != nil {
		slog.Warn("power: no model", "err", err)
		return nil
	}
	warned := map[string]bool{}
	warn := func(err error) string {
		if k := err.Error(); !warned[k] {
			warned[k] = true
			slog.Warn("zone shadow: no margin", "err", err)
		}
		return ""
	}
	return func(s *d2.Snapshot, area int, diff string, potions int) string {
		p := s.Player
		z, err := tab.Find(area, diff)
		if err != nil {
			return "" // a town or an area the profile does not cover: nothing to say
		}
		hero, a, err := heroOf(m, db, s, potions, warn)
		if err != nil {
			return warn(err)
		}
		zm, err := m.Zone(hero, a, z, m.P.DefaultDial)
		if err != nil {
			return warn(err)
		}
		return zm.ShadowLine(p.Level)
	}
}

// heroOf reads the hero and puts the build's main attack on it for the skill damage estimate. An attack the data does
// not know is reported through note and leaves the constants in place (flagged by the estimate).
func heroOf(m *power.Model, db *data.DB, s *d2.Snapshot, potions int, note func(error) string) (power.Hero, *power.Archetype, error) {
	a, spec, err := archetypeOf(m, db, s.Player)
	if err != nil {
		return power.Hero{}, nil, err
	}
	hero := power.HeroFromPlayer(s.Player, s.Merc, potions)
	if err := m.Attach(&hero, spec); err != nil {
		note(err)
	}
	return hero, a, nil
}

// archetypeOf resolves the hero's archetype from the build the leveler picks (data, no class names here).
func archetypeOf(m *power.Model, db *data.DB, p *d2.Player) (*power.Archetype, power.BuildSpec, error) {
	build := leveler.SelectBuild(strings.ToLower(p.Class.String()), p.Level, leveler.OwnedSkills(db, p))
	spec, err := power.SpecFromBuild(plandata.FS, "build-"+build+".json", p.Level, classAttacks)
	if err != nil {
		return nil, spec, err
	}
	a, err := m.P.Resolve(spec)
	return a, spec, err
}
