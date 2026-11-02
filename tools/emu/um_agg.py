# SPDX-License-Identifier: GPL-3.0-or-later
"""UNIQMODS: aggregate the um_all.sh logs ($D2_SCRATCH/um_d<diff>_a<area>.txt) from umod_spawn.py.

Prints, per difficulty: random uniques (flags 9 = unique, no superunique bit) and their mod counts, minions per unique group,
champion packs (flag 4 kinds), the life multiplier by class (maxhp / hpcreate), exp ratio, and the superunique lines.
"""
import glob
import os
import re
from collections import Counter, defaultdict

d = os.environ.get("D2_SCRATCH", "/tmp")
LINE = re.compile(r"seed (\d+) lvl (\d+): (\d+) f([0-9a-f]+) mods ([0-9.]*) statlvl (\d+) -> lvl (\d+) exp (\d+)->(\d+) maxhp ([\d.]+) hpcreate ([\d.]+) dmg% (-?\d+) th% (-?\d+) vel% (-?\d+)")
for diff in (0, 1, 2):
    print(f"=== difficulty {diff}")
    uniq_mods = Counter(); champ_mods = Counter(); lvladd = defaultdict(Counter); expx = defaultdict(Counter)
    minions = []; packs = []; su = []
    for f in sorted(glob.glob(f"{d}/um_d{diff}_a*.txt")):
        area = re.search(r"_a(\d+)", f).group(1)
        units = []
        for ln in open(f):
            m = LINE.search(ln)
            if not m: continue
            seed, lvl, cls, fl, mods, sl, lv, e0, e1, hp, hc, dm, th, vel = m.groups()
            units.append((int(seed), int(cls), int(fl, 16), mods, int(sl), int(lv), int(e0), int(e1), float(hp), float(hc), int(dm), int(th), int(vel), area))
        for u in units:
            seed, cls, fl, mods, sl, lv, e0, e1, hp, hc, dm, th, vel, area = u
            n = len(mods.split(".")) if mods else 0
            ratio = round(e1 / e0, 2) if e0 else 0
            cl = "champ" if fl & 4 else "super" if (fl & 8 and fl & 2) else "unique" if fl & 8 else "minion" if fl & 0x10 else "?"
            if fl & 8 and (fl & 2):
                su.append((area, seed, cls, mods, sl, lv, e0, e1, hp, hc, dm, th, vel, hex(fl)))
            lvladd[cl][lv - sl] += 1; expx[cl][ratio] += 1
            if cl == "unique": uniq_mods[n] += 1
            if cl == "champ": champ_mods[mods] += 1
    print(" unique mod counts (flag 8 w/o 2):", dict(uniq_mods))
    print(" champion kind lists (top):", champ_mods.most_common(8))
    for cl in ("champ", "unique", "minion", "super"):
        print(f" {cl}: level add {dict(lvladd[cl])}  exp ratio {dict(expx[cl])}")
    print(" flagged-2 / superunique-looking rows:")
    for s in su[:12]: print("  ", s)
