## Why

SKARMOR (docs/research/sk-armor-trace.md, findings B1-B3) found that the armor recast of card fixrecast never runs inside one long clear and is blind on d2d: `Fighter.Buff` runs only on need-buff edges, so Frozen Armor lapses after 120 s (132 s at level 2) in a long fight (B1); `buffStates` knew only Frozen and Bone Armor (B2); the d2d backend never filled `Player.States`, so the state branch of fixrecast could not fire there (B3).

## What Changes

- internal/combat: `fightBuff` in the fight loop (after survival, hard rules and the lost-fight check, before the target logic) runs the same `buffDue` check on the snapshot the loop already has; skipped with `crowdAdjacentMax` or more attackers in reach. `Buff` is split into read + aura + `buffPass`.
- internal/d2: state ids 20, 30, 88 and `buffStates` entries for Chilling Armor, Energy Shield, Shiver Armor.
- internal/d2dctl `Client.Buffs` (bare read-only `debug states`) and internal/d2dsrc `Options.BuffStates` (on in d2denv.Build) fill `Player.States` for the tracked armors.
- docs/fixbuff.md with the live proof recipe.

## Impact

Offline only. No spec names buff timing (grep openspec/specs): no delta. Live proof pending.

Overlap (declared, as fixcrowd does): internal/combat/necro.go is owned by the open card fixrecast and internal/combat/fighter.go by the open card fixportalmelee (both code-merged, open for their live proofs). This card edits them (buffStates; Buff split and fightBuff) and does not list them in owns.
