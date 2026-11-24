#!/usr/bin/env python3
"""nearobj: the objects (`debug objectlist` and `debug objects`) near a spot of a level (card toolstrade50, Inner Cloister to Cathedral).
Usage: nearobj.py D2D DATA LEVEL SEED X,Y(tiles) [RADIUS_TILES]. `debug goto`, `debug warp X Y`, wait, print rows within the radius. SETUP verbs. Stdlib only."""
import os, sys, time, math
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session
d2d, data, level, seed, a = sys.argv[1], sys.argv[2], sys.argv[3], int(sys.argv[4]), sys.argv[5]
rad = float(sys.argv[6]) if len(sys.argv) > 6 else 6
p, cmd = session(d2d, data, seed)
try:
    cmd("debug goto %s" % level); time.sleep(2)
    cmd("debug warp %s" % a.replace(",", " ")); time.sleep(3)
    x0, y0 = [float(v) for v in a.split(",")]
    for verb in ("debug objectlist", "debug objects"):
        print(verb)
        for l in cmd(verb).split("\n"):
            f = l.split("\t")
            if len(f) > 3:
                try:
                    x, y = float(f[2]), float(f[3])
                except ValueError:
                    continue
                if math.hypot(x - x0, y - y0) <= rad:
                    print("  ", l[:200])
finally:
    p.kill()
