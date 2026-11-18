# SPDX-License-Identifier: GPL-3.0-or-later
"""BOSS1 (decomp19): Act 1 boss table and the Paladin exposure model (tag I: arithmetic over D / V run inputs; it does NOT run game.exe).
    python3 bs_model.py <dir with monstats.txt and monlvl.txt (tools/emu/mpq-table.sh)> [table|fight]
table: per boss the Normal / Nightmare / Hell life, damage, to-hit, experience from the tables (rule: row percent x MonLvl L-column at the
level, V run in monster-stats-trace.md; mods from the V run of tools/emu/bs_spawn.sh). fight: Paladin at clvl 8/12/16/20 (plan v2 stats,
tools/emu/pm_dps.py for the dps) against each boss: boss dps on the hero, time to kill, life lost, with the assumptions printed."""
import csv
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import pm_dps  # noqa: E402


def load(d, f):
    r = list(csv.reader(open(os.path.join(d, f), encoding="latin-1"), delimiter="\t"))
    return [dict(zip(r[0], x)) for x in r[1:]]


def num(v, default=0):
    try:
        return int(v)
    except ValueError:
        return default


# V run (tools/emu/bs_spawn.sh seeds 1-4, Normal): mods applied at spawn: level add, life factor, experience factor, damage %, AR % (stat 0x19, 0x77)
# life factor: unique / superunique x4 (MonUMod constants 7..9, V run hpcreate x4); act boss (Blood Raven, Andariel) x1.
BOSSES = [
    # name, MonStats Id, area, level add, life x, exp x, dmg%, th%, defence (V run stat 0x1f), melee period ticks, note
    ("Corpsefire", "zombie1", "Den of Evil", 3, 4, 5, 0, 0, 5),
    ("Bishibosh", "fallenshaman1", "Cold Plains", 3, 4, 5, 0, 0, 10),
    ("Blood Raven", "bloodraven", "Burial Grounds", 0, 1, 1, 0, 0, 38),
    ("Coldcrow", "cr_archer1", "Cave 1 (Cold Plains)", 3, 4, 5, 0, 0, 20),
    ("Rakanishu", "fallen2", "Stony Field", 3, 4, 5, 0, 0, 25),
    ("Treehead WoodFist", "brute2", "Dark Wood", 3, 4, 5, 135, 90, 25),
    ("Griswold", "griswold", "Tristram", 3, 4, 5, 0, 0, 25),
    ("The Countess", "corruptrogue3", "Tower Cellar 5", 3, 4, 5, 0, 0, 40),
    ("The Smith", "smith", "Barracks", 3, 4, 5, 135, 90, 50),
    ("Andariel", "andariel", "Catacombs 4", 0, 1, 1, 0, 0, 60),
]
DIFF = {"N": ("Level", "", "L-"), "NM": ("Level(N)", "(N)", "L-"), "H": ("Level(H)", "(H)", "L-")}


def rows(d):
    ms = {m["Id"]: m for m in load(d, "monstats.txt")}
    ml = {num(r["Level"]): r for r in load(d, "monlvl.txt")}
    return ms, ml


def stats(m, ml, diff, ladd, lifex):
    lvlcol, sfx, _ = DIFF[diff]
    lvl = num(m[lvlcol])
    col = lambda base: m[base + sfx] if (base + sfx) in m else ""
    hpcol = {"N": "L-HP", "NM": "L-HP(N)", "H": "L-HP(H)"}[diff]
    dmcol = {"N": "L-DM", "NM": "L-DM(N)", "H": "L-DM(H)"}[diff]
    thcol = {"N": "L-TH", "NM": "L-TH(N)", "H": "L-TH(H)"}[diff]
    xpcol = {"N": "L-XP", "NM": "L-XP(N)", "H": "L-XP(H)"}[diff]
    hp = (num(ml[lvl][hpcol]) * num(m["minHP" if diff == "N" else "MinHP" + sfx]) // 100,
          num(ml[lvl][hpcol]) * num(m["maxHP" if diff == "N" else "MaxHP" + sfx]) // 100)
    dl = lvl + ladd                                  # damage and AR are read at the raised level (pr-mods-proof.md)
    a = {}
    for k in ("A1", "A2", "S1"):
        lo = num(m[k + "MinD" + sfx]); hi = num(m[k + "MaxD" + sfx]); th = num(m[k + "TH" + sfx])
        a[k] = (num(ml[dl][dmcol]) * lo // 100, num(ml[dl][dmcol]) * hi // 100, num(ml[dl][thcol]) * th // 100)
    xp = num(ml[lvl][xpcol]) * num(m["Exp" + sfx]) // 100
    return lvl, dl, (hp[0] * lifex, hp[1] * lifex), a, xp


def table(d):
    ms, ml = rows(d)
    print("boss | area | level N / NM / H (mods +3 on a unique) | life N / NM / H | A1 damage N | A1 AR N | exp N / NM / H")
    for name, mid, area, ladd, lifex, expx, dm, th, df in BOSSES:
        m = ms[mid]
        out = []
        for diff in ("N", "NM", "H"):
            lx = lifex if diff == "N" else ({4: 3, 1: 1}[lifex] if diff == "NM" else {4: 2, 1: 1}[lifex])
            out.append(stats(m, ml, diff, ladd, lx))
        lv = " / ".join(str(o[0] + ladd) for o in out)
        hpv = " / ".join("%d" % ((o[2][0] + o[2][1]) // 2) for o in out)
        a1 = out[0][3]["A1"]
        xps = " / ".join(str(o[4] * expx) for o in out)
        print("%s | %s | %s | %s | %d-%d | AR %d | %s" % (name, area, lv, hpv, a1[0] * (100 + dm) // 100, a1[1] * (100 + dm) // 100, a1[2] * (100 + th) // 100, xps))


# Paladin plan v2 (internal/plandata/build-paladin-normal30-v2.json, cumulative stat rows): clvl -> (str, dex, vit, Might, Zeal, Sacrifice)
HERO = {8: (38, 21, 46, 4, 0, 1), 12: (46, 21, 58, 8, 1, 1), 16: (51, 21, 73, 8, 4, 1), 20: (55, 23, 87, 8, 4, 1)}
ARMOR = {8: 25, 12: 50, 16: 80, 20: 110}   # I: defence from armour (Leather / Buckler grade up to Chain / Large Shield grade); no block, no resist
PERIOD = 32                                 # I: ticks per monster attack = A1 frames (about 16) + aidel 15 (D); 25 ticks a second
RESIST = 0


def hero_life(c, vit):
    return 55 + 2 * (c - 1) + 3 * (vit - 25)   # S: Paladin 55 at clvl 1 (Vit 25), +2 a level, +3 a Vit point (CharStats LifePerLevel 8, LifePerVitality 12 in fourths)


def hitp(ar, defense, mlvl, clvl):
    return min(95, max(5, ar / (ar + defense) * 100 * 2 * mlvl / (mlvl + clvl))) / 100


def fight(d):
    ms, ml = rows(d)
    print("hero: Paladin v2 stats, no block, no resist; armour defence %s; attack = plain Attack under Might (the bot's band below clvl 12) / Zeal from clvl 12; period %d ticks" % (ARMOR, PERIOD))
    for c, (s_, d_, v_, mg, ze, sa) in HERO.items():
        life = hero_life(c, v_)
        hdef = d_ // 4 + ARMOR[c]
        print("\n== clvl %d: life %d, defence %d" % (c, life, hdef))
        for name, mid, area, ladd, lifex, expx, dm, th, df in BOSSES:
            m = ms[mid]
            lvl, dl, hp, a, xp = stats(m, ml, "N", ladd, lifex)
            hpa = (hp[0] + hp[1]) / 2
            sk = {"Might": mg, "Sacrifice": sa, "Zeal": ze}
            atk = "Zeal" if ze else "Attack"
            dps, _, _ = pm_dps.per_target(c, s_, d_, pm_dps.SSD, dl, df, sk, atk)
            dps_d2d = pm_dps.per_target(c, s_, d_, pm_dps.SSD, dl, df, sk, atk, d2d=True)[0]
            lo, hi, ar = a["A1"]
            ar = ar * (100 + th) // 100
            dmg = (lo + hi) / 2 * (100 + dm) / 100
            bdps = hitp(ar, hdef, dl, c) * dmg * 25 / PERIOD
            ttk = hpa / dps
            print("%-18s lvl %2d life %5.0f def %2d | hero %5.1f dps (d2d %5.1f) ttk %5.1f s | boss melee %4.1f/hit at %3.0f%% -> %4.1f dps | life lost standing %5.1f = %3.0f%% of %d"
                  % (name, dl, hpa, df, dps, dps_d2d, ttk, dmg, 100 * hitp(ar, hdef, dl, c), bdps, bdps * ttk, 100 * bdps * ttk / life, life))


if __name__ == "__main__":
    {"table": table, "fight": fight}[sys.argv[2] if len(sys.argv) > 2 else "table"](sys.argv[1])
