package env

// RespecTuning is the price model of a respec (docs/autoalloc.md section 3). The numbers are PLACEHOLDERS, chosen only
// so the preview on the Points page and the tuning screen looks sane; the real numbers are deliberately not decided yet
// (user, 2026-11-21: only the knobs matter now).
//
//	price = clamp(MinGold, MaxGold, BaseGold + PerLevel x clvl) x (1 + PerRespecPct/100 x paid respecs of this hero)
//
// The first FirstFree respecs of a hero cost nothing (the cheap first one); the escalation counts paid respecs only.
type RespecTuning struct {
	BaseGold     int
	PerLevel     int
	PerRespecPct int
	FirstFree    int
	MinGold      int
	MaxGold      int
}

func DefaultRespecTuning() RespecTuning {
	return RespecTuning{BaseGold: 500, PerLevel: 250, PerRespecPct: 50, FirstFree: 0, MinGold: 1000, MaxGold: 100000}
}

func respecKnobs() []Knob {
	const tag = " PLACEHOLDER number: the model matters, the value is not decided (docs/autoalloc.md)."
	k := func(key, label, unit, desc string, def, max float64) Knob {
		return Knob{Key: "respec." + key, Label: label, Unit: unit, Desc: desc + tag, Default: def, Min: 0, Max: max, NoHandicap: true}
	}
	return []Knob{
		k("baseGold", "Respec base price", "gold", "Fixed part of the price of a respec (all skill and stat points refunded). price = clamp(minGold, maxGold, baseGold + perLevel x level) x (1 + perRespecPct/100 x paid respecs).", 500, 1000000),
		k("perLevel", "Respec price per level", "gold", "Gold added per character level before the escalation.", 250, 100000),
		k("perRespecPct", "Respec escalation", "%", "Percent added to the price for every respec this hero already paid for.", 50, 1000),
		k("firstFree", "Free respecs", "count", "How many of a hero's first respecs cost nothing (they do not count for the escalation). 0 = every respec is paid.", 0, 10),
		k("minGold", "Respec price floor", "gold", "The price before the escalation is at least this.", 1000, 1000000),
		k("maxGold", "Respec price cap", "gold", "The price before the escalation is at most this (0 = no cap).", 100000, 10000000),
	}
}
