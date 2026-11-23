#!/usr/bin/env python3
"""presets: the `debug presets` rows of a level that match a class or name (card qdmatrix, chests the board cannot find).
Usage: presets.py D2D DATA LEVEL SEED PATTERN. Lab d2d, `debug goto` is a SETUP verb (cheated probe). Stdlib only."""
import os, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session

d2d, data, level, seed, pat = sys.argv[1], sys.argv[2], int(sys.argv[3]), int(sys.argv[4]), sys.argv[5]
p, cmd = session(d2d, data, seed)
try:
    for qs in os.environ.get("QSET", "").split(","):  # "10:0x000c,9:1": quest words set before the goto (SETUP)
        if qs:
            r, v = qs.split(":")
            print("quest", r, v, cmd("debug quest set %s %s" % (r, v)).strip()[:50])
    print(cmd("debug goto %d" % level).strip()[:60])
    time.sleep(3)
    lv = cmd("debug level").split()
    w, h = int(float(lv[4])) * 5, int(float(lv[5])) * 5
    print("level", lv[:6])
    seen = {}
    # the preset list streams in with the rooms near the hero: warp over a lattice (SETUP) and keep every row once
    for y in range(20, h, 60):
        for x in range(20, w, 60):
            if cmd("debug warp %d %d" % (x, y)).startswith("err"):
                continue
            time.sleep(0.15)
            for l in cmd("debug presets").split("\n"):
                if l and not l.startswith(("ok", "err")):
                    seen[l] = True
    print("distinct rows", len(seen))
    for idx in [v for v in os.environ.get("OPEN", "").split(",") if v]:  # "2,3": open these preset indexes and print the ground (SETUP)
        row = [l for l in seen if l.split("\t")[0] == idx]
        if row:
            f = row[0].split("\t")
            cmd("debug warp %d %d" % (int(float(f[2]) * 5) + 2, int(float(f[3]) * 5) + 3))
        time.sleep(1)
        print("interact", idx, cmd("cmd interact %s" % idx).strip()[:60])
        time.sleep(4)
        print("ground", [l[:100] for l in cmd("ground").split("\n") if l][:6])
    for l in seen:
        if pat in l:
            print(l[:200])
finally:
    p.kill()
