# fixmouth2: hold the arrival guard for the stay, guard approaches at 12, cap evading in a mouth

> status: IMPLEMENTED OFFLINE, live proof pending. Evidence: DRILLS16 section of docs/probes/drills/README.md (branch agent/drills16, logs 2026-11-01-drills16-m31/m41). Follows docs/fixmouth.md. V = read from those logs, U = unverified.

## Decision (numbers, V from the DRILLS16 logs)

- Exit (35,135), arrival (40,145) at 11.2. fixmouth's first-walk radius 16 worked: the step-away to (49,163) fired nothing in 4 of 4 entries.
- Bounce 1 (DW.m3, Fighter.move): an engage approach to a Zombie at (43,141) is 9.4 from the exit, the next walk starts at 8.9. `mouthR` 9 let the target through; `move` then walked back and fired the trigger. Now `NearMouth` uses `mouthApproachR` 12 (2.6 margin over 9.4; also covers 11.2). `move` and `reachable` already skip such points (the approach counter then ignores the target).
- Bounce 2 (DW.m4, roam): position (267,150) is 13 from the exit (255,145); `arrival` had been consumed by Clear's first walk, 13 > 9. Now `arrival` is not consumed by a walk: it is set on a level change and ends when `mouthStep` finds the walk starting beyond 16 of every unwanted exit. Ordinary levels keep `mouthR` 9. Chosen over "whole stay, never reset" because that would step away from exits on every walk of a long stay.
- Spin (25,565 `stepping away` lines at 3.2 from the exit): every walk starting inside the mouth steps away first, evade then re-targets. `Fighter.evade` started inside a guarded point is capped: after `mouthEvadeCap` 3 in `mouthEvadeWindow` 10 s it logs `evade: capped inside a cave mouth, fighting in place`, turns evading off for `pinHold` (like a pin) and spends a tick. All other decisions are unchanged; policy cards stay paused.
- Not tried: a bigger NearMouth (14+); it widens fight filtering. The real trigger bound stays U.

## Tests

`go test -p 1 -count=1 ./internal/nav ./internal/combat`: `TestArrivalHoldsForTheStay` (13 from the exit steps; flag held; 28 away clears it), `TestNearMouthApproachRadius`, `TestMoveSkipsDenZombieApproach` (43,141 and 42,140 not walked), `TestEvadeInsideMouthIsCapped`, `TestClearWithPackNearMouthFinishes` (sim); `TestRoamFirstWalkAfterArrivalStepsAway` now expects the flag held.

## Live proof (not run: no game in this task)

Same commands as DRILLS16 from the fixmouth2 worktree, on the worker host:

```
D2D_AGENT=fixmouth2 D2D_BIN=<fork>/bin/d2d.lab-cheats3 tools/agentrun.sh fixmouth2 --ttl 300 -- tools/d2dslot.sh run -n 0 -t 200 -- <built>/d2drill -group F -only DW.m3 -label fixmouth2-m3-N -v
```

N = 1, 2, 3, then `-only DW.m4 -label fixmouth2-m4`. PASS: Clear finishes before the TTL (DRILL line present), no `Entered` cycle within 1 s of an arrival, no `the level changed during the walk` storm, no thousands of `stepping away` lines. FAIL: attach the filtered logs; look at whether `evade: capped` fired and at the distances in `nav mouth exit`.

## Overlap (declared)

Card fixmouth is still open on internal/nav/walk.go and internal/nav/fixmouth_test.go (its constants, `walk`, `NearMouth`, `mouthStep`). This change edits the same functions (arrival consumption, `NearMouth` radius, end of `mouthStep`) and one fixmouth test assertion. It also touches internal/combat/fight.go (`evade`) and fighter.go (two fields). Merge fixmouth first; the conflict is textual.

## Shared-file lines for the merger

- docs/BACKLOG.md: FIXMOUTH2: arrival radius held for the stay, NearMouth 12, evade cap in a mouth (DRILLS16 bounces); offline tests pass, live DW.m3 x3 + DW.m4 pending (docs/fixmouth2.md).
- docs/INDEX.md: docs/fixmouth2.md.
- CHANGELOG: none until the live proof passes.
