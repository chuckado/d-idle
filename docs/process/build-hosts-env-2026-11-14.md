# Tower builder environment: lock hygiene, light lane, gate, status, prune (2026-11-14)

> status: applied 2026-11-14 (user go: "make whatever improvements to the tower build environment that are appropriate"). Follows `build-hosts-2026-11-14.md` and `build-hosts-stage3-2026-11-14.md`. Tags: V = verified today (the drill is named), S = secondary, U = unverified.

## 1. How agents use the builder (read this first)

1. Batch your probes. One `labrun tower -- sh -c 'a; b; c'` beats three jobs; every job is a ssh, an rsync and a queue slot.
2. A probe (`which`, `ls`, `cat`, `go version`, `git`, `rsync`, `df`, `ps`) goes with `--light`: `labrun tower --light -- 'which go; ls /gamedata | head'`. It has its own two slots, nice 19, a 5-minute limit, and never waits for a gate. Anything else is refused with a message (the allow-list is a guard rail, not a sandbox).
3. The whole-repo gate is ONE job: `labrun tower --gate [--race PKG,PKG|none]`. It runs gofmt, `tools/ci/check.sh` (hygiene, contracts, backlog, uilint, vet), `go test -short -p 4 ./...`, `go test -race` on the named packages (default `./internal/watch/...`), and `leakcheck -mode fail`, under one lock, and prints one `TOWERGATE PASS` or `TOWERGATE FAIL` line with the step times; exit non-zero on failure. Do not hand-chain vet and tests as separate jobs.
4. See who has the slot: `labrun tower-status` (lock holder with job id, age and command, queue, load, memory, disk).
5. NEVER start a daemon, server or `&` background process inside a job. A job ends when its command ends; what it leaves behind is logged (`/data/lock.d/orphans.log`) and, if it is old, killed. If you need a server, start it with `tools/agentrun.sh` on the sandbox host. Detach a long job with `labrun --detach --notify`, not with a daemon.
6. Do not run raw `ssh tower-build flock ...` or builds outside `labrun`; they bypass the queue (the C++ fork builds do this today, see section 8).

## 2. What changed (all in the appdata volume, no image change, no container restart)

| What | Where | Source |
|---|---|---|
| `lockrun` (lock wrapper, both lanes), `towerctl` (status, watchdog, prune), `towergate` (the gate) | builder `/data/bin/` (the d2-builder `/data` volume, appdata on the NVMe pool) | private ops repo `builder/bin/` |
| `labrun`: `tower` wrap is now `/data/bin/lockrun`; new flags `--light`, `--gate`, `--race`; new verbs `tower-status`, `tower-install`; the job id is exported as `LABRUN_JOB` | `~/.local/bin/labrun` | backup `~/.local/bin/labrun.bak-towerenv-20261114` |
| Lock state | builder `/data/lock.d/` (`queue/`, `holder/`, `watchdog.log`, `orphans.log`, `prune.log`) | runtime |

Nothing touches the d2-builder template, Dockerfile or entrypoint, so nothing new is needed in the Unraid UI; the files sit in the container's own data volume next to `gate.lock`. The old command line `flock -w 3600 /data/gate.lock ...` still works and excludes the new wrapper through the same lock file (V: drill 1b).

## 3. Lock hygiene

Cause (V, 2026-11-14): `flock FILE CMD` runs CMD with the lock file descriptor open; a daemon started by the job inherits it, so the lock stays held after the job ends. A daemon held the slot for 3 h while about 10 jobs queued.

- `lockrun` takes the lock with `flock -o` (closes the descriptor for the job), so only the flock process holds it and it ends with the job (V: drill 1, a job leaves `sleep` daemons behind, the next job starts in 0 s; control with plain `flock`: the next job is blocked, rc 1 after 5 s).
- The job runs in its own session (`setsid`), under `nice -n 10 timeout -k 30 45m`. After it ends, processes left in its session are written to `orphans.log`; on a timeout they are stopped.
- Watchdog (`towerctl watchdog`, run by every queued `lockrun` every 30 s, so whoever waits frees the slot, and by hand): (A) a lock holder whose parent is init, older than 10 minutes, with no live job record: killed. (B) any holder or job older than its limit plus 5 minutes: whole process tree killed. (C) a holder record whose process is gone: removed. Every action is logged to `watchdog.log` (V: drill 2, 2b, 3; ages shortened with `TOWER_*` environment overrides, which exist for drills only).
- "Idle CPU" is not a separate rule: the 45-minute timeout is the age limit, and rule A covers abandoned daemons. U: a job that is alive, busy-looping and under 45 minutes is left alone (as before).

## 4. The light lane

`labrun tower --light`: lock files `/data/light.0.lock` and `/data/light.1.lock` (two slots, separate from `gate.lock`), `nice -n 19`, 5-minute limit, 2-minute wait. Only probe commands pass the allow-list (`which ls cat head tail wc grep echo printf test stat df du ps top uptime free nproc date sleep cd git rsync sort uniq cut tr sed awk find file diff sha256sum go version|env|list|help`, shell loops and assignments); `go test`, builds, `rm`, `python3` are refused (exit 2). A light job whose mirror has a gate running in it skips the rsync (it sees the tree as last synced). V: drill 4 and a real `labrun tower --light` while the gate queue was full (7 s including the sync; a refused `go test ./...` returns exit 2 in 2 s).

## 5. Leakcheck and CI parity

The mirror is not a git checkout (a worktree's `.git` is a pointer file), and the builder has no credentials, so there is no clone. `towergate` builds a throwaway git index of the synced tree in `/data/gateidx/<pid>` (`GIT_DIR` and `GIT_WORK_TREE` set, `main` and `origin/main` point at that one commit, deleted at the end). `git ls-files` therefore lists exactly what the agent has (`.gitignore` honoured), so `leakcheck` and `tools/ci/check.sh` run unchanged. Limit: check steps that compare with real history (`--changed`, drift) see an empty diff on the builder; the merge worker's own gate is still the authority for those. Result of the first run: section 7.

## 6. Status and disk policy

- `labrun tower-status [--json]` (V): one screen; also lists old-style holders and waiters that have no record.
- `towerctl prune` (V dry run; applies only with `--apply`, `--force`, or `--if-needed` when `/data` is over 85 percent, today 72 percent): in `/data/build` keep the newest 6 directories, in `/data/src` 8, in the labrun mirrors `/data/labrun/herdr/code/d2lod-wt` 10; only entries older than a day, never `vp-bins` or `prebuilt`, never a directory a running process or a holder record refers to; finished job logs older than 3 days; the Go build cache entries unused for 7 days, only when the disk is over the threshold (or `--force`). `lockrun` runs `prune --if-needed` after every main-lane job, so no cron is needed. The dry run today lists 10 old fork build and source directories, about 560 MB.
- Optional, needs the user (root on the NAS): an Unraid User Scripts entry `docker exec d2-builder /data/bin/towerctl prune --if-needed` hourly, visible in the plugin's UI. Not done: no root access here.

## 7. Drills and the first gate run

Run on the builder with `TOWER_LD=/data/lock-drill TOWER_MAIN_LOCK=/data/drill.lock` (a private lock; the real `gate.lock` was never touched by a drill), 2026-11-14 about 16:40 to 16:43 EST:

1. Daemon inside a job: job A leaves two `sleep` daemons; job B starts and ends in 0 s; `orphans.log` names the one left in A's session. V.
1b. Control, old style (plain `flock`, daemon inherits the fd): the next `flock -w 5` fails rc 1 after 5 s. V.
2. Lock held by an old-style daemon: `status` shows it as HELD with no record; `watchdog` kills it (rule A); the next job runs in 0 s. V.
2b. A stuck holder (limit shortened to 6 s plus 2 s grace): `watchdog` kills the job tree (rule B). V.
3. A waiter frees the stuck holder itself: the waiting `lockrun` ran after 33 s without a manual watchdog. V.
4. A main-lane job holds the lock for 40 s; the light probe ran in 0 s; `go test` and `rm` were refused. V.

First gate run (`labrun tower --gate`, this worktree, V): `TOWERGATE PASS: gofmt ok 0s check.sh ok 24s test ok 159s race[./internal/watch/...] ok 36s leakcheck ok 24s`; the job waited about 15 min in the queue behind other gates, 1261 s in all.

## 8. Rollback

- labrun: `cp ~/.local/bin/labrun.bak-towerenv-20261114 ~/.local/bin/labrun` restores the old wrap `flock -w 3600 /data/gate.lock nice -n 10 timeout 45m` at once; nothing else needs undoing, because the old line works with or without the new files.
- Builder: `rm -rf /data/bin /data/lock.d /data/lock-drill /data/gateidx /data/drill*.lock /data/drill-light*` removes everything added (files only, owned by the builder user).
- Ops repo: `git revert` of the `builder/bin` commit.

## 9. Not done, needs a decision

- Unraid User Scripts entry for prune (root on the NAS).
- The C++ fork builds (`vp-*`, run by `build-fork.sh` through plain ssh) do not go through the lock; they compete with gates for the 6 CPUs. They should call `lockrun` too (one line in `tools/build-fork.sh`, ops repo, owned by the fork-build card).
- `docs/process/dispatcher-brief.md` belongs to TASK-310; the two lines to add there are in the task card result.
- No container restart was needed or done.
