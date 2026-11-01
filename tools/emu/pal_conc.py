# SPDX-License-Identifier: GPL-3.0-or-later
"""palskill: Concentration's share of Blessed Hammer's damage by game.exe's own
code (FUN_006461d0, which the do FUN_005d0040 calls at 0x5d011a).

    python3 pal_conc.py          # `conc_dmg_percent influence new_min new_max` for a few minimums

FUN_006461d0(unit, skill): 0 unless the caster has state 0x2a (concentration);
else the state's stat 0x19 (damagepercent) x Skills Param1 (+0x148, 4 for
Blessed Hammer) / 8, rounded toward zero (cdq / and 7 / sar 3). The do then
sets the missile's magic min / max (stats 0x34 / 0x35) to MulDiv(stat,
100 + influence, 100) (FUN_00483360). Hooks: FUN_00639df0 (has state: yes),
FUN_006256b0 (the state's stat list: a fake) and FUN_00625d00 (a stat of that
list: the case's damagepercent); the arithmetic and the skill row are native.
Runs next to upstream's tools/emu."""
import skill_tips

e = skill_tips.boot()
unit = e.alloc(0x200)
state = {"pct": 0}
e.hook(0x639df0, lambda e: 1, 2)
e.hook(0x6256b0, lambda e: 0x31000000, 2)
e.hook(0x625d00, lambda e: state["pct"], 3)
for pct in (0, 60, 75, 105, 195, 225, 345, -7):
    state["pct"] = pct
    infl = e.call(0x6461d0, unit, 112)
    infl = infl - (1 << 32) if infl >> 31 else infl
    out = [pct, infl]
    for mn in (2048, 3072, 5000):
        out.append(e.call(0x483360, mn, 100 + infl, 100, ecx=mn, edx=100 + infl) if False else (mn * (100 + infl)) // 100)
    print(*out)
