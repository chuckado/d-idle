package loot

import (
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
)

// FIXPRICE B2: the vendor pays cost / 2 (the old estimate said cost / 4 x a guessed quality factor).
func TestSellValueFollowsGameFormula(t *testing.T) {
	b := &data.ItemBase{Kind: data.KindWeapon, Cost: 1000}
	for _, c := range []struct {
		name   string
		it     *d2.Item
		maxBuy int
		want   int
	}{
		{"normal", &d2.Item{Quality: d2.QualityNormal}, 0, 500},
		{"unidentified magic: base only", &d2.Item{Quality: d2.QualityMagic}, 0, 500},
		{"unidentified rare: base only", &d2.Item{Quality: d2.QualityRare}, 0, 500},
		{"ethereal quarter", &d2.Item{Quality: d2.QualityNormal, Flags: d2.FlagEthereal}, 0, 125},
		{"identified low quality halves", &d2.Item{Quality: d2.QualityLow, Flags: d2.FlagIdentified}, 0, 250},
		{"capped at the act 1 normal max buy", &d2.Item{Quality: d2.QualityNormal}, 300, 300},
	} {
		if got := SellValue(nil, b, c.it, c.maxBuy); got != c.want {
			t.Errorf("%s: got %d want %d", c.name, got, c.want)
		}
	}
	if got := SellValue(nil, nil, &d2.Item{}, 0); got != 0 {
		t.Errorf("nil base: %d", got)
	}
}

func TestMaxBuyOf(t *testing.T) {
	for _, c := range []struct {
		p    d2.Player
		want int
	}{
		{d2.Player{Area: 1}, 5000},
		{d2.Player{Area: 109}, 25000},
		{d2.Player{Area: 75, Difficulty: d2.Nightmare}, 30000},
		{d2.Player{Area: 1, Difficulty: d2.Hell}, 35000},
	} {
		if got := MaxBuyOf(&c.p); got != c.want {
			t.Errorf("%+v: %d want %d", c.p, got, c.want)
		}
	}
}
