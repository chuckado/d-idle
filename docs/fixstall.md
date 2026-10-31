# fixstall: NPC approach stall (DRILLS5 F1) and pocket escape (F2)

## F1 H.pause: root cause
Evidence (docs/probes/drills/2026-10-31-drills5-h.txt): the lab's own stepping Pause also needed 147, 417, 540 and 950 ticks (6 to 38 s), so the long walk is not a wall-clock or Pause artifact; the bot's Pause failed only because the game-time deadline (15 s = 375 ticks, claim.go 20 s) is shorter than those walks. The error line always names the arrival cell (25.5,22.5): `cmd interact` walks in a straight line and stands still for good at the first camp object (known since 2026-10-30, comment on `ApproachNPC`); the old loop noticed this only after a full 3 s round, then side-stepped blind (3.5 cells, nearest to the NPC first) and repeated rounds of 3 s plus up to 4.5 s of side steps. Several rounds, each mostly dead time, explain 16 to 38 s. Not isolated: which object blocks at (25.5,22.5) (no map in d2dctl); that is UNVERIFIED.

## What changed
- `ApproachNPC` (internal/d2dctl/town_verbs.go): after 4 polls (1.2 s) of a round the position is read; no movement means stalled and the side step starts at once. The budget is `timeout + 1 s per cell` of the straight line to the NPC, on the game clock (`Now`). The callers keep 15 s and 20 s as the base. The wall limit of the adapter call is 90 s.
- Ownership note: internal/d2dctl/town_verbs.go, internal/d2dtown/adapter.go, internal/nav/walk.go and nav.go are owned by the open cards waits-audit, fix-bank and fix-nav, so this card owns only its new files (internal/d2dctl/fixstall_test.go, internal/nav/pinned_test.go); the code edits are listed in the proposal.
- Tests (internal/d2dctl/fixstall_test.go): `TestApproachNPCBudgetGrowsWithDistance` (a 30 s walk of a 30 cell approach passes with a 15 s base), `TestApproachNPCStallDetectedEarly` (side step within 2 s of game time), the existing deadline and side-step tests unchanged.

## F2 D6.d
`nav.walk` (internal/nav/walk.go) gets `walkOpts.escape`, set by `GoTo` and `GoToUnitOf` only (exit, explore and roam walks keep their own logic: `exitUnstick` tries other approach cells before its escape). When the ladder has reached rung 3 and the character is still within 3 subtiles of the first stall, `Escape` (the bot's town portal, `tw.ToTown`) runs once per walk; a new level ends the walk with `ErrLevelChanged` (callers re-plan), a failed or refused escape (no portal) is logged and the ladder goes on, and the escape is tried once more before `ErrStuck`. The stall timer is halved for the rungs after the first when nothing moves (navigator clock, no wall sleeps). Test: `TestPinnedGoToEscapesByHook` (pinned in solid walls: one escape, under 40 s of game time; without the hook still ErrStuck).

## Verified
Offline only (go test of internal/nav, d2dctl, d2dtown with the sim and the d2d stub). VERIFIED LIVE: nothing. UNVERIFIED: that the shorter detection brings the Kashya walk under the base budget; the blocking object; that the escape works from a real pin (the drill's `debug pin` ring).

## Live proof recipe
On the worker host, one hidden slot, as in the DRILLS5 section of docs/probes/drills/README.md: `-only H.pause` (expect the bot's Pause to open Kashya in both tries; ticks well under 950) and `-only D6.d` (expect the character out of the pocket, `nav: stalled in a pocket, escaping by town portal` in the log, GoTo ending in ErrLevelChanged; the drill's verdict logic may need to accept it).
