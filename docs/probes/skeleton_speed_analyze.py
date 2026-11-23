#!/usr/bin/env python3
"""Analyse trace_*.csv from skeleton_speed_probe.py: speeds in cells/s and subtiles/tick, snaps, distances. Usage: analyze.py trace.csv [foes.csv]"""
import csv, collections, math, statistics as st, sys
rows = list(csv.DictReader(open(sys.argv[1])))
foes = {}
if len(sys.argv) > 2:
    for r in csv.DictReader(open(sys.argv[2])):
        foes[(r["phase"], int(r["tick"]))] = int(r["n_live"])
W = 5   # window in ticks
by = collections.OrderedDict()
for r in rows:
    by.setdefault(r["phase"], []).append(r)
print("phase | who | n | moving median c/s | max-window c/s | sub/tick | snaps(>1.5 c/tick) | mean dist to hero (cells)")
for ph, rs in by.items():
    series = collections.defaultdict(list)
    for r in rs:
        t = int(r["tick"])
        series["hero"].append((t, float(r["hero_x"]), float(r["hero_y"]), "-", float(r["hero_x"]), float(r["hero_y"])))
        if r["minion_id"] not in ("", "0"):
            series[r["minion_id"]].append((t, float(r["mx"]), float(r["my"]), r["mode"], float(r["hero_x"]), float(r["hero_y"])))
    for k, s in series.items():
        if k == "hero":
            s = list({x[0]: x for x in s}.values())
        vs, snaps = [], 0
        for i in range(len(s) - W):
            a, b = s[i], s[i + W]
            d = math.dist(a[1:3], b[1:3]); vs.append(d / (b[0] - a[0]))
        for i in range(1, len(s)):
            if math.dist(s[i - 1][1:3], s[i][1:3]) / max(1, s[i][0] - s[i - 1][0]) > 1.5: snaps += 1
        mv = [v for v in vs if v > 0.01]
        dist = st.mean(math.dist(x[1:3], x[4:6]) for x in s) if k != "hero" else 0
        if mv:
            print(f"{ph:18s} | {k:5s} | {len(s):4d} | {st.median(mv)*25:5.2f} | {max(vs)*25:5.2f} | {st.median(mv)*5:.4f} | {snaps} | {dist:.1f}")
        else:
            print(f"{ph:18s} | {k:5s} | {len(s):4d} | idle | | | {snaps} | {dist:.1f}")
