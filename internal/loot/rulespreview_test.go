package loot

import (
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/lootrules"
)

// The dry run decides the pack and stash as they are, shows what a candidate would change against the active set, and
// acts on nothing.
func TestPreviewShowsTheChangeAndTouchesNothing(t *testing.T) {
	for _, cls := range []d2.Class{d2.Sorceress, d2.Amazon} {
		l, db := synthLooter(t)
		l.e.Tune.Bot.LootShadow = 1
		sh := NewShadow(lootrules.Starter())
		l.SetShadow(sh)
		if _, err := sh.Preview(nil, lootrules.Starter()); err == nil {
			t.Fatal("a preview before any pass must say so")
		}
		ring := synthUnit(t, db, 31, "rin", d2.QualityRare, 0, d2.ContInventory)
		rune1 := synthUnit(t, db, 32, "r01", d2.QualityNormal, d2.FlagIdentified, d2.ContStash)
		hp := synthUnit(t, db, 33, "hp1", d2.QualityNormal, d2.FlagIdentified, d2.ContInventory)
		s := &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Class: cls, Level: 12, Gold: 500, Area: 2}, Inventory: &d2.Inventory{Items: []d2.Unit{ring, rune1, hp}}}
		if l.ShadowPass(s, 30, 10) == nil {
			t.Fatal("no pass")
		}
		active := lootrules.Starter()
		doc, err := sh.Preview(nil, active)
		if err != nil || doc.Schema != lootrules.PreviewSchema || len(doc.Items) != 3 || doc.Changed != 0 {
			t.Fatalf("%v: preview of the active set: %v %+v", cls, err, doc)
		}
		// candidate: the jewelry sell rule off, and the rune rule stops stashing
		cand := lootrules.Starter()
		for i := range cand.Rules {
			if cand.Rules[i].ID == "s-jewelry-sell" || cand.Rules[i].ID == "t-runes" {
				cand.Rules[i].Off = true
			}
		}
		doc, err = sh.Preview(cand, active)
		if err != nil {
			t.Fatal(err)
		}
		var ringRow *lootrules.PreviewRow
		for i := range doc.Items {
			if doc.Items[i].ID == 31 {
				ringRow = &doc.Items[i]
			}
		}
		if ringRow == nil || !ringRow.Changed || ringRow.WasSell != lootrules.Sell || ringRow.Sell == nil || ringRow.Sell.Action != lootrules.Keep || !strings.Contains(ringRow.Name, "Ring") && ringRow.Name == "" {
			t.Errorf("%v: ring row %+v", cls, ringRow)
		}
		if doc.Changed < 1 || doc.Items[0].Changed == false {
			t.Errorf("changed rows must come first: %+v", doc.Items)
		}
		// a broken candidate is a field error, not a panic
		if _, err := sh.Preview(&lootrules.RuleSet{Schema: "lootrules/9"}, active); err == nil {
			t.Error("a broken candidate previewed")
		}
		// nothing was counted by the dry run
		if r := sh.Report(true); r.Decisions != 0 {
			t.Errorf("the dry run touched the shadow counters: %d", r.Decisions)
		}
	}
}
