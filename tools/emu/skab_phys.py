# SPDX-License-Identifier: GPL-3.0-or-later
"""SKABRECHECK claim 2: a missile's PHYSICAL roll and its hit on a monster, by game.exe's own code.

    python skab_phys.py gen N SEED > cases.tsv
    python skab_phys.py game cases.tsv > game.tsv           # game.exe 1.14d under the emulator
    test_skab_phys --phys cases.tsv game.tsv > d2d.tsv       # the REAL d2d rules::missile_blow (skab_phys_test.cpp)
    python skab_phys.py cmp cases.tsv game.tsv d2d.tsv

A case is `id lo hi res elo ehi eres seed_lo seed_hi`: the missile's physical
min / max (stats 0x15 / 0x16, 256ths), the monster's physical resistance (stat
0x24), an optional fire pair (stats 0x30 / 0x31, `0 0` = none) and the
monster's fire resistance (stat 0x27), and the missile's seed (+0x20 / +0x24).

`game` prints `id rp re lost`: FUN_005a89a0 rolls the record from the missile
stats on that seed (rp = record +8, the physical, re = record +0x10, the fire,
256ths), then FUN_0057c6c0 applies it (FUN_0057c1e0 -> FUN_0057bf80 ->
FUN_0057be00, MulDiv FUN_00483360); lost = the monster's life lost, 256ths.
Hooked (not the claim): the stat getter (a dict) and life_apply.py's unit
kinds, events and side effects, as sk_missile_hit.py does.

`d2d` prints `id A B`: A = the life d2d's missile_blow takes (damage x 256 +
frac) on the SAME seed; B = the same on the roll game.exe made (min = max =
rp / re, so no draw): B isolates the whole-point arithmetic from the roll range.
"""
import os
import sys

sys.path.insert(0, os.getcwd())

MASK = 0xFFFFFFFF
BIG_LIFE = 30000


def xorshift(state):
    x = state[0]
    x ^= (x << 13) & MASK
    x ^= x >> 17
    x ^= (x << 5) & MASK
    state[0] = x & MASK
    return state[0]


def gen(n, seed):
    st = [(seed & MASK) or 1]
    cid = 0
    out = []
    edges = (0, 1, 2, 255, 256, 257, 300, 511, 512, 1000, 2560, 25600)
    ress = (-150, -100, -50, 0, 1, 33, 50, 75, 99, 100, 150)

    def seeds():
        return xorshift(st), xorshift(st)

    for lo in edges:
        for hi in edges:
            if hi < lo:
                continue
            for res in ress:
                cid += 1
                out.append(f"{cid} {lo} {hi} {res} 0 0 0 {' '.join(map(str, seeds()))}")
    for lo in (1, 256, 300, 1000):
        for hi in (lo, lo + 1, lo + 300, lo + 5000):
            for res in (-100, 0, 50, 99):
                for elo, ehi in ((1, 1), (256, 768), (100, 4000)):
                    for eres in (-100, 0, 50, 100):
                        cid += 1
                        out.append(f"{cid} {lo} {hi} {res} {elo} {ehi} {eres} {' '.join(map(str, seeds()))}")
    for _ in range(n):
        cid += 1
        big = xorshift(st) % 10 == 0
        top = 0x180000 if big else 0x8000
        lo = xorshift(st) % top
        hi = lo + xorshift(st) % (top // 2 + 1)
        res = xorshift(st) % 301 - 150
        if xorshift(st) % 10 < 3:
            elo = 1 + xorshift(st) % top
            ehi = elo + xorshift(st) % (top // 2 + 1)
            eres = xorshift(st) % 301 - 150
        else:
            elo = ehi = eres = 0
        out.append(f"{cid} {lo} {hi} {res} {elo} {ehi} {eres} {' '.join(map(str, seeds()))}")
    return out


def parse(path):
    for line in open(path):
        f = line.split()
        if len(f) == 9:
            yield tuple(int(x) for x in f)


def s32(v):
    v &= MASK
    return v - (1 << 32) if v >> 31 else v


def game(path):
    import life_apply
    g = life_apply.Game()
    e = g.e
    missile = e.alloc(0x100)
    for cid, lo, hi, res, elo, ehi, eres, sl, sh in parse(path):
        g.reset(BIG_LIFE, 0, 0)
        g.stats[(g.target, 0x29)] = 0
        g.stats[(g.target, 0x24)] = res & MASK
        g.stats[(g.target, 0x27)] = eres & MASK
        g.stats[(missile, 0x15)] = lo & MASK
        g.stats[(missile, 0x16)] = hi & MASK
        g.stats[(missile, 0x30)] = elo & MASK
        g.stats[(missile, 0x31)] = ehi & MASK
        e.w32(missile + 0x20, sl & MASK)
        e.w32(missile + 0x24, sh & MASK)
        e.mu.mem_write(g.rec, b"\0" * 0x80)
        e.call(0x5a89a0, missile, 0, g.rec)
        rp, re_ = e.s32(g.rec + 8), e.s32(g.rec + 0x10)
        e.w32(g.rec, 1)
        e.w32(g.rec + 4, 1)
        before = g.life()
        e.call(0x57c6c0, g.target, 1, g.rec, ecx=g.game, edx=g.attacker)
        print(cid, rp, re_, before - g.life())


def cmp(cases, gpath, dpath):
    cs = {c[0]: c for c in parse(cases)}
    g = {int(l.split()[0]): l.split()[1:] for l in open(gpath) if l.strip()}
    d = {int(l.split()[0]): l.split()[1:] for l in open(dpath) if l.strip()}
    n = 0
    rows = {}
    for cid, (_, lo, hi, res, elo, ehi, eres, sl, sh) in cs.items():
        if cid not in g or cid not in d:
            continue
        n += 1
        rp, re_, gl = (int(x) for x in g[cid])
        a, b = int(d[cid][0]), int(d[cid][1])
        kind = "zero-min" if lo <= 0 or hi <= 0 else ("phys+fire" if elo > 0 else "phys")
        rb = ("res<-100" if res < -100 else "res -100..-1" if res < 0 else "res 0" if res == 0 else "res 1..99" if res < 100 else "res>=100")
        r = rows.setdefault((kind, rb), [0, 0, 0, 0, 0])
        r[0] += 1
        r[1] += a != gl
        r[2] += b != gl
        r[3] += gl - b
        r[4] += gl - a
    print(f"cases {n}; per class: n, A (same seed) differ, B (game's roll) differ, mean game-minus-d2d in points: B, A")
    tn = ta = tb = 0
    for k in sorted(rows):
        r = rows[k]
        print("  %-9s %-13s %7d %7d %7d  B %+8.3f  A %+8.3f" % (*k, r[0], r[1], r[2], r[3] / r[0] / 256, r[4] / r[0] / 256))
        tn += r[0]; ta += r[1]; tb += r[2]
    print(f"total {tn}: A differ {ta}, B differ {tb}")


def main():
    what = sys.argv[1] if len(sys.argv) > 1 else ""
    if what == "gen":
        print("\n".join(gen(int(sys.argv[2]), int(sys.argv[3]))))
    elif what == "game":
        game(sys.argv[2])
    elif what == "cmp":
        cmp(sys.argv[2], sys.argv[3], sys.argv[4])
    else:
        raise SystemExit(__doc__)


if __name__ == "__main__":
    main()
