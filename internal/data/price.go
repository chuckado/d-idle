package data

// Vendor prices as game.exe 1.14d computes them (FUN_0062efb0). VERIFIED: tools/emu/price_model.py equals game.exe
// on 100,000 emulator cases (docs/research/price-trace.md). Not modelled, so the results are lower bounds for items
// that carry them (tag I in the trace): the stat cost of identified magic, rare, crafted, set and unique items,
// staff skill mods, missing charges, sockets, and the stack terms of throwing weapons.

// PriceFacts is what the price depends on besides the base row.
type PriceFacts struct {
	Identified, Ethereal bool
	ClassItem            bool // item type with a class restriction (sells for another quarter)
	Quiver               bool // ammo: priced by quantity / 1024
	Low                  bool // low quality (identified: its price loses half)
	Defence              int  // the item's own defence stat (31); 0 = unknown, the base maximum is taken
	Dur, MaxDur          int
	Qty                  int // stack size, 0 = 1
}

func cdiv(a, b int) int { return a / b } // Go truncates toward zero like game.exe's idiv

// scale is base x m / 1024 as game.exe does it: above 0xffff the base is divided first.
func scale(base, m int) int {
	if base > 0xffff && m != 0 {
		return cdiv(base, 1024) * m
	}
	return cdiv(m*base, 1024)
}

// base is the cost after the armor and identified-quality terms (same for buy, sell and repair).
func (b *ItemBase) priceBase(f PriceFacts) int {
	p := b.Cost
	if b.Kind == KindArmor && b.MaxAC > 0 && f.Defence > 0 {
		p = cdiv(b.Cost*f.Defence, b.MaxAC)
	}
	if f.Identified && f.Low {
		p -= cdiv(p, 2)
	}
	return p
}

func (b *ItemBase) qty(f PriceFacts) int { return max(f.Qty, 1) }

// SellPrice is what a vendor pays: cost / 2, an ethereal item and a class item a quarter of that each, capped at
// the NPC max buy (maxBuy <= 0: no cap), never below 1. Normal-quality items only count their base; see the file note.
func (b *ItemBase) SellPrice(f PriceFacts, maxBuy int) int {
	if b == nil {
		return 0
	}
	s := b.priceBase(f)
	if f.Quiver {
		s = cdiv(b.Cost*b.qty(f), 1024)
	}
	if f.Ethereal {
		s = cdiv(s, 4)
	}
	if f.ClassItem {
		s = cdiv(s, 4)
	}
	zero := f.Ethereal && !b.NoDurability && b.Durability > 0 && f.MaxDur != 0 && f.Dur <= 0
	if zero {
		s = 0
	} else {
		s = scale(s, 512) // npc.txt "buy mult" 512 at every vendor
	}
	if !f.Quiver {
		s *= b.qty(f)
	}
	if maxBuy > 0 && s > maxBuy {
		s = maxBuy
	}
	return max(s, 1)
}

// RepairPrice is what a smith charges for one item: cost x (max - dur) / max x 128 / 1024. 0 when the item cannot be
// repaired (unidentified, ethereal, no durability, quiver); a whole item prices 1 (Repair All skips it).
func (b *ItemBase) RepairPrice(f PriceFacts) int {
	if b == nil || !f.Identified || f.Ethereal || f.Quiver || b.NoDurability || b.Durability <= 0 || f.MaxDur <= 0 {
		return 0
	}
	r := b.priceBase(f)
	if f.Dur < f.MaxDur {
		r = cdiv((f.MaxDur-max(f.Dur, 0))*r, f.MaxDur)
	} else {
		r = 0
	}
	return max(scale(r, 128), 1)
}

// BuyPrice is what a vendor asks: cost x its sell mult / 1024 x quantity (not for quivers); questA applies the
// discount of the Act quest (922 / 1024). vendorMult is VendorBuyMult(npc class).
func (b *ItemBase) BuyPrice(f PriceFacts, vendorMult int, questA bool) int {
	if b == nil {
		return 0
	}
	p := b.priceBase(f)
	if f.Quiver {
		p = cdiv(b.Cost*b.qty(f), 1024)
	}
	p = ScalePrice(p, vendorMult, questA)
	if !f.Quiver {
		p *= b.qty(f)
	}
	return max(p, 1)
}

// ScalePrice applies a vendor multiplier (out of 1024) and, with questA, the act quest discount (922 / 1024) to a price.
func ScalePrice(p, vendorMult int, questA bool) int {
	p = scale(p, vendorMult)
	if questA {
		p = scale(p, 922)
	}
	return max(p, 1)
}

// VendorBuyMult is the npc.txt "sell mult" (out of 1024) of an NPC class: what the vendor charges relative to cost.
// Charsi 960, Gheed 1088, the Act 5 vendors (Larzuk, Malah, Drehya, Nihlathak) 2048, every other vendor 1024.
func VendorBuyMult(npcClass int) int {
	switch npcClass {
	case 154:
		return 960
	case 147:
		return 1088
	case 511, 512, 513, 514:
		return 2048
	}
	return 1024
}

// ActOfArea is the act (1..5) of a level id, by the town ids 1, 40, 75, 103, 109.
func ActOfArea(area int) int {
	switch {
	case area >= 109:
		return 5
	case area >= 103:
		return 4
	case area >= 75:
		return 3
	case area >= 40:
		return 2
	}
	return 1
}

// MaxBuy is the most a vendor pays for one item (npc.txt "max buy", difficulty 0 Normal, 1 Nightmare, 2 Hell).
func MaxBuy(act, difficulty int) int {
	switch difficulty {
	case 1:
		return 30000
	case 2:
		return 35000
	}
	return [...]int{5000, 5000, 10000, 15000, 20000, 25000}[min(max(act, 0), 5)]
}

// PriceFactsOf derives the type terms (class item, quiver) from the item types table.
func (db *DB) PriceFactsOf(b *ItemBase) (f PriceFacts) {
	if db == nil || b == nil {
		return f
	}
	if t := db.ItemType(b.Type); t != nil {
		f.ClassItem = t.Class != ""
		f.Quiver = t.Quiver != ""
	}
	return f
}
