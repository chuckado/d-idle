package bulkid

import (
	"encoding/json"
	"strings"
	"testing"

	"d2lod-bot/internal/env"
)

func mustFilter(t *testing.T) *Filter {
	t.Helper()
	f, err := DefaultFilter()
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// items: a hero-neutral mix. The two "builds" differ only in what they carry; the planner has no class input.
func itemsFor(build string) []Item {
	switch build {
	case "Sorceress":
		return []Item{
			{ID: 1, Code: "rin", Types: []string{"ring", "jewelry"}, Quality: "magic", Ilvl: 12, Value: 90, Place: PlacePack},
			{ID: 2, Code: "ssd", Types: []string{"sword", "weap"}, Quality: "rare", Ilvl: 30, Value: 400, Place: PlaceStash},
			{ID: 3, Code: "cap", Types: []string{"helm", "armo"}, Quality: "magic", Ilvl: 10, Value: 20, Place: PlacePack},
		}
	default: // Paladin
		return []Item{
			{ID: 1, Code: "amu", Types: []string{"amul", "jewelry"}, Quality: "rare", Ilvl: 20, Value: 300, Place: PlaceStash},
			{ID: 2, Code: "gsd", Types: []string{"sword", "weap"}, Quality: "rare", Ilvl: 14, Value: 60, Place: PlacePack},
			{ID: 3, Code: "cm1", Types: []string{"char"}, Quality: "magic", Ilvl: 25, Value: 100, Place: PlacePack},
		}
	}
}

func TestPolicyNeverAndManualDoNothing(t *testing.T) {
	f := mustFilter(t)
	for _, build := range []string{"Sorceress", "Paladin"} {
		for _, pol := range []int{env.IdentifyNever, env.IdentifyManual} {
			tu := env.DefaultIdentifyTuning()
			tu.Policy = pol
			p := Make(f, tu, Hero{CainOpen: true, Charges: 50, Gold: 99999, ScrollPrice: 80}, itemsFor(build))
			if len(p.Picks) != 0 || p.NeedBuy != 0 || p.Reason == "" {
				t.Fatalf("%s policy %d: %+v", build, pol, p)
			}
		}
	}
}

func TestCainFreeWaitsForCainAndSpendsNothing(t *testing.T) {
	f := mustFilter(t)
	tu := env.DefaultIdentifyTuning()
	for _, build := range []string{"Sorceress", "Paladin"} {
		p := Make(f, tu, Hero{CainOpen: false, Charges: 50, Gold: 99999, ScrollPrice: 80}, itemsFor(build))
		if len(p.Picks) != 0 || p.Reason != WhyNoCain {
			t.Fatalf("%s: %+v", build, p)
		}
		p = Make(f, tu, Hero{CainOpen: true}, itemsFor(build))
		if len(p.Picks) != 2 || p.NeedBuy != 0 {
			t.Fatalf("%s: want the 2 filtered items via Cain, got %+v", build, p)
		}
		for _, k := range p.Picks {
			if k.Via != ViaCain {
				t.Fatalf("via %q", k.Via)
			}
		}
	}
}

func TestTripCapAndOrderAreDeterministic(t *testing.T) {
	f := mustFilter(t)
	tu := env.DefaultIdentifyTuning()
	tu.TripCap = 2
	var items []Item
	for i := uint32(1); i <= 6; i++ {
		items = append(items, Item{ID: i, Types: []string{"ring"}, Quality: "magic", Value: 10, Place: PlacePack})
	}
	items = append(items, Item{ID: 9, Types: []string{"ring"}, Quality: "unique", Value: 1, Place: PlacePack})
	a := Make(f, tu, Hero{CainOpen: true}, items)
	if len(a.Picks) != 2 || a.Picks[0].Item.ID != 9 || a.Picks[1].Item.ID != 1 {
		t.Fatalf("%+v", a.Picks)
	}
	if len(a.Skips) != 5 || a.Skips[0].Reason != WhyCap {
		t.Fatalf("%+v", a.Skips)
	}
	b := Make(f, tu, Hero{CainOpen: true}, items)
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	if string(ja) != string(jb) {
		t.Fatal("not deterministic")
	}
}

func TestScrollsRespectCapMinValueAndGoldFloor(t *testing.T) {
	f := mustFilter(t)
	tu := env.DefaultIdentifyTuning()
	tu.Policy = env.IdentifyScrolls
	tu.MinValue = 50
	tu.MaxScrolls = 10
	for _, build := range []string{"Sorceress", "Paladin"} {
		items := itemsFor(build)
		// Gold 700, floor 600, scroll 80: one purchase would leave 620 (ok), two would leave 540 (no).
		p := Make(f, tu, Hero{Gold: 700, ScrollPrice: 80}, items)
		if p.NeedBuy != 1 || len(p.Picks) != 1 {
			t.Fatalf("%s: gold floor: %+v", build, p)
		}
		spent := p.NeedBuy * 80
		if 700-spent < tu.GoldFloor {
			t.Fatalf("%s: broke the floor", build)
		}
		// Held charges are used before any purchase.
		p = Make(f, tu, Hero{Charges: 5, Gold: 0, ScrollPrice: 80}, items)
		if p.NeedBuy != 0 || len(p.Picks) != 2 {
			t.Fatalf("%s: held charges: %+v", build, p)
		}
		// Per-visit cap.
		tu2 := tu
		tu2.MaxScrolls = 1
		p = Make(f, tu2, Hero{Charges: 5, ScrollPrice: 80}, items)
		if len(p.Picks) != 1 {
			t.Fatalf("%s: cap: %+v", build, p)
		}
		// Unknown price buys nothing.
		p = Make(f, tu, Hero{Gold: 99999}, items)
		if p.NeedBuy != 0 || len(p.Picks) != 0 {
			t.Fatalf("%s: unknown price: %+v", build, p)
		}
	}
}

func TestPlanJSONHasNoAffixData(t *testing.T) {
	f := mustFilter(t)
	p := Make(f, env.DefaultIdentifyTuning(), Hero{CainOpen: true}, itemsFor("Sorceress"))
	b, _ := json.Marshal(p)
	for _, w := range []string{"prefix", "suffix", "affix", "props", "stat", "raw", "norm"} {
		if strings.Contains(strings.ToLower(string(b)), w) {
			t.Fatalf("plan JSON names %q: %s", w, b)
		}
	}
}

func TestFilterTableValidation(t *testing.T) {
	bad := map[string]string{
		"schema":  `{"schema":"x","rules":[]}`,
		"unknown": `{"schema":"bulkid.filter/1","rules":[],"nope":1}`,
		"empty":   `{"schema":"bulkid.filter/1","rules":[{"id":"a","types":[],"quality":["magic"]}]}`,
		"repeat":  `{"schema":"bulkid.filter/1","rules":[{"id":"a","types":["*"],"quality":["magic"]},{"id":"a","types":["*"],"quality":["rare"]}]}`,
	}
	for name, doc := range bad {
		if _, err := ParseFilter([]byte(doc)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
