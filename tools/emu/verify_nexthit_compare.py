# SPDX-License-Identifier: GPL-3.0-or-later
"""VERIFYNEXTHIT compare: d2d (test_verify_nexthit) vs game.exe (verify_nexthit.py).

    python verify_nexthit_compare.py d2d.txt g.txt [d2d2.txt g2.txt ...]   (pairs: one per seed, tags are per file)

Prints: cases, cases differing (any per-frame hit line), monster-casts touched, monster-casts whose
total differs, hits per touched monster (d2d, game.exe), split by skill and by moving / second cast.
"""
import sys
from collections import defaultdict


def load(path, prefix, frames, totals, info, touched):
    with open(path) as source:
        lines = source.readlines()
    for line in lines:
        f = line.split()
        if not f: continue
        tag = prefix + f[0][1:]
        if line[0] == "K":
            moving = any(not p.endswith("/0.00,0.00") for p in f[5:])
            info[tag] = (f[1], moving, f[4] != "0")
        elif line[0] == "D":
            frames[tag].append(" ".join(f[1:]))
        elif line[0] == "S":
            totals[tag] = [int(x) for x in f[1:]]
        elif line[0] == "I":
            for c in f[2:]:
                for u in c.split(":")[1].split("."): touched[tag].add(int(u))


def main():
    df, dt, info, touched = defaultdict(list), {}, {}, defaultdict(set)
    gf, gt = defaultdict(list), {}
    for k in range(1, len(sys.argv), 2):
        load(sys.argv[k], f"{k}:", df, dt, info, touched)
        load(sys.argv[k + 1], f"{k}:", gf, gt, {}, defaultdict(set))
    agg = defaultdict(lambda: [0, 0, 0, 0, 0, 0])   # cases, differ, touched, total differ, d2d hits, game hits
    for tag, (skill, moving, two) in info.items():
        keys = ["all", skill, ("moving" if moving else "standing"), ("2 casts" if two else "1 cast")]
        differ = df.get(tag, []) != gf.get(tag, [])
        d, g = dt.get(tag, []), gt.get(tag, [])
        for k in keys:
            a = agg[k]
            a[0] += 1; a[1] += differ
            for i in touched[tag]:
                a[2] += 1; a[3] += d[i] != g[i]; a[4] += d[i]; a[5] += g[i]
    print(f"{'slice':10} {'cases':>7} {'differ':>7} {'touched':>8} {'tot-diff':>8} {'d2d/t':>6} {'exe/t':>6}")
    for k in ("all", "nova", "frostnova", "standing", "moving", "1 cast", "2 casts"):
        a = agg[k]
        if a[0]: print(f"{k:10} {a[0]:7} {a[1]:7} {a[2]:8} {a[3]:8} {a[4]/max(a[2],1):6.2f} {a[5]/max(a[2],1):6.2f}")


if __name__ == "__main__":
    main()
