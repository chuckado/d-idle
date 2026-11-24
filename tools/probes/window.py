#!/usr/bin/env python3
"""window: print the walkable grid around a cell and the object rows near it (card qdmatrix, Kurast Bazaar to Ruined Temple).
Usage: window.py D2D DATA LEVEL SEED CX CY R. One char per subtile cell ('.' walkable, '#' blocked, 'X' the cell itself).
`debug goto` is a SETUP verb (cheated probe). Stdlib only."""
import os, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session, grid

d2d, data, level, seed, cx, cy, r = sys.argv[1], sys.argv[2], int(sys.argv[3]), int(sys.argv[4]), int(sys.argv[5]), int(sys.argv[6]), int(sys.argv[7])
p, cmd = session(d2d, data, seed)
try:
    print(cmd("debug goto %d" % level).strip()[:60])  # the grid is level-wide and level-local; no warp needed
    time.sleep(3)
    rows = grid(cmd)
    for y in range(max(0, cy - r), min(len(rows), cy + r + 1)):
        line = "".join("X" if (x, y) == (cx, cy) else rows[y][x] for x in range(max(0, cx - r), min(len(rows[0]), cx + r + 1)))
        print("%4d %s" % (y, line))
    for l in cmd("debug objectlist").split("\n"):
        f = l.split("\t")
        if len(f) > 3 and not l.startswith(("ok", "err", "room")):
            try:
                if abs(float(f[2]) - cx) <= r + 10 and abs(float(f[3]) - cy) <= r + 10:
                    print(l[:200])
            except ValueError:
                pass
finally:
    p.kill()
