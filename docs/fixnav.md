# FIXNAV: mouth bounce, touched waypoints, act 3-5 town waypoints (card fix-nav)

> status: current. Offline fix; nothing here is verified live (UNVERIFIED until the drills below). Source: docs/probes/postfix-soak1.md, log of the soak read with streaming greps (V = counted in the log).

## Log findings (V, postfix1, Sorceress, d2d)
- 33 `taking the waypoint to town`: 16 failed `level changed during the walk 6 -> 20`, 2 `none on level 20` (the same bounce, the retry ran inside the tower), 15 not logged as failures (arrival or later fallback). No timeout, no panel, no not-activated failure.
- The 6 -> 20 cases: the character leaves the Forgotten Tower into Black Marsh and stands in the tower mouth's trigger. The step-away fired with identical `from`/`to` (334,140 -> 330,160) at least 5 times, each bounced into the tower 3 to 50 ms later; the waypoint walk never started. Not a path crossing an exit: the plan avoids exits (TestGoToAvoidsExitTile); the start point is inside the trigger.
- The same family hits ExitTo gates: `left level 17 for 3 but ended in 19` 40, `2 for 3 ended in 8` 32, `3 for 17 ended in 9` 29 (exits near the goal are not penalized by design; not changed, report only).
- Picker: 78 `leveling target not reachable` (Black Marsh 29, Dark Wood 31, Cave Level 1 13...). Black Marsh was touched at 15:23:46 by the town's waypoint trip (`wp: used 6 -> 4`) but still refused 16 times after: nav touched it, the runner never learned (only goal.go travel and the in-passing touch call markTouched). Dark Wood: 0 touches in 11 visits; the in-passing touch needs a sweepPass with nothing hostile and the waypoint within 70 subtiles; why it never fired there is not provable from the log (UNVERIFIED), the 70 limit is left alone. The picker rule itself is right.
- `gold under the kit cost` (143 empty-source attempts): economy, not changed.

## Fix (shared nav and bot; the trigger geometry is d2d)
- walk.go: a step-away ending in ErrLevelChanged counts a bounce for that exit (level, exit position, 10 min); the next step-away turns by 0, 90, -90, 45, -45, 135, -135, 180 degrees. The trigger area is seed-dependent (docs/d2d-loop.md), so no fixed direction is right.
- waypoint.go: `noteActivated` on a panel opened and on a successful travel, `DrainActivated`; bot/waypoints.go `drainTouched` feeds `markTouched` before activeWaypoints and World read it.
- act2hints.go: `knownWaypointRel` towns 75 (159,49), 103 (44,19), 109 (114,69), measured level-relative (DRILLS4); used only with a level Origin (d2d). Game.exe keeps `knownWaypointPos`.
- Tests: TestMouthStepTurnsAfterABounce, TestKnownWaypointRelTowns, TestNavigatorTouchedWaypointIsActive.

## Drill variants to add (group DW, another agent after merge)
1. DW.m1: Black Marsh arrival from the Forgotten Tower, then `TakeWaypoint(1)` with an activated waypoint: expect arrival in town and at most 2 bounces.
2. DW.m2: same from Stony Field/Cold Plains cave mouths (Den, Cave Level 1) after `debug goto` into the mouth.
3. DW.a12: Act 1 area with an unactivated waypoint in view (Dark Wood, Black Marsh): touch in passing, then a leveling pick must not say `no activated waypoint`.
4. DW.a13: never-visited towns 75, 103, 109 via `debug act N` and `debug goto`: no `none on level N`.
