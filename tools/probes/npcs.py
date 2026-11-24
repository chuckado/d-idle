#!/usr/bin/env python3
"""npcs: the `npcs` listing of a town after `debug goto` (card qdmatrix, Act 3 Meshif). Usage: npcs.py D2D DATA LEVEL SEED.
`debug goto` is a SETUP verb (cheated probe). Stdlib only."""
import os, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session
p, cmd = session(sys.argv[1], sys.argv[2], int(sys.argv[4]))
try:
    print(cmd("debug goto %s" % sys.argv[3]).strip()[:60])
    time.sleep(3)
    print(cmd("npcs")[:3000])
finally:
    p.kill()
