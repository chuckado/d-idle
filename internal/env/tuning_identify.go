package env

// IdentifyTuning holds the bulk identify knobs (docs/design/bulk-identify-2026-11-18.md section 3).
type IdentifyTuning struct {
	Policy         int // IdentifyManual, IdentifyCainFree (default), IdentifyScrolls, IdentifyNever
	MinValue       int // policy 2: skip items whose estimated sell value is below this (0: no bar)
	MaxScrolls     int // identify scrolls or tome charges one town visit may spend
	GoldFloor      int // gold that must stay in the purse after an identify purchase
	TripCap        int // items one Cain visit identifies at most
	Chunk          int // items moved from the stash to the pack per batch
	AutoSellRest   int // 1: the sell chore may sell identified items the rules rate sell (0: keep them for the player)
	EtherPerReveal int // Ether one reveal costs in the vault when no Cain or scroll is used (0: Ether is not a payer)
}

// Identify policy codes (knob identify.policy).
const (
	IdentifyManual   = 0
	IdentifyCainFree = 1
	IdentifyScrolls  = 2
	IdentifyNever    = 3
)

func DefaultIdentifyTuning() IdentifyTuning {
	return IdentifyTuning{Policy: IdentifyCainFree, MaxScrolls: 10, GoldFloor: 600, TripCap: 40, Chunk: 10}
}

func identifyKnobs() []Knob {
	d := DefaultIdentifyTuning()
	k := func(key, label, unit, desc string, def, min, max int) Knob {
		return Knob{Key: "identify." + key, Label: label, Unit: unit, Desc: desc, Default: float64(def), Min: float64(min), Max: float64(max), NoHandicap: true}
	}
	return []Knob{
		k("policy", "Bulk identify policy", "code", "0 manual (only the player's commands), 1 cain-free (default: Deckard Cain identifies on a town visit that happens anyway, free; nothing is bought; manual while Cain is not rescued), 2 scrolls-above-value (Cain when open, else scrolls or the tome for items worth identify.minValue), 3 never. Forced to manual in measured runs.", d.Policy, 0, 3),
		k("minValue", "Scroll bar", "gold", "Policy 2: items whose estimated sell value is below this are not identified with a scroll. 0 = no bar.", d.MinValue, 0, 99999),
		k("maxScrolls", "Scrolls per visit", "count", "Identify scrolls or tome charges one town visit may spend.", d.MaxScrolls, 0, 100),
		k("goldFloor", "Gold kept after buying", "gold", "Gold that must stay in the purse after an identify purchase; a purchase that would break it is skipped.", d.GoldFloor, 0, 99999),
		k("tripCap", "Items per Cain visit", "count", "Most items one Cain visit identifies; the rest follows at the next visit.", d.TripCap, 1, 200),
		k("chunk", "Stash batch size", "count", "Items taken out of the stash per batch before Cain; each batch is verified before the next.", d.Chunk, 1, 40),
		k("autoSellRest", "Sell the rest", "bool", "1: the sell chore may sell identified items the loot rules rate sell. 0 (default): identified items are kept for the player to see first.", d.AutoSellRest, 0, 1),
		k("etherPerReveal", "Ether per reveal", "ether", "Ether one vault reveal costs when neither Cain nor a scroll pays; 0 = Ether is not used for reveals (U: price not decided).", d.EtherPerReveal, 0, 99999),
	}
}
