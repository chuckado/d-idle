## Why

SKARMOR (docs/research/sk-armor-trace.md, findings B1-B3) found that the armor recast of card fixrecast never runs inside one long clear and is blind on d2d: `Fighter.Buff` runs only on need-buff edges, so Frozen Armor lapses after 120 s (132 s at level 2) in a long fight (B1); `buffStates` knew only Frozen and Bone Armor (B2); the d2d backend never filled `Player.States`, so the state branch of fixrecast could not fire there (B3).

## What Changes

- internal/combat: `fightBuff` in the fight loop (after survival, hard rules and the lost-fight check, before the target logic) runs the same `buffDue` check on the snapshot the loop already has; skipped with `crowdAdjacentMax` or more attackers in reach. `Buff` is split into read + aura + `buffPass`.
- internal/d2: state ids 20, 30, 88 and `buffStates` entries for Chilling Armor, Energy Shield, Shiver Armor.
- internal/d2dctl `Client.Buffs` (bare read-only `debug states`) and internal/d2dsrc `Options.BuffStates` (on in d2denv.Build) fill `Player.States` for the tracked armors.
- internal/data: a cache lacking `skills.auralencalc` is re-extracted by `CacheCurrent`/`Open`, refused by `Load` with the column named (DRILLS13 F1: a stale lab cache silently disabled fixrecast).
- docs/fixbuff.md with the live proof recipe.

## Impact

Offline only. No spec names buff timing (grep openspec/specs): no delta. Live proof pending.

Overlap (declared, as fixcrowd does): cmd/d2bot/d2dbackend_test.go (synthetic cache header, one line) is not owned by an open card; internal/data/data_test.go (fixture column and one test) is owned by the archived card fixrecast; internal/combat/necro.go is owned by the card fixrecast (archived) and internal/combat/fighter.go by the card fixportalmelee (archived). This card edits them (buffStates; Buff split and fightBuff) and does not list them in owns.
