# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp28: join game.exe's population (ms_spawn.py output) with the rooms of the same level and seed (dc28_rooms.py output).
    python3 dc28_perseed.py POP.txt ROOMS.txt [POP2.txt ...] > perseed.csv   ->  lid,seed,units,plain,unique,minion,rooms,tiles"""
import os, sys
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import am_model
rooms = {}
for line in open(sys.argv[2]):
    f = line.split()
    if f and f[0] == "S":
        rooms[(int(f[1]), int(f[2]))] = (int(f[3]), int(f[4]))
print("lid,seed,units,plain,unique,minion,rooms,tiles")
for path in [sys.argv[1]] + sys.argv[3:]:
    for lid, seeds in sorted(am_model.load(path).items()):
        for seed, v in sorted(seeds.items()):
            k = {}
            for kind, _, _ in v:
                k[kind] = k.get(kind, 0) + 1
            r = rooms.get((lid, seed), (0, 0))
            print("%d,%d,%d,%d,%d,%d,%d,%d" % (lid, seed, len(v), k.get("plain", 0), k.get("unique", 0), k.get("minion", 0), r[0], r[1]))
