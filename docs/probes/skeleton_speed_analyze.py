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

# --- pathing checks: stuck while moving, reversals, snaps with their context, leash
print("\npathing: who | phase | stuck runs (moving mode, <0.005 c/tick, >=10 ticks) | direction reversals (>120 deg between 5-tick headings, moving) | snaps: tick hero_dist_before -> after | max dist to hero")
allm = collections.defaultdict(list)
for r in rows:
    if r["minion_id"] not in ("", "0"):
        allm[r["minion_id"]].append((int(r["tick"]), float(r["mx"]), float(r["my"]), r["mode"], float(r["hero_x"]), float(r["hero_y"]), r["phase"]))
for k, s in allm.items():
    stuck = 0; run = 0; rev = 0; snaps = []; mx = 0
    last_head = None
    for i in range(1, len(s)):
        t0, x0, y0, m0, *_ = s[i - 1]; t, x, y, m, hx, hy, ph = s[i]
        d = math.dist((x0, y0), (x, y)); mx = max(mx, math.dist((x, y), (hx, hy)))
        if m in ("WL", "RN") and d < 0.005: run += 1
        else:
            if run >= 10: stuck += 1
            run = 0
        if d > 1.5:
            snaps.append(f"{t}:{math.dist((x0,y0),(s[i-1][4],s[i-1][5])):.1f}->{math.dist((x,y),(hx,hy)):.1f}")
        if i >= 5:
            a, b = s[i - 5], s[i]
            if math.dist(a[1:3], b[1:3]) > 0.3:
                head = math.atan2(b[2] - a[2], b[1] - a[1])
                if last_head is not None and abs((head - last_head + math.pi) % (2 * math.pi) - math.pi) > 2.09: rev += 1
                last_head = head
    print(f"minion {k}: stuck_runs={stuck} reversals={rev} snaps={snaps} max_dist_to_hero={mx:.1f}")

print("\nroute length per leg (cells travelled; net = start->end distance of the same minion)")
legs = collections.OrderedDict()
for r in rows:
    legs.setdefault(r["phase"], []).append(r)
for ph, rs in legs.items():
    if ph.endswith("_stop") or ph.startswith("A_"): continue
    trk = collections.defaultdict(list)
    for r in rs:
        trk["hero"].append((int(r["tick"]), float(r["hero_x"]), float(r["hero_y"])))
        if r["minion_id"] not in ("", "0"): trk[r["minion_id"]].append((int(r["tick"]), float(r["mx"]), float(r["my"])))
    out = []
    for k, s in trk.items():
        s = list({x[0]: x for x in s}.values())
        ln = sum(math.dist(s[i - 1][1:], s[i][1:]) for i in range(1, len(s)) if math.dist(s[i - 1][1:], s[i][1:]) <= 1.5)
        out.append(f"{k}: {ln:.1f} (net {math.dist(s[0][1:], s[-1][1:]):.1f})")
    print(ph, " | ".join(out))
