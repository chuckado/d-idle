# SPDX-License-Identifier: GPL-3.0-or-later
"""game.exe's MonUMod eligibility (FUN_005a03e0) as the oracle: may this mod row be rolled on this monster?

(Scratch oracle of card decomp36. It imports skill_tips / price_oracle of upstream's tools/emu: run it on the builder with
that directory on PYTHONPATH; the d2d side is the fork branch named in docs/research/dc36-lane-b.md.)

    python3 dc36_umod.py game > game.tsv          # `monster_row mod_id allowed`
    ../../build/tests/test_umod_can_take --umod-can-take | grep -E '^[0-9]+ [0-9]+ [01]$' > d2d.tsv
    diff game.tsv d2d.tsv

Every MonStats row x every MonUMod row (the table at tables+0xc50, 0x20 bytes a row), in an expansion game
(the second argument of FUN_005a03e0 is game +0x70), through FUN_005a03e0 itself, natively: the row's
enabled byte (+6), its exclusions (+0xa / +0xc through FUN_005a0070, the monster type matrix) and fPick (+9:
1 FUN_0046c140 flag 4, 2 the multishot flags of MonStats +0xc / +0xe, 3 FUN_0046c140 flag 2). The monster
is a fake unit {type 1, class = the row}. A row's weight (cpick / upick) is checked by the caller
(FUN_005a0500 / FUN_005a0600), not here.
"""
import sys

import skill_tips


def main():
    a = sys.argv[1:]
    if a and a[0] == "probe": pass
    e = skill_tips.boot()
    tables = e.r32(0x744304)
    rows, nrows = e.r32(tables + 0xa78), e.r32(tables + 0xa80)
    mods, nmods = e.r32(tables + 0xc50), e.r32(tables + 0xc54)
    unit = e.alloc(0x100)
    e.w32(unit, 1)
    if a and a[0] == "probe":
        print("monsters", nrows, "umod rows", nmods, hex(mods))
        for r in range(nmods):
            rec = mods + 0x20 * r
            print(r, "id/word4", e.r16(rec + 4), "en", e.read(rec + 6, 1)[0], "fpick", e.read(rec + 9, 1)[0], "ex", e.r16(rec + 0xa), e.r16(rec + 0xc), "w8", e.read(rec + 8, 1)[0])
        return
    for m in range(nrows):
        e.w32(unit + 4, m)
        for r in range(nmods):
            rec = mods + 0x20 * r
            ok = e.call(0x5a03e0, rec, 1, regs={"ebx": unit}) & 0xff
            print(m, r, ok)


if __name__ == "__main__":
    main()
