# SPDX-License-Identifier: GPL-3.0-or-later
"""mercstats: which items pass the merc's potion filter of FUN_0054d230 (it tests FUN_00629bb0(item, type) for types 0x4c, 0x51,
0x50), run natively over every item class id of game.exe's own loaded item tables. FUN_00629bb0 walks game.exe's
item type matrix (the closure of ItemTypes Equiv1 / Equiv2), so a type that lists a potion type as its parent passes too.

    mst_potion.py       prints per tested type the item class ids that pass (class id = row of weapons + armor + misc)
"""
import drlg


def main():
    e = drlg.boot()
    item = e.alloc(0x40)
    types = {0x4c: "hpot", 0x4d: "mpot", 0x4e: "rpot", 0x4f: "spot", 0x50: "apot", 0x51: "wpot"}
    t = e.r32(0x744304)
    count = e.r32(t + 0xbfc)
    print("item types known to the matrix:", count)
    total = 0
    while e.call(0x6335f0, total):             # the item record of a class id: null past the last (as drops.py counts them)
        total += 1
    print("item classes:", total)
    for tid, name in types.items():
        hits = []
        for cls in range(total):
            e.w32(item + 4, cls)
            got = e.call(0x629bb0, item, tid) & 0xffffffff
            if got:
                hits.append(cls)
        print("type %#x (%s):" % (tid, name), *hits)


main()
