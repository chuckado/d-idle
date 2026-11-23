#!/usr/bin/env python3
"""Skeleton speed probe: a Necromancer raises skeletons on a lockstep d2d, the hero walks, we record per tick positions.
Usage: skeleton_speed_probe.py <d2d> <outdir> [seed]. Reads only plus SETUP verbs (--lab-verbs, throwaway char, no save)."""
import json, os, re, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from skeleton_speed_lib import Session
d2d, out = sys.argv[1], sys.argv[2]
seed = int(sys.argv[3]) if len(sys.argv) > 3 else 5
os.makedirs(out, exist_ok=True)
s = Session(d2d, ["--lab-verbs", "--read-combat", "--lockstep", "--no-save", "--toggle", "autoloot=off,autoparty=off"], cls=1, seed=seed)

def hero():
    m = re.search(r"cam=([-\d.]+),([-\d.]+)", s.cmd("state"))
    return float(m.group(1)), float(m.group(2))
def mins():
    return json.loads(s.cmd("debug minions").splitlines()[0])
def step(n=1):
    return s.cmd(f"step {n}")
def log(*a):
    print(*a, flush=True)
try:
    log(s.cmd("debug points 70 20"), s.cmd("debug points 69 5"), s.cmd("debug goto 2"))
    step(100)
    log(s.cmd("debug mana full"))
    log("level", s.cmd("debug level").strip())
    log(s.cmd("spawn fallen 3") if False else s.cmd("debug spawn 0 3"))
    step(25)
    log(s.cmd("monsters"))
    log(s.cmd("debug kill"))
    step(10)
    log("corpses", s.cmd("debug corpses"))
    hx, hy = hero()
    log("hero", hx, hy)
    for tgt in ("near", f"{hx+1} {hy}"):
        log("cast", tgt, s.cmd(f"debug cast 70 {tgt}"))
        step(30)
        log(json.dumps(mins()))
finally:
    s.close()
