#!/usr/bin/env python3
"""stepdrop_real: does the REAL (level-placed) superunique leave its quest drop after `debug kill`? (card stepgrounds)
Usage: stepdrop_real.py D2D DATA SEED AREA NAME   (lab d2d, --lab-verbs; goto, warp, kill are SETUP verbs: CHEATED mechanics
probe). Warps over a lattice until `monsters` lists a unit whose name contains NAME, then kills the level and prints the
ground rows with a quest code (hdm bks bkd leg) and the row count. Env QSET="3:0x000c" sets quest words first. Stdlib only."""
import sys, time, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session
d2d, data, seed, area, name = sys.argv[1], sys.argv[2], int(sys.argv[3]), int(sys.argv[4]), sys.argv[5]
p, cmd = session(d2d, data, seed)
def find():
    return [l for l in cmd("monsters").split("\n") if name.lower() in l.lower()]
try:
    for q in filter(None, os.environ.get("QSET", "").split(",")):
        r, w = q.split(":"); print("quest set", r, w, cmd("debug quest set %s %s" % (r, w)).strip()[:60])
    print("goto", cmd("debug goto %d" % area).strip()[:80]); time.sleep(3)
    lv = cmd("debug level").split(); w, h = int(float(lv[4])) * 5, int(float(lv[5])) * 5
    hit = find()
    for y in range(20, h, 25):
        for x in range(20, w, 25):
            if hit:
                break
            if cmd("debug warp %d %d" % (x, y)).startswith("err"):
                continue
            time.sleep(0.25)
            hit = find()
    print("found", hit[:2])
    if hit:
        print("kill", cmd("debug kill").strip()[:80])
        for _ in range(4):
            cmd("step 25")
        gr = [l for l in cmd("ground").split("\n") if l and not l.startswith("ok")]
        print("ground rows", len(gr), "quest rows", [l for l in gr if l.split("\t")[0] in ("hdm", "bks", "bkd", "leg")])
        print("monster after", [l for l in find()][:2])
finally:
    p.kill()
