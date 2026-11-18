#!/usr/bin/env python3
"""Compare two result files line by line (the game.exe side against a d2d build's) and count the differing columns.

    python3 -I dc36_cmp.py GAME.tsv D2D.tsv [column names...]

Each line is `id v1 v2 ...`; the ids must be the same in both files, in order. Prints `lines N differ M` and,
with column names, how many lines differ in each. Exit 0 either way (the numbers are the result).
"""
import sys


def main():
    game = [l.split() for l in open(sys.argv[1]) if l.strip()]
    d2d = [l.split() for l in open(sys.argv[2]) if l.strip()]
    names = sys.argv[3:]
    if len(game) != len(d2d) or any(a[0] != b[0] for a, b in zip(game, d2d)):
        sys.exit("ids differ between the two files")
    per = [0] * (len(game[0]) - 1)
    bad = 0
    for a, b in zip(game, d2d):
        d = [x != y for x, y in zip(a[1:], b[1:])]
        bad += any(d)
        for i, x in enumerate(d): per[i] += x
    print(f"lines {len(game)} differ {bad}")
    for i, n in enumerate(per):
        if n: print(f"  column {names[i] if i < len(names) else i + 1}: {n}")


if __name__ == "__main__":
    main()
