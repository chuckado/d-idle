#!/usr/bin/env python3
"""Read-only soak-log classifier for "nav waypoint failed" (d2lod-bot, card decompwp): per line the area of the last
status/minute line, the target and the err class; also counts "wp: used A -> B" successes. Usage: wp_soaklines.py LOG..."""
import sys, json, re, glob, collections
fail = collections.Counter(); used = collections.Counter(); since = collections.Counter(); opened = collections.Counter()
for path in sys.argv[1:]:
    area = None; changed = None; lastwpinteract = None
    with open(path, errors="replace") as f:
        for line in f:
            if '"msg"' not in line: continue
            try: d = json.loads(line)
            except Exception: continue
            m = d.get("msg", ""); t = d.get("time", "")[11:19]
            if m in ("status", "minute") and "area" in d:
                if d["area"] != area: area, changed = d["area"], t
            elif m.startswith("nav waypoint failed"):
                err = re.sub(r"\d+", "N", d.get("err", ""))
                fail[(path.split("/")[-3] if "/logs/" in path else path, area, d.get("to"), err)] += 1
            elif m.startswith("wp: used"):
                mm = re.match(r"wp: used (\d+) -> (\d+)", m); used[(mm.group(1), mm.group(2))] += 1
            elif m == "nav waypoint interact":
                opened[d.get("err", "")[:60]] += 1
for k, v in sorted(fail.items(), key=lambda x: -x[1]): print("FAIL", v, *k)
for k, v in sorted(used.items(), key=lambda x: -x[1])[:20]: print("USED", v, *k)
for k, v in opened.most_common(8): print("INTERACT", v, k)
