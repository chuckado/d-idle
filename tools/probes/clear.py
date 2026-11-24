#!/usr/bin/env python3
"""clear: is a goal connected to a start through corridors of a given width (card qdmatrix, Arcane Sanctuary)? Erodes the
walkable grid of `debug grid` by R cells and flood-fills 4- and 8-connected. Usage: clear.py D2D DATA LEVEL SEED X,Y X,Y [R..].
`debug goto` is a SETUP verb (cheated probe). Stdlib only."""
import os, sys, time
from collections import deque
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session, grid
p, cmd = session(sys.argv[1], sys.argv[2], int(sys.argv[4]))
def conn(rows, a, b, r, diag):
    h, w = len(rows), len(rows[0])
    ok = [[rows[y][x] == "." for x in range(w)] for y in range(h)]
    for _ in range(r):
        ok = [[ok[y][x] and all(0 <= y + dy < h and 0 <= x + dx < w and ok[y + dy][x + dx] for dy in (-1, 0, 1) for dx in (-1, 0, 1)) for x in range(w)] for y in range(h)]
    def snap(pt):
        x, y = pt
        for d in range(0, 14):
            for dy in range(-d, d + 1):
                for dx in range(-d, d + 1):
                    if 0 <= y + dy < h and 0 <= x + dx < w and ok[y + dy][x + dx]:
                        return (x + dx, y + dy)
    sa, sb = snap(a), snap(b)
    if not sa or not sb:
        return "no walkable cell near " + ("start" if not sa else "goal")
    seen = {sa}; q = deque([sa])
    steps = [(1, 0), (-1, 0), (0, 1), (0, -1)] + ([(1, 1), (1, -1), (-1, 1), (-1, -1)] if diag else [])
    while q:
        x, y = q.popleft()
        if (x, y) == sb:
            return "connected"
        for dx, dy in steps:
            n = (x + dx, y + dy)
            if 0 <= n[0] < w and 0 <= n[1] < h and ok[n[1]][n[0]] and n not in seen:
                seen.add(n); q.append(n)
    return "cut (reached %d cells)" % len(seen)
try:
    cmd("debug goto %s" % sys.argv[3]); time.sleep(3)
    rows = grid(cmd)
    a = tuple(int(v) for v in sys.argv[5].split(",")); b = tuple(int(v) for v in sys.argv[6].split(","))
    for r in (int(v) for v in (sys.argv[7:] or ["0", "1", "2"])):
        print("erode", r, "4-conn:", conn(rows, a, b, r, False), "| 8-conn:", conn(rows, a, b, r, True))
finally:
    p.kill()
