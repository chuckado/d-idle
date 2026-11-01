# SPDX-License-Identifier: GPL-3.0-or-later
"""SKPASS: replay the leveler's point spending (internal/leveler/plan.go Plan +
leveler.go rule: reqlevel, reqskills, maxlvl) over a build table, clvl by clvl,
and compute the lightning skills' damage with the traced formula
(docs/research/skill-damage-trace.md: brackets FUN_00644b70, synergy MulDiv
with the min gate FUN_00644d50 / FUN_00644e40, mastery once FUN_00644c90).
Every number this prints is tag I (computed from D columns), not an oracle run.

    python3 sk_pass_plan.py PLAN.json [quests]   # quests: Den +1 at clvl 7, Radament +1 at 20, Izual +2 at 33
    python3 sk_pass_plan.py PLAN.json quests|none cold   # the END table's cold synergies

The Skills.txt columns are copied from `sk_pass_tbl.py skills` (patched 1.14d).
"""
import json
import sys

# name: (reqlevel, reqskills); maxlvl 20 for all sorceress rows
REQ = {
    "Fire Bolt": (1, []), "Warmth": (1, []), "Charged Bolt": (1, []), "Ice Bolt": (1, []),
    "Frozen Armor": (1, []), "Static Field": (6, []), "Telekinesis": (6, []), "Frost Nova": (6, []),
    "Ice Blast": (6, ["Ice Bolt"]), "Lightning": (12, ["Charged Bolt"]), "Chain Lightning": (18, ["Lightning"]),
    "Teleport": (18, ["Telekinesis"]), "Glacial Spike": (18, ["Ice Blast"]),
    "Blizzard": (24, ["Frost Nova", "Glacial Spike"]), "Lightning Mastery": (30, []), "Cold Mastery": (30, []),
}

# EMin, EMinLev1..5, EMax, EMaxLev1..5, HitShift, synergy (skills, par8)
DMG = {
    "Charged Bolt": (4, (1, 1, 2, 3, 4), 8, (1, 1, 2, 3, 4), 7, (["Lightning"], 6)),
    "Lightning": (1, (0, 0, 0, 0, 0), 40, (8, 12, 20, 28, 36), 8, (["Charged Bolt", "Chain Lightning", "Nova"], 8)),
    "Chain Lightning": (1, (0, 0, 0, 0, 0), 40, (11, 13, 15, 15, 15), 8, (["Charged Bolt", "Lightning", "Nova"], 4)),
}


def bonus(lvl, lev):
    l1, l2, l3, l4, l5 = lev
    if lvl <= 8:
        return (lvl - 1) * l1
    if lvl <= 16:
        return 7 * l1 + (lvl - 8) * l2
    if lvl <= 22:
        return 7 * l1 + 8 * l2 + (lvl - 16) * l3
    if lvl <= 28:
        return 7 * l1 + 8 * l2 + 6 * l3 + (lvl - 22) * l4
    return 7 * l1 + 8 * l2 + 6 * l3 + 6 * l4 + (lvl - 28) * l5


def muldiv(a, b, c):
    return a * b // c if a <= 0x100000 else a // c * b


def damage(name, pts, lm):
    """(min, max) in whole points per hit at base levels pts, Lightning Mastery level lm."""
    lvl = pts.get(name, 0)
    if lvl < 1:
        return None
    emin, lmin, emax, lmax, shift, (syn, par8) = DMG[name]
    calc = sum(pts.get(s, 0) for s in syn) * par8
    mastery = 50 + 12 * (lm - 1) if lm > 0 else 0
    lo = (emin + bonus(lvl, lmin)) << shift
    hi = (emax + bonus(lvl, lmax)) << shift
    if calc and (lo > 0x100 or lmin[0] != 0):
        lo += muldiv(lo, calc, 100)
    if calc:
        hi += muldiv(hi, calc, 100)
    lo += muldiv(lo, mastery, 100) if mastery else 0
    hi += muldiv(hi, mastery, 100) if mastery else 0
    return lo / 256, hi / 256, calc


def rule(name, clvl, lv):
    req, skills = REQ.get(name, (1, []))
    if req > clvl or lv.get(name, 0) >= 20:
        return False
    return all(lv.get(s, 0) > 0 for s in skills)


def plan(rows, have, clvl, pts_free):
    levels, owned = dict(have), dict(have)
    blocked, out, spent = set(), [], 0
    for r in rows:
        for sp in r["skills"]:
            for _ in range(sp["points"]):
                n = sp["skill"]
                if owned.get(n, 0) > 0:
                    owned[n] -= 1
                    continue
                if spent >= pts_free or n in blocked:
                    continue
                if not rule(n, clvl, levels):
                    blocked.add(n)
                    continue
                out.append(n)
                levels[n] = levels.get(n, 0) + 1
                spent += 1
    return out


# EDmgSymPerCalc / ELenSymPerCalc / auralencalc synergies of the cold skills (D, patched Skills.txt)
SYN = {
    "Ice Bolt": (["Frost Nova", "Ice Blast", "Glacial Spike", "Blizzard", "Frozen Orb"], 15),
    "Ice Blast": (["Ice Bolt", "Blizzard", "Frozen Orb"], 8),
    "Glacial Spike": (["Ice Bolt", "Ice Blast", "Frozen Orb"], 5),
    "Blizzard": (["Ice Bolt", "Ice Blast", "Glacial Spike"], 5),
    "Frost Nova": (["Blizzard", "Frozen Orb"], 10),
}


def cold(rows, quests):
    """END table: points and each cold skill's damage synergy % (I), Cold Mastery pierce % (20 + 5 (L-1))."""
    qp = {7: 1, 20: 1, 33: 2} if quests else {}
    have, free = {}, 0
    names = ("Ice Bolt", "Ice Blast", "Glacial Spike", "Blizzard", "Frost Nova")
    print("clvl spent | " + " ".join(n.split()[-1] for n in names) + " CM Warmth | synergy % " + " / ".join(names) + " | CM pierce")
    for clvl in range(2, 61):
        free += 1 + qp.get(clvl, 0)
        got = plan(rows, have, clvl, free)
        for n in got:
            have[n] = have.get(n, 0) + 1
        free -= len(got)
        pts = " ".join(str(have.get(n, 0)) for n in names + ("Cold Mastery", "Warmth"))
        syn = " / ".join(("+%d" % (sum(have.get(x, 0) for x in SYN[n][0]) * SYN[n][1])) if have.get(n, 0) else "-" for n in names)
        cm = have.get("Cold Mastery", 0)
        print("%2d %-34s | %s | %s | %s" % (clvl, ",".join(got) or "(none, %d saved)" % free, pts, syn, ("-%d" % (20 + 5 * (cm - 1))) if cm else "-"))


def main():
    rows = json.load(open(sys.argv[1]))["rows"]
    if len(sys.argv) > 3 and sys.argv[3] == "cold":
        cold(rows, sys.argv[2] == "quests")
        return
    quests = len(sys.argv) > 2 and sys.argv[2] == "quests"
    qp = {7: 1, 20: 1, 33: 2} if quests else {}
    have, free = {}, 0
    print("clvl spent_this_level | CB L CL LM TK TP FA FN | CB/bolt (syn%) | Lightning (syn%) | Chain Lightning (syn%)")
    for clvl in range(2, 61):
        free += 1 + qp.get(clvl, 0)
        got = plan(rows, have, clvl, free)
        for n in got:
            have[n] = have.get(n, 0) + 1
        free -= len(got)
        lm = have.get("Lightning Mastery", 0)
        cols = []
        for s in ("Charged Bolt", "Lightning", "Chain Lightning"):
            d = damage(s, have, lm)
            cols.append("-" if d is None else "%.1f-%.1f (+%d%%)" % d)
        short = " ".join(str(have.get(s, 0)) for s in ("Charged Bolt", "Lightning", "Chain Lightning",
                                                       "Lightning Mastery", "Telekinesis", "Teleport", "Frozen Armor", "Frost Nova"))
        print("%2d %-40s | %s | %s" % (clvl, ",".join(got) or "(none, %d saved)" % free, short, " | ".join(cols)))


if __name__ == "__main__":
    main()
