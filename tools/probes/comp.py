#!/usr/bin/env python3
"""comp: connected component of given level-local subtile cells in `debug grid` (card qdmatrix, Arcane Sanctuary pads).
Usage: comp.py D2D DATA LEVEL SEED X,Y [X,Y ...]. Prints each cell's component id and size and the walkable state of the
cell. `debug goto` is a SETUP verb (cheated probe). Stdlib only."""
import os, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session, grid
from pocket import comps
p, cmd = session(sys.argv[1], sys.argv[2], int(sys.argv[4]))
try:
    cmd("debug goto %s" % sys.argv[3]); time.sleep(3)
    rows = grid(cmd)
    lab, n = comps(rows)
    sizes = {}
    for r in lab:
        for v in r:
            sizes[v] = sizes.get(v, 0) + 1
    print("grid", len(rows), "x", len(rows[0]), "components", n)
    for a in sys.argv[5:]:
        x, y = (int(v) for v in a.split(","))
        v = lab[y][x] if 0 <= y < len(lab) and 0 <= x < len(lab[0]) else None
        near = {}
        for dy in range(-6, 7):
            for dx in range(-6, 7):
                yy, xx = y + dy, x + dx
                if 0 <= yy < len(lab) and 0 <= xx < len(lab[0]) and lab[yy][xx] >= 0:
                    near[lab[yy][xx]] = near.get(lab[yy][xx], 0) + 1
        print(a, "cell", rows[y][x] if v is not None else "?", "component", v, "size", sizes.get(v), "near", dict(sorted(near.items(), key=lambda kv: -kv[1])[:3]))
finally:
    p.kill()
