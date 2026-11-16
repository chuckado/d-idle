# Recipe: kills per minute against coverage (farm loop decay curve)

> status: current. Tool `tools/farmloop/measure.py` (stdlib Python, reads stdin). Tags: V = ran once (docs/probes/farmloop-on-20261115.md), U = unverified.

1. On the worker host: `docker logs --since 6h d2play-20261108 2>&1 > clears.log` (use the window since the last clears; a fresh hero or a bot restart resets the cumulative counters, the script treats a drop as a reset and skips that minute).
2. `python3 -I tools/farmloop/measure.py --bin 10 < clears.log` prints, per area and coverage bin, the minutes, mean kills per minute and mean XP per minute from the `minute` lines (`explored`, `kills`, `xp`, `area`), then every `farm reset:` line (zone, coverage, rate against fresh, overhead).
3. Compare the high-coverage bins with the low ones: the ratio is the decay the farm loop's value rule assumes (docs/farmloop.md section 3). A bin needs at least 10 minutes before it means anything. `explored` is -1 outside a roam zone and those minutes are skipped.
4. Rerun after a few clears with the loop on; reset lines appear when the loop triggers.
