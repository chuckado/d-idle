# Operator quest control

> status: current (card opsquest). Code: `internal/ops`, `cmd/d2bot/ops*.go`, `tools/questloop.sh`.

## What it is
A way for the person (or agent) who runs the container to press the quest board verbs without the viewer's passcode: Start, Hand in, Retry, Stop and the auto switch, plus a read of the board. It is a local unix socket in the instance state directory, served by the bot, and a client in the same binary (`d2bot ops quest ...`). It calls the same `questboard.Desk` functions the viewer's POST routes call (`internal/watch/questsapi.go`), so the refusals (busy, locked, not_ready, paused...) are the viewer's.

## Turn it on or off
- Off by default. It exists only when the container has `D2BOT_OPERATOR=1` in its environment. Not in the friends pack defaults.
- Never in a measured run: with `D2BOT_RUN` set no socket is created, and a request that arrives later is refused (`measured_run`).
- Off again: remove the variable and recreate the container (the socket file is removed at start of the next run; it is mode 0600).
- Env only; there is no knob and no flag.

## Verbs (the allowlist is in code, `internal/ops/ops.go`)
`start`, `retry`, `cancel`, `handin`, `auto` (change the board) and `status` (read). Every other verb is refused (`verb`); there is no sell, vault, hero, character, tuning, respec, order, mode or dismiss verb.

```
docker exec <container> <d2bot> ops quest status [--json]
docker exec <container> <d2bot> ops quest start|retry|cancel|handin <quest> [--chunk ID] --why "reason"
docker exec <container> <d2bot> ops quest auto <quest> --on|--off --why "reason"
docker exec <container> <d2bot> ops quest wait <quest> [--chunk ID] [--timeout 600]
```
`--why` is required for every verb that changes something (1 to 200 characters, one line). Exit codes: 0 done, 1 refused or not finished, 2 usage or a `wait` timeout. The socket is found by `--sock`, then `$D2BOT_OPERATOR_SOCK`, then `/data/instances/main/state/operator.sock` (the container layout).

`wait` polls `status` every 5 s until the chunk is done (`finished`), parked (`blocked`), back to ready after running (`stopped`), was not running at all (`idle`) or the timeout passes (`timeout`). It prints one line `RESULT outcome=... quest=... chunk=... state=... step=... code=... text="..."` with the last step and the park reason.

Finding the binary in the container: `/data/didle/current/bin/d2bot` when the `current` link exists (a self-updated release), else `ls -d /data/didle/releases/*/bin/d2bot | sort | tail -1`, else `d2bot` on the PATH. `tools/questloop.sh` does this lookup itself (`--bin` or `D2BOT_BIN` overrides).

## Audit
Every verb the desk ran adds a line to the hero's quest audit ring (`<Hero>.audit.jsonl`, the same file the viewer writes) with `remote:"operator"`, `agent:"ops-cli"`, `path:"operator/<verb>"`, the desk's code and reason, and `why` (the caller's reason; the field is new and empty for viewer lines). A refusal before the desk (bad verb, missing reason, rate limit) is logged, not written to the ring, so a runaway loop cannot push real events out of it.

## Limits (in the bot, per process)
At most 6 mutating verbs per quest per hour, 30 per hour for the whole board, and a 60 s cooldown after a `retry` during which `retry` and `start` of that quest are refused (`cooldown`; `cancel` still works). Excess is `rate_limited` with the reason in words.

## Threat model
1. Anyone with `docker exec` on the host already owns the container, its state and its save; this adds no power they lack.
2. The socket is not on the network: a unix socket, mode 0600, owned by the bot's user, created under a temporary name and renamed into place so it is never looser.
3. It is not exposed through the viewer, a port or the answer listener; the viewer's passcode and token are neither read nor needed.
4. The scope is the quest board only; the allowlist is tested, anything else is refused.
5. Every action is audited with its reason; the limits stop a runaway loop.
6. Off by default and never in measured runs, so gates and soaks cannot be touched by it.
7. Caveat: the player's own Stop/Start history shows operator lines too; the board does not distinguish them visually (only the audit does).
8. Caveat: the limits live in process memory and reset at a bot restart.
9. Caveat: the quest board is class-agnostic, so nothing here depends on the hero's class or build.
10. Do not put the socket on a shared mount; the instance state directory must stay private to the container.

## Driving quests overnight
`tools/questloop.sh <container> <quest|--all> [--max-tries 3] [--wait 300] [--stall 6] [--backoff 65] [--budget 10800] [--tries-backoff 1200] [--deadline 07:30] [--bin PATH]`.

- Per quest: status, then start (retry when parked; nothing when a chunk already runs, so restarting the loop is safe), then `wait` in `--wait` second windows with one heartbeat line per window (elapsed, try, step, progress).
- Slow progress is not a failure: `status` rows carry a `progress` counter (the server counts each change of state or step it sees for a chunk), and `wait` prints it. A window that ends with a moved step or counter just continues. `--stall` windows without any move, a park, a cancel or a refusal count as a failed try.
- After `--max-tries` tries the loop backs off `--tries-backoff` s (20 min, under the 6 per quest per hour limit) and goes on with the same quest until the per-chunk `--budget` (3 h) or the `--deadline` (HH:MM local time, default 07:30, or `@EPOCH`) passes. A refused `cooldown` or `rate_limited` waits and is not a try; `not_ready`/`locked`/`unavailable` skips the quest with a line.
- A finished chunk moves on (the next chunk of the quest, then the next quest). `--all` walks the board in order, skips done and locked quests, repeats until nothing is startable, and exits 1 if any quest used up its budget or the deadline passed.
- Hero deaths are not exposed by the board, so they are not a stop condition. The area is not exposed either; progress is the step and state counter.
- Offline test: `bash tools/questloop_test.sh` (a fake `docker` shim and a fake clock).
