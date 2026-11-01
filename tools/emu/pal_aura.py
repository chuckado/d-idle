# SPDX-License-Identifier: GPL-3.0-or-later
"""palskill: the Paladin auras' calcs by game.exe's own calc machine.

    python3 pal_aura.py gen N SEED > cases.tsv
    python3 pal_aura.py game cases.tsv > game.tsv
    pal_test --aura cases.tsv > d2d.tsv          # tools/emu/pal_test.cpp, the real d2d rules
    diff game.tsv d2d.tsv

A case is `id skill lvl clvl m_fire m_ltng m_cold m_pois n (skill level)*n`
(skill_damage.py's format). `game` prints per case `id range perdelay auralen
a1..a6 p1..p5`: aurarangecalc (+0x64), perdelay calc (+0x128), auralencalc
(+0x60), the six aurastatcalc (+0x68, a slot with no aurastat id +0x54 prints
x) and the five passivecalc (+0xa4, no passivestat +0x98 prints x), each by
FUN_00646ca0(unit, calc, skill, lvl): the calc machine the aura do 65 / 66
call (FUN_005cf010 at 0x5cf0f1, 0x5cf1f6 and 0x5cf195). Hooks as
skill_damage.py (stat getter, skill record by id, skill level)."""
import random
import sys

import skill_damage

AURAS = [98, 99, 100, 102, 103, 104, 105, 108, 109, 110, 111, 113, 114, 115, 118, 119, 120, 122, 123, 124, 125, 117]
OTHERS = [96, 97, 98, 99, 100, 102, 103, 104, 105, 108, 109, 110, 112, 113, 114, 115, 118, 119, 120, 121, 122, 123, 124, 125]


def s32(v):
    return v - (1 << 32) if v >> 31 else v


def run(o, skill, lvl, clvl, m, pairs):
    e = o.e
    o.levels = dict(pairs)
    o.levels[skill] = lvl
    o.stats = {0xc: clvl, 0x149: m[0], 0x14a: m[1], 0x14b: m[2], 0x14c: m[3]}
    r = o.skills + skill * skill_damage.RECORD
    out = []

    def calc(off):
        ptr = e.r32(r + off)
        return s32(e.call(0x646ca0, o.unit, ptr, skill, lvl)) if ptr else 0
    out += [calc(0x64), calc(0x128), calc(0x60)]
    for k in range(6):
        sid = e.read(r + 0x54 + 2 * k, 2)
        out.append(calc(0x68 + 4 * k) if int.from_bytes(sid, "little", signed=True) >= 0 else "x")
    for k in range(5):
        sid = e.read(r + 0x98 + 2 * k, 2)
        out.append(calc(0xa4 + 4 * k) if int.from_bytes(sid, "little", signed=True) >= 0 else "x")
    return out


def gen(n, seed):
    rng = random.Random(seed)
    out = []
    for s in AURAS:
        for lvl in range(1, 41):
            out.append((s, lvl, 30, [0, 0, 0, 0], [(t, 0) for t in OTHERS[:0]]))
            out.append((s, lvl, 99, [50, 50, 50, 50], [(t, 20) for t in OTHERS]))
    for _ in range(n):
        s = rng.choice(AURAS)
        k = rng.randint(0, 8)
        pairs = [(rng.choice(OTHERS), rng.randint(0, 30)) for _ in range(k)]
        out.append((s, rng.choice([rng.randint(1, 20), rng.randint(1, 60)]), rng.randint(1, 99), [rng.choice([0, rng.randint(0, 300)]) for _ in range(4)], pairs))
    for i, (s, lvl, clvl, m, pairs) in enumerate(out):
        print(i, s, lvl, clvl, *m, len(pairs), *(v for p in pairs for v in p))


def main():
    a = sys.argv[1:]
    if a[0] == "gen":
        gen(int(a[1]), int(a[2]))
    elif a[0] == "game":
        o = skill_damage.Oracle()
        with open(a[1]) as f:
            for line in f:
                cid, s, lvl, clvl, m, pairs = skill_damage.parse(line)
                print(cid, *run(o, s, lvl, clvl, m, pairs))


if __name__ == "__main__":
    main()
