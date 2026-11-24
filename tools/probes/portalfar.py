#!/usr/bin/env python3
"""portalfar: does `cmd interact -2000-which` take a town portal from far away (card qdmatrix, live "portal click did not change
the level" with the hero 70 subtiles from the portal)? Opens a portal in level 2 with a scroll, enters it, warps the hero N cells
from the standing portal in town and sends the interact. Usage: portalfar.py D2D DATA SEED [CELLS ...]. `debug goto`, `debug give`,
`debug warp` are SETUP verbs (cheated probe). Stdlib only."""
import os, re, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session
p, cmd = session(sys.argv[1], sys.argv[2], int(sys.argv[3]))
dists = [float(x) for x in sys.argv[4:]] or [4.5, 8, 14]
def step(n):
    for _ in range(n):
        cmd("step 25")
def level():
    return int(cmd("debug level").split()[1])
def portals():
    return [l.split() for l in cmd("debug portals").split("\n") if l and l[0].isdigit()]
try:
    print("lockstep", cmd("lockstep on").strip()[:30])
    print("goto", cmd("debug goto 2").strip()[:40]); step(4)
    print("give", cmd("debug give tsc pack").strip()[:60])
    items = cmd("items")
    m = re.search(r"\[tsc[^\]]*?id=(\d+)", items) or re.search(r"id=(\d+)[^\]]*tsc", items)
    print("items:", items[:400].replace("\n", " | "))
    if not m:
        sys.exit("no tsc id")
    iid = int(m.group(1))
    print("key i", cmd("key i").strip()[:20]); step(4)
    xy = cmd("debug itemat %d" % iid).split()
    x, y = int(xy[1]), int(xy[2])
    for c in ("move %d %d" % (x - 3, y), "move %d %d" % (x, y), "rclick %d %d" % (x, y)):
        print(c, cmd(c).strip()[:30]); step(6)
    print("key i", cmd("key i").strip()[:20]); step(10)
    print("portals", portals())
    print("enter", cmd("cmd interact -2000").strip()[:40]); step(80)
    print("level", level())
    for d in dists:
        ps = [q for q in portals() if int(q[1]) == level()]
        if not ps:
            print("no portal in this level", portals()); break
        px, py = float(ps[0][2]), float(ps[0][3])
        which = int(ps[0][0])
        print("dist", d, "portal", ps[0], "warp", cmd("debug warp %d %d" % (round(px + d), round(py))).strip()[:30])
        step(6)
        if os.environ.get("KICK"):  # the priming click the d2dctl client sends after a warp (verbs.go Kick)
            print(" kick", cmd("click 400 300").strip()[:20]); step(6)
        n = 0
        print(" pos after warp", cmd("debug position").strip()[:90])
        if os.environ.get("GOPROBE"):  # the bot's own EnterPortal (d2dctl), on a real-time instance
            print(" lockstep off", cmd("lockstep off").strip()[:30])
            import subprocess
            r = subprocess.run([os.environ["GOPROBE"], "-sock", cmd.sock], capture_output=True, text=True, timeout=200)
            print(" GOPROBE", r.returncode, r.stdout.strip(), r.stderr.strip()[:200])
            break
        print(" interact", cmd("cmd interact %d" % (-2000 - which)).strip()[:30])
        step(40)
        print(" pos after 1000 ticks", cmd("debug position").strip()[:90])
        for n in range(1, 61):
            step(4)
            if level() != 1:
                break
        print(" ->", "level", level(), "after", n * 4 * 25, "ticks")
        if level() != 1:
            print(" back in town:", cmd("cmd interact -2000").strip()[:20]); step(80); print(" level", level())
finally:
    p.kill()
