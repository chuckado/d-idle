# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp29: game.exe's animation rate for a player as a table (FUN_00623f50 under unicorn, the skiasgame2_rate.py rig).
    tools/emu/run-on-worker.sh dc29_rate.py > rate.tsv
Lines `A ias wsm rate` (attack mode A1 = 7, AnimData speed 256; stat 68 = 100 - WSM) and `C fcr rate` (cast mode SC = 10, stat 105 = FCR
raw, speed 256). The unit's skill record is the Rig's stand-in (rate only; FCR/IAS go through FUN_00621740's 120-factor rows as game.exe)."""
import sys

import skiasgame2_rate as r

rig = r.Rig(r.EXE)
for wsm in (0, -10, -20, 10, 20, 30):
    for ias in range(0, 201):
        print("A", ias, wsm, rig.rate(7, 256, ias, 100 - wsm))
for fcr in range(0, 201):
    rig.stats[105] = fcr
    rig.stats[93] = 0
    rig.stats[0x44] = 100
    rig.e.w32(rig.unit, 0)
    rig.e.w32(rig.unit + 4, 1)
    rig.e.w32(rig.unit + 0x10, 10)
    rig.e.w32(rig.unit + 0x50, rig.rec)
    rig.e.w32(rig.rec + 0xc, 256)
    rig.e.call(0x623f50, rig.unit)
    print("C", fcr, rig.e.r16(rig.unit + 0x4c))
