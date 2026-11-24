#!/usr/bin/env python3
"""stepon: does standing on a tile exit's cell change the level in the lab d2d (card qdmatrix, Kurast Bazaar to Ruined Temple)?
Usage: stepon.py D2D DATA LEVEL SEED X Y [DX DY]. `debug goto` and `debug warp` are SETUP verbs (cheated probe): warps next to
(X,Y), then steps with `debug walk`-less clicks are not available, so only the arrival level after the warp is read."""
import os, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session
d2d, data, level, seed, x, y = sys.argv[1], sys.argv[2], int(sys.argv[3]), int(sys.argv[4]), int(sys.argv[5]), int(sys.argv[6])
p, cmd = session(d2d, data, seed)
try:
    r = cmd("debug goto %d" % level).split(); time.sleep(3)
    ox, oy = int(float(r[1]) * 5), int(float(r[2]) * 5)  # level origin (tiles) to absolute subtiles
    print("origin", ox, oy)
    print("lockstep", cmd("lockstep on").strip()[:40])
    print("warp", cmd("debug warp %d %d" % (x + ox - 6, y + oy + 2)).strip()[:40]); time.sleep(1)
    print("move", cmd("cmd move %d %d" % (x + ox, y + oy)).strip()[:60])
    for i in range(8):
        print(" step", cmd("step 25").strip()[:40], "|", cmd("state").replace("\n"," ")[:260], cmd("debug level").split()[1])
finally:
    p.kill()
