#!/usr/bin/env python3
"""mercstats: compare two oracle outputs line by line (log lines and labrun lines are ignored).

    python3 -I tools/emu/mst_cmp.py game.txt d2d.txt [first-N-mismatches-to-print]

A data line starts with a hex seed, a decimal id or a level. Prints the case count, the mismatch count and the first
mismatches. Columns of the levelup output can be masked with --skip NAME (repeatable), e.g. --skip regen.
"""
import re
import sys

LINE = re.compile(r"^([TA] \d|[0-9a-f]{8} \d+ \d+ \d+:|\d+ \d+ str |\d+ \d+$)")


def rows(path):
    return [l.rstrip("\n") for l in open(path) if LINE.match(l)]


def main():
    args = sys.argv[1:]
    skip = []
    while "--skip" in args:
        i = args.index("--skip")
        skip.append(args[i + 1])
        del args[i:i + 2]
    a, b = rows(args[0]), rows(args[1])
    show = int(args[2]) if len(args) > 2 else 8

    def mask(line):
        t = line.split()
        out, i = [], 0
        while i < len(t):
            if t[i] in skip:
                i += 2
                continue
            out.append(t[i]); i += 1
        return " ".join(out)
    bad = [(x, y) for x, y in zip(a, b) if mask(x) != mask(y)]
    print(f"game lines {len(a)}, d2d lines {len(b)}, compared {min(len(a), len(b))}, mismatches {len(bad)}")
    for x, y in bad[:show]:
        print("G", x)
        print("D", y)


main()
