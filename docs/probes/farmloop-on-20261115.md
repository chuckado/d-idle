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

## Rolled back: thrash (2026-11-16 02:44Z)

The loop thrashed. `Save and Exit requested` lines at 02:16, 02:23, 02:27 and 02:34 (about every 7 minutes, one fresh game each), then a burst of 5 leave/re-enter cycles in 20 seconds at 02:42:31 to 02:42:47 (area 8, explored 100 percent, town, and straight out again). The user reported the bot flaking out. Rolled back the same way it was set: current file saved as `tuning.json.bak-thrash`, `tuning.json.bak-farmloopon` copied over `tuning.json` (no `bot.farmReset` entry, default 0), `docker restart d2play-20261108`. V after 3.5 minutes: healthy, no `Save and Exit requested` apart from the restart itself, 0 error lines, hero alive, clvl 5 to 6, xp 13996 -> 14275, 0 deaths. Cause is for the farm loop card: the guards (cooldown 5 minutes, "no fresh game within 180 s") did not stop the burst at 100 percent coverage (U which guard failed).
