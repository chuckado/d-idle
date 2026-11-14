package lootrules

import (
	"strings"
	"testing"
)

func ptr[T any](v T) *T { return &v }

func mustParse(t *testing.T, s string) *RuleSet {
	t.Helper()
	rs, err := Parse([]byte(s))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return rs
}

const hdr = `"schema":"lootrules/1","name":"t","limits":{"maxActions":5}`

func TestParseErrorsNamePathAndLine(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"syntax", "{\n \"schema\": \n}}", "line 3"},
		{"unknown key", `{` + hdr + `,"rules":[{"id":"a","action":"stash","reason":"x","wen":1}]}`, "wen"},
		{"bad action", `{` + hdr + `,"rules":[{"id":"a","action":"equip","reason":"x"}]}`, "rules[0].action"},
		{"no reason", `{` + hdr + `,"rules":[{"id":"a","action":"stash"}]}`, "rules[0].reason"},
		{"dup id", `{` + hdr + `,"rules":[{"id":"a","action":"stash","reason":"x"},{"id":"a","action":"keep","reason":"x"}]}`, "duplicate id"},
		{"bad field", `{` + hdr + `,"rules":[{"id":"a","action":"keep","reason":"x","when":{"f":"colour","op":"eq","v":1}}]}`, "rules[0].when.f"},
		{"bad group", `{` + hdr + `,"rules":[{"id":"a","action":"keep","reason":"x","when":{"all":[{"f":"group","op":"in","v":["pets"]}]}}]}`, "rules[0].when.all[0].v"},
		{"two forms", `{` + hdr + `,"rules":[{"id":"a","action":"keep","reason":"x","when":{"all":[{"f":"ilvl","op":"ge","v":1}],"f":"ilvl"}}]}`, "exactly one"},
		{"pickup in pack", `{` + hdr + `,"rules":[{"id":"a","action":"pickup","scope":["pack"],"reason":"x"}]}`, "not valid in scope pack"},
		{"schema", `{"schema":"lootrules/2","name":"t","limits":{"maxActions":1},"rules":[]}`, "this build reads"},
		{"stat", `{` + hdr + `,"rules":[{"id":"a","action":"keep","reason":"x","when":{"f":"stat:luck","op":"ge","v":1}}]}`, "unknown field"},
		{"plus", `{` + hdr + `,"rules":[{"id":"a","action":"keep","reason":"x","when":{"f":"ilvl","op":"ge","v":1,"plus":{"lvl":1}}}]}`, "only usable_by"},
	}
	for _, c := range cases {
		_, err := Parse([]byte(c.in))
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: got %v, want it to mention %q", c.name, err, c.want)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	rs := Starter()
	b, err := rs.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	rs2, err := Parse(b)
	if err != nil || len(rs2.Rules) != len(rs.Rules) {
		t.Fatalf("round trip: %v", err)
	}
}

func decide(t *testing.T, rs *RuleSet, it Item, cx Ctx) Decision {
	t.Helper()
	c, err := Compile(rs, cx.Class, cx.Build)
	if err != nil {
		t.Fatal(err)
	}
	return c.Decide(&it, &cx)
}

func TestFirstMatchBooleansAndDefault(t *testing.T) {
	rs := mustParse(t, `{`+hdr+`,"defaults":{"pack":"sell"},"rules":[
	 {"id":"a","scope":["pack"],"action":"stash","reason":"good ring","when":{"all":[{"f":"type","op":"in","v":["ring"]},{"any":[{"f":"stat:fcr","op":"ge","v":10},{"f":"stat:life","op":"ge","v":20}]},{"not":{"f":"ethereal","op":"eq","v":true}}]}},
	 {"id":"b","scope":["pack"],"action":"keep","reason":"any ring","when":{"f":"type","op":"in","v":["ring"]}}]}`)
	cx := Ctx{Class: "Paladin"}
	ring := Item{Place: PlacePack, Types: []string{"ring"}, Quality: "rare", Identified: true, Stats: map[string]int{"fcr": 10}}
	if d := decide(t, rs, ring, cx); d.Action != Stash || d.Rule != "a" {
		t.Fatalf("good ring: %+v", d)
	}
	ring.Stats = map[string]int{"fcr": 5}
	if d := decide(t, rs, ring, cx); d.Action != Keep || d.Rule != "b" {
		t.Fatalf("weak ring: %+v", d)
	}
	if d := decide(t, rs, Item{Place: PlacePack, Quality: "normal", Value: 7}, cx); d.Action != Sell || d.Rule != "default" {
		t.Fatalf("default: %+v", d)
	}
}

func TestUnidentifiedIsExplicit(t *testing.T) {
	rs := mustParse(t, `{`+hdr+`,"rules":[
	 {"id":"stat","action":"stash","reason":"good","when":{"all":[{"f":"type","op":"in","v":["ring"]},{"f":"stat:fcr","op":"ge","v":10}]}},
	 {"id":"id","scope":["pack"],"action":"identify","reason":"find out","when":{"all":[{"f":"type","op":"in","v":["ring"]},{"f":"identified","op":"eq","v":false}]}},
	 {"id":"junk","scope":["pack"],"action":"sell","reason":"junk","when":{"f":"type","op":"in","v":["ring"]}}]}`)
	cx := Ctx{}
	un := Item{Place: PlacePack, Types: []string{"ring"}, Quality: "magic", Identified: false}
	d := decide(t, rs, un, cx)
	if d.Action != Identify || !d.Pending {
		t.Fatalf("unidentified magic ring must be identified, not sold: %+v", d)
	}
	un.Identified, un.Stats = true, map[string]int{"fcr": 3}
	if d := decide(t, rs, un, cx); d.Action != Sell || d.Pending {
		t.Fatalf("identified weak ring: %+v", d)
	}
	// NOT of an unknown stays unknown: a "not good" rule must not catch an unidentified ring
	rs2 := mustParse(t, `{`+hdr+`,"rules":[{"id":"x","scope":["pack"],"action":"sell","reason":"weak","when":{"not":{"f":"stat:fcr","op":"ge","v":10}}}]}`)
	un.Identified, un.Stats = false, nil
	if d := decide(t, rs2, un, cx); d.Action != Ignore || !d.Pending {
		t.Fatalf("not(unknown): %+v", d)
	}
	// ground: identified is unreadable, onUnknown match takes it
	rs3 := mustParse(t, `{`+hdr+`,"rules":[{"id":"g","scope":["ground"],"action":"keep","reason":"id later","onUnknown":"match","when":{"f":"identified","op":"eq","v":false}}]}`)
	g := Item{Place: PlaceGround, Types: []string{"ring"}, Quality: "magic", Hidden: HIdentified}
	if d := decide(t, rs3, g, cx); d.Action != Pickup || !d.Pending || d.Fate != Keep {
		t.Fatalf("ground unknown: %+v", d)
	}
}

func TestGroundDerivesPickupFromFate(t *testing.T) {
	rs := mustParse(t, `{`+hdr+`,"rules":[
	 {"id":"r","action":"stash","reason":"rune","prio":90,"when":{"f":"group","op":"in","v":["runes"]}},
	 {"id":"s","action":"sell","reason":"gem","when":{"f":"group","op":"in","v":["gems"]}},
	 {"id":"s2","action":"sell","pickup":true,"reason":"gear","when":{"f":"group","op":"in","v":["gear"]}}]}`)
	cx := Ctx{}
	if d := decide(t, rs, Item{Place: PlaceGround, Types: []string{"rune"}}, cx); d.Action != Pickup || d.Prio != 90 {
		t.Fatalf("rune: %+v", d)
	}
	if d := decide(t, rs, Item{Place: PlaceGround, Types: []string{"gem"}}, cx); d.Action != Ignore || d.Fate != Sell {
		t.Fatalf("gem: %+v", d)
	}
	if d := decide(t, rs, Item{Place: PlaceGround, Types: []string{"weap"}}, cx); d.Action != Pickup {
		t.Fatalf("gear: %+v", d)
	}
}

func TestInvariants(t *testing.T) {
	sellAll := mustParse(t, `{`+hdr+`,"rules":[{"id":"all","scope":["pack"],"action":"sell","reason":"everything"}]}`)
	cx := Ctx{Class: "Paladin"}
	for _, c := range []struct {
		name string
		it   Item
		want string
	}{
		{"equipment", Item{Place: "equipment", Quality: "normal"}, Ignore},
		{"belt", Item{Place: "belt", Quality: "normal"}, Ignore},
		{"keep mark", Item{Place: PlacePack, Quality: "normal", Mark: "keep"}, Ignore},
		{"start kit", Item{Place: PlacePack, Quality: "normal", StartItem: true}, Keep},
		{"quest", Item{Place: PlacePack, Quality: "normal", Quest: true}, Keep},
		{"tome", Item{Place: PlacePack, Quality: "normal", Code: "tbk", Types: []string{"book"}}, Keep},
		{"rune without confirmation", Item{Place: PlacePack, Quality: "normal", Types: []string{"rune"}}, Ignore},
		{"unique without confirmation", Item{Place: PlacePack, Quality: "unique", Types: []string{"swor"}}, Sell},
	} {
		if d := decide(t, sellAll, c.it, cx); d.Action != c.want {
			t.Errorf("%s: %+v want %s", c.name, d, c.want)
		}
	}
	// drop of a magic item: never without the rule's confirm AND the set's acknowledgement
	mk := func(confirm bool, ack string) *RuleSet {
		return mustParse(t, `{`+hdr+ack+`,"rules":[{"id":"d","scope":["pack"],"action":"drop","reason":"clutter","confirm":`+map[bool]string{true: "true", false: "false"}[confirm]+`}]}`)
	}
	magic := Item{Place: PlacePack, Quality: "magic", Value: 3}
	white := Item{Place: PlacePack, Quality: "normal", Value: 3}
	if d := decide(t, mk(false, ""), magic, cx); d.Action != Ignore {
		t.Errorf("unconfirmed drop of magic: %+v", d)
	}
	if d := decide(t, mk(true, ""), magic, cx); d.Action != Ignore {
		t.Errorf("confirm without ack: %+v", d)
	}
	if d := decide(t, mk(true, `,"ack":["drop-valuable"]`), magic, cx); d.Action != Drop {
		t.Errorf("confirmed and acknowledged: %+v", d)
	}
	if d := decide(t, mk(false, ""), white, cx); d.Action != Drop {
		t.Errorf("white junk drops: %+v", d)
	}
	unreadable := Item{Place: PlacePack, Quality: "normal", Hidden: HQuality}
	if d := decide(t, mk(false, ""), unreadable, cx); d.Action != Ignore {
		t.Errorf("unreadable quality is never dropped: %+v", d)
	}
	// no equip action exists
	if _, err := Parse([]byte(`{` + hdr + `,"rules":[{"id":"e","action":"equip","reason":"x"}]}`)); err == nil {
		t.Error("equip must not parse")
	}
}

func TestPlanCapOrderAndDeterminism(t *testing.T) {
	rs := Starter()
	c, _ := Compile(rs, "Sorceress", "")
	cx := &Ctx{Class: "Sorceress", PackFullPct: 90}
	var items []PlanItem
	for i := 0; i < 20; i++ {
		items = append(items, PlanItem{uint32(100 - i), &Item{Types: []string{"rune"}, Quality: "normal", Value: i % 4}})
	}
	items = append(items, PlanItem{7, &Item{Types: []string{"club"}, Quality: "normal", Value: 5}})
	a := c.Plan(items, cx)
	if len(a) != rs.Limits.MaxActions {
		t.Fatalf("cap: %d", len(a))
	}
	for _, e := range a {
		if e.Action != Stash {
			t.Fatalf("stash first, drop is last and cut by the cap: %+v", e)
		}
	}
	b := c.Plan(items, cx)
	for i := range a {
		if a[i].ID != b[i].ID {
			t.Fatal("plan is not deterministic")
		}
	}
	if a[0].ID != 100-3 && a[0].ID >= 100 {
		t.Fatalf("best value first, ties by id: %d", a[0].ID)
	}
}

func TestExplainShowsEveryRule(t *testing.T) {
	c, _ := Compile(Starter(), "Paladin", "")
	tr := c.Explain(&Item{Place: PlacePack, Types: []string{"rune"}, Quality: "normal"}, &Ctx{})
	hit := 0
	for _, x := range tr {
		if x.Result == "match" {
			hit++
		}
	}
	if len(tr) < 10 || hit == 0 {
		t.Fatalf("trace %v", tr)
	}
}

func TestClassAndBuildOverlay(t *testing.T) {
	rs := mustParse(t, `{`+hdr+`,"rules":[
	 {"id":"am","scope":["pack"],"classes":["Amazon"],"builds":["javazon"],"action":"keep","reason":"javelins","when":{"f":"group","op":"in","v":["javelins"]}},
	 {"id":"s","scope":["pack"],"action":"sell","reason":"javelins sell","when":{"f":"group","op":"in","v":["javelins"]}}]}`)
	jav := Item{Place: PlacePack, Quality: "normal", Types: []string{"jave"}}
	if d := decide(t, rs, jav, Ctx{Class: "Amazon", Build: "javazon"}); d.Rule != "am" {
		t.Fatalf("amazon: %+v", d)
	}
	for _, cl := range []Ctx{{Class: "Amazon", Build: "bowazon"}, {Class: "Paladin", Build: "javazon"}, {Class: "NoSuchClass"}} {
		if d := decide(t, rs, jav, cl); d.Rule != "s" {
			t.Fatalf("%+v: %+v", cl, d)
		}
	}
}

func TestRequirementsHeroMercAndLater(t *testing.T) {
	rs := mustParse(t, `{`+hdr+`,"rules":[
	 {"id":"now","scope":["pack"],"action":"keep","reason":"you can wear it","when":{"f":"usable_by","op":"eq","v":"hero"}},
	 {"id":"merc","scope":["pack"],"action":"keepmerc","reason":"for the mercenary","when":{"f":"usable_by","op":"eq","v":"merc"}},
	 {"id":"later","scope":["pack"],"action":"keep","reason":"wearable by clvl 30","when":{"f":"hero_usable_from","op":"le","v":30}},
	 {"id":"soon","scope":["pack"],"action":"keep","reason":"two levels away","when":{"f":"usable_by","op":"eq","v":"hero","plus":{"lvl":2}}},
	 {"id":"orb","scope":["pack"],"action":"stash","reason":"another class's item","when":{"f":"class_restricted","op":"in","v":["Sorceress"]}},
	 {"id":"rest","scope":["pack"],"action":"sell","reason":"cannot use it","when":{"f":"value","op":"ge","v":0}}]}`)
	armor := Item{Place: PlacePack, Quality: "normal", ReqLvl: 24, ReqStr: 50, MercTypes: []string{"Desert Guard"}, Value: 80, Types: []string{"tors"}}
	orb := Item{Place: PlacePack, Quality: "normal", ClassOnly: "Sorceress", ReqLvl: 10, Value: 40, Types: []string{"orb"}}
	pal := &Actor{Class: "Paladin", Level: 22, Str: 60, Dex: 30}
	sor := &Actor{Class: "Sorceress", Level: 22, Str: 20, Dex: 25}
	guard := &Actor{Class: "Desert Guard", Level: 25, Str: 70, Dex: 40}
	got := func(it Item, hero, merc *Actor) Decision {
		return decide(t, rs, it, Ctx{Class: hero.Class, Hero: hero, Merc: merc})
	}

	if d := got(armor, pal, guard); d.Rule != "merc" || d.Action != KeepMerc {
		t.Errorf("the guard (level 25) can wear it, the level 22 paladin cannot yet: %+v", d)
	}
	// a merc of another hireling type cannot
	if d := got(armor, pal, &Actor{Class: "Rogue Scout", Level: 40, Str: 90, Dex: 90}); d.Rule != "later" {
		t.Errorf("a Rogue Scout cannot wear it: %+v", d)
	}
	if d := got(armor, pal, guard); d.Rule == "now" {
		t.Errorf("level 24 armor at clvl 22: %+v", d)
	}
	pal.Level = 24
	if d := got(armor, pal, nil); d.Rule != "now" {
		t.Errorf("clvl 24 paladin wears it: %+v", d)
	}
	pal.Level = 22
	if d := got(armor, pal, nil); d.Rule != "later" {
		t.Errorf("clvl 22 paladin, level 24 armor, wearable by 30: %+v", d)
	}
	sor.Level = 22
	if d := got(armor, sor, nil); d.Rule == "later" || d.Rule == "now" {
		t.Errorf("sorceress with 20 str never wears 50 str armor (U: stat points could change it): %+v", d)
	}
	if d := got(orb, sor, nil); d.Rule != "now" {
		t.Errorf("sorceress uses her orb: %+v", d)
	}
	if d := got(orb, pal, nil); d.Rule != "orb" {
		t.Errorf("paladin cannot use a sorceress orb, class rule: %+v", d)
	}
	// unread hero stats: the requirement rules answer unknown, never "cannot wear"
	if d := decide(t, rs, armor, Ctx{Hero: &Actor{Class: "Paladin"}}); !d.Pending {
		t.Errorf("unread hero must leave the decision pending: %+v", d)
	}
	c, _ := Compile(rs, "Paladin", "")
	if !c.UsesActors {
		t.Error("UsesActors")
	}
	a := Ctx{Hero: pal, Merc: guard}
	f1 := a.Fingerprint()
	pal.Level++
	if a.Fingerprint() == f1 {
		t.Error("level up must change the fingerprint")
	}
}

func TestDurabilityFields(t *testing.T) {
	rs := mustParse(t, `{`+hdr+`,"rules":[
	 {"id":"b","scope":["pack"],"action":"sell","reason":"broken junk","when":{"all":[{"f":"broken","op":"eq","v":true},{"f":"quality","op":"in","v":["normal"]}]}},
	 {"id":"e","scope":["pack"],"action":"keep","reason":"worn out","when":{"f":"durpct","op":"lt","v":20}}]}`)
	if d := decide(t, rs, Item{Place: PlacePack, Quality: "normal", MaxDur: 20, Durability: 0}, Ctx{}); d.Rule != "b" {
		t.Fatalf("%+v", d)
	}
	if d := decide(t, rs, Item{Place: PlacePack, Quality: "magic", MaxDur: 20, Durability: 2}, Ctx{}); d.Rule != "e" {
		t.Fatalf("%+v", d)
	}
}

func BenchmarkStarterGround(b *testing.B) {
	c, _ := Compile(Starter(), "Sorceress", "")
	cx := &Ctx{Class: "Sorceress", CLvl: 12, Gold: 400}
	items := make([]Item, 0, 40)
	for i := 0; i < 40; i++ {
		items = append(items, Item{Place: PlaceGround, Code: "cap", Types: []string{"helm", "armo"}, Quality: []string{"normal", "magic", "rare", "unique"}[i%4], Value: i * 7, ILvl: 10, Hidden: HIdentified})
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range items {
			c.Decide(&items[j], cx)
		}
	}
}
