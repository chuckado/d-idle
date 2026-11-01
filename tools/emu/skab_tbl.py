# SPDX-License-Identifier: GPL-3.0-or-later
"""SKABRECHECK claim 1 (data): who writes the resist-pierce stats 333..336.

    python skab_tbl.py          # run next to upstream's tools/emu (emu.py), like sk_pass_tbl.py

Reads the patched 1.14d tables as game.exe's MPQ stack serves them: every
Skills.txt row (all classes) whose passivestat / aurastat / *stat column names
a pierce stat, the ItemStatCost rows 333..336, and every Properties.txt row
whose stat columns name one.
"""
import os
import sys

sys.path.insert(0, os.getcwd())
import emu  # noqa: E402


def table(e, name):
    t = e.mpq_file("data\\global\\excel\\%s.txt" % name).decode("latin-1").splitlines()
    head = t[0].split("\t")
    return head, [dict(zip(head, ln.split("\t"))) for ln in t[1:]]


def main():
    e = emu.Emu()
    _, isc = table(e, "ItemStatCost")
    names = {}
    for r in isc:
        try:
            i = int(r.get("*ID") or r.get("ID") or -1)
        except ValueError:
            continue
        if 329 <= i <= 336:
            names[i] = r.get("Stat")
    print("ItemStatCost 329..336:", names)
    wanted = {n for i, n in names.items() if 333 <= i <= 336}
    h, rows = table(e, "Skills")
    print("Skills rows naming", sorted(wanted), "in any stat column:")
    n = 0
    for r in rows:
        hits = [(k, v) for k, v in r.items() if v in wanted]
        if hits:
            n += 1
            print("  ", r.get("Id") or r.get("skill"), r.get("charclass"), hits)
    print("  (%d rows of %d)" % (n, len(rows)))
    print("Skills rows with passivestat/aurastat 329..332 (masteries):")
    for r in rows:
        hits = [(k, v) for k, v in r.items() if "stat" in k and v in ("passive_fire_mastery", "passive_ltng_mastery", "passive_cold_mastery", "passive_pois_mastery")]
        if hits:
            print("  ", r.get("Id") or r.get("skill"), r.get("charclass"), hits)
    for sk in ("Fire Mastery", "Lightning Mastery", "Cold Mastery", "Poison and Bone Mastery"):
        for r in rows:
            if r.get("skill") == sk:
                print(sk, {k: v for k, v in r.items() if k.startswith("passive") and v})
    h, props = table(e, "Properties")
    print("Properties rows with a stat column naming a pierce stat:")
    for r in props:
        hits = [(k, v) for k, v in r.items() if v in wanted]
        if hits:
            print("  ", r.get("code"), hits)


if __name__ == "__main__" and len(sys.argv) == 1:
    main()


def rowdump(ids):
    e = emu.Emu()
    h, rows = table(e, "Skills")
    for r in rows:
        if (r.get("Id") or "") in ids:
            print("---", r.get("skill"), {k: v for k, v in r.items() if v and (k.startswith("passive") or k in ("Id", "charclass", "reqlevel", "srvdofunc", "EType", "srvstfunc", "srvmissile"))})


if __name__ == "__main__" and len(sys.argv) > 1:
    rowdump(sys.argv[1:])
