# SPDX-License-Identifier: GPL-3.0-or-later
"""um_tab.py (card uniqmods): MonUMod and MonStats facts the trace quotes, from the 1.14d patch-layer tables.

    python3 um_tab.py monumod MONUMOD.TXT           # the mod table: weights by difficulty, fPick, excludes
    python3 um_tab.py weights MONUMOD.TXT           # sum of the upick weights per difficulty (Normal / NM / Hell)
    python3 um_tab.py rows MONSTATS.TXT HCIDX...    # id, MonType, minions, party sizes, levels, boss flags
Columns are read by header name (docs/process/mpq-data-reading.md). Pure table reading: tag D.
"""
import csv
import sys


def rows(path):
    with open(path, newline="") as f:
        r = list(csv.reader(f, delimiter="\t"))
    return r[0], r[1:]


def monumod(path):
    h, rs = rows(path)
    ix = {k: h.index(k) for k in h}
    cols = ["uniquemod", "id", "enabled", "version", "champion", "fPick", "exclude1", "cpick", "upick", "upick (N)", "upick (H)", "constants"]
    print("|".join(cols))
    for x in rs:
        if len(x) > 5:
            print("|".join(x[ix[c]] if ix[c] < len(x) else "" for c in cols))


def weights(path):
    h, rs = rows(path)
    ix = {k: h.index(k) for k in h}
    tot = [0, 0, 0]
    cnt = [0, 0, 0]
    for x in rs:
        if len(x) < 10 or x[ix["champion"]] == "1" or x[ix["enabled"]] != "1":
            continue
        for d, c in enumerate(("upick", "upick (N)", "upick (H)")):
            v = x[ix[c]] if ix[c] < len(x) else ""
            if v:
                tot[d] += int(v)
                cnt[d] += 1
    print("unique pick weights (Normal NM Hell):", tot, "rows:", cnt)


def monrows(path, ids):
    h, rs = rows(path)
    ix = {k: h.index(k) for k in h}
    want = {int(i) for i in ids}
    cols = ["Id", "hcIdx", "BaseId", "MonType", "minion1", "minion2", "SetBoss", "PartyMin", "PartyMax", "MinGrp", "MaxGrp",
            "Velocity", "Run", "Level", "Level(N)", "Level(H)", "boss", "demon", "lUndead", "hUndead"]
    print("|".join(cols))
    for x in rs:
        if len(x) > 20 and x[ix["hcIdx"]].isdigit() and int(x[ix["hcIdx"]]) in want:
            print("|".join(x[ix[c]] if ix[c] < len(x) else "" for c in cols))


if __name__ == "__main__":
    {"monumod": lambda: monumod(sys.argv[2]), "weights": lambda: weights(sys.argv[2]),
     "rows": lambda: monrows(sys.argv[2], sys.argv[3:])}[sys.argv[1]]()
