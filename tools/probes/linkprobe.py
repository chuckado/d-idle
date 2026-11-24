#!/usr/bin/env python3
"""linkprobe: when does d2d link a neighbour level? (card toolstrade50, Inner Cloister to Cathedral, live 2026-11-24)
Usage: linkprobe.py D2D DATA LEVEL SEED X,Y[;X,Y...] [SECONDS]. Per position: `debug goto LEVEL`, `debug warp X Y`, then poll `debug nearby` every
0.5 s and print the first time a level appears (or none within SECONDS), with `debug position`. `debug goto` and `debug warp` are SETUP
verbs (cheated probe). Stdlib only."""
import os, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session
d2d, data, level, seed, spots = sys.argv[1], sys.argv[2], sys.argv[3], int(sys.argv[4]), sys.argv[5]
secs = float(sys.argv[6]) if len(sys.argv) > 6 else 20
p, cmd = session(d2d, data, seed)
try:
    for sp in spots.split(";"):
        x, y = sp.split(",")
        cmd("debug goto %s" % level); time.sleep(2)
        w = cmd("debug warp %s %s" % (x, y)).strip()[:40]
        t0 = time.time(); found = None
        while time.time() - t0 < secs:
            near = cmd("debug nearby").strip()
            if "\t" in near or any(c.isdigit() for c in near.replace("ok", "")):
                found = (round(time.time() - t0, 1), near.replace("\n", " | ")[:120]); break
            time.sleep(0.5)
        print("spot", sp, "warp", w, "->", found or "no link in %ds" % secs, "pos", cmd("debug position").strip()[:60])
finally:
    p.kill()
