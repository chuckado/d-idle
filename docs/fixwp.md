# FIXWP: waypoint primitive fixes (card fix-wp, P23)

> status: current. Findings W1 to W4 of docs/research/waypoint-trace.md; live proof is drill group DW (docs/probes/drills/2026-10-31-dw.txt).

## Finding and fix
- W1 (d2d only): d2d lists a waypoint as an `npcs` row named "Waypoint" (V live: debug presets shows object 119 for the Rogue Encampment). d2dsrc filed it with class -1, so `nav.nearestWaypointUnit` never found it (295 of 295 failures). Now any name containing "waypoint" gets `d2dsrc.WaypointClass` (119; `d2map.IsWaypointObject` accepts it; the exact objects.txt id of other acts is not in npcs and nav does not need it). Game.exe's reader is untouched (it classifies by the real object class).
- W2 (shared nav): rows are game.exe's hit-box centres y 135 171 207 243 280 316 352 388 424 at x 230, tabs x 112 + 64 per act at y 77. V live: d2d's `dump ui` reports `wp_rows` rects (x 98..377, tops 121 157 193 229 266 302 338 374 410, 29 px) and `wp_tabs` [80,60,64,31]. The old pitch of 34 missed rows 7 and 8; the icon-column fallback (a symptom of that error) is gone.
- W5 (found by the drill, d2d only): `knownWaypointPos` (Lut Gholein 5069,5084) is act-absolute, d2d positions are level-relative (`debug level 40` origin 1000 1000 tiles), so the walk to a waypoint whose preset has not streamed in aimed 5000 subtiles away ("no path ... {5069 5084}"). `d2map.LevelMap.Origin` (set by d2dworld for OriginNone) is subtracted in nav. Game.exe: Origin is zero.
- W3 untouched: `WpCooldownSec` stays (d2d refuses for 10 s after any arrival; fixed upstream by #153, merged 2026-11-01, the knob stays until the lab binary has it); comment points at the trace.

## Tests
`TestWaypointRowHasWaypointClass` (d2dsrc: row mapped, other names stay -1, with and without Tables), `TestNearestWaypointUnit` (class -1 not found, nearest wins), `TestWaypointRowCentresHitBoxes`, `TestWaypointClicksEveryLevel` (all 39 levels: tab, row centre), `TestDW` (drills against a fake).

## Live drills (DW, slot with the click lab binary, throwaway Sorceress, `debug wp all` and `debug goto` ledgered; each run twice, identical; P35 gate clean)
| drill | verdict | observed |
|---|---|---|
| DW.a1 | PASS | Rogue Encampment to Cold Plains (row 1): area 3, 193 ticks, 2 clicks (tab, row), arrival 1 subtile from the waypoint |
| DW.a2 | PASS | to Catacombs level 2 (act 1 row 8): area 35, 193 ticks |
| DW.a4 | PASS | Lut Gholein to Canyon of the Magi (act 2 row 8): area 46, 193 ticks, after W5 |
| DW.a5 | PASS | Lut Gholein to Palace Cellar 1 (act 2 row 6): area 52 |
| DW.b | PASS | not activated: error "level 3 not reached (waypoint not activated?)", area stays 1, panel closed, 487 ticks |
| DW.c | PASS | immediate second use: waited, Stony Field reached after 285 ticks ("waited 5840 ms" line) |
| DW.a3 | BLOCKED | cross-act row: d2d lists only the current act's tab until the act is unlocked (`wp_tabs` one rect after `debug wp all`; clicks on tab 2/3 do nothing). Missing: a cheat that unlocks act 2 (`debug act 2`) |
First run (before W5 and the harness fix): a2 and c PASS, a1 flaky, a4/a5 FAIL with the W5 signature. Harness note: `debug goto` is an arrival for d2d even into the same level, so the lab starts a fresh navigator per goto.

## What remains
DW.a3 and acts 3 to 5 (towns 75, 103, 109 have no known position; the preset streams in late, as in Lut Gholein). The tab x for acts 4 and 5 and the post-fix soak (W4) are unmeasured. The 10 s wait can drop when the d2d patch lands. The lockstep hover blind spot did not bite: the verb interacts through `cmd interact`.
