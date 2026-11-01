# fixf17: DRILLS9 finding F17 (DW.m1 BLOCKED in run 1 of the full F group)

> status: harness fix, offline only; not verified live (recipe below). Source logs: docs/probes/drills/2026-11-01-drills9-f*.txt, `dwm1-h`, `dwm1-i`.

## Timeline of the failing run (VERIFIED LIVE, from the logs)

- The failing call is not the trip to town: it is `TakeWaypoint(outer)` of `fMouth`, the first call of DW.m1 (`debug goto 1`, `Tick 300`, then the navigator walks the town from (141,103) to the town waypoint at (84,69) and takes it to Cold Plains). The error text says "the waypoint trip in".
- Run 1 (`dwm1-i`, uncapped): clicks to waypoints 1, 2, then `nav stuck pos=139,103 rung=0` and `stuck recovered method=reclick` (cur=4, pos 133,97), then NO log line (no click, no stuck, no plan) until the next `nav walk` (run 2). In `f` the same walk reaches cur=9 (pos 102,86) before it goes silent. So the walk dies at a different point of the same route (cur 4 in `dwm1-i`, cur 9 in `f`), not at a fixed cell: no geometry stall.
- Run 2 (passes): the same first lines (stuck at 139,103, reclick), then all 11 waypoints and `wp: used 1 -> 3`.
- Every walk start in every log shows a `nav stuck` after two clicks with the character moved about 2 cells (town 139,103; Cold Plains 114,342; 221,140; 221,159). A real stall needs 1.2 s (StuckMs) of virtual time; two 100 ms polls cannot give that. The navigator's clock must have jumped at the first poll.

## Cause

The lab steps the game with `Client.Step` (`lab.Tick`, the object drills' `Step(250)`), which the Pacer does not see: it only follows `game_ms` at its next own step (`Pacer.follow`). Its virtual time therefore lags every tick the drills stepped since its last step, and the first `Sleep` of the next walk jumps by all of them (the walk's `deadline = Now()+2m` and `anchorTime` were taken before). The `stuck at the first poll` of every walk is that jump (it is at least 1.2 s). In the F group the lag before DW.m1 is large (the chest and shrine drills each step about 300 ticks per attempt, up to 4 attempts, 2 repeats, plus DW.m1's own 325): when it reaches about 110 s, only 4 to 9 s of the 120 s budget are left and the walk (about 10 s) times out mid-route, which is exactly the observed spread. Run 2 starts with only 325 ticks of lag. The product never steps outside the Pacer (grep `.Step(` in cmd/internal: only Pacer and the standalone d2dspawn/d2dcasts tools), so the bot is not affected.

Evidence tags: the silent loop, the stuck at poll 2 of every walk and the variable stop point are VERIFIED LIVE (logs). The size of the lag (about 110 s) is UNVERIFIED: no run logged the Pacer; `syncPacer` now logs it.

## Hypotheses (ranked)

1. Pacer lag from direct steps (above): explains silence, stuck at poll 2, variable stop point, run 2 passing, why only a long sequence of stepping drills triggers it. Reproduced offline: `TestTickKeepsPacerCurrent` (fake lockstep d2d, 2 minutes stepped, deadline expires after one poll) fails with the old `Tick` and passes now.
2. A panel left open (the walk's `continue` branches log nothing): UNVERIFIED, not excluded, but it would also hit run 2 and the clicks before the silence worked.
3. Stale navigator state: ruled out (`Goto` builds a fresh navigator; the stuck memory repeats appear in passing runs too).
4. Leftover monsters, corpses, a chest or shrine on the path, a shrine buff: the stop point varies along the route, in town where the object drills did not go; no evidence.
5. A camp-walk product stall (fixstall family) and timeouts too short: the route is 58 cells, about 10 s; the walk is silent, not stalled.

## What changed

- `lab.Tick` calls `syncPacer` after each step; the object drills use `Tick`; `TakeWaypoint` syncs first and logs `drill: the Pacer lagged the game before a waypoint trip lag=...` when the lag exceeds 1 s.
- No product code, no numeric gate. Other direct `Client.Step` calls in cmd/d2drill (game_lab, hover_lab) are not routed; the waypoint trip syncs itself, so they cannot cause F17.

## Live proof recipe (worker host)

`-group F` x3 invocations (as DRILLS9 `f`): DW.m1 must PASS in run 1 of each (`runs: 2, identical`). Grep the logs for `the Pacer lagged`: a line with a lag near or above 100 s settles the cause (it should now be 0 after `Tick`; a remaining line means another direct step path). If DW.m1 still fails with no lag line, capture `debug dump ui` at the silence and the next probe is hypothesis 2.
