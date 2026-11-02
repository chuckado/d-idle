# SPDX-License-Identifier: GPL-3.0-or-later
"""mercstats: game.exe 1.14d's own hireling code as the oracle (the worker host, upstream's emulator modules).

    mst_oracle.py hire FIRST-LAST     FUN_006637f0, the hire offer (upstream hire.dump): seed act diff clvl: id level life str dex cost exp def dmin dmax
    mst_oracle.py revive              FUN_006637b0, the resurrection price, merc level 0..255: `level cost`
    mst_oracle.py levelup             FUN_00572840, the level-up (stats 0 2 7 31 23 24 19 39 41 43 45 74, skill levels),
                                      every hireling Id of the LoD rows x level 1..98
    mst_oracle.py maxlevel            FUN_00611830(0) and the experience step FUN_00663790 for a few inputs

Each line is plain text; tools/emu/mst_d2d_test.cpp prints d2d's rules on the same keys (diff the two).
"""
import struct
import sys

import drlg
from unicorn.x86_const import UC_X86_REG_EDX


def revive(e):
    cur = [0]
    e.hook(0x625480, lambda _e: cur[0], 3)
    unit = e.alloc(0x40)
    for level in range(256):
        cur[0] = level
        print(level, e.call(0x6637b0, unit) & 0xffffffff)


def hire(e, span):
    import hire as upstream
    first, last = (int(v, 0) for v in span.split("-"))
    for s in range(first, last + 1):
        print(upstream.dump(e, s * 0x9e3779b1 & 0xffffffff))


def table_ids(e):
    base, count = e.r32(0x96bdd0), e.r32(0x96bdd4)
    ids = []
    for k in range(count):
        v, i = struct.unpack_from("<Hxxi", e.read(base + k * 0x118, 8), 0)
        if v == 100 and i not in ids:
            ids.append(i)
    return sorted(ids)


def levelup(e):
    sets, skills = {}, []
    cur = {"hid": 0, "level": 1}
    merc, player, game = e.alloc(0x400), e.alloc(0x400), e.alloc(0x400)
    e.w32(game + 0x70, 1)                                    # expansion
    e.w32(merc + 0xc, 77)

    def get(_e):
        unit, stat = _e.arg(0), _e.arg(1)
        if unit == merc and stat == 0xc:
            return cur["level"] - 1
        return 0
    e.hook(0x625480, get, 3)
    e.hook(0x5de330, lambda _e: 0, 3)

    def put(_e):
        sets[_e.arg(1)] = struct.unpack("<i", struct.pack("<I", _e.arg(2) & 0xffffffff))[0]
        return 0
    e.hook(0x627260, put, 4)

    def record(_e):
        rec = _e.alloc(0x20)
        _e.w32(rec + 8, cur["hid"])
        return rec
    e.hook(0x574bd0, record, 0)
    e.hook(0x6253b0, lambda _e: sets.get(7, 0), 3)
    e.hook(0x56deb0, lambda _e: skills.append((_e.mu.reg_read(UC_X86_REG_EDX), _e.arg(0))) or 0, 2)
    for hid in table_ids(e):
        for level in range(1, 99):
            sets.clear(); skills.clear()
            cur["hid"], cur["level"] = hid, level
            e.call(0x572840, merc, level, ecx=game, edx=player)
            row = e.call(0x6562f0, 1, hid, level)
            ids = [struct.unpack("<i", e.read(row + 0x78 + 4 * k, 4))[0] for k in range(6)]
            have = dict(skills)
            lv = [have.get(i, 0) if i > 0 else 0 for i in ids]
            life = sets.get(7, 0)
            assert life % 256 == 0, (hid, level, life)
            g = lambda s: sets.get(s, 0)
            print(hid, level, "str", g(0), "dex", g(2), "life", life >> 8, "def", g(0x1f), "dmg", g(0x17), g(0x18), "ar", g(0x13),
                  "res", g(0x27), g(0x29), g(0x2b), g(0x2d), "regen", g(0x4a), "skills", *lv, "next", g(0x1e), "life6", g(6) >> 8)


def maxlevel(e):
    print("611830(0)", e.call(0x611830, 0), "611830(1)", e.call(0x611830, 1), "611830(6)", e.call(0x611830, 6))
    for lv, per in ((1, 100), (9, 110), (50, 105), (98, 130)):
        print("663790", lv, per, e.call(0x663790, lv, per) & 0xffffffff)


if __name__ == "__main__":
    mode = sys.argv[1]
    e = drlg.boot()
    if mode == "hire":
        hire(e, sys.argv[2])
    elif mode == "revive":
        revive(e)
    elif mode == "levelup":
        levelup(e)
    elif mode == "maxlevel":
        maxlevel(e)
    else:
        sys.exit(2)
