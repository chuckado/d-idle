# SPDX-License-Identifier: GPL-3.0-or-later
"""VERIFYNEXTHIT compare: d2d (test_verify_nexthit) vs game.exe (verify_nexthit.py).

    python verify_nexthit_compare.py d2d.txt g.txt

Prints: cases, cases differing (any per-frame hit line), monster-casts touched, monster-casts whose
total differs, hits per touched monster (d2d, game.exe), split by skill and by moving / second cast.
"""
import sys
from collections import defaultdict


def load(path):
    frames, totals, info, touched = defaultdict(list), {}, {}, defaultdict(set)
    for line in open(path):
        f = line.split()
        if not f: continue
        tag = f[0][1:]
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
    return frames, totals, info, touched


def main():
    df, dt, info, touched = load(sys.argv[1])
    gf, gt, _, _ = load(sys.argv[2])
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
