# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp23: the outcome table of the warp walk: d2d's real tick (the dump's W lines) against game.exe's walk model
(ww_path.py game: path request + frames, the distance test when a walk ends, ONE more walk, then idle).
    ww_outcome.py DUMP GAMETSV [--list]
"""
import collections
import sys


def main():
    dump, gamefile = sys.argv[1], sys.argv[2]
    W, D, G = {}, {}, {}
    for line in open(dump):
        f = line.split()
        if f[0] == "W" and f[6] != "-1":
            W[(f[1], f[2], f[3], f[5])] = (int(f[6]), int(f[7]), int(f[8]))      # taken, ticks, unit blocked
        elif f[0] == "D":
            D[tuple(f[1:5])] = f[8:]
    for line in open(gamefile):
        c = line.rstrip("\n").split("\t")
        G[tuple(c[:4])] = c
    table = collections.Counter()
    rows = []
    for key, g in G.items():
        d2d, ticks, unit_blocked = W[key]
        game = min(int(g[4]), 1)
        same_path = g[6].split() == D[key]
        table[(unit_blocked, d2d, game)] += 1
        rows.append((key, d2d, game, same_path, ticks, g[7]))
    print("unit_blocked d2d_taken game_taken: starts")
    for k in sorted(table): print(*k, ":", table[k])
    tot = collections.Counter()
    for (ub, d, g), n in table.items():
        tot["d2d"] += n * d; tot["game"] += n * g; tot["n"] += n
        if d != g: tot["differ"] += n
    print(f"d2d takes {tot['d2d']} of {tot['n']}, game.exe model takes {tot['game']}, outcome differs in {tot['differ']}")
    if "--list" in sys.argv:
        for r in sorted(rows):
            if r[1] != r[2]: print(*r)


if __name__ == "__main__":
    main()
