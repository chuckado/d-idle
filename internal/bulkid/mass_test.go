package bulkid

import (
	"encoding/json"
	"strings"
	"testing"

	"d2lod-bot/internal/vault"
)

func qn(q int) string {
	return [...]string{"?", "low", "normal", "superior", "magic", "set", "rare", "unique", "crafted"}[q]
}

func rec(id string, chain []string, q, ilvl, value int) vault.Rec {
	return vault.Rec{Item: vault.Item{ID: id, Raw: "deadbeef", Chain: chain, Quality: q, Ilvl: ilvl, Value: value}, State: vault.Recorded}
}

func vaultRecs(build string) []vault.Rec {
	if build == "Sorceress" {
		return []vault.Rec{
			rec("a", []string{"ring", "jewl"}, 4, 10, 50),
			rec("b", []string{"amul"}, 6, 30, 200),
			rec("c", []string{"helm", "armo"}, 4, 10, 20),                                                             // not selected
			{Item: vault.Item{ID: "d", Chain: []string{"ring"}, Quality: 4, Identified: true}, State: vault.Recorded}, // identified
			{Item: vault.Item{ID: "e", Chain: []string{"ring"}, Quality: 4}, State: vault.Recorded, RevQ: true},       // queued
		}
	}
	return []vault.Rec{
		rec("a", []string{"char"}, 4, 25, 80),
		rec("b", []string{"swor", "weap"}, 6, 25, 300),
		rec("c", []string{"swor", "weap"}, 6, 10, 90), // below the rare gear ilvl
		rec("z", []string{"jewl"}, 4, 5, 10),
	}
}

func TestMassRevealCostModes(t *testing.T) {
	f := mustFilter(t)
	for _, build := range []string{"Sorceress", "Paladin"} {
		recs := vaultRecs(build)
		cain := MassReveal(f, recs, qn, Price{Free: true, Charges: 0})
		if cain.Mode != ViaCain || cain.Count != 2 && build == "Sorceress" || !strings.Contains(cain.Line, "free at Cain") {
			t.Fatalf("%s cain: %+v", build, cain)
		}
		scroll := MassReveal(f, recs, qn, Price{Charges: 1, ScrollGold: 80})
		if scroll.Mode != ViaScroll || scroll.Count != 1 || scroll.Skipped == 0 || scroll.Gold != 80 {
			t.Fatalf("%s scroll: %+v", build, scroll)
		}
		none := MassReveal(f, recs, qn, Price{Charges: 0})
		if none.Mode != "none" || none.Count != 0 || len(none.IDs) != 0 || none.Skipped == 0 {
			t.Fatalf("%s none: %+v", build, none)
		}
		unknown := MassReveal(f, recs, qn, Price{Charges: -1})
		if unknown.Count != 0 {
			t.Fatalf("%s unknown charges must not queue: %+v", build, unknown)
		}
	}
}

// Best first, and identified or already queued records are never listed.
func TestMassRevealOrderAndExclusions(t *testing.T) {
	f := mustFilter(t)
	m := MassReveal(f, vaultRecs("Sorceress"), qn, Price{Free: true})
	if len(m.IDs) != 2 || m.IDs[0] != "b" || m.IDs[1] != "a" {
		t.Fatalf("%+v", m)
	}
}

// Must-fail: the preview of unidentified records carries ids and words only, never record bytes or property text.
func TestMassPreviewLeaksNothing(t *testing.T) {
	f := mustFilter(t)
	recs := vaultRecs("Paladin")
	recs[0].Props = "dmg:0:99;res:0:40"
	b, _ := json.Marshal(MassReveal(f, recs, qn, Price{Free: true}))
	for _, w := range []string{"deadbeef", "dmg", "res", "Props", "raw"} {
		if strings.Contains(string(b), w) {
			t.Fatalf("preview names %q: %s", w, b)
		}
	}
}
