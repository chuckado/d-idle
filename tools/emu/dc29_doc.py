#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp29: markdown tables of docs/research/decomp29-skilldamage.md from internal/power/skilldamage.json.
    python3 tools/emu/dc29_doc.py internal/power/skilldamage.json > tables.md"""
import json
import sys

d = json.load(open(sys.argv[1]))
S, R, F = d["skills"], d["rate"], d["frames"]
LV = [1, 5, 10, 15, 20, 30]


def ceil_div(a, b):
    return -(-a // b)


def cast_ticks(n, fcr, seq=False):
    t = ceil_div(256 * n, R["cast"][fcr])
    return t if seq else t - 1


def att_ticks(n, ias):
    return ceil_div(256 * n, R["attack_wsm0"][ias])


def bps(fn, hi=200):
    out, prev = [], None
    for x in range(hi + 1):
        t = fn(x)
        if t != prev:
            out.append((x, t))
            prev = t
    return out


def pt(v):
    v = v / 256
    return ("%g" % round(v, 2))


def at(e, key, sub, i):
    return e[key]["none"][sub][e["levels"].index(i)]


FOCUS_E = ["Fire Bolt", "Charged Bolt", "Lightning", "Chain Lightning", "Frost Nova", "Blizzard", "Meteor", "Fire Ball", "Ice Bolt", "Glacial Spike", "Frozen Orb", "Nova", "Static Field",
           "Holy Bolt", "Blessed Hammer", "Fist of the Heavens", "Holy Fire", "Holy Shock",
           "Magic Arrow", "Fire Arrow", "Cold Arrow", "Lightning Bolt", "Lightning Fury", "Poison Javelin", "Plague Javelin", "Charged Strike", "Power Strike", "Immolation Arrow",
           "Teeth", "Bone Spear", "Bone Spirit", "Poison Nova", "Poison Dagger",
           "Fire Trauma", "Fists of Fire", "Blades of Ice", "Claws of Thunder", "Shock Field", "Lightning Sentry", "Wake of Fire Sentry",
           "Firestorm", "Eruption", "Arctic Blast", "Fire Claws", "Rabies", "Hurricane", "Volcano", "Tornado", "Molten Boulder", "Shock Wave", "Twister", "Armageddon"]
FOCUS_P = ["Attack", "Static Field", "Corpse Explosion", "Sacrifice", "Zeal", "Smite", "Vengeance", "Charge", "Bash", "Double Swing", "Frenzy", "Concentrate", "Whirlwind", "Berserk", "Double Throw", "Leap Attack", "War Cry",
           "Jab", "Impale", "Fend", "Strafe", "Multiple Shot", "Guided Arrow", "Tiger Strike", "Dragon Claw", "Dragon Talon", "Cobra Strike", "Royal Strike", "Blade Fury", "Fury", "Maul", "Feral Rage", "Hunger", "Shock Wave"]

print("### T1 Element damage (points) per skill level, no synergy; `+syn` = every named synergy at 20 (level 20 row)\n")
print("| Skill (id, class) | anim | mana L1 / L10 / L20 / L30 | L1 | L5 | L10 | L15 | L20 | L30 | L20 +syn | length L1 / L20 (ticks) |")
print("|---|---|---|---|---|---|---|---|---|---|---|")
for n in FOCUS_E:
    e = S.get(n)
    if not e or not any(e["elem"]["none"]["max"]):
        continue
    m = [e["mana"][e["levels"].index(i)] for i in (1, 10, 20, 30)]
    cells = []
    for i in LV:
        k = e["levels"].index(i)
        cells.append("%s-%s" % (pt(e["elem"]["none"]["min"][k]), pt(e["elem"]["none"]["max"][k])))
    k = e["levels"].index(20)
    syn = "%s-%s" % (pt(e["elem"]["all20"]["min"][k]), pt(e["elem"]["all20"]["max"][k]))
    ln = "%d / %d" % (e["elem"]["none"]["len"][0], e["elem"]["none"]["len"][k]) if any(e["elem"]["none"]["len"]) else "0"
    print("| %s (%d, %s, %s) | %s | %s | %s | %s | %s |" % (n, e["id"], e["class"], e["etype"], e["anim"], " / ".join("%g" % x for x in m), " | ".join(cells), syn, ln))

print("\n### T2 Weapon skills: srcdam, to-hit bonus and the calc slots (levels 1 / 10 / 20; none | +syn at 20)\n")
print("| Skill (id, class) | anim | srcdam | ToHit + per level | mana L1 / L20 | calc1 | calc2 | calc3 | calc4 | calc source (calc1 ; calc2) |")
print("|---|---|---|---|---|---|---|---|---|---|")
for n in FOCUS_P:
    e = S.get(n)
    if not e:
        continue
    lv = e["levels"]

    def c(j):
        a = [e["calc"]["none"][lv.index(i)][j] for i in (1, 10, 20) if i in lv]
        b = e["calc"]["all20"][lv.index(20)][j] if 20 in lv else a[-1]
        if not any(a) and not b:
            return "0"
        return "/".join(str(x) for x in a) + " \\| " + str(b)
    mana = "%s / %s" % (e["mana"][0], e["mana"][lv.index(20)] if 20 in lv else e["mana"][-1])
    srcs = "; ".join((x or "-").replace("|", "\\|") for x in e["calc_src"][:2])
    print("| %s (%d, %s) | %s | %s | %s + %s | %s | %s | %s | %s | %s | `%s` |" % (n, e["id"], e["class"], e["anim"], e["srcdam"], e["tohit_pct"], e["tohit_per_level"], mana, c(0), c(1), c(2), c(3), srcs[:90]))

print("\n### T3 Cast frames and FCR breakpoints (plain cast = `ceil(256 N / rate) - 1`, V run rate, see doc)\n")
print("| Class | SC frames (HTH) | breakpoints (FCR>= : ticks) |")
print("|---|---|---|")
for cls, nm in (("ama", "Amazon"), ("sor", "Sorceress"), ("nec", "Necromancer"), ("pal", "Paladin"), ("bar", "Barbarian"), ("dru", "Druid"), ("ass", "Assassin")):
    sc = F.get(cls, {}).get("SC", {}).get("HTH")
    if not sc:
        print("| %s | none | |" % nm)
        continue
    n = sc["frames"]
    note = "" if sc["speed"] == 256 else " (anim speed %d)" % sc["speed"]
    # druid speed != 256: rate scales
    def tk(f, n=n, sp=sc["speed"]):
        return ceil_div(256 * n, sp * (100 + min(120 * f // (120 + f), 75)) // 100) - 1 if sp != 256 else ceil_div(256 * n, R["cast"][f]) - 1
    print("| %s | %d%s | %s |" % (nm, n, note, ", ".join("%d: %d" % b for b in bps(tk))))

print("\n### T4 Attack frames (A1) per weapon class and IAS breakpoints (WSM 0; `ceil(256 N / rate)`)\n")
print("| Class | wclass | A1 frames | breakpoints (IAS>= : ticks) |")
print("|---|---|---|---|")
for cls, nm in (("ama", "Amazon"), ("sor", "Sorceress"), ("nec", "Necromancer"), ("pal", "Paladin"), ("bar", "Barbarian"), ("dru", "Druid"), ("ass", "Assassin")):
    for wc, v in sorted(F.get(cls, {}).get("A1", {}).items()):
        if wc not in ("HTH", "1HS", "2HS", "STF", "BOW", "XBW"):
            continue
        n = v["frames"]
        def tk(f, n=n, sp=v["speed"]):
            return ceil_div(256 * n, sp * (100 + min(120 * f // (120 + f), 75)) // 100) if sp != 256 else ceil_div(256 * n, R["attack_wsm0"][f])
        print("| %s | %s | %d | %s |" % (nm, wc, n, ", ".join("%d: %d" % b for b in bps(tk))))
