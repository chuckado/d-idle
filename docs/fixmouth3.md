# fixmouth3: a stay guard for unwanted exits, trigger radius 15, a step-away that leaves

> status: IMPLEMENTED OFFLINE, live proof pending. Evidence: DRILLS17 section of docs/probes/drills/README.md (branch agent/drills17, logs 2026-11-01-drills17-m31/m41). Follows docs/fixmouth.md and docs/fixmouth2.md. V = read from those logs, S = simulated here, U = unverified.

## Root cause (V, DRILLS17 logs)

| Bounce site | Numbers | Guard that should have caught it | Why it did not |
|---|---|---|---|
| DW.m3 engage approach (`GoTo < move < engage < fight`) to the Zombie (43,141); the click that fired is (42,146) | exit (35,135): target 10.0, fired click 13.0 | `NearMouth` (radius 12, and the target 10.0 is inside) | it answers only when `mouthGuard`, `avoiding()` or `arrival`; the first is off by default, and `arrival` ended at the first step-away (the next walk started 28.7 from the exit, beyond 16) |
| DW.m4 explore, `roam < explore` at (267,150), first click (265,151) | exit (255,145): position 13, fired click 11.7; `arrival=false r=9` | `mouthStep` (radius 9 off the flag) | 13 > 9, flag already off; the roam target filter (Chebyshev 7) kept targets near the mouth, and `penalizeExits` ignores an exit whose walk goal is within 16 (`exitAvoidR+2`), so a target at 8 to 16 got no detour either (S: the new sweep test clicked 12.1 from the exit with the old filter) |
| DW.m4 step-away from (256,148) | 3.2 from the exit, `arrival=true r=16`; the SECOND waypoint click (260,156) fired | the step itself | it walked with the normal waypoint clicks: every waypoint inside the zone is another click that can fire; fired clicks reach 11.2 to 13.0 from the listed exit position |

Arrival flag: V, it is a one-step guard (see DRILLS17). New V fact used for the radius: clicks fired 13.0 (Den) and 11.7 and 11.2 (Cave Level 1) from the exit; clicks 16.2 to 18 away never did (DRILLS16). The real bound is U; the arrival offset 11.2 itself does not fire without a click.

## Design

1. Stay guard (level property). `Navigator.stayLevel`: set in `mouthStep` whenever a walk starts inside the guard radius of an exit it does not want (`Dest != wantExit`, the same predicate `NearMouth`, `mouthStep` and `penalizeExits` use) in a non-town level; held until the level differs (`NearMouth` compares with the current level, `mouthStep` with the observed one). It does not depend on the one-shot `arrival` flag, which stays as it was. While it holds, `NearMouth` answers for every point within `mouthApproachR` and `mouthStep` uses `mouthArrivalR` (16). Levels where no walk ever started inside a mouth are unchanged (mouthGuard still decides). Why not "every level with an exit": overworld levels and towns list exits at their edges; guarding them all would stop approaches to monsters at map edges for no observed failure.
2. `mouthApproachR` 12 to 15: covers 13.0 with margin 2, stays below the 16.2 that never fired. Walk-start radius stays `mouthR` 9 (16 during the stay), as in fixmouth.
3. Roam target filter (`roamExitClear`) is Euclidean `exitAvoidR + 3` = 17 (was 7 Chebyshev). Filtering the target, not widening every walk guard to 17, keeps fights and passing walks as they were; 17 also lies above the 16 at which `penalizeExits` stops detouring, so a roam walk always gets the cost disc around the mouth. Cost: coverage points within 17 of a mouth are not visited.
4. Step-away: for the turned direction (`mouthTurns[k]`) and then its opposite, the target is accepted only if `leavesExit` (the straight path never gets nearer to the exit than the start, minus 0.5 slack) and, for the first try, ends at least 8 farther; else the flood fallback; else the old log line. The walk is `oneClick`: while the goal is in line of sight the click goes to the goal itself (about 20 subtiles), not to waypoints, so no waypoint click lands in the zone. U: the game walks a long straight click through the zone without firing; DRILLS16 saw 4 of 4 safe multi-click step-aways from 11.2, none yet from 3.2 that was one click.
5. Wanted exit: unchanged predicate, the stay guard never applies to `Dest == wantExit`; tested both ways. Evade cap of fixmouth2 stays. Policy cards stay paused: no area choice changes.

## Tests (`go test -p 1 -count=1 ./internal/nav ./internal/combat`)

`TestStayGuardOutlivesArrival` (10.0 and 13.0 guarded after the flag ended, 17 not, other level not), `TestNoStayGuardWithoutMouthContact`, `TestWantedExitIsNotGuarded`, `TestRoamTargetsFilteredAtTriggerRadius`, `TestLeavesExit`, `TestStepAwayFrom32HeadsAwayInOneClick`, `TestStepAwayBlockedPicksSafeDirection`, `TestExploreSweepNeverClicksInTheMouth` (S: an Explore sweep from the arrival offset never clicks within 15 of the exit after the step-away), combat `TestMoveSkipsLoggedBounceTargets`. The sim has no tile warps, so "zero level changes" is proxied by "zero clicks inside the trigger radius".

## Live proof (not run: no game in this task)

From the fixmouth3 worktree, `d2drill` built there, on the worker host (same as DRILLS17):

```
D2D_AGENT=fixmouth3 D2D_BIN=<fork>/bin/d2d.lab-cheats3 tools/agentrun.sh fixmouth3 --ttl 300 -- tools/d2dslot.sh run -n 0 -t 200 -- <built>/d2drill -group F -only DW.m3 -label fixmouth3-m3-N -v
```

N = 1, 2, 3, then `-only DW.m4 -label fixmouth3-m4`. PASS: zero `level changed during the walk` errors in Clear and Explore and zero `Entered` cycles within 1 s, in all runs. Baseline DRILLS17: 3 of 4 and 4 of 4 entries bounced in DW.m3 (18 level-change errors per run), 2 of 4 in DW.m4. FAIL: attach the filtered logs; check `nav mouth check` (now with `stay=`), whether the step-away was one `nav click`, and the distance of the fired click.

## Overlap (declared)

internal/nav/walk.go and internal/nav/nav.go are owned by the open card fixmouth; fixmouth2 edited the same functions. This change edits the constants block, `NearMouth`, `mouthStep`, the walk follow step (one `oneClick` branch), `walkOpts`, and one Navigator field; the textual conflict is near `mouthStep`. Merge fixmouth and fixmouth2 first. internal/nav/roam.go is owned here.

## Shared-file lines for the merger

- docs/BACKLOG.md: FIXMOUTH3: stay guard held until the level changes, trigger radius 15, one-click away-checked step-away, roam targets filtered at 17 (DRILLS17 bounces); offline tests pass, live DW.m3 x3 + DW.m4 pending (docs/fixmouth3.md).
- docs/INDEX.md: docs/fixmouth3.md.
- CHANGELOG: none until the live proof passes.
