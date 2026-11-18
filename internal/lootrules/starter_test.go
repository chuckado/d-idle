package lootrules

import (
	"testing"

	"d2lod-bot/internal/tidy"
)

// The Starter pack rules must answer exactly like today's tidy policy (docs/tidy.md): same action for every item, same
// why text for every table row, for two classes and an Amazon override.
func TestStarterPackEqualsTidyPolicy(t *testing.T) {
	pol, err := tidy.Default()
	if err != nil {
		t.Fatal(err)
	}
	rs := Starter()
	name := map[string]string{}
	for _, r := range rs.Rules {
		name[r.ID] = r.Name
	}
	type row struct {
		label string
		f     tidy.Facts
	}
	mk := func(label, typ, quality string, v int, mod func(*tidy.Facts)) row {
		f := tidy.Facts{Code: "xxx", Type: typ, Quality: quality, Place: "pack", Identified: true, SellValue: v}
		if mod != nil {
			mod(&f)
		}
		return row{label, f}
	}
	rows := []row{
		mk("tome", "book", "normal", 5, func(f *tidy.Facts) { f.Code = "tbk" }),
		mk("key", "key", "normal", 1, func(f *tidy.Facts) { f.Code = "key" }),
		mk("scroll", "scro", "normal", 1, func(f *tidy.Facts) { f.Code = "isc" }),
		mk("potion", "hpot", "normal", 30, nil),
		mk("rune", "rune", "normal", 10, nil),
		mk("gem", "gemb", "normal", 50, nil),
		mk("skull", "skul", "normal", 50, nil),
		mk("charm", "scha", "magic", 40, nil),
		mk("grand charm", "lcha", "rare", 40, nil),
		mk("jewel", "jewl", "magic", 40, nil),
		mk("unique sword", "swor", "unique", 90, nil),
		mk("set boots", "boot", "set", 90, nil),
		mk("rare ring", "ring", "rare", 150, nil),
		mk("crafted amulet", "amul", "crafted", 150, nil),
		mk("rare circlet", "circ", "rare", 150, nil),
		mk("magic ring", "ring", "magic", 60, nil),
		mk("magic amulet unidentified", "amul", "magic", 60, func(f *tidy.Facts) { f.Identified = false }),
		mk("magic circlet", "circ", "magic", 60, nil),
		mk("4os white armor", "tors", "normal", 40, func(f *tidy.Facts) { f.Sockets = 4 }),
		mk("4os superior shield", "shld", "superior", 15, func(f *tidy.Facts) { f.Sockets = 4 }),
		mk("4os magic armor", "tors", "magic", 15, func(f *tidy.Facts) { f.Sockets = 4 }),
		mk("2os white armor worth 30", "tors", "normal", 30, func(f *tidy.Facts) { f.Sockets = 2 }),
		mk("white club worth 5", "club", "normal", 5, nil),
		mk("low club worth 5", "club", "low", 5, nil),
		mk("magic wand worth 5", "wand", "magic", 5, nil),
		mk("magic sword worth 100 unidentified", "swor", "magic", 100, func(f *tidy.Facts) { f.Identified = false }),
		mk("white bow worth 25", "bow", "normal", 25, nil),
		mk("javelins", "jave", "normal", 40, nil),
		mk("amazon javelins", "ajav", "superior", 60, nil),
		mk("worthless unknown", "xxxx", "normal", 0, nil),
		mk("worn rune", "rune", "normal", 10, func(f *tidy.Facts) { f.Place = "equipment" }),
		mk("stash rune", "rune", "normal", 10, func(f *tidy.Facts) { f.Place = "stash" }),
		mk("keep-marked rune", "rune", "normal", 10, func(f *tidy.Facts) { f.Mark = "keep" }),
		mk("sell-marked sword", "swor", "magic", 80, func(f *tidy.Facts) { f.Mark = "sell" }),
		mk("start kit", "swor", "normal", 60, func(f *tidy.Facts) { f.StartItem = true }),
		mk("quest item", "ques", "normal", 60, func(f *tidy.Facts) { f.Quest = true }),
	}
	n := 0
	for _, class := range []string{"Paladin", "Sorceress", "Amazon"} {
		for _, pack := range []int{40, 90} {
			for _, stash := range []int{10, 95} {
				cx := tidy.Ctx{Class: class, PackFullPct: pack, StashFullPct: stash}
				c, err := Compile(rs, class, "")
				if err != nil {
					t.Fatal(err)
				}
				for _, r := range rows {
					want := pol.Decide(r.f, cx)
					it := &Item{Place: r.f.Place, Code: r.f.Code, Types: []string{r.f.Type, "misc"}, Quality: r.f.Quality, Identified: r.f.Identified,
						Sockets: r.f.Sockets, Quest: r.f.Quest, StartItem: r.f.StartItem, Value: r.f.SellValue, Mark: r.f.Mark}
					if r.f.Type == "ajav" {
						it.Types = []string{"ajav", "jave"}
					}
					got := c.Decide(it, &Ctx{Class: class, PackFullPct: pack, StashFullPct: stash})
					n++
					// tidy answers "drop" or "ignore" where the engine floors differ only in text; actions must be equal
					if string(want.Action) != got.Action {
						t.Errorf("%s/%s pack %d stash %d: tidy %s (%s), rules %s (%s)", class, r.label, pack, stash, want.Action, want.Rule, got.Action, got.Rule)
						continue
					}
					if name[got.Rule] != "" && name[got.Rule] != want.Rule || name[got.Rule] == "" && got.Rule != "invariant:"+want.Rule && want.Rule != "default" && want.Rule != "place" && want.Rule != "mark" && want.Rule != "start" && want.Rule != "quest" {
						// the amazon override carries tidy name "javelins"
						if !(got.Rule == "amazon-javelins" && want.Rule == "javelins") {
							t.Errorf("%s/%s: rule tidy %q, rules %q", class, r.label, want.Rule, got.Rule)
						}
					}
					if name[got.Rule] != "" && want.Why != got.Why {
						t.Errorf("%s/%s: why tidy %q, rules %q", class, r.label, want.Why, got.Why)
					}
				}
			}
		}
	}
	t.Logf("%d decisions compared", n)
}

func TestStarterGroundSpotChecks(t *testing.T) {
	c, err := Compile(Starter(), "Sorceress", "")
	if err != nil {
		t.Fatal(err)
	}
	cx := &Ctx{Class: "Sorceress", CLvl: 10, Gold: 100, Needs: func(k string, _ *Item) bool { return k == "potion" }}
	for _, tc := range []struct {
		label, want string
		it          Item
	}{
		{"rune", Pickup, Item{Types: []string{"rune"}, Quality: "normal"}},
		{"needed healing potion", Pickup, Item{Types: []string{"hpot"}, Quality: "normal"}},
		{"unique", Pickup, Item{Types: []string{"swor"}, Quality: "unique", Hidden: HIdentified}},
		{"white club, poor and worth 25", Pickup, Item{Code: "clb", Types: []string{"club", "weap"}, Quality: "normal", Value: 25}},
		{"white club, worth 5", Ignore, Item{Code: "clb", Types: []string{"club", "weap"}, Quality: "normal", Value: 5}},
		{"unidentified magic ring (label only)", Pickup, Item{Types: []string{"ring"}, Quality: "magic", Hidden: HIdentified}},
		{"magic sword (sells)", Ignore, Item{Types: []string{"swor", "weap"}, Quality: "magic", Hidden: HIdentified | HStats}},
	} {
		tc.it.Place = PlaceGround
		if d := c.Decide(&tc.it, cx); d.Action != tc.want {
			t.Errorf("%s: %+v", tc.label, d)
		}
	}
	cx.Gold = 5000
	it := Item{Place: PlaceGround, Code: "clb", Types: []string{"club", "weap"}, Quality: "normal", Value: 25}
	if d := c.Decide(&it, cx); d.Action != Ignore {
		t.Errorf("not poor any more: %+v", d)
	}
}
