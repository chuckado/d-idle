# Merge queue watchdog (card mgwatchdog, TASK-752)

> status: current, rolled out 2026-11-21. Design: `docs/design/fable-review-2026-11-21/proving-ground-gate.md` section 4.7. The code is in the private ops repo (the merge queue tool, the merge script, user timers on the head box); this page describes it by role.

## What it does

| piece | behaviour |
|---|---|
| heartbeat | the queue worker writes a heartbeat file at the top of every loop turn (about every 30 s while idle, between merges while one runs) |
| alarm 1 | the worker is not running, or (no merge running) the heartbeat is older than 10 minutes |
| alarm 2 | one queue item has been `running` for more than 45 minutes (the line names the phase merge.sh was in) |
| alarm 3 | items are eligible, nothing started or finished for 30 minutes, and CI of main is green |
| timer | a user timer runs `mergeq watchdog` every 5 minutes: evaluates the alarms, restarts a dead worker (at most 3 times an hour; NOT after a deliberate `mergeq daemon stop`, which leaves a stop flag that `daemon start` removes), sends the existing phone push once per alarm until it clears, and sets one tile on the board (blocked while an alarm holds, done when healthy) |
| status line | `mergeq status` prints `watchdog: ok` or `watchdog: ALARM: ...` and exits 2 while an alarm holds, so the coordinator's own tick notices without reading text |
| phase budgets | `merge.sh` records its phase (preflight, merge, gates, push, ci-wait) in a phase file; the worker kills a merge whose phase exceeds its budget (preflight 5 min, merge 5 min, gates 30 min, push 3 min, ci-wait 12 min) and fails the item naming the phase (`phase gates exceeded its budget of 1800s`); the old single timeout stays as an outer cap (50 minutes). Budgets: `MERGEQ_PHASE_BUDGETS=gates=2400,...` |
| service unit | a unit with `Restart=always` for the worker is in the ops repo but NOT enabled: switching needs a stopped worker and an empty queue; until then the timer restarts a dead worker within 5 minutes |

## Verification

- V: ops `tools/tests/test_mergeq_watchdog.py` (6 tests, offline, fake merge script and fake CI): a merge hung in the gates phase is killed and named when over budget (the test fails against the old single timeout, which would wait 60 s of the fake hang), `status` alarms and exit code for a dead worker, a deliberate stop, a running item over 45 minutes, a stale heartbeat, an idle queue with green CI, a restart by the timer command and no restart after a stop.
- V: the existing queue tests (`test_mergeq.py`, `test_merge_wt.py`) pass with the change (the status calls now allow exit 2).
- V: the timer is installed and the first run printed `watchdog: ok` against the live queue.
- U: the heartbeat, the phase budgets and the nightly red hold live in the worker process: they take effect when the worker is restarted (the running worker predates them; `merge.sh` already writes the phase file when the worker passes the variable). Status, the watchdog command, the timer and the stop flag are active now.
- Not done here: storing failing test names and the flake ledger per failed gate (card mgonejob).
