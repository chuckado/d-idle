package leveler

import (
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
)

// FIXPRICE B3: quality extras only for identified items, armor scaled by defence / max defence.
func TestEstimatePriceFollowsGameRules(t *testing.T) {
	weap := &data.ItemBase{Kind: data.KindWeapon, Cost: 100}
	armor := &data.ItemBase{Kind: data.KindArmor, Cost: 400, MaxAC: 100}
	for _, c := range []struct {
		name string
		b    *data.ItemBase
		it   *d2.Item
		st   d2.Stats
		want int
	}{
		{"unidentified magic: base", weap, &d2.Item{Quality: d2.QualityMagic}, nil, 100},
		{"identified magic: heuristic x2.5", weap, &d2.Item{Quality: d2.QualityMagic, Flags: d2.FlagIdentified}, nil, 250},
		{"armor at half defence", armor, &d2.Item{Quality: d2.QualityNormal}, d2.Stats{stDefence: 50}, 200},
		{"armor defence unknown: base", armor, &d2.Item{Quality: d2.QualityNormal}, nil, 400},
	} {
		if got := estimatePrice(c.b, c.it, c.st); got != c.want {
			t.Errorf("%s: got %d want %d", c.name, got, c.want)
		}
	}
}
