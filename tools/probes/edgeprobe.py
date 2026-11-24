#!/usr/bin/env python3
"""edgeprobe: do the top rows of a level open over time when the hero stands near the edge? (card toolstrade50)
Usage: edgeprobe.py D2D DATA LEVEL SEED X,Y(tiles) [SECONDS]. `debug goto`, `debug warp`, then every 3 s the first 6 rows of `debug grid 5`
around the hero's column ('.' walkable '#' blocked), `debug nearby` and the hero position. SETUP verbs (cheated probe). Stdlib only."""
import os, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session, grid
d2d, data, level, seed, spot = sys.argv[1], sys.argv[2], sys.argv[3], int(sys.argv[4]), sys.argv[5]
secs = float(sys.argv[6]) if len(sys.argv) > 6 else 30
p, cmd = session(d2d, data, seed)
try:
    cmd("debug goto %s" % level); time.sleep(1)
    x, y = spot.split(",")
    cmd("debug warp %s %s" % (x, y))
    t0 = time.time()
    while time.time() - t0 < secs:
        rows = grid(cmd)
        cx = int(float(x) * 5)
        print("t=%4.1f" % (time.time() - t0), "nearby", cmd("debug nearby").strip().replace("\n", "|")[:40], "top rows around x", cx)
        for r in rows[:6]:
            print("   ", r[max(0, cx - 12):cx + 13])
        time.sleep(3)
finally:
    p.kill()
