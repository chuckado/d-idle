#!/usr/bin/env python3
"""mercstats: how often each pick comes out of the merc's skill roll (FUN_005e4d30, rules::merc_attack's pick), exactly, from the
hireling.txt table and the skill levels the level-up gives (Skills.txt reqlevel gates a skill).

    python3 -I tools/emu/mst_picks.py hireling.txt

The roll is rand(total + 1) on 0..total: under DefaultChance the plain action (the Rogue's MonStats Skill1 shot, the Desert
Mercenary's swing); else the first skill whose running sum covers it (a running aura is left out of the sum). `use` is the
long-run share of the merc's thinks that reach the roll at all (FUN_005e5050: chance min(growth + 40 + 2 x level, 95), growth
+10 after each miss, reset on a hit; 98 for the melee Desert Mercenary). Derived from the V run rules; not a run itself.
"""
import sys

REQ = {"Fire Arrow": 1, "Cold Arrow": 6, "Inner Sight": 1, "Jab": 1, "Prayer": 1, "Defiance": 6, "Blessed Aim": 12,
       "Thorns": 6, "Holy Freeze": 18, "Might": 1}
AURAS = ("Prayer", "Defiance", "Blessed Aim", "Thorns", "Holy Freeze", "Might")


def main():
    lines = [l.rstrip("\n").split("\t") for l in open(sys.argv[1]) if not l.startswith("labrun")]
    head, rows = lines[0], lines[1:]
    col = head.index

    def row_for(hid, level):
        found = None
        for r in rows:
            if len(r) < 30 or r[2] != "100" or int(r[3]) != hid:
                continue
            if found is not None and level < int(r[col("Level")]):
                return found
            found = r
        return found

    def use_share(level, melee):
        if melee:
            return 0.98
        probs, growth = [], 0
        while True:
            p = min(growth + 40 + 2 * level, 95) / 100
            probs.append(p)
            if p >= 0.95:
                break
            growth += 10
        surv = expected = 1.0
        expected = 0.0
        for p in probs[:-1]:
            expected += surv
            surv *= 1 - p
        expected += surv / probs[-1]
        return 1 / expected

    def picks(hid, level, aura_on):
        r = row_for(hid, level)
        gained = max(level - int(r[col("Level")]), 0)
        default = total = int(r[col("DefaultChance")])
        cum = []
        for k in range(1, 7):
            name = r[col("Skill%d" % k)]
            if not name:
                break
            skill_level = int(r[col("Level%d" % k)]) + (int(r[col("LvlPerLvl%d" % k)]) * (level - int(r[col("Level")])) >> 5)
            if level < REQ[name] or skill_level < 1 or (aura_on and name in AURAS):
                cum.append((name, None))
                continue
            total += int(r[col("Chance%d" % k)]) + int(r[col("ChancePerLvl%d" % k)]) * gained // 4
            cum.append((name, total))
        count = {}
        for roll in range(total + 1):
            if roll < default:
                count["plain"] = count.get("plain", 0) + 1
                continue
            for name, c in cum:
                if c is not None and c >= roll:
                    count[name] = count.get(name, 0) + 1
                    break
            else:
                count["plain"] = count.get("plain", 0) + 1
        return {k: round(100 * v / (total + 1), 1) for k, v in count.items()}

    for hid, label, melee in ((0, "Rogue Fire", False), (1, "Rogue Cold", False)):
        for level in (5, 6, 10, 15, 20, 30, 35, 36, 50, 67):
            print(label, "level", level, "roll %:", picks(hid, level, False), "share of thinks that roll:", round(100 * use_share(level, melee), 1))
    for hid, label in ((6, "Desert Comb-Normal"), (7, "Desert Def-Normal")):
        for level in (5, 6, 9, 15, 20, 30, 43):
            print(label, "level", level, "roll %:", picks(hid, level, False), "| aura running:", picks(hid, level, True), "| roll share 98")


main()
