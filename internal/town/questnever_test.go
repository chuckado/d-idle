package town_test

import (
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/town"
)

// A data build whose items.txt quest column is 0 for the Act 1 chain items must still never sell them (sandbox hero
// 2026-11-21: Scroll of Inifuss bks and Key to the Cairn Stones bkd were not in the town's never-sell list). The check
// reads the base row only, so it is the same for every class (the class-keyed behaviour is in internal/task).
func TestNeverSellQuestItemsWithoutQuestColumn(t *testing.T) {
	u := d2.Unit{Class: 1, Item: &d2.Item{}}
	for _, code := range []string{"bks", "bkd", "hdm", "leg"} {
		if r := town.NeverSellReason(&data.ItemBase{Code: code}, u); r == "" {
			t.Errorf("%s may be sold", code)
		}
	}
	if r := town.NeverSellReason(&data.ItemBase{Code: "crs"}, u); r != "" {
		t.Errorf("a plain sword is refused: %s", r)
	}
}
