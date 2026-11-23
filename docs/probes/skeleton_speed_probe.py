#!/usr/bin/env python3
"""Skeleton speed probe: a Necromancer raises skeletons on a lockstep d2d, the hero walks, we record per tick positions.
Usage: skeleton_speed_probe.py <d2d> <outdir> [seed]. Reads only plus SETUP verbs (--lab-verbs, throwaway char, no save)."""
import json, math, os, re, sys, time
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
    return s.cmd("debug grid 5").splitlines()[:-1]
def far_path(g, hx, hy, maxd=360):
    """4-neighbour BFS over the SUBTILE grid (debug grid 5), obstacles grown by one subtile; path to the farthest cell
    within maxd steps. Returns cells (floats) every subtile."""
    import collections
    H, W = len(g), len(g[0])
    free = [[g[y][x] == "." for x in range(W)] for y in range(H)]
    ok = [[free[y][x] and all(free[min(H-1, max(0, y+dy))][min(W-1, max(0, x+dx))] for dx in (-1,0,1) for dy in (-1,0,1)) for x in range(W)] for y in range(H)]
    sx, sy = int(hx * 5), int(hy * 5)
    par = {(sx, sy): None}; dist = {(sx, sy): 0}; q = collections.deque([(sx, sy)])
    while q:
        x, y = q.popleft()
        if dist[(x, y)] >= maxd:
            continue
        for dx, dy in ((1,0),(-1,0),(0,1),(0,-1)):
            n = (x+dx, y+dy)
            if 0 <= n[0] < W and 0 <= n[1] < H and n not in dist and ok[n[1]][n[0]]:
                dist[n] = dist[(x, y)] + 1; par[n] = (x, y); q.append(n)
    far = max(dist, key=lambda c: (dist[c], c))
    path = []; c = far
    while c is not None:
        path.append(((c[0] + 0.5) / 5, (c[1] + 0.5) / 5)); c = par[c]
    return path[::-1]
def hop(path, phase, hopn=25, per_hop=200):
    """walk the hero along the path in hops of hopn subtiles (a fresh cmd move each), recording every tick"""
    for i in range(hopn, len(path) + hopn, hopn):
        c = path[min(i, len(path) - 1)]
        s.cmd(f"cmd move {c[0]:.2f} {c[1]:.2f}")
        for _ in range(per_hop // 10):
            rec(phase, 10)
            if math.dist(hero(), c) < 0.5:
                break
def scene(area, tag, goto_wait=100):
    s.cmd("debug points 70 3"); s.cmd("debug points 69 5")
    log(tag, s.cmd(f"debug goto {area}").strip()); step(goto_wait)
    s.cmd("debug kill"); step(20)
    s.cmd("debug spawn fallen 3"); step(25); s.cmd("debug kill"); step(40)
    for row in [l.split("\t") for l in s.cmd("monsters").splitlines() if l.count("\t") > 10 and "Fallen" in l and not os.environ.get("NOPETS")]:
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
        g = grid(); path = far_path(g, hx, hy)
        log("path cells", len(path), "from", path[0], "to", path[-1])
        hop(path, "C_walk_far"); rec("C_walk_far_stop", 100)
        s.cmd("key r"); hop(path[::-1], "E_run_back"); s.cmd("key r"); rec("E_run_back_stop", 100)
        open(os.path.join(out, f"grid_area{area}.txt"), "w").write("\n".join(g))
        open(os.path.join(out, f"path_area{area}.txt"), "w").write(" ".join(f"{x:.1f},{y:.1f}" for x, y in path[::5]))
    dump(f"trace_area{area}{os.environ.get('TAG', '')}.csv")
    log("final", json.dumps(mins()))
finally:
    s.close()
