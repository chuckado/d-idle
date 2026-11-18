# SPDX-License-Identifier: GPL-3.0-or-later
#!/usr/bin/env python3
"""MELBLOCK (card decomp27): may a monster's melee swing be blocked?  game.exe 1.14d FUN_0057ec10 against d2d's rules::monster_blow.

Run from a checkout of upstream tools/emu (needs emu.py); EMU = that directory, GAMEEXE = game.exe 1.14d:

    EMU=... GAMEEXE=... python3 mel_block.py game N > game.tsv     # FUN_0057ec10 natively under unicorn
    d2d: test_melee_block > d2d.tsv                                 # fork branch fid/melee-block-r1 (oracle commit, no claim)
    python3 mel_block.py compare game.tsv d2d.tsv

`game`: FUN_0057ec10 (the melee hit entry every monster swing and every monster skill's hit goes through) is run for real with a
monster attacker (unit type 1) and a player defender (unit type 0) in each of the modes 0..18 and each defender block chance of the grid.
FUN_0057dfb0 and its callees FUN_00622d00 (the moving test) run for real; hooked: the block chance FUN_00622720 (the case's chance),
the dodge FUN_0057dd60 (0), the to-hit FUN_0057d9b0 (1: the melee to-hit is the earlier monster_hit.py oracle's), the reach FUN_00622c40,
the bookkeeping FUN_00554010 / FUN_005347c0 / FUN_00554200 / FUN_00639df0. The defender's seed (+0x20 / +0x24) takes TRIALS distinct
values. Output `G mode chance blocked trials hit` (hit: the result's bit 0 set, i.e. the swing hit through the block). A hook on
FUN_0057dfb0 also reads its argument list: `A mode can_block missile` (must be 1 / 0 for every case: the melee entry's constants).

`compare`: d2d's `D mode chance blocked trials` (monster_blow, a Fighter with the chance as Block, to-hit sure, defender motion by mode:
WL walking, RN running, every other mode standing; town walk TW is excluded: d2d has no town fight). A cell passes when d2d's blocked
count and game.exe's both lie within 5 sigma of the exact expectation p = chance (a third, truncating, for RN and TW) and chance 0 is exactly 0
on both sides. The game's expectation is also checked against the seed-exact LCG (FUN_0057dfb0's roll: the defender's 64-bit LCG, rand % 100)
so a deviation of the game from the read rule shows as `GAMEREAD`.
"""
import math
import os
import sys

NAMES = "DT NU WL RN GH TN TW A1 A2 BL SC TH KK S1 S2 S3 S4 DD KB".split()
CHANCES = (0, 1, 5, 9, 10, 20, 25, 33, 50, 75)
TRIALS = 400
M32 = 0xFFFFFFFF


def lcg_roll(seed):
    """FUN_0057dfb0 0x57e021..0x57e04b: seed pair (lo = +0x20, hi = +0x24); lo*0x6ac690c5 + hi (64-bit), new hi = high word, new lo = low word."""
    lo, hi = seed
    prod = lo * 0x6ac690c5 + hi
    nlo, nhi = prod & M32, (prod >> 32) & M32
    return (nlo % 100, (nlo, nhi))


def seeds(n):
    x = 0x12345679
    out = []
    for _ in range(n):
        x ^= (x << 13) & M32; x ^= x >> 17; x ^= (x << 5) & M32
        a = x
        x ^= (x << 13) & M32; x ^= x >> 17; x ^= (x << 5) & M32
        out.append((a, x))
    return out


def effective(chance, mode):
    """The read rule: a third (signed truncation) when 622d00 holds and the mode is not WL; 622d00 is run for real, so the game decides."""
    return chance // 3 if mode in (3, 6) else chance


def game(n_trials):
    sys.path.insert(0, os.environ["EMU"])
    import emu
    e = emu.Emu(os.environ["GAMEEXE"])
    st = {"chance": 0, "args": None}
    gm = e.alloc(0x2000)
    mon = e.alloc(0x200)
    ply = e.alloc(0x200)
    e.w32(mon, 1); e.w32(mon + 4, 60)
    e.w32(ply, 0); e.w32(ply + 4, 3)
    e.hook(0x622720, lambda e_: st["chance"], 2)
    e.hook(0x57dd60, lambda e_: 0, 1)
    e.hook(0x57d9b0, lambda e_: 1, 2)
    e.hook(0x622c40, lambda e_: 1, 3)
    e.hook(0x554010, lambda e_: 0, 4)
    e.hook(0x5347c0, lambda e_: 0, 0)
    e.hook(0x554200, lambda e_: 1, 1)
    e.hook(0x639df0, lambda e_: 0, 2)
    e.hook(0x625480, lambda e_: 0, 3)
    sd = seeds(n_trials)
    for mode in range(0, 19):
        for chance in CHANCES:
            st["chance"] = chance
            blocked = hit = 0
            for lo, hi in sd:
                e.w32(ply + 0x10, mode)
                e.w32(ply + 0x20, lo); e.w32(ply + 0x24, hi)
                ax = e.call(0x57ec10, ply, 0, 0, ecx=gm, edx=mon) & 0xffff
                blocked += 1 if ax & 0x10 else 0
                hit += ax & 1
            print("G %d %d %d %d %d" % (mode, chance, blocked, n_trials, hit))
    # the argument list FUN_0057dfb0 gets from the melee entry: hook it once, run every mode
    seen = set()
    e.hook(0x57dfb0, lambda e_: seen.add((e_.arg(1), e_.arg(2))) or 0, 3)
    for mode in range(0, 19):
        e.w32(ply + 0x10, mode)
        e.call(0x57ec10, ply, 0, 0, ecx=gm, edx=mon)
    for missile, can_block in sorted(seen):
        print("A %d %d" % (can_block, missile))
    # the read rule against the seed-exact LCG, per cell
    for mode in range(0, 19):
        for chance in CHANCES:
            c = effective(chance, mode)
            exp = 0
            for s in sd:
                r, _ = lcg_roll(s)
                exp += 1 if (c > 0 and r < c) else 0
            print("R %d %d %d" % (mode, chance, exp))


def sigma_ok(count, trials, p):
    mean = trials * p / 100.0
    sd = math.sqrt(trials * (p / 100.0) * (1 - p / 100.0))
    return abs(count - mean) <= 5 * sd + 1e-9


def compare(gpath, dpath):
    g, r, a, d = {}, {}, set(), {}
    for line in open(gpath):
        f = line.split()
        if f and f[0] == "G" and len(f) == 6: g[(int(f[1]), int(f[2]))] = (int(f[3]), int(f[4]))
        elif f and f[0] == "R" and len(f) == 4: r[(int(f[1]), int(f[2]))] = int(f[3])
        elif f and f[0] == "A" and len(f) == 3: a.add((int(f[1]), int(f[2])))
    for line in open(dpath):
        f = line.split()
        if f and f[0] == "D" and len(f) == 5: d[(int(f[1]), int(f[2]))] = (int(f[3]), int(f[4]))
    if not g or not d:
        print("no cases"); return 1
    bad = n = 0
    if a != {(1, 0)}:
        bad += 1
        print("ARGS: FUN_0057dfb0 got %s from the melee entry, expected only can_block 1 / missile 0" % sorted(a))
    for (mode, chance), (gb, gt) in sorted(g.items()):
        if mode == 6:
            print("excluded TW (town walk: d2d has no town fight): game %d of %d at %d" % (gb, gt, chance)) if chance == 75 else None
            continue
        if (mode, chance) not in d:
            bad += 1; print("MISSING d2d cell mode %s chance %d" % (NAMES[mode], chance)); continue
        db, dt = d[(mode, chance)]
        eff = effective(chance, mode)
        n += 1
        why = []
        if r.get((mode, chance)) != gb: why.append("GAMEREAD game %d, read rule %s" % (gb, r.get((mode, chance))))
        if eff == 0 and (gb or db): why.append("chance 0 blocked: game %d d2d %d" % (gb, db))
        if not sigma_ok(gb, gt, eff): why.append("game %d of %d outside 5 sigma of %d%%" % (gb, gt, eff))
        if not sigma_ok(db, dt, eff): why.append("d2d %d of %d outside 5 sigma of %d%%" % (db, dt, eff))
        if why:
            bad += 1; print("DIFF mode %s chance %d: %s" % (NAMES[mode], chance, "; ".join(why)))
    print("%d cells, %d differ" % (n, bad))
    return 1 if bad or n == 0 else 0


if __name__ == "__main__":
    if len(sys.argv) >= 2 and sys.argv[1] == "game":
        game(int(sys.argv[2]) if len(sys.argv) > 2 else TRIALS)
    elif len(sys.argv) == 4 and sys.argv[1] == "compare":
        sys.exit(compare(sys.argv[2], sys.argv[3]))
    else:
        sys.exit(__doc__)
