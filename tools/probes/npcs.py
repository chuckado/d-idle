#!/usr/bin/env python3
"""npcs: the `npcs` listing of a town over a lattice of warps (card qdmatrix, Act 3 Meshif): the d2d lists only the NPCs of
the rooms that are up, so the lattice walks the whole town. Usage: npcs.py D2D DATA LEVEL SEED. `debug goto` and `debug warp`
are SETUP verbs (cheated probe). Stdlib only."""
import os, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session
p, cmd = session(sys.argv[1], sys.argv[2], int(sys.argv[4]))
try:
    print(cmd("debug goto %s" % sys.argv[3]).strip()[:60])
    time.sleep(3)
    lv = cmd("debug level").split()
    ox, oy, w, h = (int(float(v) * 5) for v in lv[2:6])
    seen = {}
    for y in range(20, h, 40):
        for x in range(20, w, 40):
            if cmd("debug warp %d %d" % (ox + x, oy + y)).startswith("err"):
                continue
            time.sleep(0.2)
            for l in cmd("npcs").split("\n"):
                f = l.split("\t")
                if len(f) >= 6 and not l.startswith(("ok", "err")):
                    seen[f[0]] = l
    for l in sorted(seen.values()):
        print(l[:120])
finally:
    p.kill()
