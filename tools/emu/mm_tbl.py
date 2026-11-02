#!/usr/bin/env python3
"""MONSTERMISSILE table dump (card monstermissile): every monster missile row that can reach the player.

    python3 mm_tbl.py TABLES_DIR [acts]      # TABLES_DIR holds monstats.txt, missiles.txt, skills.txt (1.14d patch layer, mpq-cat)

Prints one line per (monster, source, missile): source is MissA1 / MissA2 / MissS1 / MissC or Skill<k> -> srvmissilea.
Columns are read by header name (docs/process/mpq-data-reading.md). Pure table reading: tag D.
"""
import csv
import os
import sys


def rows(d, name):
    r = list(csv.reader(open(os.path.join(d, name), encoding="latin-1"), delimiter="\t"))
    return [dict(zip(r[0], x)) for x in r[1:] if x and x[0] and x[0] != "Expansion"]


def main():
    d = sys.argv[1]
    mons = rows(d, "monstats.txt")
    mis = {m["Missile"]: m for m in rows(d, "missiles.txt")}
    sk = {s["skill"]: s for s in rows(d, "skills.txt")}
    out = []
    for m in mons:
        if m.get("enabled") != "1":
            continue
        src = []
        for c in ("MissA1", "MissA2", "MissS1", "MissC", "MissS2", "MissS3", "MissS4", "MissSQ"):
            if m.get(c): src.append((c, m[c], ""))
        for k in range(1, 9):
            s = sk.get(m.get(f"Skill{k}", ""))
            if s and s.get("srvmissilea") in mis:
                src.append((f"Skill{k}", s["srvmissilea"], f"{s['skill']}/mode {m.get(f'Sk{k}mode','')}/lvl {m.get(f'Sk{k}lvl','')}"))
        for c, mn, note in src:
            r = mis.get(mn)
            if not r: continue
            out.append((int(m["hcIdx"]), m["Id"], c, mn, note, r, m))
    for hc, mid, c, mn, note, r, m in sorted(out, key=lambda x: (x[0], x[2])):
        print("\t".join(str(x) for x in (
            hc, mid, c, mn, note, r["Id"], "vel", r["Vel"], "velLev", r["VelLev"], "rng", r["Range"], "lrng", r["LevRange"], "size", r["Size"],
            "ck", r["CollideKill"], "cf", r["CollideFriend"], "tohit", r["ToHit"], "pierce", r["Pierce"], "srcdmg", r["SrcDamage"],
            "min", r["MinDamage"], "max", r["MaxDamage"], "shift", r["HitShift"], "etype", r["EType"], "emin", r["EMin"], "emax", r["Emax"], "elen", r["ELen"],
            "skill", r["Skill"], "srvdmg", r["pSrvDmgFunc"], "hitfn", r["pSrvHitFunc"], "dofn", r["pSrvDoFunc"], "expl", r["Explosion"], "hitclass", r["HitClass"],
            "A1", f"{m['A1MinD']}-{m['A1MaxD']}/{m['A1TH']}", "A2", f"{m['A2MinD']}-{m['A2MaxD']}/{m['A2TH']}", "S1", f"{m['S1MinD']}-{m['S1MaxD']}/{m['S1TH']}")))


if __name__ == "__main__": main()
