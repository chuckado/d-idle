# Reset: leave the game and rejoin

> status: current (botreset + resetinstant, 2026-11-15). Offline tested only; not verified live (checklist at the end).

## What it does
One tap makes the bot leave the current game and re-enter the same character; it then carries on with its plan. A fresh game costs about 15 s (the same Cycler.NewGame the fresh-game leveling and the recycler use). Reset is a break-glass control (user feedback 2026-11-15: "resetting should be instant"), so it never waits and never asks.
- Soft (default, INSTANT): the press cancels the running action (errReset) and the loop starts the new game on its very next tick (tick 100 ms, `bot.tickMs`). It does NOT wait for nothing hostile near: logging out of an offline d2d game is itself the escape. `wait: true` (programmatic callers only, the UI never sends it) keeps the old bounded wait (`resetPatience` 20 s). A paused hero is reset and stays paused; a dead hero waits for its recovery first. Measured on the fake clock (`TestPlayerResetStartsOnTheNextTick`, hostile in reach): 150 ms from press to the new game (V, offline test).
- Escalation: the loop marks every tick as an acknowledgement. When a soft press is not acknowledged within `bot.ResetAckWithin` (3 s) the loop is hung (the failure the watchdog only catches after 600 s), so the wire escalates by itself to the hard path, logs `reset escalated: loop unresponsive`, and the view becomes `resetting` + `hard` with the reason "the bot was stuck, restarting it" (the UI says "Resetting... the bot was stuck, restarting it"). `D2BOT_HARD_RESET=0` switches the escalation off with the hard level.
- Hard: the d2bot process exits with exitStorm, exactly the watchdog's "runner loop made no progress, exiting for a restart" path (cmd/d2bot/watchdog.go), so whatever supervises d2bot restarts it (`--supervise`, the container's restart policy). It is also the fallback when a backend has no Cycler (the reply says so in `reason`). `D2BOT_HARD_RESET=0` switches it off (409 `hard_off`). Without a supervisor or restart policy a hard reset just stops the bot: documented, not detected.

## What is kept
Everything the runner already persists or holds in the same process: the goal chain and journey step, the active and queued orders, until-level orders, the Loop, parked steps, the player's pause (`playerpause.json`). A soft reset does not even restart the process; a hard one reads the same state files as the watchdog restart. Per-game memory (cleared areas, step counters) is forgotten, as for any new game. After rejoining the hero is in town and walks back out by its plan. Safety, equipment and combat code are untouched. A bot-started new game counts as a recycle on the minute line.

## API (additive, `watch.reset/1`)
- `GET /api/reset` (viewer): `{"schema","state":"idle|resetting|rejoining|failed","since","reason","hard","last":{"result":"done|failed","at","hard","reason"}}`. `since`, `reason` and `hard` are absent while idle.
- `POST /api/reset {"hard":false,"wait":false}` (answer listener, same gates as pause; both fields optional, unknown fields refused: the old `force` field is gone, instant is the default). It answers `202` at once with the same body as the GET, plus `"already":true` when a reset was running (a double press changes nothing); the request never blocks for the reset. Refusals are 409 `{"code","error"}`: `measured_run` (D2BOT_RUN is set: a measured run is never reset), `hard_off`. 503 `off` without a runner.
- Goldens: `internal/watch/testdata/reset/` (new: `view-resetting-instant`, `view-resetting-escalated`). State file: `<state dir>/playerreset.json` (a hard reset in flight shows as `rejoining` in the new process until its first in-game snapshot; the last result survives).
- Log lines: `reset requested` (hard, wait), `reset escalated: loop unresponsive`, `reset done`, `reset failed`. The status sentence says "Resetting" (band word) and the whyline code `player_reset` ("You pressed Reset, so I am leaving the game and rejoining.").

## UI (one tap)
A small always-visible Reset button sits in the cockpit header next to the "..." menu (`#btn-reset`, label Reset, aria-label "Reset bot (leaves the game and rejoins)", a 44 px hit area on the 32 px button on a phone). The "..." menu entry, the Orders More entry and the Needs you chip (stuck over 5 min, or loop heartbeat older than 2 min) send the same press. No confirm sheet: the soft reset is cheap and reversible. Within 100 ms (before the reply) the button reads "Resetting...", the strip shows "Resetting... leaving the game" and a toast says "Reset sent"; the server's state then drives the strip until idle, and a toast says done or failed. A 3 s cooldown (and the server's `already`) make a second press a no-op. A refused press gives the button back with the reason as a toast. Only after a FAILED reset the strip offers "Restart the bot", which opens the one confirm sheet ("Restart the bot?", Cancel and Restart, focus kept inside, Esc cancels) for the hard variant. Fit fixtures: `reset-*` and `phonedefect-resetclip` (must fail) pages of `tools/uishots.sh --fit` at the eight sizes plus 412x780 and 360x640.

## Not done
No Undo (a reset cannot be taken back, and needs none); no scheduled or automatic reset (the watchdog and stall ladder keep their own); the hard level cannot tell whether a supervisor exists.

## Live checklist for the sandbox (U)
1. Hero in town, idle: tap Reset in the header. Expect the button and strip to change at once, log `reset requested`, game leaves and re-enters in about 15 s, `reset done`, the plan continues. Orders and the pause are still listed.
2. During a fight (hostile in reach): tap Reset. It leaves immediately (no wait); life never below the chicken line because of the reset.
3. During a hang (no debug hook blocks the loop yet, U): block `Runner.step` in a lab build, tap Reset; expect after 3 s `reset escalated: loop unresponsive`, the process exit, the supervisor restart, then `reset done`.
4. Double tap: the second is ignored (cooldown), one new game only.
5. With D2BOT_RUN set: 409 `measured_run`, a toast says so, nothing happens.
6. After a failed reset: "Restart the bot" shows in the strip and confirms before the hard restart.
