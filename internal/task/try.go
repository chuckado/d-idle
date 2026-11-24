package task

import "sync/atomic"

// The owner's one-shot "try anyway" start (docs/questboard.md, ADR 0030 amended 2026-11-24). While it is in force the
// READINESS checks pass: the level and life floors of the boss gates, the Barracks and road floors, the tower floor, the
// Andariel unattended hold. The kit checks (potions, portal, mana, antidotes), the death-spiral guards (no weapon, no armour,
// banned areas) and every combat reaction stay as they are. The bot sets it while the board runs the try chunk; one process,
// one hero, so a process-wide switch is enough (like MercAlive).
var tryActive atomic.Bool

// SetTry turns the readiness override on or off.
func SetTry(on bool) { tryActive.Store(on) }

// TryActive reports the override.
func TryActive() bool { return tryActive.Load() }
