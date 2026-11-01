## Why

DRILLS14 named the Den of Evil bounce: the first walk after a warp arrival fires the return trigger because `mouthStep` (mouthR 9) skips an arrival 11.2 subtiles from the listed exit. Details: docs/fixmouth.md.

## What Changes

- `internal/nav`: the first walk after a warp arrival uses `mouthArrivalR` (16) in `mouthStep`; `NearMouth` answers after an arrival until that walk has consulted `mouthStep`.
- Tests in `internal/nav` and `internal/combat`; docs/fixmouth.md with the live proof recipe.

## Impact

Movement primitive only; no roam or leveling policy change. Needs a live DW.m3 x3 and DW.m4 rerun before the fix is claimed.
