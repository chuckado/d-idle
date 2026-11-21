package agency

import (
	"fmt"
	"strings"
	"testing"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/leveler"
	"d2lod-bot/internal/plandata"
)

// testEnv has a nominal one-hand weapon (average 20, U) so weapon skills have an estimate; the weapon-unread cases
// clear it.
func testEnv(t *testing.T) Env {
	t.Helper()
	env, err := NewEnv()
	if err != nil {
		t.Fatal(err)
	}
	r := fixtureRules()
	env.Rules = &r
	env.Weapon = Weapon{Class: "1HS", WSM: 0, Avg: 20}
	return env
}

func alloc(class string, clvl int, sk map[string]int) Allocation {
	return Allocation{Class: class, Clvl: clvl, Skills: sk}
}

func mustDerive(t *testing.T, a Allocation, env Env) Plan {
	t.Helper()
	p, err := Derive(a, env)
	if err != nil {
		t.Fatalf("Derive(%+v): %v", a, err)
	}
	return p
}

func warnCodes(p Plan) string {
	var c []string
	for _, w := range p.Warnings {
		c = append(c, w.Code)
	}
	return strings.Join(c, ",")
}

func attackName(p Plan) string {
	if p.Attack == nil {
		return "-"
	}
	return p.Attack.Skill
}

// allocFromTable is the allocation a plan table implies at clvl.
func allocFromTable(t *testing.T, table string, clvl int) (Allocation, string) {
	t.Helper()
	raw, err := plandata.FS.ReadFile("build-" + table + ".json")
	if err != nil {
		t.Fatal(err)
	}
	b, err := leveler.ParseBuild(raw)
	if err != nil {
		t.Fatal(err)
	}
	a := Allocation{Class: b.Class, Clvl: clvl, Skills: map[string]int{}}
	for _, r := range b.Rows {
		if r.Clvl > clvl {
			break
		}
		a.Str += r.Attr[leveler.Str]
		a.Dex += r.Attr[leveler.Dex]
		a.Vit += r.Attr[leveler.Vit]
		a.Ene += r.Attr[leveler.Ene]
		for _, s := range r.Skills {
			a.Skills[s.Skill] += s.Points
		}
	}
	return a, b.Combat
}

// botPick is the bot's current choice: the first attack of the profile band that the character has points in
// (classdata AttackNames; "Attack", the plain swing, needs no points).
func botPick(t *testing.T, class, build string, clvl int, sk map[string]int) string {
	t.Helper()
	cls, err := d2Class(class)
	if err != nil {
		t.Fatal(err)
	}
	prof, err := combat.LookupProfile(cls, combat.Build(build))
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range prof.AttackNames(clvl, func(s string) bool { return sk[s] > 0 }) {
		if sk[n] > 0 || n == "Attack" {
			return n
		}
	}
	return "-"
}

func dpsOf(p Plan, skill string) (float64, bool) {
	for _, c := range p.Candidates {
		if c.Skill == skill {
			return c.DPS, true
		}
	}
	return 0, false
}

func TestAgreementTable(t *testing.T) {
	env := testEnv(t)
	tables := []string{"sorceress_normal30", "fallback-lightning-v2", "paladin-zealot-lite", "paladin_hammerdin_endgame90", "barbarian_melee_normal30", "necromancer_summoner_normal30"}
	agree, total := 0, 0
	var lines []string
	for _, tb := range tables {
		for _, clvl := range []int{6, 12, 18, 24, 30} {
			a, comb := allocFromTable(t, tb, clvl)
			env.Combat = combat.Build(comb)
			p := mustDerive(t, a, env)
			bot := botPick(t, a.Class, comb, clvl, a.Skills)
			got := attackName(p)
			verdict, why := "agree", ""
			if got != bot {
				verdict = "differ"
				bd, inCand := dpsOf(p, bot)
				switch {
				case got == "-":
					why = "derive finds no usable attack"
				case !inCand:
					why = fmt.Sprintf("bot's %s is not a derive candidate (no points, or unsupported tier)", bot)
				case bd == 0:
					why = fmt.Sprintf("bot's %s has no damage estimate; derive %s %.1f dps", bot, got, p.Attack.DPS)
				default:
					why = fmt.Sprintf("derive %s %.1f dps beats bot's %s %.1f dps (the bot's band order is not an argmax)", got, p.Attack.DPS, bot, bd)
				}
			}
			total++
			if verdict == "agree" {
				agree++
			}
			lines = append(lines, fmt.Sprintf("%-30s clvl %2d  bot %-15s derive %-15s %-6s %s [%s] arch=%s", tb, clvl, bot, got, verdict, why, warnCodes(p), p.Archetype))
		}
	}
	t.Logf("AGREEMENT TABLE (%d of %d agree)\n%s", agree, total, strings.Join(lines, "\n"))
}

func TestOffPlan(t *testing.T) {
	env := testEnv(t)
	t.Run("sorceress all fire ball", func(t *testing.T) {
		p := mustDerive(t, alloc("sorceress", 25, map[string]int{"Fire Bolt": 1, "Fire Ball": 20, "Warmth": 3}), env)
		if attackName(p) != "Fire Ball" || !p.Has(WSingleElement) || p.Secondary != nil || p.Has(WNoAttack) {
			t.Fatalf("attack %s warnings %s", attackName(p), warnCodes(p))
		}
		if p.Archetype != "caster" {
			t.Errorf("archetype %q", p.Archetype)
		}
	})
	t.Run("paladin tourist", func(t *testing.T) {
		sk := map[string]int{}
		for _, n := range []string{"Sacrifice", "Smite", "Zeal", "Holy Bolt", "Might", "Holy Fire", "Prayer", "Thorns", "Blessed Aim", "Concentration", "Resist Fire", "Defiance"} {
			sk[n] = 1
		}
		p := mustDerive(t, alloc("paladin", 18, sk), env)
		if !p.Has(WTourist) || p.Has(WNoAttack) {
			t.Fatalf("warnings %s", warnCodes(p))
		}
		low := mustDerive(t, alloc("paladin", 11, sk), env)
		if low.Has(WTourist) {
			t.Errorf("tourist must not fire below clvl 12: %s", warnCodes(low))
		}
	})
	t.Run("barbarian whirlwind maxed", func(t *testing.T) {
		p := mustDerive(t, alloc("barbarian", 30, map[string]int{"Bash": 1, "Double Swing": 1, "Whirlwind": 20, "Leap Attack": 1}), env)
		if a := attackName(p); a != "Bash" && a != "Double Swing" {
			t.Fatalf("fallback attack %s", a)
		}
		if !p.Has(WUnsupportedPts) || p.Unplayed.Skills["Whirlwind"] != 20 || p.Unplayed.Count < 20 {
			t.Fatalf("unplayed %+v warnings %s", p.Unplayed, warnCodes(p))
		}
		if !strings.Contains(p.Attack.Why, "Whirlwind") {
			t.Errorf("why does not name the skipped skill: %q", p.Attack.Why)
		}
	})
	t.Run("necromancer summoner", func(t *testing.T) {
		p := mustDerive(t, alloc("necromancer", 30, map[string]int{"Raise Skeleton": 15, "Skeleton Mastery": 10, "Clay Golem": 3, "Teeth": 1}), env)
		if p.Archetype != "summoner" || p.Main == nil || p.Main.Skill != "Raise Skeleton" || p.Main.Role != RoleSummon || p.Has(WNoAttack) {
			t.Fatalf("archetype %q main %+v warnings %s", p.Archetype, p.Main, warnCodes(p))
		}
		if p.Has(WTourist) {
			t.Errorf("a pure summoner is not a tourist: %s", warnCodes(p))
		}
		only := mustDerive(t, alloc("necromancer", 30, map[string]int{"Raise Skeleton": 20, "Skeleton Mastery": 9}), env)
		if only.Has(WNoAttack) || only.Main == nil || attackName(only) != "Attack" {
			t.Errorf("pets only: main %+v warnings %s", only.Main, warnCodes(only))
		}
	})
	t.Run("zero attack skills", func(t *testing.T) {
		p := mustDerive(t, alloc("sorceress", 20, map[string]int{"Frozen Armor": 1, "Warmth": 5}), env)
		if !p.Has(WNoAttack) || p.Main != nil {
			t.Fatalf("main %+v warnings %s", p.Main, warnCodes(p))
		}
		pal := mustDerive(t, alloc("paladin", 20, map[string]int{"Might": 10}), env)
		if attackName(pal) != "Attack" || pal.Has(WNoAttack) {
			t.Errorf("paladin with no attack skill plays the plain swing: %s %s", attackName(pal), warnCodes(pal))
		}
	})
	t.Run("weapon unread", func(t *testing.T) {
		e := testEnv(t)
		e.Weapon = Weapon{}
		p := mustDerive(t, alloc("paladin", 18, map[string]int{"Sacrifice": 1, "Smite": 1, "Zeal": 10}), e)
		if !p.Has(WWeaponUnread) || attackName(p) != "Zeal" {
			t.Fatalf("attack %s warnings %s", attackName(p), warnCodes(p))
		}
	})
	t.Run("mana starved and energy", func(t *testing.T) {
		a := alloc("sorceress", 30, map[string]int{"Ice Bolt": 1, "Ice Blast": 1, "Glacial Spike": 1, "Blizzard": 20})
		p := mustDerive(t, a, env)
		if attackName(p) != "Blizzard" || !p.Has(WManaStarved) {
			t.Fatalf("attack %s warnings %s mana %.1f/%.1f", attackName(p), warnCodes(p), p.ManaPerSec, p.ManaBudget)
		}
		a.Ene = 200
		if q := mustDerive(t, a, env); q.ManaBudget <= p.ManaBudget {
			t.Errorf("energy must raise the budget: %.1f -> %.1f %s", p.ManaBudget, q.ManaBudget, warnCodes(q))
		}
	})
	t.Run("prereq only", func(t *testing.T) {
		p := mustDerive(t, alloc("paladin", 20, map[string]int{"Sacrifice": 1, "Smite": 1, "Zeal": 15, "Might": 3}), env)
		if !p.Has(WPrereqOnly) {
			t.Fatalf("warnings %s", warnCodes(p))
		}
	})
	t.Run("bolts vs hammer picks by damage", func(t *testing.T) {
		p := mustDerive(t, alloc("paladin", 30, map[string]int{"Holy Bolt": 1, "Blessed Hammer": 20, "Concentration": 20, "Zeal": 1}), env)
		if attackName(p) != "Blessed Hammer" {
			t.Errorf("attack %s", attackName(p))
		}
		if !strings.Contains(p.Attack.Why, "dps") {
			t.Errorf("why: %q", p.Attack.Why)
		}
	})
}

func TestValidate(t *testing.T) {
	r := fixtureRules()
	codes := func(ps []Problem) string {
		var c []string
		for _, p := range ps {
			c = append(c, p.Code+":"+p.Skill)
		}
		return strings.Join(c, " ")
	}
	cases := []struct {
		name string
		a    Allocation
		want string
	}{
		{"clean", alloc("paladin", 18, map[string]int{"Sacrifice": 5, "Smite": 5, "Zeal": 5}), ""},
		{"prereq", alloc("paladin", 18, map[string]int{"Zeal": 3}), "prereq:Zeal prereq:Zeal"},
		{"req level", alloc("paladin", 5, map[string]int{"Sacrifice": 1, "Smite": 1}), "req_level:Smite"},
		{"cap", alloc("sorceress", 60, map[string]int{"Fire Bolt": 21}), "max_level:Fire Bolt"},
		{"budget", alloc("sorceress", 3, map[string]int{"Fire Bolt": 2, "Charged Bolt": 1}), "budget:"},
		{"quest points raise the budget", Allocation{Class: "sorceress", Clvl: 3, Skills: map[string]int{"Fire Bolt": 2, "Charged Bolt": 1}, QuestSkillPts: 1}, ""},
		{"no rules row", alloc("sorceress", 10, map[string]int{"Warmth": 1}), "no_rules:Warmth"},
	}
	for _, c := range cases {
		if got := codes(Validate(c.a, r)); got != c.want {
			t.Errorf("%s: problems %q, want %q", c.name, got, c.want)
		}
	}
}

func TestCanAdd(t *testing.T) {
	r := fixtureRules()
	for _, c := range []struct {
		name  string
		a     Allocation
		skill string
		ok    bool
		why   string
	}{
		{"free", alloc("paladin", 10, map[string]int{"Sacrifice": 2}), "Smite", true, ""},
		{"no prereq", alloc("paladin", 20, map[string]int{"Sacrifice": 1}), "Zeal", false, "needs a point in Smite"},
		{"level", alloc("paladin", 5, map[string]int{"Sacrifice": 1}), "Smite", false, "needs clvl 6"},
		{"cap", alloc("sorceress", 40, map[string]int{"Fire Bolt": 20}), "Fire Bolt", false, "cap"},
		{"no point", alloc("sorceress", 3, map[string]int{"Fire Bolt": 2}), "Charged Bolt", false, "no unspent"},
		{"no row", alloc("sorceress", 9, nil), "Hydra", false, "no skills.txt row"},
	} {
		ok, why := CanAdd(c.a, r, c.skill)
		if ok != c.ok || !strings.Contains(why, c.why) {
			t.Errorf("%s: %v %q", c.name, ok, why)
		}
	}
}

func TestPreview(t *testing.T) {
	env := testEnv(t)
	a := alloc("sorceress", 14, map[string]int{"Fire Bolt": 10, "Ice Bolt": 0})
	b, af, err := Preview(a, env, "Fire Bolt")
	if err != nil {
		t.Fatal(err)
	}
	if d := Diff(b, af); d.DPS <= 0 {
		t.Errorf("one more Fire Bolt point must add damage: %+v", d)
	}
	if a.Skills["Fire Bolt"] != 10 {
		t.Error("Preview changed the allocation")
	}
	b, af, err = Preview(a, env, "Ice Bolt")
	if err != nil {
		t.Fatal(err)
	}
	d := Diff(b, af)
	if !b.Has(WSingleElement) || af.Has(WSingleElement) || len(d.GoneWarnings) != 1 || d.MainChanged {
		t.Errorf("a point in a second element must clear single_element: before %s after %s", warnCodes(b), warnCodes(af))
	}
	t.Logf("preview Ice Bolt: dps %+.1f, gone %v, secondary %+v", d.DPS, d.GoneWarnings, af.Secondary)
}

func TestErrorsNameTheirCause(t *testing.T) {
	env := testEnv(t)
	if _, err := Derive(alloc("wizard", 10, nil), env); err == nil || !strings.Contains(err.Error(), "wizard") {
		t.Errorf("unknown class: %v", err)
	}
	if _, err := Derive(alloc("sorceress", 10, map[string]int{"Banana Strike": 3}), env); err == nil || !strings.Contains(err.Error(), "Banana Strike") {
		t.Errorf("unknown skill: %v", err)
	}
	if _, err := Derive(alloc("sorceress", 0, nil), env); err == nil {
		t.Error("clvl 0 must be an error")
	}
}

func TestCapabilityTable(t *testing.T) {
	c, err := LoadCaps()
	if err != nil {
		t.Fatal(err)
	}
	counts := c.Counts()
	for _, cl := range []string{"sorceress", "paladin"} {
		n := 0
		for _, v := range counts[cl] {
			n += v
		}
		if n != 30 {
			t.Errorf("%s has %d rows, want all 30 skills", cl, n)
		}
	}
	for _, cl := range Classes {
		if _, ok := counts[cl]; ok {
			t.Logf("tiers %-12s supported %2d experimental %2d unsupported %2d", cl, counts[cl][Supported], counts[cl][Experimental], counts[cl][Unsupported])
		}
	}
	// a skill missing from the table is derived from power's kinds and marked experimental; one in neither is not found.
	env := testEnv(t)
	// every skill power knows has a row since dlv-data, so the derived path is exercised with a made-up kind.
	if d, ok := c.Lookup("Made Up Arrow", map[string]string{"Made Up Arrow": "missile"}); !ok || d.Tier != Experimental || !d.Derived || d.Role != RoleAttack {
		t.Errorf("derived row: %+v %v", d, ok)
	}
	if _, ok := c.Lookup("Banana Strike", env.Profiles.SkillKinds); ok {
		t.Error("a skill in neither source must not get a row")
	}
	// every attack row that is not unsupported should have damage data, otherwise Derive cannot rank it.
	for n, s := range c.Skills {
		if s.Role == RoleAttack && s.Tier != Unsupported {
			if _, err := env.Skills.Skill(n); err != nil {
				t.Logf("no skilldamage row for usable attack %s: %v", n, err)
			}
		}
	}
	_ = d2.Sorceress
}

func TestUnplayedLog(t *testing.T) {
	env := testEnv(t)
	a, comb := allocFromTable(t, "sorceress_normal30", 12)
	env.Combat = combat.Build(comb)
	p := mustDerive(t, a, env)
	t.Logf("sorceress_normal30 clvl 12 unplayed %v warnings %v", p.Unplayed.Skills, p.Warnings)
}
