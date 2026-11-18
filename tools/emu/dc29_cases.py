# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp29: cases for skill_damage.py's `game` and d2d's `test_game --skill-damage`: every skill id 6..280 x level 1..30 x synergy set
{none, all others at 1, all others at 20}, clvl 30, masteries 0 (the line format of skill_damage.py: id skill lvl clvl m1 m2 m3 m4 n (skill lvl)*n)."""
import skill_damage

o = skill_damage.Oracle()
i = 0
for s in range(6, 281):
    for lvl in range(1, 31):
        for v in (0, 1, 20):
            pairs = [(t, v) for t in range(o.count) if t != s] if v else []
            print(i, s, lvl, 30, 0, 0, 0, 0, len(pairs), *(x for p in pairs for x in p))
            i += 1
