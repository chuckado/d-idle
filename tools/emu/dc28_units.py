# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp28: per area units per clear from ms_spawn.py output (same counting as am_model.py: units with experience > 0).
    python3 dc28_units.py POP.txt  ->  CSV lid,seeds,units_mean,units_min,units_max,plain,champion,unique,minion"""
import os, sys, statistics as st
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import am_model
pop = am_model.load(sys.argv[1])
print("lid,seeds,units_mean,units_min,units_max,plain,champion,unique,minion,mlvl_mean")
for lid, seeds in sorted(pop.items()):
    u = [len(v) for v in seeds.values()]
    kinds = {}
    lv = []
    for v in seeds.values():
        for k, _, l in v:
            kinds[k] = kinds.get(k, 0) + 1
            lv.append(l)
    n = len(seeds)
    g = lambda k: kinds.get(k, 0) / n
    print("%d,%d,%.1f,%d,%d,%.1f,%.1f,%.1f,%.1f,%.2f" % (lid, n, st.mean(u), min(u), max(u), g("plain"), g("champion"), g("unique"), g("minion"), st.mean(lv)))
