# SPDX-License-Identifier: GPL-3.0-or-later
"""Per-claim mismatch counts of d2d's vendor prices against game.exe's.

    uv run python price_claims.py cases.tsv game.tsv d2d.tsv [d2d2.tsv ...]

For each d2d output (a build: upstream, a PR head, a patch) prints the
buy / sell / repair mismatches against game.exe (price_oracle.py game):

- all: every case;
- bigmul / nodur: the columns whose price_model.py price moves when that
  game.exe term is switched to d2d's old behaviour (the npc.txt mults
  divide by 1024 first above 0xffff; the repair gate's ItemTypes Repair and
  FUN_00629930 durability check), i.e. the cases the claim is about;
- clean: the columns no term d2d still lacks moves (not a stack or a
  quiver, whole-item repair unmoved); with every claimed term fixed this
  is 0 / 0 / 0.
"""
import collections
import sys

import price_model
import price_oracle

COLS = ("buy", "sell", "repair")


def main():
    o = price_oracle.Oracle()
    m = price_model.Model(o)
    game = {l.split()[0]: l.split()[1:] for l in open(sys.argv[2]) if l.strip()}
    builds = [(name, {l.split()[0]: l.split()[1:] for l in open(name) if l.strip()}) for name in sys.argv[3:]]
    sets = collections.defaultdict(list)          # set name -> [(case id, column)]
    for line in open(sys.argv[1]):
        if not line.strip(): continue
        case = price_oracle.parse(line)
        m.off = set()
        normal = m.price(case)
        moved = {}
        for term in ("bigmul", "nodur", "whole"):
            m.off = {term}
            moved[term] = [a != b for a, b in zip(m.price(case), normal)]
        m.off = set()
        k = m.kind(case[1])
        for i in range(3):
            sets["all"].append((case[0], i))
            if moved["bigmul"][i]: sets["bigmul"].append((case[0], i))
            if moved["nodur"][i]: sets["nodur"].append((case[0], i))
            if not (k["stack"] or k["quiver"] or moved["whole"][i]): sets["clean"].append((case[0], i))
    for name, d2d in builds:
        print(name)
        for s in ("all", "bigmul", "nodur", "clean"):
            size, bad = collections.Counter(), collections.Counter()
            for cid, i in sets[s]:
                size[i] += 1
                bad[i] += game[cid][i] != d2d[cid][i]
            print(f"  {s:7s} " + " / ".join(f"{bad[i]}" for i in range(3)) + "  of " + " / ".join(f"{size[i]}" for i in range(3)) + "  (buy / sell / repair)")


if __name__ == "__main__":
    main()
