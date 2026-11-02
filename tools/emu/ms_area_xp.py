# SPDX-License-Identifier: GPL-3.0-or-later
"""Per area: the real experience game.exe gave the plain monsters of a populated level (ms_spawn.py), and what
a kill pays a character of level CLVL after the level-difference step (FUN_0057e2f0, tables 0x6e1668 / 0x6e1694).

    python3 ms_area_xp.py spawn0.tsv meta.tsv "CLVL,CLVL,.."  [lids]

Per lid: units (plain, flags & 0x1c == 0 and rows read at the unit's own level), mean base experience, mean
level, mean life, and the mean award at each CLVL (integer MulDiv as the game: exp x table / 256, truncated),
next to the bot's estimate (internal/bot/recoveryxp.go: 10 x mlvl x level-difference share).
"""
import sys

BELOW = [256, 256, 256, 256, 256, 256, 207, 159, 110, 61, 13]   # 0x6e1668: clvl >= mlvl, index min(clvl - mlvl, 10)
ABOVE = [256, 256, 256, 256, 256, 256, 225, 174, 92, 38, 5]     # 0x6e1694: mlvl > clvl, clvl < 25, index min(mlvl - clvl, 10)


def award(exp, clvl, mlvl):
    if clvl >= mlvl: t = BELOW[min(clvl - mlvl, 10)]
    elif clvl >= 25: return exp * clvl // mlvl
    else: t = ABOVE[min(mlvl - clvl, 10)]
    return exp if t == 256 else exp * t // 256


def bot_pct(clvl, mlvl):
    d = mlvl - clvl
    if clvl >= 25 and d > 0: return 100 * clvl / mlvl
    if d >= 10: return 1.95
    if d >= 6: return [87.9, 68.0, 35.9, 14.8][d - 6]
    if d >= -5: return 100
    if d >= -9: return [80.9, 62.1, 43.0, 23.8][-d - 6]
    return 5.08


def main():
    spawn, clvls = sys.argv[1], [int(x) for x in sys.argv[2].split(",")]
    only = {int(x) for x in sys.argv[3].split(",")} if len(sys.argv) > 3 else None
    areas = {}
    for line in open(spawn):
        f = line.split()
        if not f or f[0] != "U": continue
        lid, flags, statlvl = int(f[1]), int(f[5], 16), f[6]
        if flags & 0x1c or statlvl == "None": continue
        st = {}
        for tok in f[8:]:
            k, _, v = tok.partition(":")
            st[int(k, 16)] = int(v)
        if st.get(0xc, 0) != int(statlvl) or st.get(0xd, 0) <= 0: continue
        areas.setdefault(lid, []).append((st[0xd], st[0xc], st.get(7, 0) / 256))
    print("lid units mean_exp mean_lvl mean_hp | per clvl: game award / bot estimate (10 x mlvl x pct)")
    for lid, rows in sorted(areas.items()):
        if only and lid not in only: continue
        n = len(rows)
        me, ml, mh = sum(r[0] for r in rows) / n, sum(r[1] for r in rows) / n, sum(r[2] for r in rows) / n
        cells = []
        for c in clvls:
            g = sum(award(e, c, l) for e, l, _ in rows) / n
            b = sum(10 * l * bot_pct(c, l) / 100 for _, l, _ in rows) / n
            cells.append(f"{c}:{g:.1f}/{b:.1f}")
        print(f"{lid} {n} {me:.1f} {ml:.1f} {mh:.1f} | " + " ".join(cells))


main()
