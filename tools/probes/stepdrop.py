#!/usr/bin/env python3
"""stepdrop: does a monster made by `debug spawn` and killed by `debug kill` leave its quest drop on the ground? (card stepgrounds)
Env QSET="3:0x000c" sets quest words first.
Usage: stepdrop.py D2D DATA SEED AREA MONSTER [MOD]   (lab d2d, --lab-verbs; goto, spawn and kill are SETUP verbs: a CHEATED
mechanics probe, never evidence about play). Prints the verb list, the spawn reply, the monster rows and the ground rows
before and after the kill, plus the `ground` reply the bot's reader sees. Stdlib only."""
import sys, time, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session
d2d, data, seed, area, mon = sys.argv[1], sys.argv[2], int(sys.argv[3]), int(sys.argv[4]), sys.argv[5]
mod = sys.argv[6] if len(sys.argv) > 6 else ""
p, cmd = session(d2d, data, seed)
def show(tag):
    # only the rows that matter: the spawned class in `monsters`, the quest codes and a count in `ground`
    ms = [l for l in cmd("monsters").split("\n") if mon.lower() in l.lower()]
    gr = [l for l in cmd("ground").split("\n") if l and not l.startswith("ok")]
    q = [l for l in gr if l.split("\t")[0] in ("hdm", "bks", "bkd", "leg")]
    print(tag, "| monsters of", mon, ":", ms[:3], "| ground rows:", len(gr), "| quest rows:", q)

try:
    for q in filter(None, os.environ.get("QSET", "").split(",")):  # "3:0x000c": SETUP quest words before the spawn
        r, w = q.split(":"); print("quest set", r, w, cmd("debug quest set %s %s" % (r, w)).strip()[:60])
    print("goto", cmd("debug goto %d" % area).strip()[:80]); time.sleep(3)
    print("spawn", cmd(("debug spawn %s 1 %s" % (mon, mod)).strip()).strip()[:200]); time.sleep(0.5)
    show("before")
    print("kill", cmd("debug kill").strip()[:200])
    for _ in range(4):
        cmd("step 25")
    show("after")
finally:
    p.kill()
