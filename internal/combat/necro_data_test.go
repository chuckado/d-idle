package combat

import (
	"context"
	"errors"
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
)

// Every Necromancer profile carries the necro block; no other class may (validateNecro).
func TestNecroBlockOnlyOnNecromancerProfiles(t *testing.T) {
	n := 0
	for _, p := range Profiles() {
		if (p.Class == "necromancer") != (p.Necro != nil) {
			t.Errorf("profile %q: necro block %v", p.ProfileID(), p.Necro != nil)
		}
		if p.Necro != nil {
			n++
		}
	}
	if n < 2 {
		t.Fatalf("want the default and the summoner Necromancer profile, got %d", n)
	}
}

func TestValidateNecroErrorsNameTheProfile(t *testing.T) {
	good := ProfileFor(d2.Necromancer, "").Necro
	cases := []struct {
		name string
		p    Profile
		want string
	}{
		{"paladin with block", Profile{Class: "paladin", Necro: good}, `"paladin"`},
		{"necro without block", Profile{Class: "necromancer", Build: "x"}, `"necromancer/x"`},
		{"bad kind", Profile{Class: "necromancer", Necro: &NecroParams{Golems: []string{"g"}, Curses: good.Curses,
			Summons: []SummonRule{{Name: "a", Kind: "zombie", Max: 1}}}}, "bad summon rule"},
		{"bad when", Profile{Class: "necromancer", Necro: &NecroParams{Golems: []string{"g"}, Summons: good.Summons,
			Curses: []CurseRule{{Name: "c", When: "sometimes"}}}}, `unknown when "sometimes"`},
	}
	for _, c := range cases {
		err := validateNecro([]Profile{c.p})
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err %v, want %q", c.name, err, c.want)
		}
	}
}

// A Paladin asking for the necro data is an error that names the profile, never a fallback to Necromancer rules.
func TestNecroParamsUnknownClassIsAnError(t *testing.T) {
	h, w := newNecro(t)
	w.snap.Player.Class, w.snap.Player.Skills, h.f.Build = d2.Paladin, nil, ""
	w.snap.Player.Skills = []d2.Skill{{ID: 1, Level: 1}} // no necromancer skill: the profile is the Paladin's
	h.f.SetCatalog(MapCatalog{Skills: map[int]*data.Skill{}})
	if _, err := h.f.necroParams(w.snap.Player); err == nil || !strings.Contains(err.Error(), "paladin") {
		t.Fatalf("want an error naming paladin, got %v", err)
	}
	// and the Necromancer one works on both builds
	for _, b := range []Build{"", "summoner"} {
		if ProfileFor(d2.Necromancer, b).Necro == nil {
			t.Fatalf("build %q has no necro block", b)
		}
	}
}

// The data drives the rules: a different curse order, golem ladder and summon cap change the behaviour (two data sets).
func TestCurseAndSummonFollowTheData(t *testing.T) {
	h, w := newNecro(t)
	known := KnownSkills(w.snap.Player, h.f.cat, h.f.keys)
	p := w.snap.Player
	tg := Target{Unit: mon(2, 1, 112, 100), Name: "f", Dist: 12, LifePct: 100}
	def := ProfileFor(d2.Necromancer, "").Necro
	if sp, _ := curseFor(def, known, p, tg); sp.Name != "Amplify Damage" {
		t.Fatalf("default data picks %q", sp.Name)
	}
	alt := &NecroParams{Curses: []CurseRule{{Name: "Weaken", When: "any"}, {Name: "Amplify Damage", When: "any"}}}
	if sp, _ := curseFor(alt, known, p, tg); sp.Name != "Weaken" {
		t.Fatalf("alt data picks %q", sp.Name)
	}
	if _, ok := curseFor(&NecroParams{Curses: []CurseRule{{Name: "Amplify Damage"}}}, known, p, tg); ok {
		t.Fatal("a rule with no When must never be cast")
	}
	if got := wantMinions(SummonRule{Max: 3}, 10); got != 3 {
		t.Fatalf("Max cap: %d", got)
	}

	// golem ladder from data: the first entry that is known wins
	w.snap.Player.Skills = append(w.snap.Player.Skills, d2.Skill{ID: 80, Level: 1})
	h.f.SetCatalog(MapCatalog{Skills: map[int]*data.Skill{idClay: necroCat().Skills[idClay],
		80: {ID: 80, Key: "FireGolem", Name: "FireGolem", Summon: "firegolem", Mana: 20, ManaShift: 8, MinMana: 20}}, Monsters: necroCat().Monsters})
	h.f.SetHotkeys(map[int]string{idClay: "f8", 80: "f2"})
	w.snap.Player.RightSkill = idClay
	alt = &NecroParams{Golems: []string{"FireGolem", "Clay Golem"}, Summons: def.Summons}
	sp, _, _, ok := h.f.planSummon(w.snap, alt, false)
	if !ok || sp.Name != "FireGolem" {
		t.Fatalf("ladder from data: %v %v", sp.Name, ok)
	}
	alt.Golems = []string{"Clay Golem"}
	if sp, _, _, ok := h.f.planSummon(w.snap, alt, false); !ok || sp.Name != "Clay Golem" {
		t.Fatalf("second ladder: %v", sp.Name)
	}
}

// KeepSummons for a non-Necromancer stays a no-op (the class gate), and Curse on a Paladin does not touch the data.
func TestNonNecromancerNeverReadsNecroData(t *testing.T) {
	h, w := newNecro(t)
	w.snap.Player.Class, w.snap.Player.Skills, h.f.Build = d2.Paladin, nil, ""
	if n, err := h.f.KeepSummons(context.Background()); n != 0 || err != nil {
		t.Fatalf("paladin summons: %d %v", n, err)
	}
	if done, err := h.f.Curse(context.Background(), w.snap, nil); done || err != nil {
		t.Fatalf("paladin curse: %v %v", done, err)
	}
}

// necro5: the engine read replaces the snapshot census only with the knob on, a reader and a profile with summon data; it
// fails to the snapshot census otherwise. Must fail if census ignored the reader or the knob.
func TestCensusEngineRead(t *testing.T) {
	h, w := newNecro(t)
	w.snap.Monsters = append(w.snap.Monsters, d2.Unit{ID: 600, Type: d2.TypeMonster, Class: clsGolem, Mode: d2.MonModeStand,
		X: 100, Y: 100, Monster: &d2.MonsterInfo{LifePct: 100, States: []uint8{d2.StateMinion}}})
	rows := []MinionRow{{Skill: idRaise}, {Skill: idRaise}, {Skill: idClay}}
	h.f.MinionRead = func() ([]MinionRow, error) { return rows, nil }
	if m := h.f.census(w.snap); m.Total() != 1 || m.Golems != 1 {
		t.Fatalf("knob 0 must use the snapshot: %+v", m)
	}
	h.f.e.Tune.Combat.MinionRead = 1
	if m := h.f.census(w.snap); m.Skeletons != 2 || m.Golems != 1 {
		t.Fatalf("knob 1 must use the read: %+v", m)
	}
	h.f.MinionRead = func() ([]MinionRow, error) { return nil, errors.New("no flag") }
	if m := h.f.census(w.snap); m.Total() != 1 {
		t.Fatalf("a failed read must fall back: %+v", m)
	}
	// A Paladin has no necro block: the read is never consulted.
	w.snap.Player.Class, w.snap.Player.Skills, h.f.Build = d2.Paladin, nil, ""
	h.f.MinionRead = func() ([]MinionRow, error) { return rows, nil }
	if m := h.f.census(w.snap); m.Skeletons != 0 {
		t.Fatalf("a Paladin must keep the snapshot census: %+v", m)
	}
}

// necro3: the Poison Nova caster profile exists and curses differently from the summoner (two builds, one engine).
func TestPoisonNovaProfileCurses(t *testing.T) {
	h, w := newNecro(t)
	known := KnownSkills(w.snap.Player, h.f.cat, h.f.keys)
	p := w.snap.Player
	tg := Target{Unit: mon(2, 1, 112, 100), Name: "f", Dist: 12, LifePct: 100}
	caster, err := LookupProfile(d2.Necromancer, "poisonnova")
	if err != nil {
		t.Fatal(err)
	}
	if sp, _ := curseFor(caster.Necro, known, p, tg); sp.Name != "Lower Resist" {
		t.Fatalf("caster curses %q", sp.Name)
	}
	summ, _ := LookupProfile(d2.Necromancer, "summoner")
	if sp, _ := curseFor(summ.Necro, known, p, tg); sp.Name != "Amplify Damage" {
		t.Fatalf("summoner curses %q", sp.Name)
	}
	if _, err := LookupProfile(d2.Paladin, "poisonnova"); err == nil || !strings.Contains(err.Error(), "paladin") {
		t.Fatalf("a Paladin must not get a necromancer profile: %v", err)
	}
	if caster.AttackNames(36, func(n string) bool { return true })[0] != "Poison Nova" ||
		caster.AttackNames(20, func(n string) bool { return n != "Poison Nova" })[0] != "Bone Spear" {
		t.Fatal("caster attack bands: Poison Nova from 30, Bone Spear from 18")
	}
}

// necro nova: the Poison Nova caster profile reads its attack order from data; Poison Nova leads from clvl 30, a poison-immune
// target falls back to Bone Spear, and a Paladin never gets a Necromancer attack (two classes, one rule).
func TestPoisonNovaAttackAndImmuneFallback(t *testing.T) {
	const idNova, idSpear, idTeeth = 92, 84, 67
	cat := MapCatalog{Skills: map[int]*data.Skill{
		idNova:  {ID: idNova, Key: "Poison Nova", Name: "Poison Nova", EType: "pois", Mana: 20, ManaShift: 8, MinMana: 20},
		idSpear: {ID: idSpear, Key: "Bone Spear", Name: "Bone Spear", EType: "mag", Mana: 7, ManaShift: 8, MinMana: 7},
		idTeeth: {ID: idTeeth, Key: "Teeth", Name: "Teeth", EType: "mag", Mana: 3, ManaShift: 8, MinMana: 3},
	}}
	p := &d2.Player{Class: d2.Necromancer, Level: 40, Mana: 200, MaxMana: 200,
		Skills: []d2.Skill{{ID: idNova, Level: 20}, {ID: idSpear, Level: 10}, {ID: idTeeth, Level: 5}}}
	keys := map[int]string{idNova: "f3", idSpear: "f2", idTeeth: "f1"}
	known := KnownSkills(p, cat, keys)
	pr, err := LookupProfile(d2.Necromancer, "poisonnova")
	if err != nil {
		t.Fatal(err)
	}
	var order []SkillSpec
	for _, n := range pr.AttackNames(40, func(n string) bool { _, ok := known[n]; return ok }) {
		if sp, ok := known[n]; ok {
			order = append(order, sp)
		}
	}
	normal := &Target{HasRes: true}
	if sp, why := PickByResist(order, normal, 200, nil); why != PickOK || sp.Name != "Poison Nova" || sp.Kind != KindSelf {
		t.Fatalf("a normal pack gets Poison Nova: %v %v", sp.Name, why)
	}
	immune := &Target{HasRes: true, Res: [6]int{0, 0, 0, 0, 0, 0}}
	immune.Res[data.Poison] = 100
	if sp, why := PickByResist(order, immune, 200, nil); why != PickOK || sp.Name != "Bone Spear" {
		t.Fatalf("a poison-immune target falls back to Bone Spear: %v %v", sp.Name, why)
	}
	// Below clvl 30 the band is Bone Spear first: the profile, not Go, holds the level.
	if got := pr.AttackNames(20, func(n string) bool { return n != "Poison Nova" }); len(got) == 0 || got[0] != "Bone Spear" {
		t.Fatalf("clvl 20 order %v", got)
	}
	// The same rule for a Paladin: no necro profile, the poison skills are not in its attack order.
	if _, err := LookupProfile(d2.Paladin, "poisonnova"); err == nil {
		t.Fatal("a Paladin must not get the poisonnova profile")
	}
	for _, n := range ProfileFor(d2.Paladin, "").AttackNames(40, func(string) bool { return true }) {
		if n == "Poison Nova" {
			t.Fatalf("the Paladin order holds Poison Nova")
		}
	}
}
