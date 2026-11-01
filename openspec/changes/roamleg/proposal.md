## Why

DRILLS21: after the mouth gate holds a walk, Navigator.roam folded ErrMouthGate into the failed-target case: 10 failed legs and 10 step-aways per episode, 50 `roam leg failed` lines per DW.m3 run, and the failed-target memory forgot good targets. The hold is about where the character stands, not about the target (docs/roamleg.md).

## What Changes

- internal/nav/roam.go: a leg that ends with ErrMouthGate does not mark its target failed; roam ends the episode with ErrRoamHeld (wraps ErrExplored, so callers are unchanged), one debug line and a counter.
- ErrNoPath, ErrStuck, ErrTimeout, exits, sampled points and gate semantics are unchanged.

## Impact

Offline only; live proof pending (recipe in docs/roamleg.md). No delta spec.

Overlap (declared): internal/nav/roam.go and mouthgate.go are owned by archived cards; this card edits roam.go only (one switch case, one error value, one counter) and does not list it in owns.
