package main

import (
	"path/filepath"
	"strings"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/loot"
	"d2lod-bot/internal/lootrules"
	"d2lod-bot/internal/watch"
)

// lootRulesDesk is the bot side of /api/lootrules (docs/lootrules.md, wave 2a): the player's rule set on disk, handed to
// the engine and the shadow stage on start and after every save.
type lootRulesDesk struct {
	st   *lootrules.Store
	knob func(string) (float64, string, error)
}

func (d lootRulesDesk) View() any {
	stage := 0
	if d.knob != nil {
		if v, _, err := d.knob("loot.rulesEngine"); err == nil {
			stage = int(v)
		}
	}
	return d.st.View(stage, loot.DefaultShadow.Recent())
}

func (d lootRulesDesk) Put(b []byte) (any, error) {
	rs, err := d.st.Put(b)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

// heroClasses are the class names the per-class presets are made for, read from the d2 package (no list here).
func heroClasses() []string {
	var out []string
	for c := d2.Class(0); ; c++ {
		n := c.String()
		if strings.HasPrefix(n, "Class(") {
			return out
		}
		out = append(out, n)
	}
}

// lootRulesAPI opens the store beside the other state files and activates its set.
func lootRulesAPI(stateDir, profile string, knob func(string) (float64, string, error)) *watch.LootRulesAPI {
	st := lootrules.OpenStore(filepath.Join(stateDir, "lootrules-"+profile+".json"), heroClasses())
	loot.SetActiveRules(st.Active())
	st.OnChange(loot.SetActiveRules)
	api := watch.NewLootRulesAPI()
	api.Bind(lootRulesDesk{st, knob})
	return api
}
