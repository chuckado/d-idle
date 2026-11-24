#!/usr/bin/env python3
"""pads: `debug pads` of a level (devp26 lab verb, card qdmatrix): each pad up with its target and landing spot.
Usage: pads.py D2D DATA [LEVEL [SEED]]. `debug goto` is a SETUP verb (cheated probe). Stdlib only."""
import os, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session
level = sys.argv[3] if len(sys.argv) > 3 else "74"
p, cmd = session(sys.argv[1], sys.argv[2], int(sys.argv[4]) if len(sys.argv) > 4 else 11)
try:
    print(cmd("debug goto %s" % level).strip()[:40]); time.sleep(3)
    print(cmd("debug pads")[:3000])
    print(cmd("debug position")[:300])
finally:
    p.kill()
