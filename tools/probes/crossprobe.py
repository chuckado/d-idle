#!/usr/bin/env python3
"""crossprobe: which cells just across a level edge are walkable, by where the hero stands (card toolstrade50).
Usage: crossprobe.py D2D DATA LEVEL SEED SIDE(N|S|W|E) HEROPOSITIONS(tiles x,y;..) [WAIT]. Per hero position: `debug goto LEVEL`, `debug warp x y` (tiles),
wait, then `debug blocked` one tile and 0.1 tile across the edge along the whole side. Prints '.' open and '#' blocked per tile along the side.
`debug goto` and `debug warp` are SETUP verbs (cheated probe). Stdlib only."""
import os, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session
d2d, data, level, seed, side, spots = sys.argv[1], sys.argv[2], sys.argv[3], int(sys.argv[4]), sys.argv[5], sys.argv[6]
wait = float(sys.argv[7]) if len(sys.argv) > 7 else 4
p, cmd = session(d2d, data, seed)
try:
    lv = cmd("debug level").strip()
    print("level", lv[:80])
    w = int(float(lv.split()[1])) if len(lv.split()) > 1 else 18
    for sp in spots.split(";"):
        x, y = sp.split(",")
        cmd("debug goto %s" % level); time.sleep(2)
        cmd("debug warp %s %s" % (x, y)); time.sleep(wait)
        row = ""
        for i in range(0, 18):
            r = cmd("debug blocked %.2f %.2f" % (i + 0.5, -0.3)).strip()
            row += "." if r.startswith("ok 0") or r.split()[:2] == ["ok", "0"] or r.endswith(" 0") else ("#" if r.startswith("ok") else "?")
        print("hero", sp, "across the %s edge:" % side, row, "sample", cmd("debug blocked 5.5 -0.3").strip()[:40])
finally:
    p.kill()
