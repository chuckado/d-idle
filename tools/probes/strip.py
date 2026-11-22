#!/usr/bin/env python3
"""strip: the walkable cells of the border strip between two levels, as the bot's d2dworld sees them (card mapverify).

Usage: strip.py D2D DATA FROM TO SEED [WAIT_S]   (lab d2d, --headless --lab-verbs; SETUP verb `debug goto`: cheated mechanics probe)
Reads `debug layout` / `nearby` / `level` / `grid 5` at FROM after WAIT_S seconds, finds the shared edge with TO, and per tile
(5 subtiles) prints: inside walkable (grid), and whether `debug blocked` just past the edge answers open (what
internal/d2dworld border.go `across` asks). Prints raw replies of the verbs first so the layout is visible. Stdlib only."""
import os, sys, time
sys.path.insert(0, __file__.rsplit("/", 1)[0])
from gridwhole import session, grid

d2d, data, fr, to, seed = sys.argv[1], sys.argv[2], int(sys.argv[3]), int(sys.argv[4]), int(sys.argv[5])
wait = float(sys.argv[6]) if len(sys.argv) > 6 else 1.5
p, cmd = session(d2d, data, seed)
TARGET_SIZE = (80, 40)
try:
    print("goto target first (its size)", cmd("debug goto %d" % to).strip()[:60])
    time.sleep(wait)
    tl = cmd("debug level").split()
    print("target level", tl[:8], "layout lines", len(cmd("debug layout").strip().split("\n")) - 1)
    TARGET_SIZE = (int(tl[4]), int(tl[5]))   # cells = tiles
    print("goto", cmd("debug goto %d" % fr).strip()[:60])
    time.sleep(wait)
    for v in ("debug level", "debug layout", "debug nearby"):
        print("##", v); print(cmd(v).strip()[:1500])
    lv = cmd("debug level").split()
    print("level fields", lv[:10])
    g = grid(cmd)
    print("grid rows", len(g), "cols", len(g[0]) if g else 0)
    lay = {}
    t0 = time.time()
    while True:   # the layout lists levels as d2d builds them: poll until the target shows (or 40 s)
        for l in cmd("debug layout").split("\n"):
            f = l.split()
            if len(f) >= 5 and f[0].isdigit():
                lay[int(f[0])] = tuple(int(x) for x in f[1:5])
        if to in lay or time.time() - t0 > 10:
            break
        time.sleep(1)
    if fr not in lay:   # the level itself is not in the layout either (27, 32): its own `debug level` gives origin and size
        lv2 = cmd("debug level").split()
        lay[fr] = (int(lv2[6]), int(lv2[7]), int(lv2[4]), int(lv2[5]))
    if to not in lay:   # linked (nearby) but not placed by the layout verb: derive the origin from nearby, size from the visit above
        for l in cmd("debug nearby").split("\n"):
            f = l.split()
            if len(f) == 3 and f[0] == str(to) and fr in lay:
                tv = TARGET_SIZE
                lay[to] = (lay[fr][0] + int(f[1]), lay[fr][1] + int(f[2]), tv[0], tv[1])
                print("target synthesised from nearby:", lay[to])
    print("layout has target after %.1f s: %s" % (time.time() - t0, to in lay))
    print("## nearby late"); print(cmd("debug nearby").strip()[:600])
    o, t = lay.get(fr), lay.get(to)
    if os.environ.get("STRIP_FINE"):   # objects near the edge and the engine's own collision at fine steps across the strip
        for v in ("debug warps", "debug warptriggers"):
            print("##", v); print(cmd(v).strip()[:1500])
        print("## objects"); print("\n".join(cmd("debug objects").split("\n")[:60])[:3000])
        for yy in (-0.5, -0.1, 0.1, 0.3, 0.5, 0.7, 0.9, 1.1, 1.5, 2.5):
            print("blocked row %5.1f: %s" % (yy, "".join("#" if cmd("debug blocked %.2f %.2f" % (xx / 10, yy)).split()[1] == "1" else "." for xx in range(264, 300, 1))))
    print("own", o, "target", t)
    if o and t:
        side = None
        if t[1] == o[1] + o[3]: side = "S"   # neighbour below
        elif t[0] + t[2] == o[0]: side = "W"
        elif t[1] + t[3] == o[1]: side = "N"
        elif t[0] == o[0] + o[2]: side = "E"
        print("shared edge side", side)
        # tiles along the edge, in own-relative tile coordinates; grid is 5 subtiles per tile from the level origin
        if side in ("S", "N"):   # the first 12 grid rows next to the edge (or last 12), only the columns that have an open cell
            rows = range(0, 12) if side == "N" else range(len(g) - 12, len(g))
            cols = [x for x in range(len(g[0])) if any(g[y][x] == "." for y in rows)]
            if cols:
                print("window rows", rows[0], "-", rows[-1], "cols", cols[0], "-", cols[-1])
                for y in rows:
                    print("  %3d %s" % (y, "".join(g[y][cols[0]:cols[-1] + 1])[:150]))
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
