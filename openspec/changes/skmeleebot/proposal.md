## Why

The SKMELEE trace (docs/research/sk-melee-trace.md, bot findings B1 to B4) found that the resource simulator's melee always lands for a flat 3, `WeakPack` has no attack-rating term and the melee rung ignores the Blunt-vs-undead bonus. The policy's "melee is enough" choices were tuned on those.

## What Changes

- `internal/resource/melee.go`: the plain Attack's to-hit and damage as pure functions (`HitChance`, `Blow`, `Foe`), fed by data the caller supplies; Blunt through the ItemTypes chain (`IsBlunt` takes the chain walk).
- `internal/resource/sim.go`: the scenario's melee rolls to-hit and the damage range when a `Blow` is set (B1); the stock scenario sets one.
- `internal/resource/types.go`, `attack.go`: `Input.Blow` and `Input.Foe` (nil keeps today's behaviour); `weak()` adds the to-hit term (B2); `meleePack` raises the melee rung's pack limit by the Blunt gain against undead (B3).
- `docs/skmeleebot.md`: what changed, before and after numbers, what stayed approximate; B4 note.

## Impact

Offline Go and docs. No spec, ADR or glossary change. Production wiring of `Input.Blow` / `Input.Foe` from the live snapshot is not part of this card (see docs/skmeleebot.md).
