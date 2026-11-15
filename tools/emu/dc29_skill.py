# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp29: per-skill, per-level damage, calc slots and mana of every class skill by game.exe's own code (the
`oracle first` run for docs/research/decomp29-skilldamage.md).

    tools/emu/run-on-worker.sh dc29_skill.py all 1 30 > rows.jsonl     (skill ids 6..280 with a class; levels A..B)
    tools/emu/run-on-worker.sh dc29_skill.py mana  1 30 > mana.jsonl

One JSON line per (skill id, level). Per line, with the skill at `lvl` and every OTHER skill at S in {0, 1, 20}
(synergy sets "none", "all1", "all20"; a skill the calc does not name has no effect):
  elem  : FUN_00644d50 / FUN_00644e40 (min / max, 256ths, flag 1 = with the mastery stat, masteries 0) and FUN_00644f20 (length)
  phys  : FUN_00647bc0 / FUN_00647d00 (the skill's own physical min / max, 256ths, flag 0)
  calc  : FUN_00646ca0 on calc1..4 (+0x138..0x144)
`mana` mode runs sk_mana's FUN_0056c160 / FUN_0056bfe0 (the real check and payment) with plenty of mana and prints what is paid.
Hooked, not the claim: stat getter FUN_00625480 (stat 0xc = clvl, 0x149..0x14c = masteries = 0, else 0), FUN_00643810 (unit's skill
record by id), FUN_006442a0 (its level: base and with items equal). Everything else is game.exe's code (as skill_damage.py).
clvl is fixed at 30 for ulvl terms (Druid summons, Raven); masteries 0 (a mastery is a plain percent on top, FUN_00644c90)."""
import json
import sys

import skill_damage

RECORD = skill_damage.RECORD


def s32(v):
    return v - (1 << 32) if v >> 31 else v


def calcs(o, skill, lvl):
    e = o.e
    r = o.skills + skill * RECORD
    out = []
    for k in range(4):
        ptr = e.r32(r + 0x138 + 4 * k)
        out.append(0 if ptr in (0, 0xffffffff) else s32(e.call(0x646ca0, o.unit, ptr, skill, lvl)))
    return out


def one(o, skill, lvl, others, clvl=30):
    o.levels = dict(others)
    o.levels[skill] = lvl
    o.stats = {0xc: clvl, 0x149: 0, 0x14a: 0, 0x14b: 0, 0x14c: 0}
    e = o.e
    lo = s32(e.call(0x644d50, o.unit, skill, lvl, 1))
    hi = s32(e.call(0x644e40, o.unit, skill, lvl, 1))
    ln = s32(e.call(0x644f20, o.unit, skill, lvl, 1))
    plo = s32(e.call(0x647bc0, o.unit, skill, lvl, 0))
    phi = s32(e.call(0x647d00, o.unit, skill, lvl, 0))
    return {"e": [lo, hi, ln], "p": [plo, phi], "c": calcs(o, skill, lvl)}


def class_skills(o):
    out = []
    for s in range(o.count):
        r = o.skills + s * RECORD
        # +0x2 (word) is the charclass byte position in the row; a class skill has charclass != -1 (checked below by id range)
        out.append(s)
    return out


def main():
    a = sys.argv[1:]
    mode, lo_l, hi_l = a[0], int(a[1]), int(a[2])
    ids = [int(x) for x in a[3].split(",")] if len(a) > 3 else list(range(6, 281))
    if mode == "all":
        o = skill_damage.Oracle()
        for s in ids:
            if s >= o.count:
                continue
            for lvl in range(lo_l, hi_l + 1):
                rec = {"id": s, "lvl": lvl}
                for tag, v in (("none", 0), ("all1", 1), ("all20", 20)):
                    others = {t: v for t in range(o.count) if t != s} if v else {}
                    rec[tag] = one(o, s, lvl, others)
                print(json.dumps(rec, separators=(",", ":")))
    elif mode == "mana":
        import sk_mana
        o = sk_mana.Oracle()
        for s in ids:
            if s >= o.count:
                continue
            for lvl in range(lo_l, hi_l + 1):
                chk, paid, _ = o.run(s, lvl, 0x200000)
                print(json.dumps({"id": s, "lvl": lvl, "check": chk, "paid": paid}, separators=(",", ":")))
    else:
        raise SystemExit(__doc__)


if __name__ == "__main__":
    main()
