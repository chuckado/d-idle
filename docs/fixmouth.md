# fixmouth: first walk after a warp arrival steps away from the mouth

> status: IMPLEMENTED OFFLINE, live proof pending. Evidence: DRILLS14 section of docs/probes/drills/README.md (branch agent/drills14) and docs/fixdenbounce.md. Tags: V = read from the DRILLS14 logs, U = unverified.

## Decision

Two options were proposed: (A) apply `mouthStep` to the first walk after a warp arrival out to about 16 subtiles, or (B) raise `mouthR` to 14. Chosen: A.

Numbers (V, DRILLS14 logs): arrival (40,145), listed exit (35,135), offset (+5,+10), distance 11.2; `mouthR` 9 (walk.go) so `mouthStep` skipped; `roamExitClear` 7 keeps the roam target (Chebyshev 11 from the exit); one click toward (36,146) fired the trigger 12 ms later, 4 of 4 invocations. Cave Level 1 arrives at the same offset (260,155) vs (255,145). Bounce positions 16.2 to 18 from the exit did not fire (m4a log).

Why not B: `mouthR` also feeds `NearMouth` (fight approach filter) and every walk start. At 14 each walk that begins within 14 subtiles of an exit would be pushed 20 subtiles away, a behaviour change far outside the arrival case. A changes only the walk that starts right after a level change (`Navigator.arrival`), and `roamExitClear` (a roam POLICY value) stays at 7.

Constants: `mouthArrivalR` 16 (walk.go): V that 11.2 fires and 16.2 did not; the real trigger bound is U.

## Behaviour

- `walk` records the level of its first observation (`walkLevel`); a start in a different level than the previous walk sets `arrival`. A walk that ended by a level change keeps the old level in `walkLevel`.
- `mouthStep` uses radius `mouthArrivalR` while `arrival` is set; the first walk consults it once and clears `arrival`.
- `NearMouth` also answers while `arrival` is set, so `Fighter.move` skips approach targets within `mouthR` of an unwanted exit right after arrival (Cave Level 1 Clear approach) even with `mouthGuard` off. Targets elsewhere are walked to.
- U: the Cave Level 1 step-away direction fired the trigger on its first click in DRILLS14 (Result 2). This change does not alter the direction; whether skipping the approach plus the wider step avoids it is unproven.

## Tests

`go test -p 1 -count=1 ./internal/nav ./internal/combat`: `TestMouthStepArrivalRadius` (the exit (2048,2050), character (2053,2060) must step; 20 away must not; not-arrival unchanged), `TestRoamFirstWalkAfterArrivalStepsAway`, `TestNearMouthAnswersAfterArrival`, `TestMoveSkipsTargetsInsideAMouthOnly`.

## Live proof (not run: no game in this task)

Per the DRILLS14 recipe, from the fixmouth worktree, build `cmd/d2drill` and run on the worker host:

```
D2D_AGENT=fixmouth D2D_BIN=<fork>/bin/d2d.lab-cheats3 tools/agentrun.sh fixmouth --ttl 900 -- tools/d2dslot.sh run -n 0 -t 600 -- <built>/d2drill -group F -only DW.m3 -label fixmouth-m3-N -v
```

for N = 1, 2, 3, then the same with `-only DW.m4 -label fixmouth-m4`. PASS: in every entry the character stays in the cave (area 8 / 9) through idle, Clear and Explore phases; no `Entered` / `arrived in level` cycle within 1 s of the arrival; no `the level changed during the walk` error; `nav: stepping away from a cave mouth` appears before the first roam click in DW.m3. FAIL otherwise: attach the logs and revisit mouthArrivalR / the step direction.

## Overlap (declared)

internal/nav/walk.go and internal/nav/roam_test.go: the drills14 card owns rows for walk.go (debug lines in `mouthStep`, `TestMouthStepDebugLines`) and roam_test.go. This change touches walk.go (the constants block, `walk`, `NearMouth`, `mouthStep` radius) and does not edit roam_test.go (new tests live in fixmouth_test.go). The merger must reconcile walk.go if drills14 merges first or after; expected conflict is textual only, near `mouthStep`'s distance check.

## Shared-file lines for the merger

- docs/BACKLOG.md: FIXMOUTH: mouthStep arrival radius 16 for the first walk after a warp (Den bounce, DRILLS14); offline tests pass, live DW.m3 x3 + DW.m4 proof pending (docs/fixmouth.md).
- docs/INDEX.md: docs/fixmouth.md.
- CHANGELOG: none until the live proof passes.
