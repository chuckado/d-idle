## Why

docs/research/waypoint-trace.md (DECOMPWP) found that the bot had zero successful waypoint uses in all soaks: 295 of 295 "nav waypoint failed" lines are `unit not found`. d2d lists a waypoint as an `npcs` row named "Waypoint", which d2dsrc filed as an object with class -1, so `nav.nearestWaypointUnit` never saw it (W1). The nav row pitch of 34 px also missed panel rows 7 and 8 (W2).

## What Changes

- internal/d2dsrc (d2d only): an npcs row whose name contains "waypoint" becomes an object with class `WaypointClass` (119, which `d2map.IsWaypointObject` accepts); the memory reader's classification (Game.exe) is untouched.
- internal/nav (shared): `wpPanel` rows are game.exe's hit-box centres (135 171 207 243 280 316 352 388 424), tabs 112 + 64 act; the icon-column fallback goes (it was the pitch error). The 10 s post-arrival wait stays (d2d rule, W3) with a comment.
- cmd/d2drill + internal/drill: drill group D3 (D3.b inactive, D3.a1..a5 travel, D3.c second use).
- Delta: backend-contract ADDED requirements.
