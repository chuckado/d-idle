#!/usr/bin/env python3
"""gate: state of the Harrogath gate (objects.txt class 449) before and after `cmd interact` (card qdmatrix, Siege of Harrogath).
Usage: gate.py D2D DATA [SEED]. Lists `debug objectlist` rows of class 449, interacts with the nearest by its row index, steps in
lockstep and lists again. `debug goto`, `debug rooms up` are SETUP verbs (cheated probe). Stdlib only."""
import os, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session, grid
p, cmd = session(sys.argv[1], sys.argv[2], int(sys.argv[3]) if len(sys.argv) > 3 else 11)
def around(tag):
    rows = grid(cmd)
    print(tag, 'grid around (23,94), x 15..31, y 88..100')
    for y in range(88, 101):
        print('  %3d %s' % (y, rows[y][15:32]))

def gates():
    return [l for l in cmd("debug objectlist").split("\n") if "\t449\t" in l or "gate" in l.lower()]
try:
    print(cmd("debug goto 109").strip()[:50]); time.sleep(3)
    print(cmd("debug rooms up").strip()[:80])
    print(cmd("lockstep on").strip()[:40])
    around("before")
    rows = gates()
    print("before", *rows, sep="\n  ")
    for l in rows:
        f = l.split("\t")
        if f[1] == "449":
            print("interact", f[0], cmd("cmd interact %s" % f[0]).strip()[:60])
            for i in range(40):
                cmd("step 50")
            print("after 2000 ticks", *gates(), sep="\n  ")
            around("after")
            break
    print(cmd("debug position")[:200])
finally:
    p.kill()
