#!/usr/bin/env python3
"""Kills and XP per minute against zone coverage, from bot `minute` log lines (docs/probes/farmloop-measure.md).
usage: docker logs --since 6h CONTAINER 2>&1 | python3 -I tools/farmloop/measure.py [--bin 10]
Counters are cumulative per process and reset on a bot restart or a new hero: a drop is treated as a reset (delta 0).
Prints per area and coverage bin: minutes, mean kills/min, mean xp/min; plus the `farm reset:` lines."""
import json, sys, collections

bin_w = 10
if "--bin" in sys.argv:
    bin_w = int(sys.argv[sys.argv.index("--bin") + 1])
prev = None
rows = collections.defaultdict(list)
resets = []
for line in sys.stdin:
    if "farm reset:" in line:
        resets.append(line.strip()[:240])
    if '"msg":"minute"' not in line:
        continue
    try:
        d = json.loads(line[line.index("{"):])
    except ValueError:
        continue
    k, x, a, e = d.get("kills", 0), d.get("xp", 0), d.get("area"), d.get("explored", -1)
    if prev is not None and k >= prev[0] and x >= prev[1] and a == prev[2] and e >= 0:
        rows[(a, min(e, 100) // bin_w * bin_w)].append((k - prev[0], x - prev[1]))
    prev = (k, x, a)
print("area cov%  minutes  kills/min  xp/min")
for (a, b), v in sorted(rows.items()):
    n = len(v)
    print(f"{a:>4} {b:>3}-{b + bin_w:<3} {n:>5}  {sum(i[0] for i in v) / n:9.1f}  {sum(i[1] for i in v) / n:7.0f}")
print(f"farm reset lines: {len(resets)}")
for r in resets:
    print(r)
