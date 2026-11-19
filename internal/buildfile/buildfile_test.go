package buildfile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"d2lod-bot/internal/leveler"
	"d2lod-bot/internal/plandata"
)

func load(t *testing.T, name string) *Build {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	b, err := Parse(raw)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return b
}

func codes(ps []Problem) string {
	var s []string
	for _, p := range ps {
		s = append(s, p.Code)
	}
	return strings.Join(s, ",")
}

func TestExamplesValidateAndExpand(t *testing.T) {
	for _, f := range []string{"paladin-hammerdin.json", "sorceress-phases.json", "necromancer-respec.json", "barbarian-after.json"} {
		b := load(t, f)
		if ps := Validate(b, fixtureRules{}); len(ps) > 0 {
			t.Errorf("%s: %v", f, ps)
			continue
		}
		e, err := Expand(b, fixtureRules{})
		if err != nil {
			t.Errorf("%s: %v", f, err)
		}
		again, _ := Expand(b, fixtureRules{})
		if !reflect.DeepEqual(e, again) {
			t.Errorf("%s: Expand is not deterministic", f)
		}
		if e.StatUnspent != 0 && b.Stats[len(b.Stats)-1].To.Rest {
			t.Errorf("%s: rest goal left %d points", f, e.StatUnspent)
		}
	}
}

func TestParseStrict(t *testing.T) {
	for name, in := range map[string]string{
		"unknown field": `{"schema":"didle.build/1","class":"paladin","skills":[],"main_attack":"Zeal"}`,
		"wrong schema":  `{"schema":"didle.build/2","class":"paladin"}`,
		"trailing":      `{"schema":"didle.build/1","class":"paladin"} {}`,
		"bad stat":      `{"schema":"didle.build/1","class":"paladin","stats":[{"stat":"str","to":"lots"}]}`,
		"stat to float": `{"schema":"didle.build/1","class":"paladin","stats":[{"stat":"str","to":6.5}]}`,
	} {
		if _, err := Parse([]byte(in)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	b := load(t, "paladin-hammerdin.json")
	out, _ := Marshal(b)
	b2, err := Parse(out)
	if err != nil || !reflect.DeepEqual(b, b2) {
		t.Errorf("marshal round trip: %v", err)
	}
}

func TestValidateCodes(t *testing.T) {
	pal := func(mut func(*Build)) *Build {
		b := load(t, "paladin-hammerdin.json")
		mut(b)
		return b
	}
	cases := []struct {
		name string
		b    *Build
		r    Rules
		want string
	}{
		{"class", pal(func(b *Build) { b.Class = "monk" }), fixtureRules{}, PClass},
		{"unknown skill", pal(func(b *Build) { b.Skills[0].Skill = "Hammer Time" }), fixtureRules{}, PSkillUnknown},
		{"wrong class", pal(func(b *Build) { b.Skills[0].Skill = "Fire Bolt" }), fixtureRules{}, PWrongClass},
		{"target over 20", pal(func(b *Build) { b.Skills[1].To = 21 }), fixtureRules{}, PTarget},
		{"target zero", pal(func(b *Build) { b.Skills[1].To = 0 }), fixtureRules{}, PTarget},
		{"cycle", pal(func(b *Build) { b.Skills = []Goal{{Skill: "Ping", To: 1}} }), cyclicRules{}, PCycle},
		{"no goals", pal(func(b *Build) { b.Skills = nil }), fixtureRules{}, PNoGoals},
		{"too many points", pal(func(b *Build) {
			b.Quest = nil
			b.Skills = []Goal{{Skill: "Blessed Hammer", To: 20}, {Skill: "Concentration", To: 20}, {Skill: "Holy Shield", To: 20},
				{Skill: "Vigor", To: 20}, {Skill: "Fanaticism", To: 20}, {Skill: "Zeal", To: 20}, {Skill: "Smite", To: 20}}
		}), fixtureRules{}, PSkillBudget},
		{"after never met", pal(func(b *Build) {
			b.Skills = []Goal{{Skill: "Might", To: 1, When: &When{After: &After{Skill: "Vigor", Level: 5}}}}
		}), fixtureRules{}, PUnreachable},
		{"when over 99", pal(func(b *Build) { b.Skills[1].When = &When{Clvl: 120} }), fixtureRules{}, PWhen},
		{"stat unknown", pal(func(b *Build) { b.Stats[0].Stat = "luck" }), fixtureRules{}, PStat},
		{"stat below start", pal(func(b *Build) { b.Stats[0].To = StatTarget{N: 10} }), fixtureRules{}, PStatTarget},
		{"stats over budget", pal(func(b *Build) {
			b.Stats = []StatGoal{{Stat: "str", To: StatTarget{N: 400}}, {Stat: "vit", To: StatTarget{N: 400}}}
		}), fixtureRules{}, PStatBudget},
		{"goal after rest", pal(func(b *Build) { b.Stats = append(b.Stats, StatGoal{Stat: "dex", To: StatTarget{N: 90}}) }), fixtureRules{}, PRest},
		{"too many quest points", pal(func(b *Build) { b.Quest = []QuestPoint{{Clvl: 10, Skill: 20}} }), fixtureRules{}, PQuest},
		{"phase order", pal(func(b *Build) {
			b.Phases = []Phase{{Name: "a", FromClvl: 10, UntilClvl: 5, Skills: []Goal{{Skill: "Might", To: 1}}}}
		}), fixtureRules{}, PPhase},
		{"respec without phases", pal(func(b *Build) { b.RespecAt = &Respec{Clvl: 24} }), fixtureRules{}, PRespec},
		{"stat after", pal(func(b *Build) { b.Stats[1].When = &When{After: &After{Skill: "Might", Level: 1}} }), fixtureRules{}, PStatAfter},
	}
	for _, c := range cases {
		got := codes(Validate(c.b, c.r))
		if !strings.Contains(got, c.want) {
			t.Errorf("%s: want %s, got %q", c.name, c.want, got)
		}
	}
}

func skillAt(rows []leveler.Row, clvl int) map[string]int { m, _ := Cumulative(rows, clvl); return m }

func TestExpandBankingAndPrereqs(t *testing.T) {
	b := load(t, "paladin-hammerdin.json")
	e, err := Expand(b, fixtureRules{})
	if err != nil {
		t.Fatal(err)
	}
	if len(e.Pre) != 98 || e.Pre[0].Clvl != 2 || e.Pre[97].Clvl != 99 {
		t.Fatalf("rows %d, %d..%d", len(e.Pre), e.Pre[0].Clvl, e.Pre[len(e.Pre)-1].Clvl)
	}
	// Might at 2; clvl 3..: Blessed Hammer is closed until 18, Holy Shield (prereqs Smite, Charge, Defiance) is
	// spendable only from 18 too, so the points bank. First spend after Might is Concentration's chain (Blessed Aim at 12).
	if got := skillAt(e.Pre, 2)["Might"]; got != 1 {
		t.Errorf("Might at 2: %d", got)
	}
	at17 := skillAt(e.Pre, 17)
	at18 := skillAt(e.Pre, 18)
	if at17["Blessed Hammer"] != 0 {
		t.Errorf("Blessed Hammer before reqlevel: %v", at17)
	}
	if at18["Blessed Hammer"] < 5 || at18["Holy Bolt"] != 1 {
		t.Errorf("banked points not spent at 18 with Holy Bolt prerequisite: %v", at18)
	}
	// prerequisites are added as late as possible: Holy Bolt (reqlevel 6) is NOT bought at 6 but with Blessed Hammer.
	if skillAt(e.Pre, 17)["Holy Bolt"] != 0 {
		t.Errorf("Holy Bolt bought early: %v", skillAt(e.Pre, 17))
	}
	// the chain behind Fanaticism (Zeal) opens at 30, one point per level: Zeal at 30, Fanaticism at 31
	if a, b := skillAt(e.Pre, 29), skillAt(e.Pre, 31); a["Zeal"] != 0 || b["Zeal"] != 1 || b["Fanaticism"] != 1 {
		t.Errorf("Fanaticism chain: at 29 %v, at 31 %v", a, b)
	}
	// total spent never exceeds earned (98 levels + 4 quest)
	sk, _ := Cumulative(e.Pre, 99)
	sum := 0
	for _, v := range sk {
		sum += v
	}
	if sum > 98+4 {
		t.Errorf("spent %d", sum)
	}
	// stats: str to 60 first (35 points = 7 levels), dex per 2 beside vit rest
	_, at := Cumulative(e.Pre, 8)
	if at[0] != 35 {
		t.Errorf("str points at 8: %v", at)
	}
}

func TestStatPerAndRest(t *testing.T) {
	b := &Build{Class: "paladin", Stats: []StatGoal{
		{Stat: "str", To: StatTarget{N: 40}, Per: 3}, {Stat: "vit", To: StatTarget{Rest: true}}}}
	st, unspent := expandStats(b, [4]int{25, 20, 25, 15})
	if st[2] != [4]int{3, 0, 2, 0} || unspent != 0 {
		t.Errorf("clvl 2: %v unspent %d", st[2], unspent)
	}
	// str from 25 to 40 = 15 points = 5 levels at 3 per level, then all 5 into vit
	if st[6] != [4]int{3, 0, 2, 0} || st[7] != [4]int{0, 0, 5, 0} {
		t.Errorf("clvl 6 %v clvl 7 %v", st[6], st[7])
	}
	b.Stats = b.Stats[:1]
	_, unspent = expandStats(b, [4]int{25, 20, 25, 15})
	if unspent != 5*98-15 {
		t.Errorf("without rest the leftover stays banked: %d", unspent)
	}
}

func TestPhasesWithoutRespec(t *testing.T) {
	b := load(t, "sorceress-phases.json")
	e, err := Expand(b, fixtureRules{})
	if err != nil || e.Respec != 0 || e.Post != nil {
		t.Fatalf("%v %+v", err, e)
	}
	if skillAt(e.Pre, 17)["Fire Bolt"] != 12 || skillAt(e.Pre, 17)["Fire Ball"] != 0 {
		t.Errorf("clvl 17: %v", skillAt(e.Pre, 17))
	}
	// Fire Bolt stopped at 12 hard points; the 5 points earned at clvl 14..18 bank and go into Fire Ball at 18
	if got := skillAt(e.Pre, 18); got["Fire Ball"] != 5 || got["Fire Bolt"] != 12 {
		t.Errorf("clvl 18: %v", got)
	}
	// the later top-level Fire Bolt to 20 only opens after Fire Ball 20.
	if got := skillAt(e.Pre, 99); got["Fire Ball"] != 20 || got["Fire Bolt"] != 20 {
		t.Errorf("clvl 99: %v", got)
	}
}

func TestRespecPhase(t *testing.T) {
	b := load(t, "necromancer-respec.json")
	e, err := Expand(b, fixtureRules{})
	if err != nil {
		t.Fatal(err)
	}
	if e.Respec != 24 || e.Pre[len(e.Pre)-1].Clvl != 24 || e.Post[0].Clvl != 25 || e.Post[len(e.Post)-1].Clvl != 99 {
		t.Fatalf("respec %d pre..%d post %d..%d", e.Respec, e.Pre[len(e.Pre)-1].Clvl, e.Post[0].Clvl, e.Post[len(e.Post)-1].Clvl)
	}
	pre := skillAt(e.Pre, 24)
	if pre["Teeth"] != 20 || pre["Bone Spear"] != 0 {
		t.Errorf("before respec: %v", pre)
	}
	pool := map[string]int{}
	total := 0
	for _, s := range e.Pool {
		pool[s.Skill] = s.Points
		total += s.Points
	}
	if total != 23 || pool["Bone Spear"] < 1 || pool["Teeth"] != 1 || pool["Bone Armor"] != 1 {
		t.Errorf("pool (23 points earned by clvl 24): %v", pool)
	}
	lb := e.ToLeveler(b)
	if lb.Respec != 24 || len(lb.RespecPool) != len(e.Pool) || len(lb.Rows) != 98 {
		t.Errorf("leveler build: respec %d pool %d rows %d", lb.Respec, len(lb.RespecPool), len(lb.Rows))
	}
}

func TestAfterGate(t *testing.T) {
	b := load(t, "barbarian-after.json")
	e, _ := Expand(b, fixtureRules{})
	for _, r := range e.Pre {
		sk := skillAt(e.Pre, r.Clvl)
		if sk["Iron Skin"] > 0 && sk["Frenzy"] < 10 {
			t.Fatalf("Iron Skin before Frenzy 10 at clvl %d: %v", r.Clvl, sk)
		}
	}
	if skillAt(e.Pre, 99)["Iron Skin"] != 10 {
		t.Errorf("Iron Skin never reached 10")
	}
}

func TestHold(t *testing.T) {
	// kolbot satisfy=true: a goal that cannot be spent yet still blocks the ones after it
	mk := func(hold bool) *Build {
		return &Build{Schema: Schema, Class: "paladin", Skills: []Goal{
			{Skill: "Blessed Hammer", To: 3, Hold: hold}, {Skill: "Might", To: 5}}}
	}
	free, _ := Expand(mk(false), fixtureRules{})
	held, _ := Expand(mk(true), fixtureRules{})
	if skillAt(free.Pre, 10)["Might"] != 5 {
		t.Errorf("without hold Might should be done by 10: %v", skillAt(free.Pre, 10))
	}
	if skillAt(held.Pre, 17)["Might"] != 0 {
		t.Errorf("with hold nothing may be spent before Blessed Hammer unlocks: %v", skillAt(held.Pre, 17))
	}
}

func TestEncodeDecode(t *testing.T) {
	line := "didle1:pal:Zeal20,Holy Shield20@24,Fanaticism20|str=60,dex=70/3,vit=rest"
	b, err := Decode(line)
	if err != nil {
		t.Fatal(err)
	}
	if b.Class != "paladin" || len(b.Skills) != 3 || b.Skills[1].Skill != "Holy Shield" || b.Skills[1].When.Clvl != 24 || b.Stats[1].Per != 3 || !b.Stats[2].To.Rest {
		t.Fatalf("%+v", b)
	}
	if ps := Validate(b, fixtureRules{}); len(ps) > 0 {
		t.Errorf("decoded line invalid: %v", ps)
	}
	back, err := Encode(b)
	if err != nil || back != line {
		t.Errorf("round trip: %q %v", back, err)
	}
	for _, bad := range []string{"pal:Zeal20", "didle1:xyz:Zeal20", "didle1:pal:Zeal", "didle1:pal:Zeal20|str=lots"} {
		if _, err := Decode(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
	if _, err := Encode(load(t, "necromancer-respec.json")); err == nil {
		t.Error("respec file encoded to one line")
	}
}

func TestImportKolbot(t *testing.T) {
	goals, stats, err := ImportKolbot(
		[]byte(`[["Blessed Hammer",20,true],["Concentration",20],["Vigor",20,false]]`),
		[]byte(`[["strength",60],["vitality","all"]]`))
	if err != nil {
		t.Fatal(err)
	}
	if !goals[0].Hold || !goals[1].Hold || goals[2].Hold || !stats[1].To.Rest || stats[0].Stat != "str" {
		t.Errorf("%+v %+v", goals, stats)
	}
	b := &Build{Schema: Schema, Name: "kolbot", Class: "paladin", Skills: goals, Stats: stats}
	if ps := Validate(b, fixtureRules{}); len(ps) > 0 {
		t.Errorf("imported: %v", ps)
	}
	if _, _, err := ImportKolbot(nil, []byte(`[["dexterity","block"]]`)); err == nil {
		t.Error("block accepted")
	}
	if _, _, err := ImportKolbot([]byte(`[[36,20]]`), nil); err == nil {
		t.Error("numeric skill id accepted")
	}
}

// TestRoundTripTables: Expand(FromRows(table)) against the repo's real tables. Cumulative allocation at clvl 10, 20,
// 30 and the last row must match, and (stronger) every row must.
func TestRoundTripTables(t *testing.T) {
	tables := []string{
		"sorceress_blizzard_endgame90", "sorceress_normal30", "paladin_hammerdin_endgame90", "paladin-normal30",
		"necromancer_summoner_endgame90", "necromancer_leveling_normal30", "barbarian_warcry_endgame90", "barbarian_melee_normal30",
	}
	for _, name := range tables {
		raw, err := plandata.FS.ReadFile("build-" + name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		lb, err := leveler.ParseBuild(raw)
		if err != nil {
			t.Fatal(err)
		}
		rules := looseRules{lb.Class}
		b, notes := FromRows(lb.Rows, lb.Class, rules)
		if len(notes) > 0 {
			t.Logf("%s notes: %v", name, notes)
		}
		if ps := Validate(b, rules); len(ps) > 0 {
			t.Errorf("%s: FromRows output invalid: %v", name, ps)
			continue
		}
		// the file survives JSON and the one-line-free canonical form
		out, _ := Marshal(b)
		b, err = Parse(out)
		if err != nil {
			t.Fatal(err)
		}
		e, err := Expand(b, rules)
		if err != nil {
			t.Fatal(err)
		}
		lastRow := lb.Rows[len(lb.Rows)-1].Clvl
		for _, L := range []int{10, 20, 30, lastRow} {
			if L > lastRow {
				continue
			}
			ws, wa := Cumulative(lb.Rows, L)
			gs, ga := Cumulative(e.Pre, L)
			if !reflect.DeepEqual(ws, gs) || wa != ga {
				t.Errorf("%s clvl %d: table %v %v, expanded %v %v", name, L, ws, wa, gs, ga)
			}
		}
		exact := 0
		for _, r := range lb.Rows {
			ws, wa := Cumulative(lb.Rows, r.Clvl)
			gs, ga := Cumulative(e.Pre, r.Clvl)
			if reflect.DeepEqual(ws, gs) && wa == ga {
				exact++
			}
		}
		t.Logf("%s: %d table rows, %d goals, %d stat goals, %d quest rows, cumulative equal at %d/%d rows, json %d bytes",
			name, len(lb.Rows), len(b.Skills), len(b.Stats), len(b.Quest), exact, len(lb.Rows), len(out))
		if exact != len(lb.Rows) {
			t.Errorf("%s: not every row round-trips", name)
		}
	}
}

func TestStdQuestsTotal(t *testing.T) {
	n := 0
	for _, q := range StdQuests() {
		n += q.Skill
	}
	if n != MaxQuestSkill {
		t.Errorf("std quests %d, max %d", n, MaxQuestSkill)
	}
	var _ = json.Marshal
}
