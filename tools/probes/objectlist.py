#!/usr/bin/env python3
"""objectlist: every object row of a level, hidden ones included (`debug objectlist`, devp23 labverbs; card qdmatrix).
Usage: objectlist.py D2D DATA LEVEL SEED [FILTER]. Prints rows whose text contains FILTER (default: all rows with operate_fn 41,
class 356 or hidden 1). `debug goto` is a SETUP verb (cheated probe). Stdlib only."""
import os, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session

d2d, data, level, seed = sys.argv[1], sys.argv[2], int(sys.argv[3]), int(sys.argv[4])
flt = sys.argv[5] if len(sys.argv) > 5 else ""
p, cmd = session(d2d, data, seed)
try:
    print(cmd("debug goto %d" % level).strip()[:60])
    time.sleep(3)
    lv = cmd("debug level").split()
    w, h = int(float(lv[4])) * 5, int(float(lv[5])) * 5
    seen = {}
    # rooms stream in near the hero: warp over a lattice (SETUP) and keep every row once, keyed by class and position
    for y in range(20, h, 50):
        for x in range(20, w, 50):
            if cmd("debug warp %d %d" % (x, y)).startswith("err"):
                continue
            time.sleep(0.2)
            for l in cmd("debug objectlist").split("\n"):
                if l and not l.startswith(("ok", "err", "room")):
                    f = l.split("\t")
                    seen[tuple(f[1:4])] = l
    rows = list(seen.values())
    print("rows", len(rows))
    for l in rows:
        f = l.split("\t")
        keep = flt in l if flt else (len(f) > 6 and (f[1] == "356" or f[4] == "41" or f[6] == "1"))
        if keep:
            print(l[:200])
finally:
    p.kill()
