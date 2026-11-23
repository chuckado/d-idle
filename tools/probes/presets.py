#!/usr/bin/env python3
"""presets: the `debug presets` rows of a level that match a class or name (card qdmatrix, chests the board cannot find).
Usage: presets.py D2D DATA LEVEL SEED PATTERN. Lab d2d, `debug goto` is a SETUP verb (cheated probe). Stdlib only."""
import os, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session

d2d, data, level, seed, pat = sys.argv[1], sys.argv[2], int(sys.argv[3]), int(sys.argv[4]), sys.argv[5]
p, cmd = session(d2d, data, seed)
try:
    print(cmd("debug goto %d" % level).strip()[:60])
    time.sleep(3)
    rows = [l for l in cmd("debug presets").split("\n") if l and not l.startswith(("ok", "err"))]
    print("rows", len(rows), "header-ish:", rows[:2])
    for l in rows:
        if pat in l:
            print(l[:160])
finally:
    p.kill()
