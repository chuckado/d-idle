package main

import (
	"math"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/nav"
)

// wireTeleport gives the navigator Teleport as a capability of the character: the key is whatever the Fighter has
// verified for the skill named Teleport (empty for a class or build that has not learned it, so nothing hops), and a
// level the game refuses Teleport on (towns, rule 0) or allows only along a clear line (rule 2) is never cast on. Without tables every level is allowed.
func wireTeleport(nv *nav.Navigator, cb *combat.Fighter, db *data.DB) {
	nv.TeleportKeyFn = cb.TeleportKey
	if db == nil {
		return
	}
	// The hop cost follows the skill data at the learned level (24 at level 1, one less per level, V read SKTELE).
	nv.TeleportCost = func(p *d2.Player) int { return teleCost(db, p) }
	nv.TeleportOK = func(level int) bool {
		a := db.AreaByID(level)
		// Clear-line levels (Duriel's Lair, rule 2) are left to walking: a hop picks its cell without checking the line to it.
		return a == nil || data.CanTeleportIn(a) && a.TeleportRule() == data.TeleportAllowed
	}
}

// teleCost is the mana of one Teleport cast at the player's effective skill level, rounded up; 0 when unknown.
func teleCost(db *data.DB, p *d2.Player) int {
	sk := db.SkillByName("Teleport")
	if sk == nil {
		return 0
	}
	for _, s := range p.Skills {
		if s.ID == sk.ID && s.Effective() > 0 {
			return int(math.Ceil(sk.ManaCost(s.Effective())))
		}
	}
	return 0
}
