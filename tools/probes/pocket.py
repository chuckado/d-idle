#!/usr/bin/env python3
"""pocket: why a gate is "cut off" (card qdmatrix, Halls of the Dead Level 1 to 2). Per (level, seed): lab d2d, `debug goto L`,
read `debug grid 5` (walkable '.' / blocked '#', 5 subtile cells), `debug warps`, flood-fill 8-connected from the hero's cell and
print the component of every tile exit plus the shortest wall thickness between the hero's component and each exit's.
Usage: pocket.py D2D DATA LEVEL SEED. Stdlib only; reuses gridwhole.session."""
import os, sys
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session, grid

def comps(rows):
    h, w = len(rows), len(rows[0])
    lab = [[-1] * w for _ in range(h)]
    n = 0
    for y in range(h):
        for x in range(w):
            if rows[y][x] == "." and lab[y][x] < 0:
                st = [(x, y)]
                lab[y][x] = n
                while st:
                    cx, cy = st.pop()
                    for dx in (-1, 0, 1):
                        for dy in (-1, 0, 1):
                            nx, ny = cx + dx, cy + dy
                            if 0 <= nx < w and 0 <= ny < h and rows[ny][nx] == "." and lab[ny][nx] < 0:
                                lab[ny][nx] = n
                                st.append((nx, ny))
                n += 1
    return lab, n

def main():
    d2d, data, levels, seed = sys.argv[1], sys.argv[2], [int(v) for v in sys.argv[3].split(",")], int(sys.argv[4])
    p, cmd = session(d2d, data, seed)
    try:
        for level in levels:
            one(cmd, level)
    finally:
        p.kill()

def one(cmd, level):
    if True:
        print("== level", level, cmd("debug goto %d" % level).strip()[:80])
        lv = cmd("debug level").split()
        print("level", lv[:6])
        rows = grid(cmd)
        print("grid", len(rows), "x", len(rows[0]) if rows else 0)
                
        warps = [l.split() for l in cmd("debug warps").split("\n") if l and not l.startswith(("ok", "err"))]
        lab, n = comps(rows)
        print("components", n)
        x0, y0 = int(float(lv[2]) * 5), int(float(lv[3]) * 5)  # level origin in tiles -> subtile cells (one grid cell each)
        for f in warps:
            try:
                wx, wy = float(f[1]), float(f[2])
            except (ValueError, IndexError):
                continue
            cx, cy = int(wx * 5 - x0), int(wy * 5 - y0)
            def comp(x, y):
                return lab[y][x] if 0 <= y < len(lab) and 0 <= x < len(lab[0]) else None
            ux, uy = int(float(f[4]) * 5 - x0), int(float(f[5]) * 5 - y0)
            print("warp", f[:8], "cell", (cx, cy), "component", comp(cx, cy), "unit", (ux, uy), "component", comp(ux, uy))
        sizes = {}
        for r in lab:
            for v in r:
                if v >= 0:
                    sizes[v] = sizes.get(v, 0) + 1
        print("component sizes", sorted(sizes.items(), key=lambda kv: -kv[1])[:8])

main()
