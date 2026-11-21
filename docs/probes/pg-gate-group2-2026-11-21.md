# Proving Ground gate, cost-cut group (2026-11-21)

> status: record of the second rollout group of docs/design/fable-review-2026-11-21/proving-ground-gate.md (owner approval of the group on 2026-11-21; ADR 0027 answers Q3 to Q5, Q9). Group one is docs/probes/pg-gate-group1-2026-11-21.md. Private-repo parts are described by role. Evidence tags: V verified (the command ran), S read in source or secondary, U not verified.

## Cards, branches, what is enforced now

| card | task | public branch | ops side | state |
|---|---|---|---|---|
| mglanes | TASK-749 | agent/pgc1 (merged) | lane wrapper on the builder: quotas, whole-repo refusal, merge priority, wait budget, load guard, two long slots | installed on the builder; WARN-ONLY until 24 h after the install, then enforced (no restart needed); merge priority and the second long slot are active now |
| mgwatchdog | TASK-752 | agent/pgc2 (merged) | heartbeat, three alarms, phase budgets, status exit code, watchdog timer | status line, exit code, timer (5 min), restart of a dead worker and the stop flag active now; heartbeat and phase budgets need a worker restart (the worker predates them) |
| mgcache | TASK-747 | agent/pgc3 (merged) | per-merge gate stats, replay measurement, nightly uncached race, queue red hold, timers | `-count=1` NOT dropped; stats active now; nightly timer 23:10 and replay timer 01:30 (New York time) installed; the red hold needs a worker restart |
| pgledgeron | TASK-777 (new) | agent/pgc4 (merged) | pull of the builder's ledger spools onto the board host (script, test, 10-minute timer) | timer installed; no live ledger line yet |
| pgclock | TASK-739 | agent/pgc5 (merged) | none | merged code on the env clock; B3 (soak stop) open |

Commit hashes of each merge are in the queue log (`mergeq list`) and `git log --first-parent`; they are not copied here.

## Measurements

- V, queue file, 405 code merges with a gate phase (all with `-count=1`): gates phase median 280 s, p90 1041 s. The design review read 435 s and 1990 s over its window (heavier agent load). The per-package split (whole-repo seconds, cached packages, race seconds, load) exists only for merges after the stats change; the replay of the last 20 code merges without `-count=1` has NOT run yet (first run at the 01:30 timer). The decision on Q3 needs it and 20 live merges in the cached mode (`MERGE_GATE_CACHED=1`, default off).
- V, the lane drill (`test_lockrun.sh`, 14 checks) and a live main-lane job from an agent worktree that queued behind a real merge and showed class merge in the lane state.
- V (docker stats and the container list, 10:00 to 10:14 New York time): the builder container sat at 7.99 of 8 GiB with the NAS host at a load average of 276 to 357, then the container exited with code 0 (a clean stop, not started by this work) and sshd refused connections. The merge in progress (a whole-repo gate) failed on it (queue item 1060, `gate go-whole failed`, a builder outage, not a test failure). The container came back about 10:18 (uptime 9 min at 10:27; the lane files and the installed wrapper survived on its data volume; the lane policy log holds one line since the install, so no flood from the wrapper). The five cards then merged through the queue (items 1062, 1064 to 1067; the Go card passed the gate on the builder). U: the cause of the overload; the lane wrapper change was installed 09:23 and adds only a queue-file check and a 1-second poll loop for agent jobs, but it is not ruled out.
- U: the builder was unreachable over ssh from about 09:30 to at least 10:10 New York time with its container memory at the 8 GB cap (docker stats V): the merge in progress (a whole-repo gate) hung on it. Not shown to be caused by the lane change (the lane wrapper only adds a queue-file check and a poll loop); a lane wrapper in the ops repo history can be re-installed to rule it out.

## Not done, and why

- pgclock B3 (soak stop is tick based): the Pacer keeps one watcher (the guard's), the stop needs a second slot in `internal/d2dinput`; proposed, not built.
- The `determinism-run -n 3` pair that should prove the guard-reuse fix (journal pair ticks 2242 and 7838 of group one): not run, the builder was not reachable. Checklist: `labrun BUILDER --lane long --detach --notify -- tools/provingground/run.sh determinism-run -n 3 -id town.resupply,act1.burial.blood_raven -class Paladin -seed 11 -parallel 1` from a worktree with the pgclock merge; expect the first journal difference to move or vanish. Cause (S, code read): `Reader.Recent` aged a snapshot on the WALL clock (350 ms window): a quiet run reused the last snapshot, a loaded one queued a fresh `debug players` and `debug missiles` read, so the command stream differed at equal state. Fixed to the env clock; `goals.Manager` and the farm timers likewise.
- Worker restart (heartbeat, phase budgets, nightly hold): the queue was never empty; restart with `mergeq daemon stop` then `start` when it is.
- First live ledger line and status page (checklist in docs/process/pg-ledger-ops.md).
