# SPDX-License-Identifier: GPL-3.0-or-later
"""SKGROUND2: Fire Wall's `firewall` placements, game.exe (sk_gnd2_fw.py) vs a real d2d build.
    python sk_gnd2_fw_cmp.py game.txt d2d.txt [SHIFT [trim]]
SHIFT (default 0) is added to game.exe's frames (the harness counts the maker's first do as frame 1;
d2d's first tick of a new missile is one frame after the cast: SHIFT 1 lines the two clocks up).
Per case: the set of subtiles (relative to the target) that get a firewall, the frame each
subtile first gets one, and how many are made (d2d makes one every frame, game.exe one per
new subtile of the maker's path)."""
import sys


def load(path, game):
    cases = {}
    for ln in open(path):
        p = ln.split()
        if game and p and p[0] == "G":
            cid, frame, dx, dy = int(p[1]), int(p[4]), int(p[5]), int(p[6])
        elif not game and p and p[0] == "D":
            cid, frame, dx, dy = int(p[1]), int(p[2]), int(p[3]), int(p[4])
        else:
            continue
        cases.setdefault(cid, []).append((frame, dx, dy))
    return cases


def first(items):
    out = {}
    for frame, dx, dy in sorted(items):
        out.setdefault((dx, dy), frame)
    return out


g, d = load(sys.argv[1], True), load(sys.argv[2], False)
shift = int(sys.argv[3]) if len(sys.argv) > 3 else 0
trim = len(sys.argv) > 4 and sys.argv[4] == "trim"      # drop game.exe's frames after d2d's last
g = {c: [(f + shift, x, y) for f, x, y in v] for c, v in g.items()}
n = same_set = same_first = 0
last_missing = 0
only_g = only_d = 0
tot_g = tot_d = 0
bad = 0
for cid in sorted(g):
    if cid not in d:
        continue
    n += 1
    top = max(f for f, _, _ in d[cid])
    if trim:
        g[cid] = [t for t in g[cid] if t[0] <= top]
    fg, fd = first(g[cid]), first(d[cid])
    tot_g += len(g[cid])
    tot_d += len(d[cid])
    og, od = set(fg) - set(fd), set(fd) - set(fg)
    only_g += len(og)
    only_d += len(od)
    if not og and not od:
        same_set += 1
    if all(fd.get(k) == v for k, v in fg.items()):
        same_first += 1
    elif bad < 3:
        bad += 1
        k = next(k for k, v in fg.items() if fd.get(k) != v)
        print("first differs, case", cid, "subtile", k, "game frame", fg[k], "d2d", fd.get(k))
print(f"cases {n}; firewalls made: game.exe {tot_g}, d2d {tot_d}")
print(f"same subtile set {same_set}/{n}; subtiles only game.exe {only_g}, only d2d {only_d}")
print(f"every game.exe subtile first made on the same frame by d2d: {same_first}/{n}")
