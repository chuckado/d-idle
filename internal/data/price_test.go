package data

import "testing"

// Expected values are what price_model.py (equal to game.exe on 100,000 cases) gives for the same inputs.
func TestSellPrice(t *testing.T) {
	armor := &ItemBase{Kind: KindArmor, Cost: 100, MaxAC: 50, Durability: 20}
	weap := &ItemBase{Kind: KindWeapon, Cost: 1000, Durability: 20}
	for _, c := range []struct {
		name   string
		b      *ItemBase
		f      PriceFacts
		maxBuy int
		want   int
	}{
		{"plain weapon: cost/2", weap, PriceFacts{Dur: 20, MaxDur: 20}, 0, 500},
		{"armor scaled by defence/maxac", armor, PriceFacts{Defence: 40, Dur: 20, MaxDur: 20}, 0, 40},
		{"armor defence unknown: base", armor, PriceFacts{Dur: 20, MaxDur: 20}, 0, 50},
		{"ethereal quarter", weap, PriceFacts{Ethereal: true, Dur: 20, MaxDur: 20}, 0, 125},
		{"class item quarter", weap, PriceFacts{ClassItem: true, Dur: 20, MaxDur: 20}, 0, 125},
		{"ethereal class item", weap, PriceFacts{Ethereal: true, ClassItem: true, Dur: 20, MaxDur: 20}, 0, 31},
		{"ethereal at 0 durability", weap, PriceFacts{Ethereal: true, Dur: 0, MaxDur: 20}, 0, 1},
		{"low quality unidentified: no halving", &ItemBase{Cost: 100}, PriceFacts{Low: true}, 0, 50},
		{"low quality identified", &ItemBase{Cost: 100}, PriceFacts{Low: true, Identified: true}, 0, 25},
		{"max buy cap act 1 normal", &ItemBase{Cost: 100000}, PriceFacts{}, MaxBuy(1, 0), 5000},
		{"max buy hell", &ItemBase{Cost: 200000}, PriceFacts{}, MaxBuy(1, 2), 35000},
		{"minimum 1", &ItemBase{Cost: 1}, PriceFacts{}, 0, 1},
		{"stack per unit times quantity", &ItemBase{Cost: 100, Stackable: true}, PriceFacts{Qty: 3}, 0, 150},
	} {
		if got := c.b.SellPrice(c.f, c.maxBuy); got != c.want {
			t.Errorf("%s: got %d want %d", c.name, got, c.want)
		}
	}
}

func TestRepairPrice(t *testing.T) {
	staff := &ItemBase{Kind: KindWeapon, Cost: 168, Durability: 20}
	armor := &ItemBase{Kind: KindArmor, Cost: 400, MaxAC: 100, Durability: 20}
	for _, c := range []struct {
		name string
		b    *ItemBase
		f    PriceFacts
		want int
	}{
		{"D3.r live: Short Staff at 1/20 paid 19", staff, PriceFacts{Identified: true, Dur: 1, MaxDur: 20}, 19},
		{"armor scaled by defence/maxac", armor, PriceFacts{Identified: true, Defence: 50, Dur: 10, MaxDur: 20}, 12},
		{"unidentified cannot be repaired", staff, PriceFacts{Dur: 1, MaxDur: 20}, 0},
		{"ethereal cannot be repaired", staff, PriceFacts{Identified: true, Ethereal: true, Dur: 1, MaxDur: 20}, 0},
		{"whole item prices 1", staff, PriceFacts{Identified: true, Dur: 20, MaxDur: 20}, 1},
		{"no durability", &ItemBase{Cost: 100, NoDurability: true}, PriceFacts{Identified: true, Dur: 1, MaxDur: 20}, 0},
	} {
		if got := c.b.RepairPrice(c.f); got != c.want {
			t.Errorf("%s: got %d want %d", c.name, got, c.want)
		}
	}
}

func TestBuyPrice(t *testing.T) {
	b := &ItemBase{Cost: 100}
	for _, c := range []struct {
		name   string
		npc    int
		questA bool
		qty    int
		want   int
	}{
		{"Akara", 148, false, 1, 100},
		{"Charsi 0.9375", 154, false, 1, 93},
		{"Gheed 1.0625", 147, false, 1, 106},
		{"Act 5 vendor x2", 513, false, 1, 200},
		{"quest A discount", 148, true, 1, 90},
		{"quantity", 148, false, 5, 500},
	} {
		if got := b.BuyPrice(PriceFacts{Qty: c.qty}, VendorBuyMult(c.npc), c.questA); got != c.want {
			t.Errorf("%s: got %d want %d", c.name, got, c.want)
		}
	}
}

func TestMaxBuyAndAct(t *testing.T) {
	for act, want := range map[int]int{1: 5000, 2: 10000, 3: 15000, 4: 20000, 5: 25000} {
		if got := MaxBuy(act, 0); got != want {
			t.Errorf("act %d: %d", act, got)
		}
	}
	if MaxBuy(3, 1) != 30000 || ActOfArea(104) != 4 || ActOfArea(39) != 1 {
		t.Error("nightmare or act table")
	}
}
