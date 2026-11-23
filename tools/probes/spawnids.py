#!/usr/bin/env python3
"""spawnids: which monsters.txt ids does `debug spawn` accept, and what does it make? (card stepgrounds)
Usage: spawnids.py D2D DATA AREA ID[,ID..]. SETUP verbs (goto, spawn): CHEATED mechanics probe. Stdlib only."""
import sys, os, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session
d2d, data, area, ids = sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4].split(",")
p, cmd = session(d2d, data, 11)
try:
    print(cmd("debug goto " + area).strip()[:40]); time.sleep(3)
    for i in ids:
        print(i, "->", cmd("debug spawn %s 1" % i).strip()[:110])
finally:
    p.kill()
