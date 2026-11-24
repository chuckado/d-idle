package power

import (
	"fmt"
	"os"
	"testing"

	"d2lod-bot/internal/danger"
	"d2lod-bot/internal/data"
)

func TestFloorsReport(t *testing.T) {
	dir := os.Getenv("D2_MPQ_DIR")
	if dir == "" {
		t.Skip("no install")
	}
	db, err := data.Open(dir, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ov, err := danger.ShippedOverrides()
	if err != nil {
		t.Fatal(err)
	}
	dp := danger.DefaultParams()
	tb, err := danger.Build(db, dp, ov)
	if err != nil {
		t.Fatal(err)
	}
	m, err := NewModel(dp)
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("D2_FLOORS_VERBOSE") == "" {
		goto floors
	}
	for _, a := range []int{5, 6, 21, 24, 27, 28, 29, 34, 35, 36, 37} {
		z, err := tb.Find(a, "normal")
		if err != nil {
			fmt.Println("area", a, err)
			continue
		}
		fmt.Printf("ZONE %d %s mlvl=%d uniques=%.1f typical=%+v bad=%+v elite=%+v\n", a, z.Name, z.MLvl, z.Uniques, z.Encounters.Typical, z.Encounters.Bad, z.Encounters.Elite)
		for _, mo := range z.Monsters {
			if mo.Share >= 0.08 {
				fmt.Printf("   mon %s share=%.2f life=%d def=%d toHit=%d dmg=%.1f dps=%.1f reach=%s\n", mo.Key, mo.Share, mo.Life, mo.Def, mo.ToHit, mo.Dmg, mo.Dps, mo.Reach)
			}
		}
	}
	for _, k := range []string{"andariel", "the_smith", "smith"} {
		if b, err := tb.FindBoss(k, "normal"); err == nil {
			fmt.Printf("BOSS %s level=%d life=%d def=%d dmg=%.1f toHit=%d minions=%d mods=%v\n", b.Name, b.Level, b.Life, b.Def, b.Dmg, b.ToHit, b.Minions, b.Mods)
		}
	}
floors:
	type gear struct {
		name       string
		def, block int
		res        int
	}
	gears := []gear{{"as-played", 30, 15, 0}, {"plan-class", 140, 30, 25}}
	enter := 1.5 // the balanced dial (data/questboard/readiness.json)
	hero := func(lvl int, g gear) Hero {
		h := Hero{Class: "Paladin", Level: lvl, MaxLife: 279 + 11*(lvl-20), MaxMana: 44 + 4*(lvl-20), Defence: g.def, BlockPct: g.block, Resist: [4]int{g.res, g.res, g.res, g.res},
			AR: 60 + 12*lvl, DmgMin: 3 + lvl, DmgMax: 8 + 2*lvl, DmgPct: 20, Potions: 6, Merc: Merc{Hired: true, Alive: true, Level: lvl}}
		for _, term := range []string{TermLife, TermDefence, TermBlock, TermResist, TermOffence, TermPotions, TermMerc} {
			h.SetQuality(term, Measured)
		}
		return h
	}
	// the lowest clvl whose margin, divided by k (the ledger's hero correction), reaches the line; 0 = none up to 60
	floor := func(arch string, g gear, k float64, margin func(Hero, *Archetype) (best, low float64)) int {
		a, err := m.P.Archetype(arch)
		if err != nil {
			t.Fatal(err)
		}
		for lvl := 8; lvl <= 60; lvl++ {
			_, low := margin(hero(lvl, g), a)
			if low/k >= enter {
				return lvl
			}
		}
		return 0
	}
	andy, err := tb.FindBoss("andariel", "normal")
	if err != nil {
		t.Fatal(err)
	}
	zoneMargin := func(area int) func(Hero, *Archetype) (float64, float64) {
		z, err := tb.Find(area, "normal")
		if err != nil {
			t.Fatal(err)
		}
		return func(h Hero, a *Archetype) (float64, float64) {
			zm, err := m.Zone(h, a, z, "balanced")
			if err != nil {
				t.Fatal(err)
			}
			return zm.Bad.Best, zm.Bad.Low
		}
	}
	bossMargin := func(h Hero, a *Archetype) (float64, float64) {
		_, r, err := m.BossWord(h, a, andy, "balanced")
		if err != nil {
			t.Fatal(err)
		}
		return r.Best, r.Low
	}
	fmt.Printf("FLOORS clvl at which low/k >= enter %.1f (balanced); k = hero correction of the ledger\n", enter)
	for _, arch := range []string{"bruiser", "caster"} {
		for _, g := range gears {
			for _, k := range []float64{1, 1.5, 2, 3} {
				fmt.Printf("FLOORS %-7s %-10s k=%.1f barracks=%d outer-cloister=%d jail1=%d andariel=%d\n", arch, g.name, k,
					floor(arch, g, k, zoneMargin(28)), floor(arch, g, k, zoneMargin(27)), floor(arch, g, k, zoneMargin(29)), floor(arch, g, k, bossMargin))
			}
		}
	}
}

// TestAndarielDecisionFixture prints the board word for Andariel at clvl 8 to 20 as JSON lines (card toolstrade48,
// docs/probes/andariel-gate-recalibration-2026-11-24.md): archetype, gear class, k (the ledger's hero correction), dial.
// Needs the install; D2_FIXTURE=1.
func TestAndarielDecisionFixture(t *testing.T) {
	dir := os.Getenv("D2_MPQ_DIR")
	if dir == "" || os.Getenv("D2_FIXTURE") == "" {
		t.Skip("no install or no D2_FIXTURE")
	}
	db, err := data.Open(dir, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ov, _ := danger.ShippedOverrides()
	dp := danger.DefaultParams()
	tb, err := danger.Build(db, dp, ov)
	if err != nil {
		t.Fatal(err)
	}
	m, err := NewModel(dp)
	if err != nil {
		t.Fatal(err)
	}
	andy, err := tb.FindBoss("andariel", "normal")
	if err != nil {
		t.Fatal(err)
	}
	type gear struct {
		name       string
		def, block int
		res        int
	}
	gears := []gear{{"as-played", 30, 15, 0}, {"plan-class", 140, 30, 25}}
	dials := map[string][2]float64{"balanced": {1.5, 1.2}, "bold": {1.2, 1.0}}
	for _, cls := range []struct{ name, arch string }{{"Paladin", "bruiser"}, {"Necromancer", "caster"}} {
		a, err := m.P.Archetype(cls.arch)
		if err != nil {
			t.Fatal(err)
		}
		for _, g := range gears {
			for _, k := range []float64{1, 1.5, 2, 3} {
				for _, lvl := range []int{8, 9, 10, 11, 12, 14, 17, 20} {
					h := Hero{Class: cls.name, Level: lvl, MaxLife: 279 + 11*(lvl-20), MaxMana: 44 + 4*(lvl-20), Defence: g.def, BlockPct: g.block, Resist: [4]int{g.res, g.res, g.res, g.res},
						AR: 60 + 12*lvl, DmgMin: 3 + lvl, DmgMax: 8 + 2*lvl, DmgPct: 20, Potions: 6, Merc: Merc{Hired: true, Alive: true, Level: lvl}}
					for _, term := range []string{TermLife, TermDefence, TermBlock, TermResist, TermOffence, TermPotions, TermMerc} {
						h.SetQuality(term, Measured)
					}
					_, r, err := m.BossWord(h, a, andy, "balanced")
					if err != nil {
						t.Fatal(err)
					}
					for _, dn := range []string{"balanced", "bold"} {
						dl := dials[dn]
						w := "too_hard"
						switch {
						case r.Low/k >= dl[0]:
							w = "safe"
						case r.Best/k >= dl[1]:
							w = "risky"
						}
						fmt.Printf("FIXTURE {\"class\":%q,\"clvl\":%d,\"gear\":%q,\"k\":%.1f,\"dial\":%q,\"best\":%.2f,\"low\":%.2f,\"word\":%q}\n", cls.name, lvl, g.name, k, dn, r.Best/k, r.Low/k, w)
					}
				}
			}
		}
	}
}
