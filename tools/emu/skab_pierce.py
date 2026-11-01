# SPDX-License-Identifier: GPL-3.0-or-later
"""SKABRECHECK claim 1: split sk_resist.py's differential by what it proves about the pierce stats.

    python skab_pierce.py cases.tsv game.tsv d2d.tsv

cases / game / d2d are sk_resist.py's files (`gen`, `game`, `test_resist --resist`, here
test_skab_resist). Counts, per element (fire 333, lightning 334, cold 335), the
cases where the attacker has a pierce stat > 0 and the mismatches of `dealt`
(the REAL missile_blow against FUN_0057c6c0) and `effres`, in classes:
  pure      no curse (v == 0): the pierce rule alone
  curse     v < 0, base < 100 (the cut does not apply): pierce after a plain curse
  d1        v < 0, base >= 100: the known curse-cut deviation (not the claim)
Boundary subsets of `pure`: base 99 / 100 / 101 (the 100 line), result near -100.
"""
import sys


def load(path, width):
    out = {}
    for line in open(path):
        f = line.split()
        if len(f) == width:
            out[int(f[0])] = f
    return out


def main():
    cs, gm, dd = load(sys.argv[1], 6), load(sys.argv[2], 4), load(sys.argv[3], 4)
    names = {2: "fire(333)", 3: "ltng(334)", 4: "cold(335)"}
    rows = {}
    for cid, c in cs.items():
        elem, base, v, p = int(c[1]), int(c[2]), int(c[3]), int(c[4])
        if elem not in names or cid not in gm or cid not in dd or p <= 0:
            continue
        cls = "pure" if v == 0 else ("d1" if base >= 100 else "curse")
        subs = [cls]
        if cls == "pure":
            if base in (99, 100, 101):
                subs.append("pure@100line")
            if base - p <= -95:
                subs.append("pure@floor")
            if base < 100 and base - p < 100:
                subs.append("pure<100")
        for s in subs:
            r = rows.setdefault((names[elem], s), [0, 0, 0])
            r[0] += 1
            r[1] += gm[cid][3] != dd[cid][3]            # dealt
            r[2] += dd[cid][2] != "-" and gm[cid][2] != dd[cid][2]   # effres
    print("element    class          cases(pierce>0)  dealt differ  effres differ")
    for k in sorted(rows):
        print("%-10s %-14s %10d %14d %14d" % (*k, *rows[k]))
    tot = {s: [sum(r[i] for (e, c), r in rows.items() if c == s) for i in range(3)] for s in ("pure", "curse", "d1")}
    print("totals:", {s: tuple(v) for s, v in tot.items()})


if __name__ == "__main__":
    main()
