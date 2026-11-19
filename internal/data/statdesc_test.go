package data

import (
	"reflect"
	"strings"
	"testing"
)

// statDB is a hand-built table set: the columns the wording reads, with the wordings of the game's string table as far as
// the file documents them (S, no real install needed).
func statDB() *DB {
	st := func(id int, name string, prio, fn, val int, pos, neg string) *StatCost {
		return &StatCost{ID: id, Name: name, DescPriority: prio, DescFunc: fn, DescVal: val, DescStrPos: pos, DescStrNeg: neg}
	}
	db := &DB{
		StatCosts: []*StatCost{
			st(0, "strength", 100, 1, 1, "ModStr1a", "ModStr1b"),
			st(1, "plainpct", 90, 2, 1, "pct", ""),
			st(2, "plain", 90, 3, 1, "plain", ""),
			st(3, "signedpct", 95, 4, 2, "spct", ""),
			st(4, "x128", 80, 5, 1, "x128", ""),
			st(5, "perlevel", 70, 6, 1, "perlvl", ""),
			st(6, "perlevelpct", 70, 8, 1, "perlvlp", ""),
			st(7, "noval", 60, 3, 0, "indest", ""),
			st(8, "classskills", 120, 13, 1, "cls", ""),
			st(9, "tab", 119, 14, 1, "tab", ""),
			st(10, "cast", 50, 15, 1, "ModStre8a", ""),
			st(11, "aura", 50, 16, 1, "ModStre9", ""),
			st(12, "fmt", 40, 19, 1, "ModFmt", ""),
			st(13, "negpct", 40, 20, 1, "negpct", ""),
			st(14, "charges", 30, 24, 1, "ModCharges", ""),
			st(15, "oskill", 110, 28, 1, "ModStr3", ""),
			st(16, "classonly", 109, 27, 1, "ModClassOnly", ""),
			st(17, "unknownfn", 10, 99, 1, "x", ""),
			st(18, "fireresist", 30, 2, 1, "ModFire", ""),
			st(19, "lightresist", 30, 2, 1, "ModLight", ""),
			st(20, "coldresist", 30, 2, 1, "ModCold", ""),
			st(21, "poisonresist", 30, 2, 1, "ModPois", ""),
			st(22, "firemindam", 20, 1, 1, "ModFmin", ""),
			st(23, "firemaxdam", 20, 1, 1, "ModFmax", ""),
			st(24, "poisonmindam", 20, 1, 1, "ModPmin", ""),
			st(25, "poisonmaxdam", 20, 1, 1, "ModPmax", ""),
			st(26, "poisonlength", 20, 1, 1, "ModPlen", ""),
			st(27, "hitpoints", 85, 1, 1, "ModLife", ""),
			st(28, "nostr", 10, 1, 1, "missingkey", ""),
		},
		Skills:  []*Skill{{ID: 36, Name: "Fire Bolt", Class: "sor"}, {ID: 54, Name: "Teleport", Class: "sor"}},
		Classes: []*CharClass{{ID: 0, Name: "Amazon", Code: "ama"}, {ID: 1, Name: "Sorceress", Code: "sor"}},
		strs: map[string]string{
			"ModStr1a": "to Strength", "ModStr1b": "Strength", "pct": "Enhanced Damage", "plain": "Defense", "spct": "Increased Attack Speed",
			"x128": "Magic Find", "perlvl": "to Strength", "perlvlp": "Damage", "indest": "Indestructible", "cls": "to Amazon Skill Levels",
			"ModStr3b": "to Sorceress Skill Levels", "tab": "x",
			"StrSklTabItem4": "+%d to Fire Skills (Sorceress Only)",
			"ModStre8a":      "%d%% Chance to cast level %d %s on striking", "ModStre9": "Level %d %s Aura When Equipped",
			"ModFmt": "Replenish Life +%d", "negpct": "Target Defense", "ModCharges": "Level %d %s (%d/%d Charges)",
			"ModStr3": "to %s", "ModClassOnly": "to %s (%s Only)", "x": "never",
			"ModFire": "Fire Resist", "ModLight": "Lightning Resist", "ModCold": "Cold Resist", "ModPois": "Poison Resist",
			"strModAllResistances":  "All Resistances",
			"strModFireDamageRange": "Adds %d-%d fire damage", "strModFireDamage": "+%d fire damage",
			"strModPoisonDamage": "+%d poison damage over %d seconds", "ModLife": "to Life",
		},
	}
	db.indexExtra()
	db.skillByID = map[int]*Skill{}
	for _, s := range db.Skills {
		db.skillByID[s.ID] = s
	}
	return db
}

func TestStatLinesPerDescfunc(t *testing.T) {
	db := statDB()
	for _, c := range []struct {
		name, props, want string
	}{
		{"1 plus value", "0:0:20", "+20 to Strength"},
		{"1 negative uses the negative string", "0:0:-5", "-5 Strength"},
		{"2 percent", "1:0:15", "15% Enhanced Damage"},
		{"3 plain", "2:0:40", "40 Defense"},
		{"4 signed percent, value after", "3:0:10", "Increased Attack Speed +10%"},
		{"5 x100/128", "4:0:64", "50% Magic Find"},
		{"6 per level", "5:0:4", "+0.5 to Strength per level"},
		{"8 per level percent", "6:0:16", "+2% Damage per level"},
		{"descval 0 shows no value", "7:0:1", "Indestructible"},
		{"13 class skills", "8:1:2", "+2 to Sorceress Skill Levels"},
		{"14 skill tab", "9:8:3", "+3 to Fire Skills (Sorceress Only)"},
		{"15 chance to cast", "10:" + itoa(36<<6|3) + ":5", "5% Chance to cast level 3 Fire Bolt on striking"},
		{"16 aura", "11:54:12", "Level 12 Teleport Aura When Equipped"},
		{"19 printf string", "12:0:7", "Replenish Life +7"},
		{"20 negated percent", "13:0:25", "-25% Target Defense"},
		{"24 charges", "14:" + itoa(36<<6|4) + ":" + itoa(10|20<<8), "Level 4 Fire Bolt (10/20 Charges)"},
		{"28 one skill", "15:36:3", "+3 to Fire Bolt"},
		{"27 one skill with class", "16:36:2", "+2 to Fire Bolt (Sorceress Only)"},
		{"unknown function is left out", "17:0:5", ""},
		{"stat without a string is left out", "28:0:5", ""},
		{"unknown stat id is left out", "999:0:5", ""},
		{"unknown skill is left out", "15:7777:3", ""},
		{"class out of range is left out", "8:9:2", ""},
		{"skill tab out of range is left out", "9:200:3", ""},
	} {
		got := db.StatLines(c.props)
		if c.want == "" {
			if len(got) != 0 {
				t.Errorf("%s: want nothing, got %v", c.name, got)
			}
			continue
		}
		if !reflect.DeepEqual(got, []string{c.want}) {
			t.Errorf("%s: got %v, want %q", c.name, got, c.want)
		}
	}
}

func TestStatLinesGroupedAndOrdered(t *testing.T) {
	db := statDB()
	// all four resists equal: one line; unequal: four lines
	if got := db.StatLines("18:0:20;19:0:20;20:0:20;21:0:20"); !reflect.DeepEqual(got, []string{"All Resistances +20"}) {
		t.Errorf("all resist: %v", got)
	}
	if got := db.StatLines("18:0:20;19:0:20;20:0:20;21:0:30"); len(got) != 4 || got[0] != "20% Fire Resist" {
		t.Errorf("unequal resists: %v", got)
	}
	// damage ranges and the equal-value form
	if got := db.StatLines("22:0:5;23:0:12"); !reflect.DeepEqual(got, []string{"Adds 5-12 fire damage"}) {
		t.Errorf("fire range: %v", got)
	}
	if got := db.StatLines("22:0:9;23:0:9"); !reflect.DeepEqual(got, []string{"+9 fire damage"}) {
		t.Errorf("fire flat: %v", got)
	}
	// poison: 256 per frame unit, 125 frames = 5 seconds: 10*125/256
	if got := db.StatLines("24:0:10;25:0:10;26:0:125"); !reflect.DeepEqual(got, []string{"+4 poison damage over 5 seconds"}) {
		t.Errorf("poison: %v", got)
	}
	// highest descpriority first, stable otherwise; the heading skill levels come first
	got := db.StatLines("2:0:40;8:1:2;0:0:20;27:0:30")
	want := []string{"+2 to Sorceress Skill Levels", "+20 to Strength", "40 Defense", "+30 to Life"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("order: %v want %v", got, want)
	}
	// junk in the list does not stop the others
	if got := db.StatLines(";x;0:0;a:b:c;0:0:3"); !reflect.DeepEqual(got, []string{"+3 to Strength"}) {
		t.Errorf("junk: %v", got)
	}
}

func TestGfmt(t *testing.T) {
	for _, c := range []struct {
		f    string
		args []any
		want string
		ok   bool
	}{
		{"%d%% of %s", []any{5, "x"}, "5% of x", true},
		{"%+d", []any{5}, "+5", true},
		{"%d", nil, "", false},
		{"no verbs", []any{1}, "", false},
		{"%d", []any{"s"}, "", false},
		{"%x", []any{1}, "", false},
		{"", nil, "", false},
		{"50%", nil, "", false},
	} {
		got, ok := gfmt(c.f, c.args...)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("gfmt(%q,%v) = %q,%v", c.f, c.args, got, ok)
		}
	}
	if strings.Contains(statDB().StatLines("0:0:1")[0], "%") {
		t.Error("raw format left in a line")
	}
}
