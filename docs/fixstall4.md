# fixstall4: DRILLS8 finding F13 (H.pause, the last open gap of the Kashya walk)

> status: offline fixes; none verified live yet (recipe below). Follows [fixstall3.md](fixstall3.md); evidence in docs/probes/drills/README.md (DRILLS8) and `2026-10-31-drills8-h1..h3.txt`.

## Evidence and root causes

DRILLS8 `h1`: the failing try ends at 796 ticks (`not reached in 28.4s`), timeline (25 ticks of 40 ms per second): the player arrives at the camp block at t273 (26.4,20.5), then stalls and moves in short bursts (t369, t483 at 25.5,22.9, t555 at 25.5,22.5, t633, t729 at 26.1,20.1, t855 at 25.9,20.5) for 28 s until a route from (25.9,20.5) walks west along row 18.6 (t933 to t1013). The counters: 9 interacts, 7 `cmd move`, 2 `debug grid`: only two routes were planned in 5 stalls.

- Cause 1, the first leg cuts the block corner (VERIFIED offline on the real grid fixture `internal/d2dctl/testdata/camp-grid.txt`; that the live leg stopped there is UNVERIFIED, there is no live log of a single leg). The route is a breadth-first path on subtiles, but the old waypoints sat every 12 steps and each leg is a STRAIGHT `cmd move` (the d2d verbs stop at the first object, docs/fixstall.md). From (26.4,20.5) the path first goes east around the block face (x 133, rows 101 to 97) and then north-west; the chord from the stall cell to the first waypoint (126.5,91.5 subtiles) runs through the block. `TestPlanAroundLegsAreFreeLines` fails on the old code for every stall cell. In a fake world on the real grid (stop at the first blocked subtile, 2 cells/s from the DRILLS8 timeline) the old code ends `not reached in 24.6s (player at 25.4,22.9)`, the same cell the live timeline shows at t483 (25.5,22.9), because the failed first route is followed by a blind side step that walks south. From (25.9,20.5) the second route goes straight north and works: that is the route that freed the player live.
- Cause 2, the stalls are slow and the budget is a constant. The old loop routed on odd stalls only and between two routes ran a blind side step (up to 3 tries of 1.5 s), the stall check only compared with the start of the round (a player that walked 1 s and then stopped waited out the 3 s round), and the deadline never moved: timeout plus 1 s per straight-line cell (about 28 s for this walk) was spent by the failed legs and side steps before the working route began.
- Cause 3 (text): the error reported `prev`, the position of the previous poll.
- Also found: a pocket cell (26.1,20.1, an isolated free subtile at the block face, where the player stood at t681 to t729) has no route at all; and the stub grid reply of `TestApproachNPCRoutesAroundTheWall` had the `ok` line first (an empty grid, the test passed through side steps); the real reply is rows first, then `ok` (fixed in the test).

## What changed

- `internal/d2dctl/approach_path.go`: `legWaypoints` thins the path by line of sight (`lineFree`: the segment from the previous waypoint, continuous subtile positions, keeps the route's clearance and a quarter-subtile margin, except around the start, which may touch a wall); a bend always gets a waypoint; the old 12-step cap stays. `planAround` falls back, after clearance 1 and 0, to the nearest free subtile within 4 of a boxed-in start (UNVERIFIED live). `routeCells` measures the route.
- `internal/d2dctl/town_verbs.go` (`ApproachNPC`, `routeAround`): a route is planned at EVERY stall, the blind side step runs only when no route moved the player; the round also ends after 4 polls (1.2 s) without movement at any point of the round (a Level poll from the second poll on); when a route is planned the deadline grows to at least now plus 1 s per route cell plus 4 s (a 2 cells/s walk, so about 2x margin); the error reports the final position.
- Tests (`internal/d2dctl/fixstall4_test.go`, fake world on the real camp grid): `TestApproachNPCCampStallCells` (four stall cells reach Kashya within 20 s; old code fails all four), `TestApproachNPCSecondRouteFrees` (the first `cmd move` leg is dropped; the second route frees the player), `TestPlanAroundLegsAreFreeLines`, `TestApproachNPCStopMidRoundRoutesQuickly` (first leg within 2.5 s), `TestApproachNPCErrorReportsFinalPosition`, `TestPlanAroundFromPocket`. All earlier d2dctl, d2dtown and drill tests pass.

## VERIFIED vs UNVERIFIED

- VERIFIED offline: the chords of the old route cross the block on the real grid; the old code reproduces the live failure shape in the fake world; the new code reaches Kashya from the four stall cells.
- UNVERIFIED live: that the fake world's behaviour (straight leg that stops at the first blocked subtile) is what `cmd move` does; that a clearance-0 route along a wall edge is walked by the real player; the pocket escape; the speed of 2 cells/s outside the one DRILLS8 segment.

## Live recipe (worker host, lab build as DRILLS8)

H.pause x3 invocations (`-group H -only H.pause`, 2 repeats each): Kashya must open in 6 of 6 first-run bot tries AND 6 of 6 second-run tries (DRILLS8: 3 of 6 and 2 of 2 per invocation). On a failure keep the `H.pause-diag` timeline and the verb counters (`cmd move` leg count vs `debug grid`), and check whether the first leg stopped near its start.
