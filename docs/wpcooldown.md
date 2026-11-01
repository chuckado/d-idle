# Waypoint arrival cooldown removed (card wpcooldown)

> status: current.

- V (decomp, docs/research/waypoint-trace.md W3): game.exe 1.14d has no 10 s arrival cooldown in single player; it only refuses inside a PvP hostility window. The 10 s rule was d2d's; upstream PR #153 (merged 2026-11-01) removed it and the d2d binaries in use (lab2-freezeresume on upstream main 52362d52) carry that.
- Change: `nav.wpCooldownSec` stays for config compatibility but defaults to 0 and is deprecated (docs/knobs.md). 0 means no wait. A value above 0 waits that long plus 1 s after arrival, for an old d2d binary. The retry round after an eaten row click stays (at most 2 rounds, only while the panel is still open and the arrival is under 8 s old); the act-tab and row-choice logic is untouched.
- Tests: `TestWaypointNoCooldownWait`, `TestWaypointKnobWaitsOutCooldown`, `TestWaypointRetriesEatenClick`, `TestWaypointNoRetryForOldArrival` (fake clock, internal/nav/wpclick_test.go).
- live: not verified. Verified by the decomp and the upstream differential only, not on d2d with the rebuilt binary.

## Live drill (later, 3 lines)
1. Arrive in a town or area with a lit waypoint, walk to it (a fresh level change).
2. Take a waypoint-to-waypoint trip straight away with default knobs; read `wp: used A -> B (N clicks, waited 0 ms)` in the log.
3. Pass if N is 1 and the trip took no 10 s wait; if the row click is eaten, set `nav.wpCooldownSec 10` and report.

## Shared-file lines for the merger
- docs/INDEX.md: `docs/wpcooldown.md: waypoint arrival cooldown removed (default 0), live drill recipe`
- docs/BACKLOG.md: wpcooldown: live drill pending (waypoint trip timing, docs/wpcooldown.md)
- CHANGELOG.md: Waypoint trips no longer wait 10 s after arriving (`nav.wpCooldownSec` default 0, deprecated).
