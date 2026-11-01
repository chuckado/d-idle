## Why

The DRILLS16 live proof of fixmouth failed: the arrival step-away works, but later walks still bounce (roam walk 13 from the exit after the arrival flag was consumed; Fighter.move engage approach 9.4 and 8.9 from the exit; evade spin). Details: docs/fixmouth2.md.

## What Changes

- `internal/nav`: the arrival radius (16) holds for the whole stay until a walk starts clear of every unwanted exit; `NearMouth` radius 12 (was 9).
- `internal/combat`: `evade` started inside a mouth is capped (3 in 10 s), then evading is off for the pin hold.
- Tests; docs/fixmouth2.md with the live proof recipe.

## Impact

Movement and combat guard only; no change to which areas are visited or when the bot leaves (policy cards stay paused). Needs a live DW.m3 x3 and DW.m4 rerun.

Overlap (declared): internal/nav/walk.go and internal/nav/fixmouth_test.go are owned by the open card fixmouth (awaiting its live proof); this card edits them (arrival hold, mouthStep clear-start rule, new tests) and does not list them in owns.
