package main

import (
	"os"

	"d2lod-bot/internal/bot"
)

// gearDecisionsKey is the one knob behind the player-only gear policy (docs/gear-policy.md): 0 = the player decides
// equipment (the default), 1 = the bot does (legacy behaviour).
const gearDecisionsKey = "resource.gearDecisions"

// pointsDecisionsKey is the knob behind who spends stat and skill points (docs/points.md): 0 = the player (the default),
// 1 = the bot from its plan table.
const pointsDecisionsKey = "resource.pointsDecisions"

// withGearPolicy lays gear.decisions = bot and points.decisions = bot under cfg when the process is a measured or drill
// run (D2BOT_RUN is set) or D2BOT_GEAR=bot says so, so gates, soaks and drills keep the old equip, swap, buy and
// spend-the-points paths. An explicit tuning key of the config wins, per key. A copy: cfg.Tuning is not changed.
func withGearPolicy(cfg bot.Config) bot.Config {
	if os.Getenv("D2BOT_RUN") == "" && os.Getenv("D2BOT_GEAR") != "bot" {
		return cfg
	}
	m := make(map[string]float64, len(cfg.Tuning)+2)
	for k, v := range cfg.Tuning {
		m[k] = v
	}
	for _, k := range []string{gearDecisionsKey, pointsDecisionsKey} {
		if _, set := m[k]; !set {
			m[k] = 1
		}
	}
	cfg.Tuning = m
	return cfg
}
