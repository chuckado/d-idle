## Why

docs/research/waypoint-trace.md (DECOMPWP) found that the bot had zero successful waypoint uses in all soaks: 295 of 295 "nav waypoint failed" lines are `unit not found`. d2d lists a waypoint as an `npcs` row named "Waypoint", which d2dsrc filed as an object with class -1, so `nav.nearestWaypointUnit` never saw it (W1). The nav row pitch of 34 px also missed panel rows 7 and 8 (W2).

## What Changes

- internal/d2dsrc (d2d only): an npcs row whose name contains "waypoint" becomes an object with class `WaypointClass` (119, which `d2map.IsWaypointObject` accepts); the memory reader's classification (Game.exe) is untouched.
- internal/d2map LevelMap.Origin + internal/d2dworld (d2d only): level-relative coordinates carry the act origin so nav can use its act-absolute known waypoint position (W5).
- internal/nav (shared): `wpPanel` rows are game.exe's hit-box centres (135 171 207 243 280 316 352 388 424), tabs 112 + 64 act; the icon-column fallback goes (it was the pitch error). The 10 s post-arrival wait stays (d2d rule, W3) with a comment.
- cmd/d2drill + internal/drill: drill group DW (DW.b inactive, DW.a1..a5 travel, DW.c second use).
- Delta: backend-contract ADDED requirements.
