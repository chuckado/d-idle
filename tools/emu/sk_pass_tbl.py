# SPDX-License-Identifier: GPL-3.0-or-later
"""SKPASS: dump the patched 1.14d Skills.txt columns that carry synergies and
passive stats, for the Sorceress rows, as game.exe's MPQ stack serves them
(emu.Emu().mpq_file, the same reader the other sk_* oracles boot from).

    python sk_pass_tbl.py skills          # sorceress rows, non-empty synergy / passive / calc columns
    python sk_pass_tbl.py cols            # the Skills.txt header
    python sk_pass_tbl.py stats NAME...   # ItemStatCost rows by stat name
    python sk_pass_tbl.py states N...     # States.txt names by row
    python sk_pass_tbl.py func ST DO      # srvstfunc ST / srvdofunc DO addresses
    python sk_pass_tbl.py record ID OFF[:1|2]...  # a loaded Skills record field (game.exe load)
    python sk_pass_tbl.py res             # the hit element descriptors at 0x732980
"""
import os
import sys

sys.path.insert(0, os.getcwd())
import emu  # noqa: E402

WANT = ("skill", "Id", "charclass", "reqlevel", "reqskill1", "reqskill2", "reqskill3", "passive",
        "passivestate", "passiveitype", "passivestat1", "passivecalc1", "passivestat2", "passivecalc2",
        "passivestat3", "passivecalc3", "passivestat4", "passivecalc4", "passivestat5", "passivecalc5",
        "aurastat1", "aurastatcalc1", "aurastat2", "aurastatcalc2", "auralencalc", "aurastate",
        "calc1", "calc2", "calc3", "calc4", "Param1", "Param2", "Param3", "Param4", "Param5", "Param6",
        "Param7", "Param8", "EType", "EMin", "EMinLev1", "EMinLev2", "EMinLev3", "EMinLev4", "EMinLev5",
        "EMax", "EMaxLev1", "EMaxLev2", "EMaxLev3", "EMaxLev4", "EMaxLev5", "EDmgSymPerCalc",
        "ELen", "ELevLen1", "ELevLen2", "ELevLen3", "ELenSymPerCalc", "MinDam", "MaxDam",
        "DmgSymPerCalc", "ToHitCalc", "ToHit", "LevToHit", "HitShift", "SrcDam", "srvstfunc",
        "srvdofunc", "srvmissile", "srvmissilea", "cltmissile", "mana", "lvlmana", "manashift",
        "minmana", "delay", "range", "Kick", "InTown")


def table(e, name):
    t = e.mpq_file("data\\global\\excel\\%s.txt" % name).decode("latin-1").splitlines()
    head = t[0].split("\t")
    return head, [dict(zip(head, ln.split("\t"))) for ln in t[1:]]


def main():
    e = emu.Emu()
    what = sys.argv[1] if len(sys.argv) > 1 else "skills"
    if what == "cols":
        print("\t".join(table(e, "Skills")[0]))
        return
    if what == "states":
        h, rows = table(e, "States")
        for i, r in enumerate(rows):
            if str(i) in sys.argv[2:]:
                print(i, r.get("state"))
        return
    if what == "stats":
        h, rows = table(e, "ItemStatCost")
        for i, r in enumerate(rows):
            if r.get("Stat") in sys.argv[2:]:
                print(i, {k: v for k, v in r.items() if v not in ("",)})
        return
    h, rows = table(e, "Skills")
    for r in rows:
        if r.get("charclass") != "sor":
            continue
        print("----", r.get("Id"), r.get("skill"))
        for k in WANT:
            v = r.get(k, "")
            if v not in ("", "0"):
                print("  %s = %s" % (k, v))



def restable():
    """The hit's element descriptors at 0x732980 (12 x 0x2c), as FUN_0057c1e0
    walks them into FUN_0057bf80 / FUN_0057be00: +0 field offset in the damage
    record, +4 resistance stat, +8 max-resistance stat, +0xc pierce stat."""
    e = emu.Emu()
    for i in range(12):
        a = 0x732980 + i * 0x2c
        d = [e.r32(a + 4 * k) for k in range(11)]
        s = ["%d" % x if x < 0x80000000 else "%d" % (x - (1 << 32)) for x in d]
        print(i, " ".join(s))




def functab():
    """srvstfunc / srvdofunc table entries (0x732140 / 0x7322b0, 4 bytes each)."""
    e = emu.Emu()
    for base, idx in ((0x732140, int(sys.argv[2])), (0x7322b0, int(sys.argv[3]))):
        print(hex(base), idx, hex(e.r32(base + 4 * idx)))


def record():
    """Read loaded Skills records after game.exe's own table load (skill_tips.boot):
    `record ID OFF[:size]...` e.g. record 43 0x14c 0x1dc:1."""
    import skill_tips
    e = skill_tips.boot()
    tables = e.r32(0x744304)
    base = e.r32(tables + 0xb98)
    sid = int(sys.argv[2])
    for spec in sys.argv[3:]:
        off, _, size = spec.partition(":")
        a = base + sid * 0x23c + int(off, 16)
        v = e.r32(a)
        v = v & 0xff if size == "1" else v & 0xffff if size == "2" else v
        print(sid, spec, v)


if __name__ == "__main__":
    {"res": restable, "func": functab, "record": record}.get(sys.argv[1] if len(sys.argv) > 1 else "", main)()
