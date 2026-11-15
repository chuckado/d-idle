package watch

import (
	"encoding/json"
	"testing"
)

// uic3: /api/hero stats gains attrs, defense and damage. Each is absent (not zero) when the backend does not read it.
func TestHeroStatsSheetFields(t *testing.T) {
	for _, build := range []string{"sorceress", "paladin"} {
		r := newRig(t, build)
		r.sim.Hub.SetHeroStatsSource(func() (HeroStats, bool) { return HeroStats{Resists: HeroResists{Fire: 1}}, true })
		var raw map[string]json.RawMessage
		r.get(HeroRoute, &raw)
		var st map[string]json.RawMessage
		if err := json.Unmarshal(raw["stats"], &st); err != nil {
			t.Fatal(err)
		}
		for _, k := range []string{"attrs", "defense", "damage"} {
			if _, ok := st[k]; ok {
				t.Errorf("%s: %s present although not read", build, k)
			}
		}
		d := 214
		r.sim.Hub.SetHeroStatsSource(func() (HeroStats, bool) {
			return HeroStats{Attrs: &HeroAttrs{Str: 38, Dex: 25, Vit: 30, Energy: 20}, Defense: &d, Damage: &HeroDamage{Min: 4, Max: 9}}, true
		})
		var v HeroView
		r.get(HeroRoute, &v)
		if v.Stats == nil || v.Stats.Attrs == nil || v.Stats.Attrs.Vit != 30 || v.Stats.Defense == nil || *v.Stats.Defense != 214 || v.Stats.Damage == nil || v.Stats.Damage.Max != 9 {
			t.Fatalf("%s: %+v", build, v.Stats)
		}
	}
}
