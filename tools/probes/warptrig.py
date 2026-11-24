#!/usr/bin/env python3
"""warptrig: `debug warps` and `debug warptriggers` of a level (card qdmatrix, Kurast Bazaar to Ruined Temple).
Usage: warptrig.py D2D DATA [LEVEL [SEED]]. `debug goto` is a SETUP verb (cheated probe). Stdlib only."""
import os, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session
level = sys.argv[3] if len(sys.argv) > 3 else "80"
p, cmd = session(sys.argv[1], sys.argv[2], int(sys.argv[4]) if len(sys.argv) > 4 else 11)
try:
    print(cmd("debug goto %s" % level).strip()[:40]); time.sleep(3)
    for c in ("debug warps", "debug warptriggers"):
        print(c); print(cmd(c)[:6000])
finally:
    p.kill()
