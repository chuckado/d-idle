#!/usr/bin/env python3
"""game.exe 1.14d's plain Attack (the player's melee blow) as the oracle (d2lod-bot, card skmelee).

Runs game.exe's own code under unicorn through upstream's harness (psi29a/d2decomped
tools/emu/emu.py): copy this file into that directory and run it from there:

    cd <d2decomped>/tools/emu
    D2_MPQ_DIR=<mpq dir> python sk_melee.py --exe <game.exe> [cases] [seed]

Three parts, each against a port of game.exe's reading and of d2d's rule
(components/rules/combat.hpp `hit_chance`, `make_fighter`, `player_blow`, upstream main 72132b99):

- `hit`: FUN_0057d9b0 for a player (type 0) at a monster (type 1), melee (missile flag 0), as
  the plain Attack calls it (FUN_0056f070 -> FUN_0057ec10, bonus = stat 325). Hooked: the stat
  getters, FUN_00622560 (the attack rating, read: stat 19 + (dex - 7) x 5 + CharStats
  ToHitFactor), FUN_0057d8a0 (the 115 / 116 / 123 / 124 modifiers: none here), FUN_00645830
  (mastery to-hit), FUN_006223f0 (defense), FUN_00451f80 (no MonStats row: no stat 179 list).
  The seed draw runs natively on the attacker's unit + 0x20.
- `dmg`: FUN_0057b420 (ECX the player, EDX 1, stack: weapon 0, min 0, max 0, skill ED %,
  flat, SrcDam), as FUN_0057b7d0 calls it for the plain Attack (ED 0, flat 0, SrcDam 0x80).
  Hooked: the stat getters, FUN_00535bc0 (the weapon in hand), FUN_0063d340 (1: the first
  hand), FUN_00629860 / FUN_006298a0 (the weapon's StrBonus / DexBonus), FUN_00645830
  (mastery damage). The roll (FUN_0045c3e0 on unit + 0x20) runs natively.
  Gear model (reading A, the one d2d follows): the player's stat 21 / 22 = the weapon's
  base x (100 + ED) / 100 (ItemStatCost op 13 on the item, FUN_00626200) + the flat +min / +max
  of all gear; the player's stats 17 / 18 = 0. Whether 17 / 18 reach the player is U (trace Q3).
- `undead`: FUN_0057b660 (EAX target, EDI attacker, EBX the damage record): +50 to the
  record's ED (+0xc) for a Blunt weapon (ItemTypes 57, FUN_00629bb0) on an undead target
  (FUN_0063e990), plus stat 122, when the sum is over 0. Exhaustive grid.

Output: per part the case count, the lines that differ from the reading (should be 0) and from d2d.
"""
import random
import sys

import emu


def lcg(seed):
    product = seed[0] * 0x6AC690C5 + seed[1]
    seed[0], seed[1] = product & 0xFFFFFFFF, product >> 32
    return seed[0]


def rand(seed, bound):
    if bound < 1: return 0
    v = lcg(seed)
    return v & (bound - 1) if bound & (bound - 1) == 0 else v % bound


def tdiv(a, b):
    """C integer division (toward zero)."""
    q = abs(a) // abs(b)
    return q if (a >= 0) == (b > 0) else -q


def s32(v):
    v &= 0xFFFFFFFF
    return v - (1 << 32) if v >= 1 << 31 else v


def muldiv(value, times, over):
    """FUN_00483360 for the small values used here (|value x times| < 2^31): C division."""
    return tdiv(value * times, over) if over else 0


def hit_chance(attack, defense, alvl, dlvl):
    """d2d rules::hit_chance (combat.hpp:28) == FUN_0057d9b0's arithmetic."""
    if defense < 0: attack, defense = attack - defense, 0
    if attack < 0: defense, attack = defense - attack, 0
    chance = tdiv(attack * 100, attack + defense) if attack + defense else 100
    chance = tdiv(chance * 2 * alvl, max(alvl + dlvl, 1))
    return max(5, min(95, chance))


def hit_part(e, stats, rng, cases):
    st = {}
    e.hook(0x622560, lambda e: st["ar"], 1)
    e.hook(0x57d8a0, lambda e: 0, 2)
    e.hook(0x645830, lambda e: st["mastery"], 4)
    e.hook(0x451f80, lambda e: 0, 0)
    e.hook(0x535bc0, lambda e: st["weapon"], 0)
    bad_read = bad_d2d = 0
    cat = {"ar<=0": 0, "pct<0": 0, "other": 0}
    edges = []
    for ar in (-40, -1, 0, 1, 7, 99, 100, 101):
        for pct in (-90, -33, -1, 0, 1, 33):
            for d in (-10, 0, 1, 50, 100):
                edges.append((ar, pct, d))
    for case in range(len(edges) + cases):
        att, dfd = e.alloc(0x100), e.alloc(0x100)
        e.w32(att, 0); e.w32(dfd, 1)
        seed = [rng.getrandbits(32), rng.getrandbits(32)]
        e.w32(att + 0x20, seed[0]); e.w32(att + 0x24, seed[1])
        if case < len(edges):
            ar, pct, defense = edges[case]; mastery, bonus, melee_def = 0, 0, 0
        else:
            ar = rng.randrange(-40, 2500)
            defense = rng.randrange(-30, 1500)
            pct = rng.choice((0, 0, 0, 10, 25, 50, 100, 200, -10, -25))
            mastery = rng.choice((0, 0, 0, 28, 61))
            bonus = rng.choice((0, 0, 0, 0, 30, 100))          # stat 325 (progressive to-hit) passed as the bonus
            melee_def = rng.choice((0, 0, 0, 10))
        alvl, dlvl = rng.randrange(1, 99), rng.randrange(1, 99)
        weapon = rng.choice((0, 0x7000000))
        st.update(ar=ar, mastery=mastery, weapon=weapon)
        stats.clear()
        stats.update({(att, 0xc): alvl, (att, 0x77): pct, (dfd, 0xc): dlvl, (dfd, "def"): defense, (dfd, 0x21): melee_def})
        got = e.call(0x57d9b0, bonus, 0, ecx=att, edx=dfd)
        # The reading: AR' = AR + MulDiv(AR, mastery (melee, a weapon) + 119 + bonus, 100).
        total_pct = (mastery if weapon else 0) + pct + bonus
        attack = ar + muldiv(ar, total_pct, 100)
        roll = lcg(list(seed)) % 100
        want = 1 if roll < hit_chance(attack, defense + melee_def, alvl, dlvl) else 0
        # d2d: max(ar_base x (100 + ar_pct + swing.ar_pct) / 100, 1), ar_pct = 119 + 342 (mastery, any weapon).
        d2d_attack = max(tdiv(ar * (100 + pct + (mastery if weapon else 0) + bonus), 100), 1)
        d2d = 1 if roll < hit_chance(d2d_attack, defense + melee_def, alvl, dlvl) else 0
        if got != want:
            bad_read += 1
            print(f"hit case {case}: ar {ar} pct {total_pct} def {defense} lv {alvl}/{dlvl}: game {got}, reading {want}")
        if got != d2d:
            bad_d2d += 1
            cat["ar<=0" if ar + muldiv(ar, total_pct, 100) <= 0 else "pct<0" if total_pct < 0 else "other"] += 1
            if bad_d2d <= 20: print(f"hit d2d case {case}: ar {ar} pct {total_pct} def {defense}+{melee_def} lv {alvl}/{dlvl}: game {got}, d2d {d2d} (AR {attack} vs {d2d_attack})")
    print(f"hit: {len(edges)} boundary + {cases} random cases, {bad_read} differ from the reading, {bad_d2d} from d2d {cat}")


def dmg_reading(s21, s22, n111, p_gear, e_skill, s17, s18, flat, srcdam, seed):
    """FUN_0057b420 as read (stats path, EDX 1), in 256ths."""
    lo, hi = s21 << 8, s22 << 8
    lo += n111 << 8; hi += n111 << 8
    if lo < 1: lo = 0x100
    if hi <= lo: hi = lo + 0x100
    p = max(e_skill + p_gear, -90)
    out = flat
    if hi > 0:
        a = lo + muldiv(lo, s18 + p, 100)
        b = hi + muldiv(hi, s17 + p, 100)
        out = a + flat
        if b > a: out += rand(seed, b - a)
    out = max(out, 0)
    return muldiv(out, srcdam, 128) if srcdam != 0x80 else out


def d2d_damage(bm, bM, ed_min, ed_max, fmin, fmax, n111, dp, mastery, str_, dex, sb, db, weapon, e_skill, seed):
    """d2d make_fighter + player_blow's physical (no crit, SrcDam 128, no flat): the whole points
    before resistance, and the 256ths it rolled (before the >> 8)."""
    low = (tdiv(bm * (100 + ed_min), 100) if weapon else 1) + fmin + n111
    high = (tdiv(bM * (100 + ed_max), 100) if weapon else 2) + fmax + n111
    low = max(low, 1) << 8
    high = max(high << 8, low + 256)
    pct = max(dp + mastery + ((tdiv(str_ * sb, 100) + tdiv(dex * db, 100)) if weapon else 0), -90)
    pp = max(pct + e_skill, -90)
    lo = low + tdiv(low * pp, 100)
    hi = high + tdiv(high * pp, 100)
    fine = lo + rand(seed, hi - lo) if hi > lo else lo
    return max(fine >> 8, 1), fine


def dmg_part(e, stats, rng, cases):
    st = {}
    e.hook(0x535bc0, lambda e: st["weapon"], 0)
    e.hook(0x63d340, lambda e: 1, 2)
    e.hook(0x629860, lambda e: st["sb"], 1)
    e.hook(0x6298a0, lambda e: st["db"], 1)
    e.hook(0x645830, lambda e: st["mastery"], 4)
    bad_read = whole = fine_diff = 0
    by = {"weapon": [0, 0, 0], "barehanded": [0, 0, 0]}
    for case in range(cases):
        att = e.alloc(0x100)
        e.w32(att, 0)
        seed = [rng.getrandbits(32), rng.getrandbits(32)]
        e.w32(att + 0x20, seed[0]); e.w32(att + 0x24, seed[1])
        weapon = 0x7000000 if rng.random() < 0.85 else 0
        bm = rng.randrange(1, 40); bM = bm + rng.randrange(1, 60)        # Weapons.txt mindam / maxdam
        ed_min = ed_max = rng.choice((0, 0, 0, 0, 30, 50, 100, 200, 300))
        fmin = rng.choice((0, 0, 0, 1, 2, 5)); fmax = rng.choice((0, 0, 0, 1, 3, 10))
        n111 = rng.choice((0, 0, 0, 0, 1, 2))
        dp = rng.choice((0, 0, 0, 0, 20, 100, -25))                        # stat 25 (Might, gear)
        str_ = rng.randrange(10, 200); dex = rng.randrange(10, 200)
        sb, db = rng.choice(((100, 0), (100, 0), (110, 0), (0, 100), (75, 75), (0, 0)))   # StrBonus / DexBonus
        mastery = rng.choice((0, 0, 0, 0, 28, 100)) if weapon else 0
        e_skill = rng.choice((0, 0, 0, 0, 0, 50, 150))
        s21 = (tdiv(bm * (100 + ed_min), 100) if weapon else 0) + fmin
        s22 = (tdiv(bM * (100 + ed_max), 100) if weapon else 0) + fmax
        st.update(weapon=weapon, sb=sb, db=db, mastery=mastery)
        stats.clear()
        stats.update({(att, 0x15): s21, (att, 0x16): s22, (att, 0x6f): n111, (att, 0x19): dp, (att, 0): str_, (att, 2): dex})
        got = s32(e.call(0x57b420, 0, 0, 0, e_skill, 0, 0x80, ecx=att, edx=1))
        p_gear = dp + mastery + ((tdiv(str_ * sb, 100) + tdiv(dex * db, 100)) if weapon else str_)
        lo21, hi22 = (s21, s22) if weapon else (s21 if s21 > 0 else 1, s22 if s22 > 1 else 2)
        want = dmg_reading(lo21, hi22, n111, p_gear, e_skill, 0, 0, 0, 0x80, list(seed))
        if got != want:
            bad_read += 1
            print(f"dmg case {case}: weapon {bool(weapon)} {s21}-{s22} n {n111} p {p_gear} e {e_skill}: game {got}, reading {want}")
        d2d_whole, d2d_fine = d2d_damage(bm, bM, ed_min, ed_max, fmin, fmax, n111, dp, mastery, str_, dex, sb, db, weapon, e_skill, list(seed))
        kind = "weapon" if weapon else "barehanded"
        by[kind][0] += 1
        if max(got >> 8, 1) != d2d_whole:
            whole += 1; by[kind][1] += 1
            if whole <= 10: print(f"dmg d2d case {case}: {kind} {s21}-{s22} str {str_} dp {dp} e {e_skill}: game {got / 256:.3f}, d2d {d2d_whole} (rolled {d2d_fine / 256:.3f})")
        if got != d2d_whole << 8: fine_diff += 1
        by[kind][2] += got - (d2d_whole << 8)
    print(f"dmg: {cases} cases, {bad_read} differ from the reading; vs d2d: {whole} differ in whole points "
          f"(weapon {by['weapon'][1]} of {by['weapon'][0]}, barehanded {by['barehanded'][1]} of {by['barehanded'][0]}), "
          f"{fine_diff} differ in 256ths; mean game - d2d a landed blow: weapon {by['weapon'][2] / max(by['weapon'][0], 1) / 256:.4f} points, "
          f"barehanded {by['barehanded'][2] / max(by['barehanded'][0], 1) / 256:.4f}")


def undead_part(e, stats):
    st = {}
    e.hook(0x63e990, lambda e: st["undead"], 1)
    e.hook(0x535bc0, lambda e: st["weapon"], 0)
    e.hook(0x629bb0, lambda e: 1 if (st["blunt"] and e.arg(1) == 57) else 0, 2)
    bad = n = 0
    for undead in (0, 1):
        for weapon, blunt in ((0, 0), (0x7000000, 0), (0x7000000, 1)):
            for s122 in (-60, -1, 0, 1, 50):
                for ed0 in (0, 35):
                    att, tgt, rec = e.alloc(0x100), e.alloc(0x100), e.alloc(0x100)
                    e.w32(rec + 0xc, ed0)
                    st.update(undead=undead, weapon=weapon, blunt=blunt)
                    stats.clear(); stats.update({(att, 0x7a): s122})
                    e.call(0x57b660, regs={"eax": tgt, "edi": att, "ebx": rec})
                    got = s32(e.r32(rec + 0xc))
                    add = (50 if blunt else 0) + s122
                    want = ed0 + (add if undead and add > 0 else 0)
                    n += 1
                    if got != want:
                        bad += 1
                        print(f"undead {undead} weapon {bool(weapon)} blunt {blunt} s122 {s122} ed {ed0}: game {got}, reading {want}")
    print(f"undead: {n} cases (exhaustive grid), {bad} differ from the reading; d2d adds nothing in every case with a non-zero bonus")


def main():
    argv, exe, args = sys.argv[1:], None, []
    while argv:
        a = argv.pop(0)
        if a == "--exe": exe = argv.pop(0)
        else: args.append(a)
    cases = int(args[0]) if args else 2000
    seed = int(args[1]) if len(args) > 1 else 7
    rng = random.Random(seed & 0xFFFFFFFF)
    e = emu.Emu(exe) if exe else emu.Emu()
    stats = {}
    e.hook(0x625480, lambda e: stats.get((e.arg(0), e.arg(1)), 0), 3)
    e.hook(0x625500, lambda e: stats.get((e.arg(0), e.arg(1)), 0), 3)
    e.hook(0x6223f0, lambda e: stats.get((e.arg(0), "def"), 0), 1)
    e.hook(0x639df0, lambda e: 0, 2)
    e.hook(0x554010, lambda e: 0, 3)
    e.hook(0x5347b0, lambda e: 0, 0)
    e.hook(0x57d810, lambda e: 0, 0)
    print(f"sk_melee: seed {seed}, {cases} random cases a part")
    hit_part(e, stats, rng, cases)
    dmg_part(e, stats, rng, cases)
    undead_part(e, stats)


if __name__ == "__main__": main()
