package power

import (
	"testing"

	"d2lod-bot/internal/danger"
)

// A boss is judged like a pack: a stronger hero never reads worse, a weak one reads too hard, a stronger boss never reads
// safer. Both builds; an unknown dial is an error.
func TestBossWordTwoBuilds(t *testing.T) {
	m := model(t)
	rank := map[string]int{"too hard": 0, "risky": 1, "safe": 2}
	for _, tc := range []struct {
		name string
		hero func() Hero
		arch string
	}{{"paladin", paladin, "bruiser"}, {"sorceress", sorceress, "caster"}} {
		t.Run(tc.name, func(t *testing.T) {
			a := arch(t, m, tc.arch)
			boss := &danger.Boss{Key: "andariel", Level: 12, Life: 1500, Def: 100, Dmg: 30, ToHit: 150}
			weak, strong := tc.hero(), tc.hero()
			weak.MaxLife = 40
			strong.MaxLife = 2000
			ww, _, err := m.BossWord(weak, a, boss, "balanced")
			if err != nil {
				t.Fatal(err)
			}
			sw, _, err := m.BossWord(strong, a, boss, "balanced")
			if err != nil {
				t.Fatal(err)
			}
			if ww != "too hard" || rank[sw] < rank[ww] {
				t.Errorf("weak %q strong %q", ww, sw)
			}
			hard := *boss
			hard.Life, hard.Dmg = boss.Life*4, boss.Dmg*4
			hw, _, _ := m.BossWord(strong, a, &hard, "balanced")
			if rank[hw] > rank[sw] {
				t.Errorf("a stronger boss reads safer: %q vs %q", hw, sw)
			}
			if _, _, err := m.BossWord(strong, a, boss, "reckless"); err == nil {
				t.Error("unknown dial accepted")
			}
		})
	}
}
