# SPDX-License-Identifier: GPL-3.0-or-later
"""um_fpick.py (card uniqmods): which monsters game.exe's FUN_005a03e0 refuses Extra Fast / Teleportation, d2d allows.

    python3 um_fpick.py MONSTATS.TXT MONSTATS2.TXT [MONUMOD.TXT]

Rule read from the disassembly (FUN_005a03e0, V read, see docs/research/uniqmods-trace.md Q4):
fPick 3 (Extra Fast, row 6) needs FUN_0046c140(class, 2) = MonStats2 mode bit mWL;
fPick 1 (Teleportation, row 26) needs FUN_0046c140(class, 4) = mode bit mA1.
d2d's `unique_detail::can_take` (rules/uniques.hpp) does not test either (a ponytail comment).
Table reading by header name (docs/process/mpq-data-reading.md). Tag D: a list of the rows where the two
rules differ; it is the input of a differential (game side: run FUN_005a03e0), not a proof.
"""
import csv
import sys


def table(path):
    with open(path, newline="") as f:
        r = list(csv.reader(f, delimiter="\t"))
    h = r[0]
    return h, [dict(zip(h, x)) for x in r[1:] if len(x) > 3]


def main():
    _, ms = table(sys.argv[1])
    _, m2 = table(sys.argv[2])
    by2 = {x["Id"]: x for x in m2}
    n = 0
    print("hcIdx|Id|MonStatsEx|MonType|Level N|NM|H|mWL|mA1|blocked in game.exe")
    for r in ms:
        if r.get("enabled") != "1" or r.get("Level", "0") in ("", "0") or r.get("npc") == "1":
            continue
        ex = by2.get(r.get("MonStatsEx", ""))
        if ex is None:
            continue
        n += 1
        wl, a1 = ex.get("mWL") == "1", ex.get("mA1") == "1"
        blocked = [name for name, ok in (("fast", wl), ("teleport", a1)) if not ok]
        if blocked:
            print("|".join([r["hcIdx"], r["Id"], r["MonStatsEx"], r["MonType"], r["Level"], r["Level(N)"], r["Level(H)"],
                            ex.get("mWL", ""), ex.get("mA1", ""), "+".join(blocked)]))
    print("rows checked:", n)


if __name__ == "__main__":
    main()
