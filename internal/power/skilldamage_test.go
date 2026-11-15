package power

import (
	"math"
	"strings"
	"testing"

	"d2lod-bot/internal/plandata"
)

func sd(t *testing.T) *SkillData {
	t.Helper()
	d, err := LoadSkillData()
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func near(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

func hasFlag(fl []string, want string) bool {
	for _, f := range fl {
		if f == want {
			return true
		}
	}
	return false
}

func TestSkillDataLoadErrors(t *testing.T) {
	d := sd(t)
	if _, err := d.Skill("Nonesuch Bolt"); err == nil || !strings.Contains(err.Error(), "Nonesuch Bolt") {
		t.Errorf("unknown skill: %v", err)
	}
	if _, err := d.Frames("Jester", "SC", "1HS"); err == nil || !strings.Contains(err.Error(), "Jester") {
		t.Errorf("unknown class: %v", err)
	}
	if _, err := d.Frames("Paladin", "SC", "9XX"); err == nil || !strings.Contains(err.Error(), "9XX") {
		t.Errorf("unknown weapon class: %v", err)
	}
	if _, err := parseSkillData([]byte(`{"schema":"power.skilldamage/2"}`)); err == nil {
		t.Error("a newer major schema must be refused")
	}
}

// TestFireBoltSorceress: Fire Bolt level 10 (elem 17.5-22.5, 2.5 mana: decomp29 T1) at 0 FCR is 13 ticks = 0.52 s, so
// 38.46 dps and 4.8 mana per second; FCR 9 is the first breakpoint (12 ticks).
func TestFireBoltSorceress(t *testing.T) {
	d := sd(t)
	o, err := d.Estimate("Sorceress", Attack{Skill: "Fire Bolt", Level: 10, SynergyPoints: -1})
	if err != nil {
		t.Fatal(err)
	}
	if !near(o.Seconds, 0.52, 1e-9) || !near(o.DPS, 20/0.52, 0.01) || !near(o.ManaPerSec, 2.5/0.52, 0.01) {
		t.Errorf("fire bolt L10: %+v", o)
	}
	for _, f := range []string{FlagFCRDefault, FlagSynergyUnknown, FlagMastery} {
		if !hasFlag(o.Flags, f) {
			t.Errorf("flag %s missing: %v", f, o.Flags)
		}
	}
	o9, _ := d.Estimate("Sorceress", Attack{Skill: "Fire Bolt", Level: 10, SynergyPoints: 0, FCR: 9, FCRKnown: true})
	if !near(o9.Seconds, 0.48, 1e-9) || hasFlag(o9.Flags, FlagFCRDefault) || hasFlag(o9.Flags, FlagSynergyUnknown) {
		t.Errorf("fcr 9: %+v", o9)
	}
	// synergies: every named synergy at 20 points is the all20 row (level 20: 336.7-447.7 against 45.5-60.5 alone)
	alone, _ := d.Estimate("Sorceress", Attack{Skill: "Fire Bolt", Level: 20})
	full, _ := d.Estimate("Sorceress", Attack{Skill: "Fire Bolt", Level: 20, SynergyPoints: 20})
	if !near(alone.PerUse, 53, 0.01) || !near(full.PerUse, 392.2, 0.1) {
		t.Errorf("synergy sets: alone %.2f full %.2f", alone.PerUse, full.PerUse)
	}
	half, _ := d.Estimate("Sorceress", Attack{Skill: "Fire Bolt", Level: 20, SynergyPoints: 10})
	if half.PerUse <= alone.PerUse || half.PerUse >= full.PerUse {
		t.Errorf("synergy points not between the sets: %.2f", half.PerUse)
	}
}

// TestZealPaladin: a weapon skill needs a read weapon. Zeal level 10 has 36 percent ED; a 6-12 sword is 9 average, the
// swing is 15 ticks (0.6 s), so 9 x 1.36 / 0.6 = 20.4 dps before the hit chance. Without the weapon nothing is invented.
func TestZealPaladin(t *testing.T) {
	d := sd(t)
	o, err := d.Estimate("Paladin", Attack{Skill: "Zeal", Level: 10, SynergyPoints: 0, WeaponAvg: 9, WeaponClass: "1HS"})
	if err != nil {
		t.Fatal(err)
	}
	if !o.NeedsHit || !near(o.Seconds, 0.6, 1e-9) || !near(o.DPS, 20.4, 0.05) || !hasFlag(o.Flags, FlagIASDefault) || !hasFlag(o.Flags, FlagOneStrike) {
		t.Errorf("zeal L10: %+v", o)
	}
	// 9 IAS-free ticks: a 2-hand weapon class swings slower
	slow, _ := d.Estimate("Paladin", Attack{Skill: "Zeal", Level: 10, SynergyPoints: 0, WeaponAvg: 9, WeaponClass: "2HS"})
	if slow.Seconds <= o.Seconds {
		t.Errorf("2HS %.2f s not slower than 1HS %.2f s", slow.Seconds, o.Seconds)
	}
	miss, err := d.Estimate("Paladin", Attack{Skill: "Zeal", Level: 10, SynergyPoints: 0})
	if err != nil {
		t.Fatal(err)
	}
	if miss.Known() || miss.DPS != 0 || miss.Quality != Default || !hasFlag(miss.Flags, FlagWeaponUnread) {
		t.Errorf("missing weapon must give no estimate: %+v", miss)
	}
}

// TestDelaySlowsBlizzard: the skill delay is a term (decomp29 Q8: Blizzard 45 ticks, longer than the cast).
func TestDelaySlowsBlizzard(t *testing.T) {
	d := sd(t)
	o, err := d.Estimate("Sorceress", Attack{Skill: "Blizzard", Level: 20, SynergyPoints: 0})
	if err != nil {
		t.Fatal(err)
	}
	if !near(o.Seconds, 1.8, 1e-9) {
		t.Errorf("blizzard period %.2f s, want the 45-tick delay (1.8 s)", o.Seconds)
	}
}

func TestEstimateErrorsName(t *testing.T) {
	d := sd(t)
	if _, err := d.Estimate("Sorceress", Attack{Skill: "Fire Bolt"}); err == nil || !strings.Contains(err.Error(), "Fire Bolt") {
		t.Errorf("no level: %v", err)
	}
	if _, err := d.Estimate("Sorceress", Attack{Skill: "Holy Wiggle", Level: 3}); err == nil || !strings.Contains(err.Error(), "Holy Wiggle") {
		t.Errorf("unknown skill: %v", err)
	}
	if _, err := d.Estimate("Jester", Attack{Skill: "Fire Bolt", Level: 3}); err == nil || !strings.Contains(err.Error(), "Jester") {
		t.Errorf("unknown class: %v", err)
	}
}

// TestMarginUsesSkillData: two builds. A Sorceress with Fire Bolt and no damage read gets the skill estimate (source
// "skill", flagged), a Paladin with Zeal but no weapon keeps the constant (source "default", flagged weapon_unread), and
// the weapon branch takes its swing from the rate data once a weapon is read.
func TestMarginUsesSkillData(t *testing.T) {
	m := model(t)
	z := zone(8, 3, 1, 60, 10)

	sor := Hero{Class: "Sorceress", Level: 12, MaxLife: 120, MaxMana: 140, Defence: 20, Potions: 6, Attack: Attack{Skill: "Fire Bolt", Level: 8, SynergyPoints: -1}}
	sor.SetQuality(TermLife, Measured)
	zs, err := m.Zone(sor, arch(t, m, "caster"), z, "balanced")
	if err != nil {
		t.Fatal(err)
	}
	if zs.Bad.Terms.Offence != "skill" || zs.Bad.Terms.HeroDPS <= 0 {
		t.Errorf("sorceress terms %+v", zs.Bad.Terms)
	}
	for _, f := range []string{"offence_fcr_default", "offence_synergy_unknown"} {
		if !hasFlag(zs.Bad.Flags, f) {
			t.Errorf("flag %s missing: %v", f, zs.Bad.Flags)
		}
	}
	plain := sor
	plain.Attack = Attack{}
	zp, _ := m.Zone(plain, arch(t, m, "caster"), z, "balanced")
	if zp.Bad.Terms.Offence != "default" || !hasFlag(zp.Bad.Flags, "offence_mana_default") {
		t.Errorf("without a build the constants stand in, flagged: %+v %v", zp.Bad.Terms, zp.Bad.Flags)
	}
	if want := m.P.Params.DefaultDpsPerClvl * 12; !near(zp.Bad.Terms.HeroDPS, want, 1e-6) {
		t.Errorf("default dps %.2f, want %.2f", zp.Bad.Terms.HeroDPS, want)
	}

	pal := Hero{Class: "Paladin", Level: 12, MaxLife: 200, MaxMana: 40, Defence: 60, Potions: 6, AR: 120, Attack: Attack{Skill: "Zeal", Level: 1, SynergyPoints: -1}}
	pal.SetQuality(TermLife, Measured)
	zn, err := m.Zone(pal, arch(t, m, "bruiser"), z, "balanced")
	if err != nil {
		t.Fatal(err)
	}
	if zn.Bad.Terms.Offence != "default" || !hasFlag(zn.Bad.Flags, "offence_weapon_unread") {
		t.Errorf("paladin without a weapon: %+v %v", zn.Bad.Terms, zn.Bad.Flags)
	}
	pal.DmgMin, pal.DmgMax = 6, 12
	pal.SetQuality(TermOffence, Derived)
	zw, _ := m.Zone(pal, arch(t, m, "bruiser"), z, "balanced")
	if zw.Bad.Terms.Offence != "skill" {
		t.Errorf("paladin with a weapon: %+v", zw.Bad.Terms)
	}
	// no attack named, weapon read: the swing comes from the rate data (0.6 s), not the constant 1.28 s
	pal.Attack = Attack{}
	s, ok := m.swing(pal)
	if !ok || !near(s, 0.6, 1e-9) {
		t.Errorf("data swing %.2f ok=%v", s, ok)
	}
	if _, ok := m.swing(Hero{Class: "log"}); ok {
		t.Error("an unknown class must fall back to the constant")
	}

	bad := sor
	bad.Attack.Skill = "Holy Wiggle"
	if _, err := m.Margin(bad, arch(t, m, "caster"), Pack{}); err == nil || !strings.Contains(err.Error(), "Holy Wiggle") {
		t.Errorf("unknown skill must be an error naming it: %v", err)
	}
}

// TestAttackOfBuild: the level of the main attack is what the build spent by that clvl; every shipped build's attack at
// clvl 30 exists in the data or the gap is reported (not a failure: the data covers class skills, not every aura).
func TestAttackOfBuild(t *testing.T) {
	d := sd(t)
	ca := profileBands(t)
	for _, c := range []struct {
		file  string
		clvl  int
		skill string
	}{{"build-paladin-normal30.json", 12, ""}, {"build-sorceress_normal30.json", 30, ""}} {
		lo, err := SpecFromBuild(plandata.FS, c.file, 6, ca)
		if err != nil {
			t.Fatal(err)
		}
		hi, err := SpecFromBuild(plandata.FS, c.file, c.clvl, ca)
		if err != nil {
			t.Fatal(err)
		}
		if hi.Attack == "" || hi.AttackLevel < 1 {
			t.Errorf("%s@%d: attack %q level %d", c.file, c.clvl, hi.Attack, hi.AttackLevel)
		}
		if _, err := d.Skill(hi.Attack); err != nil {
			t.Errorf("%s@%d: %v", c.file, c.clvl, err)
		}
		t.Logf("%s: clvl 6 %s L%d, clvl %d %s L%d", c.file, lo.Attack, lo.AttackLevel, c.clvl, hi.Attack, hi.AttackLevel)
	}
}

// TestSkillOffenceComposesWithUnblockable: with the skill estimate in charge of the kill speed, the unblockable share of
// a ranged attacker still skips the shield's block, and block still helps against the blockable one.
func TestSkillOffenceComposesWithUnblockable(t *testing.T) {
	m := model(t)
	a := arch(t, m, "bruiser")
	h := paladin()
	h.Attack = Attack{Skill: "Zeal", Level: 5, SynergyPoints: 0, WeaponClass: "1HS"}
	h.BlockPct = 50
	pack := func(unbl float64) Pack {
		return Pack{MLvl: 8, Mons: []Mon{{Life: 60, Def: 30, ToHit: 100, Dmg: 10},
			{Life: 60, Def: 30, ToHit: 100, Dmg: 10, Ranged: true, Unblockable: unbl}}}
	}
	blockable, err := m.Margin(h, a, pack(0))
	if err != nil {
		t.Fatal(err)
	}
	free, _ := m.Margin(h, a, pack(1))
	if blockable.Terms.Offence != "skill" || free.Terms.Offence != "skill" {
		t.Fatalf("offence sources %q %q", blockable.Terms.Offence, free.Terms.Offence)
	}
	if free.Terms.Taken <= blockable.Terms.Taken || free.Best >= blockable.Best {
		t.Errorf("unblockable must take more: taken %.1f vs %.1f, margin %.2f vs %.2f", free.Terms.Taken, blockable.Terms.Taken, free.Best, blockable.Best)
	}
	h.BlockPct = 0
	nb, _ := m.Margin(h, a, pack(0))
	nf, _ := m.Margin(h, a, pack(1))
	if !near(nb.Terms.Taken, nf.Terms.Taken, 1e-9) {
		t.Errorf("with no block the share is irrelevant: %.3f vs %.3f", nb.Terms.Taken, nf.Terms.Taken)
	}
}
