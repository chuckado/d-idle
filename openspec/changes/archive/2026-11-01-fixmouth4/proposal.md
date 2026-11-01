## Why

The DRILLS18 live proof of fixmouth3 failed narrowly: DW.m4 has no bounce left, DW.m3 still bounces once per run. Three cards guarded the cave mouth at plan START (step-away target, Fighter target, roam target), while the character is moved by many other things: replans, the stuck ladder, fights inside a walk, basic-attack swings. The DRILLS18 bounce was a step-away replanned from 3.2 subtiles after swings pulled the character there. Details: docs/fixmouth4.md.

## What Changes

- `internal/nav`: a click gate wraps the Navigator's controller; while a stay guard (or the arrival flag) holds in a non-town level, every world click must keep the invariant (target at least `mouthTriggerR` 15 from each unwanted exit; from outside the segment stays outside; from inside it only moves outward); refused clicks are clipped or rewritten along the away vector, bounded, then held with `ErrMouthGate` (an `ErrNoPath`). The step-away re-checks its first click on every replan and retargets (bounded).
- `internal/combat`: a left-hand swing (a movement order) asks the mover's gate first; held swings cost a poll.
- Tests (geometry table from the logs, property test, DRILLS18 chain sims, wanted exits both ways, swing tests); docs/fixmouth4.md with the live recipe.

## Impact

Movement guard only; no change to which areas are visited or when the bot leaves (policy cards stay paused). Needs a live DW.m3 x3 and DW.m4 rerun.

Overlap (declared): internal/nav/walk.go and internal/nav/nav.go are owned by the open card fixmouth, internal/combat/fight.go by the open card fixmouth2, and fixmouth3 edited the same functions. This card edits them (nav.go: New and two fields; walk.go: constants, walkOpts, the step-away call, the replan branch; fight.go: two checks in engage and swingIntoMouth) and does not list them in owns. Merge fixmouth, fixmouth2 and fixmouth3 first.
