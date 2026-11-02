# SPDX-License-Identifier: GPL-3.0-or-later
"""um_summ.py (card uniqmods): summarise a umod_spawn.py log (real room population by game.exe 1.14d).

    python3 um_summ.py spawn.txt

Each input line: `seed S lvl L: CLASS fFLAGS mods M statlvl A -> lvl B exp E0->E1 maxhp H hpcreate C ...`.
A line with flag 10 (minion) belongs to the last non-minion line before it. Prints per (area, class,
flags): packs, minions per pack (min, max, mean), the mod lists seen, and the level / exp / life seen.
"""
import collections
import re
import sys

PAT = re.compile(r"seed (\d+) lvl (\d+): (\d+) f(\w+) mods (\S*) statlvl (\d+) -> lvl (\d+) exp (\d+)->(\d+) maxhp ([\d.]+) hpcreate ([\d.]+)")


def main(path):
    packs = collections.defaultdict(list)
    cur = None
    for line in open(path):
        m = PAT.match(line)
        if not m:
            continue
        seed, area, cls, fl, mods, sl, nl, e0, e1, hp, hc = m.groups()
        if fl == "10":
            if cur:
                cur["min"] += 1
            continue
        cur = {"min": 0, "mods": mods, "sl": int(sl), "nl": int(nl), "e0": int(e0), "e1": int(e1), "hp": float(hp), "hc": float(hc)}
        packs[(int(area), int(cls), fl)].append(cur)
    for k in sorted(packs):
        ps = packs[k]
        ms = [p["min"] for p in ps]
        lv = sorted({(p["sl"], p["nl"]) for p in ps})
        print(k, "n", len(ps), "minions", min(ms), max(ms), round(sum(ms) / len(ms), 2),
              "mods", sorted({p["mods"] for p in ps}), "lvl", lv,
              "exp x", sorted({round(p["e1"] / p["e0"], 2) for p in ps if p["e0"]}))


if __name__ == "__main__":
    main(sys.argv[1])
