## Why

docs/research/monster-missile-trace.md (B1-B4) and monster-ai-trace.md (B1-B3) found that the bot's missile dodge cannot beat arrows, has no per-missile threat model, ranks no unblockable shooter first, retreats from runners and drops a standing Fallen as unreachable.

## What Changes

- internal/combat/threat.go: missile threat table keyed by missiles.txt id (speed, damage, to-hit, blockable proxy), a dodge / ignore / kill-shooter verdict, ranking of incoming missiles, a monstats-derived shooter peak, runner and stander predicates (AI column).
- Narrow hooks: HostileCtx fields and marking in target.go, dodge filter and runner guard in fight.go, table load in SetTables (dodge.go), hostileCtx (fighter.go), hostileMissiles reuse (position.go).
- Five knobs in internal/env/tuning_combat.go (+ generated docs/knobs.md): runnerGuard defaults ON (V run in the AI trace), threatDodge, shooterKillMin, standOK default OFF (U / S).
- docs/botthreat.md: design, tags, live drill checklist.

## Impact

No spec delta. Behaviour change with defaults: no walk-away retreat while a Zombie, Brute, Blood Hawk, Corrupt Rogue, Lancer or Arach is close.

## Overlap (declared)

Edits narrow hooks in internal/combat (fight.go, target.go, dodge.go, fighter.go, position.go) and internal/env/tuning_combat.go, owned by other cards; no existing line changes behaviour with the new knobs at their defaults except runnerGuard.
