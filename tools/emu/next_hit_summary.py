# SPDX-License-Identifier: GPL-3.0-or-later
"""Hits per cast from test_next_hit's d2d output and next_hit.py's game.exe replay.

    python3 next_hit_summary.py d2d.txt game.txt

Per monster touched by at least one missile: d2d's hits vs game.exe's, by
skill, size and (boundary cases) distance; random packs by pack size.
"""
import sys
from collections import defaultdict


def read(path):
    cases, totals = {}, {}
    for line in open(path):
        f = line.split()
        if line.startswith("C"):
            mons = [tuple(map(int, m.replace("@", ",").split(","))) for m in f[4:]]
            cases[f[0][1:]] = (f[1], int(f[2]), mons)
        elif line.startswith("T"):
            totals[f[0][1:]] = list(map(int, f[1:]))
    return cases, totals


def main():
    cases, d2d = read(sys.argv[1])
    _, game = read(sys.argv[2])
    touched = set()
    for line in open(sys.argv[1]):
        if line.startswith("F"):
            f = line.split()
            for c in f[2:]:
                for u in c.split(":")[1].split("."): touched.add((f[0][1:], int(u)))
    by = defaultdict(lambda: [0, 0, 0, 0])          # monsters, d2d hits, game hits, monsters that differ
    differ_cases = 0
    for tag, (skill, lvl, mons) in cases.items():
        differ_cases += d2d[tag] != game[tag]
        for i, (size, dx, dy) in enumerate(mons):
            if (tag, i) not in touched: continue
            keys = [("all", skill), ("size", skill, size)]
            if tag[0] == "B": keys.append(("point", skill, size, max(abs(dx), abs(dy))))
            else: keys.append(("pack", skill, len(mons)))
            for k in keys:
                row = by[k]
                row[0] += 1; row[1] += d2d[tag][i]; row[2] += game[tag][i]; row[3] += d2d[tag][i] != game[tag][i]
    print(f"cases {len(cases)}, cases whose per-monster hits differ {differ_cases}")
    for k in sorted(by, key=str):
        n, a, b, d = by[k]
        print(" ".join(map(str, k)), f"monsters {n} d2d {a / n:.2f} game {b / n:.2f} per cast, differ {d}")


if __name__ == "__main__": main()
