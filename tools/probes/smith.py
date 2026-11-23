#!/usr/bin/env python3
"""smith: is The Smith in the Barracks of the lab d2d, and does the Tools of the Trade quest state decide it? (card toolstrade6)
Usage: smith.py D2D DATA SEED   (lab d2d, --headless --lab-verbs; `debug goto` is SETUP: cheated mechanics probe)
Prints the verb list, then at the Barracks the monsters and presets the verbs offer, before and after a quest-word poke
when the verb list has one. Stdlib only."""
import sys, time
sys.path.insert(0, __file__.rsplit("/", 1)[0])
from gridwhole import session
d2d, data, seed = sys.argv[1], sys.argv[2], int(sys.argv[3])
p, cmd = session(d2d, data, seed)
try:
    verbs = cmd("debug verbs")
    print("## verbs"); print(verbs.strip()[:3000])
    print("goto", cmd("debug goto 28").strip()[:80]); time.sleep(3)
    for v in ("debug level", "debug presets", "debug objects", "debug monsters", "debug units", "debug nearby"):
        print("##", v); print(cmd(v).strip()[:2500])
finally:
    p.kill()
