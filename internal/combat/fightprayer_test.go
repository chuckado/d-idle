package combat

import (
	"testing"

	"d2lod-bot/internal/data"
)

// Prayer in a fight: life under the profile's fightLifePct with the belt at or under maxPotions healing potions. It
// holds until life is fightAuraHysteresis above, never on unknown potions, never with potions in the belt, never
// without a target (that is the resting rule). Both Paladin profiles carry the row (default and hammerdin).
func TestPrayerInFightWhenHurtAndPotionsLow(t *testing.T) {
	for _, build := range []Build{"", "hammerdin"} {
		h, g := newHandGame(t)
		h.f.Build = build
		cat := palCatalog()
		cat.Skills[99] = &data.Skill{ID: 99, Key: "Prayer", Name: "Prayer", Class: "pal", Range: "none", Aura: true, ReqLevel: 6}
		h.f.SetCatalog(cat)
		p := g.p
		p.Level = 20
		p.Skills = skillsOf(98, 102, 99, 112)
		h.f.SetHotkeys(map[int]string{98: "f1", 102: "f2", 99: "f8"})
		pr := h.f.profileFor(p)
		if len(pr.Aura) == 0 || pr.Aura[0].FightLifePct == 0 {
			t.Fatalf("%q: no fight Prayer row: %+v", build, pr.Aura)
		}
		p.MaxLife = 100
		h.f.auraCrowd = 1
		tg := &Target{}
		pick := func(life, potions int, known bool, tg *Target) string {
			p.Life = life
			h.f.auraPotions, h.f.auraPotKnown = potions, known
			sp, _ := h.f.wantAura(p, pr, tg, false)
			return sp.Name
		}
		for _, c := range []struct {
			life, pots int
			known      bool
			tg         *Target
			want       string
		}{
			{80, 0, true, tg, "Might"},  // healthy
			{40, 1, true, tg, "Prayer"}, // hurt, potions low
			{40, 5, true, tg, "Might"},  // hurt but potions in the belt
			{40, 0, false, tg, "Might"}, // potions unknown
			{40, 0, true, nil, "Might"}, // no fight
		} {
			if got := pick(c.life, c.pots, c.known, c.tg); got != c.want {
				t.Errorf("%q life %d potions %d known %v: %s, want %s", build, c.life, c.pots, c.known, got, c.want)
			}
		}
		// hysteresis: with Prayer up it stays up to FightLifePct+15, then leaves
		g.aura, p.RightSkill = 99, 99
		if got := pick(60, 1, true, tg); got != "Prayer" {
			t.Errorf("%q hysteresis: %s", build, got)
		}
		if got := pick(70, 1, true, tg); got == "Prayer" {
			t.Errorf("%q still Prayer at 70", build)
		}
	}
}
