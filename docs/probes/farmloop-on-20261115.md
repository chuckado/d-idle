# Farm loop switched on in the sandbox, 2026-11-15

> status: record. Task farmloopon. Tags: V = verified in this run, U = not verified.

Request: turn the farm loop (docs/farmloop.md) on for the sandbox hero without the watch token or any POST.

## What was set (V)

The running bot reads its knob overrides from `tuning.json` in the hero's state directory inside the data volume (the file the dashboard writes; `internal/bot/tuner.go` loads it once at start, so a restart is needed, V by code reading). One line was added to `overrides`: `"bot.farmReset": 1` (`bot.farmResetPct` 85 and `bot.farmResetCooldownMin` 5 stay at their defaults). A backup copy sits next to the file as `tuning.json.bak-farmloopon` (501 bytes, the state before the edit). Then `docker restart d2play-20261108` (the hero resumes from its save).

## Checks after the restart (V, container logs)

Container healthy, one `d2bot start`, 0 error/panic/fatal and 0 watchdog lines in 4 minutes, hero alive (clvl 4, xp and kills rising, 0 deaths), the file keeps the knob. There is no startup knob dump line, so "the knob is read" rests on the file plus the loader code, not on a log line (U until a `farm reset:` line appears). The loop triggers only on a leveling farm zone at 85 percent coverage when the value rule pays, so it may take a while.

## Measurement

`tools/farmloop/measure.py`, recipe in docs/probes/farmloop-measure.md. Baseline run on the last 12 hours of logs (mixes the previous hero and the fresh one): kills per minute is about 13 to 18 at low coverage and falls to 6 to 12 above 40 to 50 percent (areas 5 and 6, a few minutes per bin: a trend, not a fit, U). No `farm reset:` line yet.

## Rollback

Set the knob back to 0 and restore the backup: `docker exec -u d2 d2play-20261108 cp -p <state dir>/tuning.json.bak-farmloopon <state dir>/tuning.json` then `docker restart d2play-20261108`. The data volume is never deleted.
