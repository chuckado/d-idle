# Reset: leave the game and rejoin

> status: current (botreset, 2026-11-15). Offline tested only; not verified live (checklist at the end).

## What it does
One press (a confirm sheet first) makes the bot leave the current game and re-enter the same character; it then carries on with its plan. A fresh game costs about 15 s (the same Cycler.NewGame the fresh-game leveling and the recycler use).
- Soft (default): the runner loop asks the Cycler for a new game at the next safe moment (nothing hostile within the clear radius, at most `resetPatience` 20 s of waiting, then it leaves anyway: a logout is itself an escape). `force: true` skips the wait. A paused hero is reset and stays paused; a dead hero waits for its recovery first.
- Hard: the d2bot process exits with exitStorm, exactly the watchdog's "runner loop made no progress, exiting for a restart" path (cmd/d2bot/watchdog.go), so whatever supervises d2bot restarts it (`--supervise`, the container's restart policy). It is the only way out when the loop itself hangs, because the soft level needs the loop. It is also the fallback when a backend has no Cycler (the reply says so in `reason`). `D2BOT_HARD_RESET=0` switches it off (409 `hard_off`). Without a supervisor or restart policy a hard reset just stops the bot: this is documented, not detected.

## What is kept
Everything the runner already persists or holds in the same process: the goal chain and journey step, the active and queued orders, until-level orders, the Loop, parked steps, the player's pause (`playerpause.json`). A soft reset does not even restart the process; a hard one reads the same state files as the watchdog restart. Per-game memory (cleared areas, step counters) is forgotten, as for any new game. After rejoining the hero is in town and walks back out by its plan. Safety, equipment and combat code are untouched. A bot-started new game counts as a recycle on the minute line.

## API (additive, `watch.reset/1`)
- `GET /api/reset` (viewer): `{"schema","state":"idle|resetting|rejoining|failed","since","reason","hard","last":{"result":"done|failed","at","hard","reason"}}`. `since`, `reason` and `hard` are absent while idle.
- `POST /api/reset {"hard":false,"force":false}` (answer listener, same gates as pause; both fields optional, unknown fields refused). Same body as the GET, plus `"already":true` when a reset was running (a double press changes nothing). Refusals are 409 `{"code","error"}`: `measured_run` (D2BOT_RUN is set: a measured run is never reset), `hard_off`. 503 `off` without a runner.
- Goldens: `internal/watch/testdata/reset/`. State file: `<state dir>/playerreset.json` (a hard reset in flight shows as `rejoining` in the new process until its first in-game snapshot; the last result survives).
- Log lines: `reset requested` (hard, force), `reset done`, `reset failed`. The status sentence says "Resetting" (band word) and the whyline code `player_reset` ("You pressed Reset, so I am leaving the game and rejoining.").

## UI
Reset... in the "..." menu, "Reset (leave the game and rejoin)" in the Orders More sheet, and a Needs you row with a Reset chip when the bot is stuck (stall detector `Runner.StuckFor` over 5 min, or the loop heartbeat older than 2 min). All three open one confirm sheet (`resetui.js`): "Leave the game and rejoin; the bot resumes its plan and orders." with Cancel and Reset (44 px targets, focus kept inside, Esc cancels); after a failed reset a quiet "Restart the bot instead" appears. A strip "Resetting..." shows until the state is idle, then a toast says done or failed. Fit fixtures: `reset-*` pages of `tools/uishots.sh --fit` at the eight sizes plus 412x780 and 360x640.

## Not done
No Undo (a reset cannot be taken back); no scheduled or automatic reset (the watchdog and stall ladder keep their own); the hard level cannot tell whether a supervisor exists.

## Live checklist for the sandbox (U)
1. Hero in town, idle: press Reset, confirm. Expect log `reset requested`, the strip, game leaves and re-enters in about 15 s, `reset done`, the plan continues. Orders and the pause are still listed.
2. Hero in a fight: press (soft). Expect it to finish or leave within 20 s; `force` leaves at once. Life never below the chicken line because of the reset.
3. During a real hang (loop not stepping): the Needs you chip appears after 2 min; soft shows no progress, press again with "Restart the bot instead" (or POST `{"hard":true}`): log `reset requested hard`, `exiting`, supervisor restart, then `reset done`, strip gone, plan resumed.
4. Double press: second press answers `already`, one new game only.
5. With D2BOT_RUN set: 409 `measured_run`, nothing happens.
