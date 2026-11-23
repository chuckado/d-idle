# SPDX-License-Identifier: GPL-3.0-or-later
"""Mana Burn on a monster's hit: what game.exe's damage build (FUN_0057b7d0) puts in the record's mana field (+0x3c).

(Scratch oracle of card decomp38; imports skill_tips of upstream's tools/emu, run on the builder.)

    python3 dc38_burn.py build N SEED > build.tsv

A case: a plain monster attacker (type 1, no hireling) whose stats 62 / 63 (manadrainmindam / manadrainmaxdam)
are `lo hi`, SrcDam byte `src`, unit seed `slo shi`; prints the case and the record's +0x3c and the seed after.
The real FUN_0057b7d0 runs; the stat getter FUN_00625480 answers 62 / 63 only, the rest is hooked to 0.
"""
import random
import sys

import skill_tips

M32 = 0xFFFFFFFF


def build(n, seed):
    e = skill_tips.boot()
    stats = {}
    e.hook(0x625480, lambda e: stats.get(e.arg(1), 0), 3)
    e.hook(0x625500, lambda e: 0, 3)
    e.hook(0x627260, lambda e: 0, 4)
    e.hook(0x535d10, lambda e: 0, 0)
    e.hook(0x63e940, lambda e: 0, 1)
    e.hook(0x63ee90, lambda e: 0, 1)
    e.hook(0x535bc0, lambda e: 0, 0)
    e.hook(0x6259b0, lambda e: 0, 1)
    e.hook(0x57b660, lambda e: 0, 0)
    e.hook(0x57b6a0, lambda e: 0, 2)
    r = random.Random(seed)
    att, tgt, rec = e.alloc(0x100), e.alloc(0x100), e.alloc(0x80)
    e.w32(att, 1); e.w32(att + 4, 5)
    e.w32(tgt, 0)
    for _ in range(n):
        lo = r.choice((0, 1, 5, 10, 30, 100, 400, r.randrange(0, 3000)))
        hi = lo + r.choice((0, 0, 1, 7, 50, r.randrange(0, 2000)))
        src = r.choice((0x80, 0x80, 0x80, 0x40, 0x100 - 1, r.randrange(1, 256)))
        slo, shi = r.getrandbits(32), r.getrandbits(32)
        stats.clear(); stats[62] = lo; stats[63] = hi
        e.mu.mem_write(rec, bytes(0x80))
        e.w32(att + 0x20, slo); e.w32(att + 0x24, shi)
        e.call(0x57b7d0, tgt, rec, 0, src, ecx=att, edx=att, limit=20000)
        print(lo, hi, src, slo, shi, e.r32(rec + 0x3c), e.r32(att + 0x20), e.r32(att + 0x24))


def heal(n, seed):
    """FUN_0057c420 on a plain monster attacker and a player defender: the record's leech fields are capped by what the
    defender has and the sum heals the attacker (stat 6, up to its max life). Prints
    `dmg life mana stam | def_mana def_stam | att_life att_max | att_life_after rec_life rec_mana rec_stam`."""
    e = skill_tips.boot()
    st = {}
    e.hook(0x625480, lambda e: st.get((e.arg(0), e.arg(1)), 0), 3)
    e.hook(0x625d10, lambda e: st.get((e.arg(0), 7), 0), 1)
    e.hook(0x627260, lambda e: st.__setitem__((e.arg(0), e.arg(1)), e.arg(2)) or 0, 4)
    e.hook(0x6272b0, lambda e: st.__setitem__((e.arg(0), e.arg(1)), st.get((e.arg(0), e.arg(1)), 0) + e.arg(2)) or 0, 4)
    e.hook(0x621e40, lambda e: 0, 2)
    e.hook(0x63ee90, lambda e: 0, 1)
    gm, att, tgt, rec = e.alloc(0x200), e.alloc(0x100), e.alloc(0x100), e.alloc(0x80)
    e.w32(att, 1); e.w32(att + 4, 5); e.w32(tgt, 0); e.w32(att + 0xc8, 0)
    e.hook(0x554010, lambda e: gm, 0)
    r = random.Random(seed)
    for _ in range(n):
        dmg = r.choice((0, 256, 2560, r.randrange(0, 50 << 8)))
        life, mana, stam = (r.choice((0, 0, r.randrange(0, 80 << 8), r.randrange(0, 5 << 8))) for _ in range(3))
        dm, ds = r.choice((0, 256, r.randrange(0, 200 << 8))), r.choice((0, r.randrange(0, 200 << 8)))
        am = r.choice((100 << 8, 500 << 8, 2000 << 8)); al = r.randrange(0, am + 1)
        st.clear(); st[(tgt, 8)] = dm; st[(tgt, 10)] = ds; st[(att, 6)] = al; st[(att, 7)] = am
        e.mu.mem_write(rec, bytes(0x80))
        e.w32(rec + 8, dmg); e.w32(rec + 0x38, life); e.w32(rec + 0x3c, mana); e.w32(rec + 0x40, stam)
        e.call(0x57c420, tgt, ecx=0, edx=gm, regs={"eax": rec, "ebx": att})
        print(dmg, life, mana, stam, dm, ds, al, am, st.get((att, 6), 0), e.r32(rec + 0x38), e.r32(rec + 0x3c), e.r32(rec + 0x40))


def chain(path):
    """The melee chain (gate FUN_0057dbf0 -> record FUN_0057b7d0, resolver FUN_0057d4f0 -> FUN_0057c6c0 -> FUN_0057aa60) with
    the real stat code replaced by a dict (identity: setter then getter of stat 62 returns what was set, V run) on a monster
    attacker whose stats 62 / 63 are `lo << 8` / `hi << 8` (what FUN_005a1f90 stores) and a player defender with 60000 mana
    points. Case line `lo hi seed` (lo / hi whole points as d2d's Monster::mana_lo / mana_hi; seed the attacker's unit
    seed low word); prints `lo hi seed points`: the mana points the one landed hit took off the defender."""
    import life_apply
    g = life_apply.Game(real=(0x57aa60, 0x57a980, 0x57aa00, 0x57a8e0, 0x57a880))
    e = g.e
    for a_, k in ((0x57d3d0, 2), (0x5d10c0, 2), (0x57c9f0, 1), (0x57cee0, 2), (0x621e40, 3), (0x554010, 4), (0x5347c0, 0), (0x5c0c30, 3)):
        e.hook(a_, lambda e: 0, k)
    e.hook(0x622c40, lambda e: 1, 3)
    e.hook(0x5a4ee0, lambda e: 1, 0)
    e.hook(0x622d00, lambda e: 0, 1)
    e.hook(0x623c20, lambda e: 0, 0)
    e.hook(0x40b430, lambda e: e.alloc(0x100), 3)
    e.hook(0x63ee90, lambda e: 0, 1)
    e.w32(g.attacker, 1); e.w32(g.attacker + 4, 5)
    e.w32(g.attacker + 0xc, 1234); e.w32(g.target + 0xc, 4321)
    e.w32(g.attacker + 0x10, 3); e.w32(g.target + 0x10, 1)
    e.w32(g.target, 0); e.w32(g.target + 4, 1)
    big = 60000 << 8
    for line in open(path):
        lo, hi, seed = (int(x) for x in line.split())
        g.reset(30000, 0, 0)
        g.stats[(g.target, 8)] = big; g.stats[(g.target, 9)] = big
        g.stats[(g.attacker, 6)] = 1000 << 8; g.stats[(g.attacker, 7)] = 2000 << 8
        g.stats[(g.attacker, 62)] = lo << 8; g.stats[(g.attacker, 63)] = hi << 8
        g.stats[(g.attacker, 21)] = 3; g.stats[(g.attacker, 22)] = 5
        e.w32(g.attacker + 0x20, seed); e.w32(g.attacker + 0x24, 666)
        rec = e.alloc(0x100)
        e.w32(rec, 0); e.w32(rec + 4, 1)
        e.call(0x57dbf0, g.target, rec, 0x80, ecx=g.game, edx=g.attacker)
        e.call(0x57d4f0, g.target, 0, ecx=g.game, edx=g.attacker)
        print(lo, hi, seed, (big - g.stats.get((g.target, 8), 0)) / 256)


if __name__ == "__main__":
    a = sys.argv[1:]
    if a[0] == "build": build(int(a[1]), int(a[2]))
    elif a[0] == "heal": heal(int(a[1]), int(a[2]))
    elif a[0] == "chain": chain(a[1])
