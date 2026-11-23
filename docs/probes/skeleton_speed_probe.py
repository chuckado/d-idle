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
rows = []
mons = []
def foes(tick):
    live = []
    for l in s.cmd("monsters").splitlines():
        f = l.split("\t")
        if len(f) > 12 and f[6] not in ("DD", "DT") and not f[5].startswith("0/"):
            live.append((f[10], f[11], f[1], f[2], f[6]))
    return live
def rec(phase, n, every=10):
    for i in range(n):
        if i % 3 == 0:
            s.cmd("debug kill")   # natural monsters (alive foes) off; pets are in hero.pets, not in the level's list
        step(1)
        h = hero(); m = mins()
        rows.append((phase, m["tick"], h, [(k["id"], k["x"], k["y"], k["mode"]) for k in m["minions"]]))
        if i % every == 0:
            mons.append((phase, m["tick"], foes(m["tick"])))
def dump(name):
    with open(os.path.join(out, name), "w") as f:
        f.write("phase,tick,hero_x,hero_y,minion_id,mx,my,mode\n")
        for ph, tk, h, ms in rows:
            for (i, x, y, md) in ms or [(0, "", "", "")]:
                f.write(f"{ph},{tk},{h[0]},{h[1]},{i},{x},{y},{md}\n")
    with open(os.path.join(out, name.replace("trace", "foes")), "w") as f:
        f.write("phase,tick,n_live,live\n")
        for ph, tk, lv in mons:
            f.write(f"{ph},{tk},{len(lv)},{'|'.join(' '.join(x) for x in lv)}\n")
def grid():
    return s.cmd("debug grid 1").splitlines()[:-1]
def far_target(g, hx, hy, minlen=30):
    """BFS over free cells from the hero; the farthest-reachable cell within 60 steps with a bend (4-neighbour)."""
    import collections
    H, W = len(g), len(g[0]); sx, sy = int(hx), int(hy)
    dist = {(sx, sy): 0}; q = collections.deque([(sx, sy)])
    while q:
        x, y = q.popleft()
        for dx, dy in ((1,0),(-1,0),(0,1),(0,-1)):
            nx, ny = x+dx, y+dy
            if 0 <= nx < W and 0 <= ny < H and (nx, ny) not in dist and g[ny][nx] == ".":
                dist[(nx, ny)] = dist[(x, y)] + 1; q.append((nx, ny))
    cand = [(d, c) for c, d in dist.items() if d <= 60]
    d, c = max(cand)
    return c, d
def scene(area, tag, goto_wait=100):
    s.cmd("debug points 70 3"); s.cmd("debug points 69 5")
    log(tag, s.cmd(f"debug goto {area}").strip()); step(goto_wait)
    s.cmd("debug kill"); step(20)
    s.cmd("debug spawn fallen 3"); step(25); s.cmd("debug kill"); step(40)
    for row in [l.split("\t") for l in s.cmd("monsters").splitlines() if l.count("\t") > 10 and "Fallen" in l]:
        s.cmd("debug mana full")
        s.cmd(f"debug cast 70 {row[1]} {row[2]}"); step(40)
    log(tag, "minions", len(mins()["minions"]), "foes", foes(0))
def walk_leg(phase, tx, ty, ticks, stop=100):
    log(phase, s.cmd(f"cmd move {tx:.2f} {ty:.2f}").strip()); rec(phase, ticks); rec(phase + "_stop", stop)
try:
    area = int(sys.argv[4]) if len(sys.argv) > 4 else 2
    scene(area, f"area{area}")
    hx, hy = hero(); log("hero", hx, hy)
    rec("A_idle", 50)
    if area == 2:
        # seed 5 Blood Moor: row y=16 is free for x 20..43; first walk to its east end
        log(s.cmd("cmd move 42.5 16.5")); rec("B_toRow", 120)
        walk_leg("C_walk_west", 20.5, 16.5, 360)
        s.cmd("key r"); walk_leg("E_run_east", 42.5, 16.5, 260); s.cmd("key r")
    else:
        g = grid(); (tx, ty), d = far_target(g, hx, hy)
        log("far target", tx, ty, "bfs steps", d)
        walk_leg("C_walk_far", tx + 0.5, ty + 0.5, int(d * 14) + 50)
        s.cmd("key r"); walk_leg("E_run_back", hx, hy, int(d * 10) + 50); s.cmd("key r")
        open(os.path.join(out, f"grid_area{area}.txt"), "w").write("\n".join(g))
    dump(f"trace_area{area}.csv")
    log("final", json.dumps(mins()))
finally:
    s.close()
