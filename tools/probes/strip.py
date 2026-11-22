#!/usr/bin/env python3
"""strip: the walkable cells of the border strip between two levels, as the bot's d2dworld sees them (card mapverify).

Usage: strip.py D2D DATA FROM TO SEED [WAIT_S]   (lab d2d, --headless --lab-verbs; SETUP verb `debug goto`: cheated mechanics probe)
Reads `debug layout` / `nearby` / `level` / `grid 5` at FROM after WAIT_S seconds, finds the shared edge with TO, and per tile
(5 subtiles) prints: inside walkable (grid), and whether `debug blocked` just past the edge answers open (what
internal/d2dworld border.go `across` asks). Prints raw replies of the verbs first so the layout is visible. Stdlib only."""
import sys, time
sys.path.insert(0, __file__.rsplit("/", 1)[0])
from gridwhole import session, grid

d2d, data, fr, to, seed = sys.argv[1], sys.argv[2], int(sys.argv[3]), int(sys.argv[4]), int(sys.argv[5])
wait = float(sys.argv[6]) if len(sys.argv) > 6 else 1.5
p, cmd = session(d2d, data, seed)
try:
    print("goto", cmd("debug goto %d" % fr).strip()[:60])
    time.sleep(wait)
    for v in ("debug level", "debug layout", "debug nearby"):
        print("##", v); print(cmd(v).strip()[:1500])
    lv = cmd("debug level").split()
    print("level fields", lv[:10])
    g = grid(cmd)
    print("grid rows", len(g), "cols", len(g[0]) if g else 0)
    lay = {}
    for l in cmd("debug layout").split("\n"):
        f = l.split()
        if len(f) >= 5 and f[0].isdigit():
            lay[int(f[0])] = tuple(int(x) for x in f[1:5])
    o, t = lay.get(fr), lay.get(to)
    print("own", o, "target", t)
    if o and t:
        side = None
        if t[1] == o[1] + o[3]: side = "S"   # neighbour below
        elif t[0] + t[2] == o[0]: side = "W"
        elif t[1] + t[3] == o[1]: side = "N"
        elif t[0] == o[0] + o[2]: side = "E"
        print("shared edge side", side)
        # tiles along the edge, in own-relative tile coordinates; grid is 5 subtiles per tile from the level origin
        if side in ("S", "N"):
            lo, hi = max(o[0], t[0]) - o[0], min(o[0] + o[2], t[0] + t[2]) - o[0]
            row = (o[3] * 5 - 1) if side == "S" else 0
            for tx in range(lo, hi):
                inside = [g[row][5 * tx + k] if row < len(g) and 5 * tx + k < len(g[0]) else "?" for k in range(5)]
                ty = o[3] if side == "S" else -1
                r = cmd("debug blocked %.1f %.1f" % (tx + 0.5, ty + 0.5)).strip()
                r2 = cmd("debug blocked %.1f %.1f" % (tx + 0.5, (ty - 1 if side == "S" else 0) + 0.5)).strip()
                print("tile x=%d inside=%s edge-row=%s past-edge blocked: %s" % (tx, "".join(inside), r2[:30], r[:30]))
        elif side in ("W", "E"):
            lo, hi = max(o[1], t[1]) - o[1], min(o[1] + o[3], t[1] + t[3]) - o[1]
            col = (o[2] * 5 - 1) if side == "E" else 0
            for ty in range(lo, hi):
                inside = [g[5 * ty + k][col] if 5 * ty + k < len(g) and col < len(g[0]) else "?" for k in range(5)]
                tx = o[2] if side == "E" else -1
                r = cmd("debug blocked %.1f %.1f" % (tx + 0.5, ty + 0.5)).strip()
                r2 = cmd("debug blocked %.1f %.1f" % ((tx - 1 if side == "E" else 0) + 0.5, ty + 0.5)).strip()
                print("tile y=%d inside=%s edge-col=%s past-edge blocked: %s" % (ty, "".join(inside), r2[:30], r[:30]))
finally:
    p.kill()
