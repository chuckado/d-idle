# SPDX-License-Identifier: GPL-3.0-or-later
"""mercstats: the merc's experience award, game.exe's FUN_0057e860 (called from the owner's experience share FUN_0057e990) run
natively, as the oracle for d2d's the merc experience rule (rules::merc_xp) + merc_stats.

    mst_xp.py CASES SEED OUTFILE      one line per case:
        K hid exp award blow merc_level owner_level ::: new_exp new_level

`award` is the experience of the slain monster at the merc's level (what FUN_0057e480 returns; not run here: it is the player's
experience function and the subject of upstream's F10 oracle). FUN_0057e990 scales it by 0x56 / 256 (to zero) unless the merc
struck the blow (read at 0x57ea28..0x57ea39, not run): this script does that scaling in Python and hands FUN_0057e860 the
result, so the scaling is V read and the rest (the gates: amount over 0, merc level under its owner's, under level 98; the
doubled addition; the level loop with FUN_00663790 and the hireling row of FUN_006562f0; the level-up call) is V run.
tools/emu/mst_d2d_test.cpp `xp` replays the line through the merc experience rule (rules::merc_xp) and rules::merc_stats.
"""
import random
import struct
import sys

import drlg
import merc_attack as ma


def threshold(level, per):
    return (level + 1) * per * level * level


def main():
    cases, seed0, path = int(sys.argv[1]), int(sys.argv[2]), sys.argv[3]
    rng = random.Random(seed0)
    outfile = open(path, "w")
    e = drlg.boot()
    rows = ma.hirelings(e)
    base = e.r32(0x96bdd0)
    per_of = {}
    for index, r in enumerate(rows):
        if r["version"] == 100 and r["id"] not in per_of:
            per_of[r["id"]] = struct.unpack("<i", e.read(base + 0x118 * index + 0x20, 4))[0]
    ids = sorted(per_of)
    cur = {}
    merc, player, game = e.alloc(0x400), e.alloc(0x400), e.alloc(0x400)
    e.w32(game + 0x70, 1)
    e.w32(merc + 0xc, 77)
    e.w32(player, 0)

    def stat_get(_e):
        unit, stat = _e.arg(0), _e.arg(1)
        if unit == player and stat == 0xc:
            return cur["owner"]
        if unit == merc and stat == 0xd:
            return cur["exp"]
        return 0
    e.hook(0x625480, stat_get, 3)
    e.hook(0x6253b0, stat_get, 3)

    def stat_put(_e):
        if _e.arg(0) == merc and _e.arg(1) == 0xd:
            cur["new_exp"] = _e.arg(2) & 0xffffffff
        return 0
    e.hook(0x627260, stat_put, 4)

    def record(_e):
        rec = _e.alloc(0x20)
        _e.w32(rec + 8, cur["hid"])
        return rec
    e.hook(0x574bd0, record, 0)
    e.hook(0x5531c0, lambda _e: 0, 4)
    e.hook(0x53bfd0, lambda _e: 0, 0)
    e.hook(0x572840, lambda _e: cur.__setitem__("new_level", _e.arg(1)) or 0, 2)
    e.hook(0x5726c0, lambda _e: 0, 1)
    e.hook(0x553380, lambda _e: 0, 1)
    e.hook(0x5c0c30, lambda _e: 0, 3)
    for _ in range(cases):
        hid = rng.choice(ids)
        per = per_of[hid]
        level = rng.randint(1, 98)
        owner = rng.choice((rng.randint(1, 99), level + 1, level, level - 1 if level > 1 else 1))
        lo, hi = threshold(level, per), threshold(level + 1, per)
        exp = lo + rng.randint(0, hi - lo - 1) if rng.random() < 0.8 else lo
        span = max(1, (hi - exp) * 2)
        award = rng.choice((rng.randint(-3, 3), rng.randint(0, span), rng.randint(0, span * 40), (hi - exp + 1) // 2, (hi - exp) // 2, (hi - exp + 1) // 2 - 1))
        blow = rng.random() < 0.5
        amount = award if blow else int(award * 0x56 / 256)          # FUN_0057e990 0x57ea28..0x57ea39, to zero
        cur.update(hid=hid, exp=exp, owner=owner, new_exp=exp, new_level=level)
        e.call(0x57e860, game, player, level, regs={"eax": amount & 0xffffffff, "edi": merc})
        print("K %d %d %d %d %d %d ::: %d %d" % (hid, exp, award, int(blow), level, owner, cur["new_exp"], cur["new_level"]), file=outfile)
    outfile.close()


main()
