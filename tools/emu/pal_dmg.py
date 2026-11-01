# SPDX-License-Identifier: GPL-3.0-or-later
"""palskill: the Paladin skills' element min / max by game.exe's own code.

    python3 pal_dmg.py gen > cases.tsv                 # skill_damage.py's case format
    python3 pal_dmg.py game cases.tsv > game.tsv       # id min max len (FUN_00644d50 / e40 / f20, flag 1)
    test_game --skill-damage cases.tsv > d2d.tsv       # the real d2d build
    diff game.tsv d2d.tsv

Runs next to upstream's tools/emu (imports skill_damage.py). Skills and the
skills their EDmgSymPerCalc reads (D, Skills.txt): Holy Bolt 101 (Blessed
Hammer 112, Fist of the Heavens 121), Blessed Hammer 112 (Vigor 115,
Blessed Aim 100), Holy Fire 102 (Resist Fire 100, Salvation 125), Holy Freeze 114,
Holy Shock 118, Fist of the Heavens 121 (D: Skills.txt order of the class)."""
import sys

import skill_damage

SYN = {"Holy Bolt": ["Blessed Hammer", "Fist of the Heavens"],
       "Blessed Hammer": ["Vigor", "Blessed Aim"],
       "Holy Fire": ["Resist Fire", "Salvation"],
       "Holy Freeze": ["Resist Cold", "Salvation"],
       "Holy Shock": ["Resist Lightning", "Salvation"],
       "Fist of the Heavens": ["Holy Bolt", "Blessed Hammer"],
       "Sacrifice": [], "Smite": [], "Zeal": [], "Vengeance": [], "Conversion": []}
IDS = {"Holy Bolt": 101, "Blessed Hammer": 112, "Fist of the Heavens": 121, "Vigor": 115, "Blessed Aim": 108,
       "Holy Fire": 102, "Holy Freeze": 114, "Holy Shock": 118, "Resist Fire": 100, "Resist Cold": 105, "Resist Lightning": 110, "Salvation": 125}


def gen():
    out = []
    for name in ("Holy Bolt", "Blessed Hammer", "Holy Fire", "Holy Freeze", "Holy Shock", "Fist of the Heavens"):
        s = IDS[name]
        syn = [IDS[x] for x in SYN[name]]
        for lvl in range(1, 41):
            for lv in (0, 1, 5, 10, 20, 30):
                for clvl in (1, 30, 99):
                    out.append((s, lvl, clvl, [0, 0, 0, 0], [(t, lv) for t in syn]))
                    out.append((s, lvl, clvl, [0, 0, 0, 0], [(syn[0], lv), (syn[1], 20 - min(lv, 20))]))
    for i, (s, lvl, clvl, m, pairs) in enumerate(out):
        print(i, s, lvl, clvl, *m, len(pairs), *(v for p in pairs for v in p))


def main():
    a = sys.argv[1:]
    if a[0] == "gen":
        gen()
    elif a[0] == "game":
        o = skill_damage.Oracle()
        for name, s in IDS.items():
            r = o.row(s)
            print("#row", s, name, r, file=sys.stderr)
        with open(a[1]) as f:
            for line in f:
                cid, s, lvl, clvl, m, pairs = skill_damage.parse(line)
                print(cid, *o.run(s, lvl, clvl, m, pairs))


if __name__ == "__main__":
    main()
