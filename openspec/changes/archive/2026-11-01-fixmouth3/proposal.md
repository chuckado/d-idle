## Why

The DRILLS17 live proof of fixmouth2 failed: Clear finishes and the evade spin is gone, but 3 of 4 (and 4 of 4) DW.m3 entries and 2 of 4 DW.m4 entries still bounce. The one-shot arrival flag ends after the first step-away, so the engage approach, the roam walks and a step-away from 3.2 are unguarded. Details and root-cause table: docs/fixmouth3.md.

## What Changes

- `internal/nav`: a stay guard (level property) set when a walk has to step away from an unwanted exit and held until the level changes; trigger radius 15; the step-away is one long click that leaves the exit, with an away-path check; roam targets filtered at 17.
- Tests; docs/fixmouth3.md with the live proof recipe.

## Impact

Movement guard only; no change to which areas are visited (policy cards stay paused). Needs a live DW.m3 x3 and DW.m4 rerun.

Overlap (declared): internal/nav/walk.go and internal/nav/nav.go are owned by the open card fixmouth (awaiting its live proof) and fixmouth2 edited the same functions; this card edits them (constants, NearMouth, mouthStep, the walk follow step, one field) and does not list them in owns.
