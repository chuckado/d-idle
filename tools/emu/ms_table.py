# SPDX-License-Identifier: GPL-3.0-or-later
"""Print game.exe's stat row (ms_stats.py game output) of named MonStats ids at their own Normal level.

    python3 ms_table.py game.tsv meta.tsv fallen1,zombie1,..

Columns: id level life(min..max) ac to-hit(A1) A1 dmg A2 dmg S1 dmg exp. A damage of 0..0 means no such attack.
First row of an id is used (a few ids repeat in the table).
"""
import sys


def main():
    game, meta, names = sys.argv[1], sys.argv[2], sys.argv[3].split(",")
    level = {}
    for line in open(meta):
        f = line.split()
        if len(f) >= 6 and f[0].isdigit() and f[1].lower() not in level: level[f[1].lower()] = int(f[5])
    rows = {}
    for line in open(game):
        f = line.split()
        if len(f) >= 16 and f[1] == "0" and f[0] in level and int(f[2]) == level[f[0]] and f[0] not in rows:
            rows[f[0]] = [int(x) for x in f[3:]]
    print("id lvl life ac th_a1 a1 a2 s1 exp th_a2 th_s1")
    for n in names:
        r = rows.get(n)
        if r:
            print(f"{n} {level[n]} {r[0]}..{r[1]} {r[2]} {r[3]} {r[4]}..{r[5]} {r[6]}..{r[7]} {r[8]}..{r[9]} {r[10]} {r[11]} {r[12]}")


main()
