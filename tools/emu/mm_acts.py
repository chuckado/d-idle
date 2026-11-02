#!/usr/bin/env python3
"""MONSTERMISSILE: Act 1 / Act 2 monster missile rows with the game.exe launch formula applied (tag D arithmetic on D tables).

    python3 mm_acts.py TABLES_DIR     # monstats.txt missiles.txt skills.txt levels.txt (1.14d patch layer, mpq-cat)

Launch record (game.exe FUN_0064b860, read; docs/research/monster-missile-trace.md Q2): for an owner that is a monster the record's
physical min / max (256ths) are  (stat21 << 8) * SrcDamage / 128 + (MinDamage + level brackets) << HitShift  (and the max alike with
stat 22), with stat 21 / 22 the owner's damage for its CURRENT MODE (FUN_005a4f50: A1 -> A1, A2 -> A2, SC / S1 -> S1).
Level brackets: the missiles' level is monster skill bonus + 1; none of the rows below has a bracket (checked), so only min / max are used.
"""
import csv
import os
import sys

DIFF = (("", "Nrm"), ("(N)", "NM"), ("(H)", "Hell"))


def rows(d, name):
    r = list(csv.reader(open(os.path.join(d, name), encoding="latin-1"), delimiter="\t"))
    return [dict(zip(r[0], x)) for x in r[1:] if x and x[0] and x[0] != "Expansion"]


def num(s):
    try: return int(s)
    except ValueError: return 0


def main():
    d = sys.argv[1]
    mons = {m["Id"]: m for m in rows(d, "monstats.txt")}
    mis = {m["Missile"]: m for m in rows(d, "missiles.txt")}
    sk = {s["skill"]: s for s in rows(d, "skills.txt")}
    lv = rows(d, "levels.txt")
    acts = {}
    for a in (0, 1):
        s = set()
        for l in lv:
            if num(l.get("Act", "")) != a or not l.get("Name"): continue
            for k in list(range(1, 11)):
                for p in ("mon", "nmon", "umon"):
                    v = l.get(f"{p}{k}", "")
                    if v in mons: s.add(v)
        acts[a] = s
    # bring the quest / boss monsters in by hand (superunique bases are in levels' umon lists or not)
    boss = {"andariel": 0, "bloodraven": 0, "countess": 0, "radament": 1, "summoner": 1, "duriel": 1}
    both = acts[0] | acts[1] | {b for b in boss if b in mons}
    print("act1_mons", len(acts[0]), "act2_mons", len(acts[1]))
    print("id\tact\tsource\tmissile\tmid\tvel\trange\tsize\tck\ttohit\tsrcdmg\tmin\tmax\tshift\tetype\temin\temax\telen\tskill\tsrvdmg\thitfn\tmonstats pct (N/NM/H)\t(no projection: the game's own record is mm_launch.py, the stat columns are PERCENTS of MonLvl damage)")
    for mid in sorted(both, key=lambda i: int(mons[i]["hcIdx"])):
        m = mons[mid]
        act = "1" if mid in acts[0] else "2" if mid in acts[1] else "B"
        srcs = []
        for c, sc in (("MissA1", "A1"), ("MissA2", "A2"), ("MissS1", "S1"), ("MissC", "S1")):
            if m.get(c): srcs.append((c, m[c], sc))
        for k in range(1, 9):
            s = sk.get(m.get(f"Skill{k}", ""))
            if s and s.get("srvmissilea") in mis: srcs.append((f"Skill{k}:{s['skill']}", s["srvmissilea"], "A1"))
        for c, mn, sc in srcs:
            r = mis.get(mn)
            if not r: continue
            S = num(r["SrcDamage"]) if r["SrcDamage"] not in ("", "-1") else 0
            lo, hi = num(r["MinDamage"]), num(r["MaxDamage"])
            sh = num(r["HitShift"])
            stat, phys = [], []
            for suf, _ in DIFF:
                a, b = num(m[f"{sc}MinD{suf}"]), num(m[f"{sc}MaxD{suf}"])
                stat.append(f"{a}-{b}")
                # 256ths -> points; trunc toward zero like (v*S + ((v*S >> 31) & 0x7f)) >> 7
                pmin = ((a << 8) * S) // 128 + (lo << sh)
                pmax = ((b << 8) * S) // 128 + (hi << sh)
                phys.append("-")
            print("\t".join(str(x) for x in (mid, act, c, mn, r["Id"], r["Vel"], r["Range"], r["Size"], r["CollideKill"], r["ToHit"], r["SrcDamage"], lo, hi, sh,
                                           r["EType"], r["EMin"], r["Emax"], r["ELen"], r["Skill"], r["pSrvDmgFunc"], r["pSrvHitFunc"], " ".join(stat), " ".join(phys))))


if __name__ == "__main__": main()
