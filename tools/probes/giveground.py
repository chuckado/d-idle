#!/usr/bin/env python3
"""giveground: does `debug give <code> <q>` put the item on the ground and does `ground` list it? (card stepgrounds)
Usage: giveground.py D2D DATA AREA CODE. SETUP verbs, CHEATED mechanics probe. Stdlib only."""
import sys, os, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session
d2d, data, area, code = sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4]
p, cmd = session(d2d, data, 11)
try:
    print(cmd("debug goto " + area).strip()[:40]); time.sleep(3)
    for g in ("debug give %s 2" % code, "debug give %s 3" % code, "debug give %s pack" % code):
        print(g, "->", cmd(g).strip()[:100])
    print("ground:", [l for l in cmd("ground").split("\n") if l.startswith(code)])
    print("items:", [l for l in cmd("items").split("\n") if code in l])
finally:
    p.kill()
