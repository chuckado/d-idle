package questboard

import (
	"d2lod-bot/internal/power"
)

// HeroFrom converts the power model's hero (read-only) into the facts readiness needs.
func HeroFrom(h power.Hero) Hero {
	low := h.Resist[0]
	for _, v := range h.Resist[1:] {
		low = min(low, v)
	}
	return Hero{Class: h.Class, Level: h.Level, MercHired: h.Merc.Hired, MercAlive: h.Merc.Alive,
		Potions: h.Potions, LowestResist: low, ResistKnown: h.Quality[power.TermResist] > power.Default}
}

// EncounterFrom turns a margin result into an Encounter. k is the survival ledger's correction (k_zone x
// k_hero, 1 when there is no ledger yet); the controller divides both margins by it the same way (V
// survival controller.go judge). An error result means no margin: pass Known false by not calling this.
func EncounterFrom(name, kind string, r power.Result, k float64) Encounter {
	if k <= 0 {
		k = 1
	}
	unread := false
	for _, f := range r.Flags {
		if f == "offence_default" || f == "offence_mana_default" || f == "offence_swing_default" {
			unread = true
		}
	}
	return Encounter{Name: name, Kind: kind, Best: r.Best / k, Low: r.Low / k, Known: true, OffenceUnread: unread}
}
