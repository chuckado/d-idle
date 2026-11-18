package main

import (
	"bytes"
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
	hero := loot.LastHeroClass()
	d.st.UseClassFallback(hero) // a hero without a saved set starts from the preset of its class
	v := d.st.View(stage, loot.DefaultShadow.Recent())
	v.Hero = hero
	return v
}

// Preview is the dry run of the sheet: the body is a PUT body (or empty for the active set); nothing is saved.
func (d lootRulesDesk) Preview(body []byte) (any, error) {
	var cand *lootrules.RuleSet
	if t := bytes.TrimSpace(body); len(t) > 0 && string(t) != "{}" {
		var err error
		if cand, err = d.st.Build(body); err != nil {
			return nil, err
		}
	}
	return loot.PreviewRules(cand)
}

func (d lootRulesDesk) Undo(seq int) (any, error)             { return d.st.Undo(seq) }
func (d lootRulesDesk) Import(b []byte, seq int) (any, error) { return d.st.Import(b, seq) }
func (d lootRulesDesk) Export() ([]byte, error)               { return d.st.Export() }

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
