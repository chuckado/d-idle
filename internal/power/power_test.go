package power

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/danger"
	"d2lod-bot/internal/plandata"
)

func model(t *testing.T) *Model {
	t.Helper()
	m, err := NewModel(danger.DefaultParams())
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// zone is a made-up area (real values: docs/survival/power.md, run on the worker host against the install).
func zone(mlvl, melee, ranged int, life, dmg float64) *danger.Zone {
	mon := func(reach string) danger.Monster {
		return danger.Monster{Key: reach, Share: 0.5, Life: int(life), Def: 30, ToHit: 60 + 5*mlvl, Dmg: dmg, Reach: reach}
	}
	return &danger.Zone{Area: 9, Name: "Test Field", Diff: "normal", MLvl: mlvl,
		Monsters: []danger.Monster{mon("melee"), mon("ranged")},
		Encounters: danger.Encounters{
			Typical: danger.Encounter{N: melee + ranged, Melee: melee, Ranged: ranged},
			Bad:     danger.Encounter{N: 2 * (melee + ranged), Melee: 2 * melee, Ranged: 2 * ranged},
			Elite:   danger.Encounter{N: melee + ranged + 1, Melee: melee + 1, Ranged: ranged, Leader: "unique"}}}
}

func paladin() Hero {
	h := Hero{Class: "Paladin", Level: 8, MaxLife: 160, MaxMana: 40, Defence: 80, BlockPct: 30, AR: 90, DmgMin: 6, DmgMax: 12, DmgPct: 20,
		Potions: 6, Resist: [4]int{10, 10, 10, 10}, Attack: Attack{WeaponClass: "1HS"}}
	for _, term := range []string{TermLife, TermDefence, TermBlock, TermResist, TermOffence, TermPotions} {
		h.SetQuality(term, Measured)
	}
	return h
}

func sorceress() Hero {
	h := Hero{Class: "Sorceress", Level: 8, MaxLife: 90, MaxMana: 120, Defence: 20, MeasuredDPS: 14, Potions: 6}
	for _, term := range []string{TermLife, TermDefence, TermOffence, TermPotions, TermResist} {
		h.SetQuality(term, Measured)
	}
	return h
}

func arch(t *testing.T, m *Model, id string) *Archetype {
	a, err := m.P.Archetype(id)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func marginOf(t *testing.T, m *Model, h Hero, a *Archetype, z *danger.Zone) float64 {
	t.Helper()
	zm, err := m.Zone(h, a, z, "balanced")
	if err != nil {
		t.Fatal(err)
	}
	return zm.Bad.Best
}

// TestMarginMonotone: more life, more damage, a living merc and a fuller belt never lower a margin; a bigger pack and a
// harder-hitting or sturdier monster never raise it. Run for the Paladin (bruiser) and the Sorceress (caster).
func TestMarginMonotone(t *testing.T) {
	m := model(t)
	for _, c := range []struct {
		name string
		h    Hero
		a    string
	}{{"paladin", paladin(), "bruiser"}, {"sorceress", sorceress(), "caster"}} {
		t.Run(c.name, func(t *testing.T) {
			a := arch(t, m, c.a)
			z := zone(8, 3, 1, 60, 10)
			base := marginOf(t, m, c.h, a, z)
			up := func(name string, mod func(*Hero)) {
				h := c.h
				mod(&h)
				if got := marginOf(t, m, h, a, z); got < base-1e-9 {
					t.Errorf("%s lowered the margin: %.3f -> %.3f", name, base, got)
				}
			}
			up("life", func(h *Hero) { h.MaxLife += 50 })
			up("potions", func(h *Hero) { h.Potions += 4 })
			up("mana", func(h *Hero) { h.MaxMana += 60 })
			up("defence", func(h *Hero) { h.Defence += 100 })
			up("block", func(h *Hero) { h.BlockPct += 20 })
			up("resist", func(h *Hero) { h.Resist = [4]int{40, 40, 40, 40} })
			up("sustain", func(h *Hero) { h.Sustain = 2 })
			up("damage", func(h *Hero) { h.DmgMax *= 2; h.MeasuredDPS *= 2 })
			up("merc", func(h *Hero) { h.Merc = Merc{Hired: true, Alive: true, Level: 8}; h.SetQuality(TermMerc, Measured) })
			down := func(name string, zz *danger.Zone) {
				if got := marginOf(t, m, c.h, a, zz); got > base+1e-9 {
					t.Errorf("%s raised the margin: %.3f -> %.3f", name, base, got)
				}
			}
			down("pack", zone(8, 4, 2, 60, 10))
			down("damage", zone(8, 3, 1, 60, 14))
			down("life", zone(8, 3, 1, 90, 10))
			down("level", zone(14, 3, 1, 60, 10))
		})
	}
}

// TestDeadMercLowersMargin: the merc term is the difference between an alive and a dead hireling.
func TestDeadMercLowersMargin(t *testing.T) {
	m := model(t)
	a := arch(t, m, "bruiser")
	z := zone(8, 4, 1, 80, 12)
	alive, dead := paladin(), paladin()
	alive.Merc = Merc{Hired: true, Alive: true, Level: 8}
	dead.Merc = Merc{Hired: true, Alive: false, Level: 8}
	if ma, md := marginOf(t, m, alive, a, z), marginOf(t, m, dead, a, z); ma <= md {
		t.Errorf("alive %.2f not above dead %.2f", ma, md)
	}
}

// TestPackSquare: danger grows faster than linearly with the pack (design 5.1, n(n+1)/2).
func TestPackSquare(t *testing.T) {
	m := model(t)
	a := arch(t, m, "bruiser")
	h := paladin()
	h.MaxLife, h.Potions = 100000, 0
	taken := func(n int) float64 {
		pk, _ := PackOf(zone(8, n, 0, 60, 10), danger.Encounter{N: n, Melee: n}, 1, 1)
		r, err := m.Margin(h, a, pk)
		if err != nil {
			t.Fatal(err)
		}
		return r.Terms.Taken
	}
	if t2, t4 := taken(2), taken(4); t4 < 2.5*t2 {
		t.Errorf("4 attackers take %.1f, 2 take %.1f: less than the quadratic growth", t4, t2)
	}
}

// TestUnreadDegrades: a term that cannot be read lowers the low estimate and says so, never invents a number.
func TestUnreadDegrades(t *testing.T) {
	m := model(t)
	a := arch(t, m, "bruiser")
	z := zone(8, 3, 1, 60, 10)
	read := paladin()
	unread := paladin()
	unread.Defence = 0
	unread.SetQuality(TermDefence, Default)
	unread.SetQuality(TermOffence, Default)
	unread.DmgMax, unread.DmgMin = 0, 0
	zr, _ := m.Zone(read, a, z, "balanced")
	zu, _ := m.Zone(unread, a, z, "balanced")
	if zu.Bad.Low >= zu.Bad.Best {
		t.Errorf("low %.2f not under best %.2f with unread terms", zu.Bad.Low, zu.Bad.Best)
	}
	if zr.Bad.Low != zr.Bad.Best {
		// resist and potions are measured in this hero, so nothing is discounted
		t.Errorf("fully measured hero: low %.2f differs from best %.2f", zr.Bad.Low, zr.Bad.Best)
	}
	if got := zu.Weakest(); got != "defence_default" {
		t.Errorf("weakest %q, want defence_default", got)
	}
	if zr.Weakest() != "" {
		t.Errorf("measured hero has weakest %q", zr.Weakest())
	}
	if _, err := m.Margin(Hero{Class: "Paladin"}, a, Pack{}); err == nil {
		t.Error("a hero without max life must be an error")
	}
}

func TestWordsAndDials(t *testing.T) {
	m := model(t)
	a := arch(t, m, "bruiser")
	h := paladin()
	for _, c := range []struct {
		z    *danger.Zone
		want string
	}{{zone(4, 1, 0, 20, 3), "safe"}, {zone(30, 8, 4, 400, 60), "too hard"}} {
		zm, _ := m.Zone(h, a, c.z, "balanced")
		if zm.Word() != c.want {
			t.Errorf("word %q (bad %.2f low %.2f), want %q", zm.Word(), zm.Bad.Best, zm.Bad.Low, c.want)
		}
	}
	z := zone(8, 3, 1, 60, 10)
	prev := 1e9
	for _, id := range []string{"cautious", "balanced", "bold"} {
		zm, err := m.Zone(h, a, z, id)
		if err != nil {
			t.Fatal(err)
		}
		if zm.Enter >= prev {
			t.Errorf("dial %s enter %.2f not below the previous", id, zm.Enter)
		}
		prev = zm.Enter
	}
	if _, err := m.Zone(h, a, z, "reckless"); err == nil {
		t.Error("unknown dial must be an error")
	}
}

// TestTwoBuilds: the same zone gives the two archetypes different margins in the direction the model says: a caster
// with a small pool and no defence is hurt by a ranged-heavy pack more than a bruiser with defence and block.
func TestTwoBuilds(t *testing.T) {
	m := model(t)
	z := zone(10, 1, 4, 70, 12)
	pal := marginOf(t, m, paladin(), arch(t, m, "bruiser"), z)
	sor := marginOf(t, m, sorceress(), arch(t, m, "caster"), z)
	if pal == sor {
		t.Errorf("both builds give %.2f: the archetype made no difference", pal)
	}
	line := func(h Hero, id string) string {
		zm, _ := m.Zone(h, arch(t, m, id), z, "balanced")
		return zm.ShadowLine(h.Level)
	}
	for _, s := range []string{line(paladin(), "bruiser"), line(sorceress(), "caster")} {
		if !strings.HasPrefix(s, "zone shadow: area=9") || !strings.Contains(s, "word=") {
			t.Errorf("shadow line shape: %s", s)
		}
	}
}

func TestProfilesValid(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(p.ArchetypeIDs(), ","); got != "bruiser,caster,hybrid_tank_caster,ranged_physical,summoner" {
		t.Errorf("archetypes %s", got)
	}
	if p.DefaultDial != "balanced" {
		t.Errorf("default dial %q (user decision 2026-11-14: Balanced)", p.DefaultDial)
	}
	bad := strings.Replace(string(profilesJSON), `"defaultDial": "balanced"`, `"defaultDial": "reckless"`, 1)
	if _, err := parse([]byte(bad)); err == nil {
		t.Error("an unknown default dial must be an error")
	}
	if _, err := parse([]byte(strings.Replace(string(profilesJSON), `"tags": {`, `"tags": {"x":"U",`, 1))); err != nil {
		t.Errorf("an extra tag is harmless: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(profilesJSON, &raw); err != nil {
		t.Fatal(err)
	}
}

// profileBands adapts classdata/profiles.json (the file the combat package embeds) to ClassAttacks.
func profileBands(t *testing.T) ClassAttacks {
	t.Helper()
	raw, err := os.ReadFile("../combat/classdata/profiles.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Profiles []struct {
			Class string `json:"class"`
			Build string `json:"build"`
			Bands []struct {
				MinLevel int      `json:"minLevel"`
				Requires string   `json:"requires"`
				Attack   []string `json:"attack"`
			} `json:"bands"`
			Melee *struct{} `json:"melee"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	return func(class string) ([]Band, bool, error) {
		for _, p := range f.Profiles {
			if p.Class != class || (p.Build != "" && p.Build != "cold") {
				continue
			}
			var bs []Band
			for _, b := range p.Bands {
				bs = append(bs, Band{MinLevel: b.MinLevel, Requires: b.Requires, Attack: b.Attack})
			}
			return bs, p.Melee != nil, nil
		}
		return nil, false, fmt.Errorf("no combat profile for class %q", class)
	}
}

// goldenArchetypes is the table every shipped build resolves to at the two levels (a change is a visible diff).
var goldenArchetypes = map[string]string{
	"amazon_bow_normal30.json@12":              "ranged_physical",
	"amazon_bow_normal30.json@30":              "ranged_physical",
	"amazon_javelin_normal30.json@12":          "ranged_physical",
	"amazon_multishot_endgame90.json@12":       "ranged_physical",
	"assassin_fireblast_normal30.json@12":      "caster",
	"assassin_fireblast_normal30.json@30":      "caster",
	"assassin_sentry_endgame90.json@12":        "caster",
	"barbarian_melee_normal30.json@12":         "bruiser",
	"barbarian_melee_normal30.json@30":         "bruiser",
	"barbarian_warcry_endgame90.json@12":       "bruiser",
	"druid_fire_normal30.json@12":              "caster",
	"druid_fire_normal30.json@30":              "caster",
	"druid_tornado_endgame90.json@12":          "caster",
	"necromancer_leveling_normal30.json@12":    "caster",
	"necromancer_leveling_normal30.json@30":    "caster",
	"necromancer_summoner_endgame90.json@12":   "summoner",
	"necromancer_summoner_normal30.json@12":    "summoner",
	"paladin-normal30-v2.json@12":              "bruiser",
	"paladin-normal30-v2.json@30":              "hybrid_tank_caster",
	"paladin-normal30.json@12":                 "bruiser",
	"paladin-zealot-lite.json@12":              "bruiser",
	"paladin_hammerdin_endgame90.json@12":      "bruiser",
	"paladin_hammerdin_endgame90.json@30":      "hybrid_tank_caster",
	"paladin_hammerdin_zeal_endgame90.json@30": "hybrid_tank_caster",
	"sorceress_blizzard_endgame90.json@12":     "caster",
	"sorceress_normal30.json@12":               "caster",
	"sorceress_normal30.json@30":               "caster",
}

func TestEveryBuildResolvesArchetype(t *testing.T) {
	m := model(t)
	ca := profileBands(t)
	ents, err := plandata.FS.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, e := range ents {
		n := e.Name()
		if !strings.HasPrefix(n, "build-") {
			continue
		}
		for _, lvl := range []int{12, 30} {
			s, err := SpecFromBuild(plandata.FS, n, lvl, ca)
			if err != nil {
				// the class-less fallback lightning tables are not builds
				if strings.Contains(err.Error(), "no class") {
					continue
				}
				t.Errorf("%s@%d: %v", n, lvl, err)
				continue
			}
			a, err := m.P.Resolve(s)
			if err != nil {
				t.Errorf("%s@%d: %v", n, lvl, err)
				continue
			}
			got[fmt.Sprintf("%s@%d", strings.TrimPrefix(n, "build-"), lvl)] = a.ID
		}
	}
	var diffs []string
	for k, want := range goldenArchetypes {
		if got[k] != want {
			diffs = append(diffs, fmt.Sprintf("%s: got %q want %q", k, got[k], want))
		}
	}
	sort.Strings(diffs)
	for _, d := range diffs {
		t.Error(d)
	}
	if t.Failed() {
		var all []string
		for k, v := range got {
			all = append(all, fmt.Sprintf("%q: %q,", k, v))
		}
		sort.Strings(all)
		t.Logf("resolved:\n%s", strings.Join(all, "\n"))
	}
}

// TestSyntheticClassGetsNearestArchetype: made-up builds on a made-up class resolve with no code change; an unknown
// skill, a missing attack and a missing class are errors naming the problem.
func TestSyntheticClassGetsNearestArchetype(t *testing.T) {
	m := model(t)
	for _, c := range []struct {
		attack  string
		contact bool
		points  map[string]int
		want    string
	}{
		{"Bash", true, nil, "bruiser"},
		{"Fire Arrow", false, nil, "ranged_physical"},
		{"Ice Blast", false, nil, "caster"},
		{"Ice Blast", true, nil, "hybrid_tank_caster"},
		{"Teeth", false, map[string]int{"Raise Skeleton": 10, "Teeth": 5}, "summoner"},
	} {
		a, err := m.P.Resolve(BuildSpec{Class: "wizard", Build: "synthetic", Attack: c.attack, Contact: c.contact, Points: c.points})
		if err != nil || a.ID != c.want {
			t.Errorf("%s: %v %v, want %s", c.attack, a, err, c.want)
		}
	}
	for _, s := range []BuildSpec{{Class: "wizard", Build: "x", Attack: "Zap"}, {Class: "wizard", Build: "x"}, {Build: "x", Attack: "Bash"}} {
		if _, err := m.P.Resolve(s); err == nil {
			t.Errorf("%+v must be an error", s)
		}
	}
}

func TestHeroFromPlayer(t *testing.T) {
	p := &d2.Player{Class: d2.Paladin, Level: 9, MaxLife: 150, MaxMana: 30, Dexterity: 27, Resist: d2.Resists{Fire: 12},
		Stats: d2.Stats{StatArmorClass: 90, StatBlock: 25, StatMinDmg: 5, StatMaxDmg: 11, StatDmgPct: 10, StatToHit: 40}}
	h := HeroFromPlayer(p, &d2.MercSheet{Hired: true, Alive: true, Level: 7}, 5)
	if h.Defence != 90 || h.q(TermDefence) != Measured || h.BlockPct != 25 || h.AR != 40+(27-7)*5 || h.DmgMax != 11 || h.Potions != 5 || !h.Merc.Alive {
		t.Errorf("hero %+v", h)
	}
	bare := HeroFromPlayer(&d2.Player{Class: d2.Sorceress, Level: 3, MaxLife: 60}, nil, -1)
	if bare.q(TermDefence) != Default || bare.q(TermOffence) != Default || bare.q(TermPotions) != Default || bare.q(TermLife) != Measured {
		t.Errorf("unread terms must be default: %+v", bare.Quality)
	}
	var r Reader = Fake{H: h}
	if got, err := r.Hero(); err != nil || got.Level != 9 {
		t.Errorf("fake reader: %v %v", got, err)
	}
}

// TestUnblockableRanged: a shield blocks a blockable ranged attacker but not one whose missile has no physical damage
// (Fallen Shaman fire, Radament's bolt); a hero with no block is unaffected by the flag.
func TestUnblockableRanged(t *testing.T) {
	m := model(t)
	flagged := func(share float64) *danger.Zone {
		z := zone(8, 1, 3, 60, 10)
		z.Monsters[1].Unblockable = share > 0
		z.Monsters[1].Share = 0.5
		return z
	}
	pal := paladin()
	pal.BlockPct = 60
	ab, as := arch(t, m, "bruiser"), arch(t, m, "caster")
	blockable, fire := marginOf(t, m, pal, ab, flagged(0)), marginOf(t, m, pal, ab, flagged(1))
	if fire >= blockable {
		t.Errorf("paladin block 60: unblockable ranged margin %.3f is not under the blockable %.3f", fire, blockable)
	}
	sor := sorceress()
	if a, b := marginOf(t, m, sor, as, flagged(0)), marginOf(t, m, sor, as, flagged(1)); math.Abs(a-b) > 1e-9 {
		t.Errorf("sorceress (no block): flag changed the margin %.3f -> %.3f", a, b)
	}
	// the zone mean carries the share of unblockable ranged spawns
	z := zone(8, 1, 3, 60, 10)
	z.Monsters = append(z.Monsters, z.Monsters[1])
	z.Monsters[1].Unblockable = true
	z.Monsters[1].Share, z.Monsters[2].Share = 0.25, 0.75
	if mm, ok := meanMon(z, true); !ok || math.Abs(mm.Unblockable-0.25/1.0) > 1e-9 {
		t.Errorf("meanMon unblockable share %v, want 0.25", mm.Unblockable)
	}
}
