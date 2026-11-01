package main

import (
	"d2lod-bot/internal/combat"
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
	nv.TeleportOK = func(level int) bool {
		a := db.AreaByID(level)
		// Clear-line levels (Duriel's Lair, rule 2) are left to walking: a hop picks its cell without checking the line to it.
		return a == nil || data.CanTeleportIn(a) && a.TeleportRule() == data.TeleportAllowed
	}
}
