package main

import (
	"d2lod-bot/internal/loot"
	"d2lod-bot/internal/watch"
)

// lootShadowAPI mounts GET /api/lootrules/shadow (docs/lootrules.md, wave 1). The report comes from the shadow every
// looter and town chore of the process feeds (loot.DefaultShadow); enabled mirrors the knob bot.lootShadow when the
// tuner is wired (true, the default, when it is not).
func lootShadowAPI(knob func(string) (float64, string, error)) *watch.LootShadowAPI {
	api := watch.NewLootShadowAPI()
	api.Bind(func() any {
		on := true
		if knob != nil {
			if v, _, err := knob("bot.lootShadow"); err == nil {
				on = v != 0
			}
		}
		return loot.DefaultShadow.Report(on)
	})
	return api
}
