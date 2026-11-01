# SPDX-License-Identifier: GPL-3.0-or-later
"""palskill: Holy Bolt's / Blessed Hammer's undead and demon test by game.exe's
own code, per monster id.

    python3 pal_flags.py game > game.tsv      # id undead demon
    pal_test --undead > d2d.tsv               # d2d's type_info.undead / demon per game.exe unit id
    diff game.tsv d2d.tsv

FUN_0063e990 (undead: MonStats +0xd & (mask 3 | mask 4)) and FUN_0063e940
(demon: +0xd & mask 5) on a fake monster unit (type 1, id at +4); both are
what the Holy Bolt hit function 7 (FUN_005a9fb0) and the Blessed Hammer
damage function 5 (FUN_005ad2f0) call. Runs next to upstream's tools/emu."""
import sys

import skill_tips

e = skill_tips.boot()
tables = e.r32(0x744304)
count = e.r32(tables + 0xa80)
unit = e.alloc(0x200)
e.w32(unit, 1)
for mid in range(count):
    e.w32(unit + 4, mid)
    print(mid, e.call(0x63e990, unit) & 1, e.call(0x63e940, unit) & 1)
