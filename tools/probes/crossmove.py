#!/usr/bin/env python3
"""crossmove: does a move click beyond a level edge change the level? (card toolstrade50, Inner Cloister to Cathedral)
Usage: crossmove.py D2D DATA LEVEL SEED X,Y(tiles) TX,TY(tiles) [SECONDS]. `debug goto`, `debug warp X Y`, `cmd move TX TY`, then every 0.5 s `debug position`.
SETUP verbs (cheated probe). Stdlib only."""
import os, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session
d2d, data, level, seed, a, b = sys.argv[1], sys.argv[2], sys.argv[3], int(sys.argv[4]), sys.argv[5], sys.argv[6]
secs = float(sys.argv[7]) if len(sys.argv) > 7 else 8
p, cmd = session(d2d, data, seed)
try:
    cmd("debug goto %s" % level); time.sleep(2)
    cmd("debug warp %s" % a.replace(",", " ")); time.sleep(1)
    print("move", cmd("cmd move %s" % b.replace(",", " ")).strip()[:50])
    t0 = time.time()
    while time.time() - t0 < secs:
        print("t=%.1f" % (time.time() - t0), cmd("debug position").strip()[:90])
        time.sleep(1)
finally:
    p.kill()
