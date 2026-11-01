## Why

no observable behaviour change: harness and docs only (live drill, no bot code).

Card armorcut (retry of a Frozen Armor cast cut by a hit) needs its live proof. DRILLS28 runs step 0 (a refresh with no monsters, tells d2d refusing to refresh a set state from a cut cast) and the D26-A rerun on both lab binaries with the armorcut code.

## What Changes

- d2drill: group V, drill D28-0 (internal/drill/d28.go, cmd/d2drill/d26_lab.go); D26-A logs the player mode, cut lines and sample gaps.
- Results in docs/probes/drills/README.md (DRILLS28) and `2026-11-01-drills28-*.txt`.

## Impact

Harness and docs only. Overlap (declared): cmd/d2drill/main.go (group switch), cmd/d2drill/d26_lab.go and internal/drill/d26.go (from drills26).
