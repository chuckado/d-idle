package resource

import "testing"

// FIXPRICE B4: potion prices follow the vendor multiplier (Charsi 960, Gheed 1088, Act 5 2048 of 1024).
func TestPricesAtVendorMult(t *testing.T) {
	for _, c := range []struct {
		name   string
		npc    int
		questA bool
		hp1    int
		hp5    int
	}{
		{"Akara", 148, false, 30, 600},
		{"Charsi", 154, false, 28, 562},
		{"Gheed", 147, false, 31, 637},
		{"Malah, Act 5", 513, false, 60, 1200},
		{"Akara after quest A", 148, true, 27, 540},
	} {
		p := PricesAt(c.npc, c.questA)
		if p.HP[0] != c.hp1 || p.HP[4] != c.hp5 {
			t.Errorf("%s: hp tier 1 %d (want %d), tier 5 %d (want %d)", c.name, p.HP[0], c.hp1, p.HP[4], c.hp5)
		}
	}
	if PricesAt(148, false) != DefaultPrices() {
		t.Error("the default is the multiplier 1.0 price")
	}
}
