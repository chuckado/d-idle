package bot

import "d2lod-bot/internal/d2"

// HeroSheet is what the character screen shows beside the resistances: the four attributes and, when the stat list
// carries them, defence and the damage range. A value the snapshot does not hold is reported absent (Has* false), never
// as a zero.
type HeroSheet struct {
	Str, Dex, Vit, Energy int
	HasAttrs              bool
	Defense               int
	HasDefense            bool
	DmgMin, DmgMax        int
	HasDamage             bool
}

// Sheet reads the hero's sheet from the newest snapshot; ok is false until the first in-game snapshot.
// Attributes: V live as the player's own fields (the stat vector ids 0 to 3, internal/d2/stats.go); whether they include
// gear is U per backend. Defence: stat 31, U live. Damage: stats 21 and 22 (S: combat/melee.go), a stat of items and
// skills, not the weapon's base damage; present only when the vector holds a maximum above zero.
func (r *Runner) Sheet() (HeroSheet, bool) {
	s := r.latest()
	if s == nil {
		return HeroSheet{}, false
	}
	return sheetOf(s.Player), true
}

func sheetOf(p *d2.Player) HeroSheet {
	h := HeroSheet{Str: p.Strength, Dex: p.Dexterity, Vit: p.Vitality, Energy: p.Energy}
	h.HasAttrs = p.Strength > 0 || p.Dexterity > 0 || p.Vitality > 0 || p.Energy > 0
	if p.Stats.Has(d2.StatArmorClass) {
		h.Defense, h.HasDefense = int(p.Stats.Get(d2.StatArmorClass)), true
	}
	if mx := int(p.Stats.Get(22)); mx > 0 && p.Stats.Has(22) {
		h.DmgMin, h.DmgMax, h.HasDamage = int(p.Stats.Get(21)), mx, true
	}
	return h
}
