#!/usr/bin/env python3
"""Independent verifier for upstream candidate fix/static-field (d2lod-bot card verifyprs).

Written without reading the author's oracle (tools/emu/static_field.py): the game.exe side
was re-derived from disasm.py reads of FUN_005c9800, FUN_005c96a0, FUN_0056c8e0,
FUN_00483360, FUN_0057c6c0, FUN_0057c1e0, FUN_0057be00, FUN_0057bf80, FUN_0063ee90.

game side: game.exe's own FUN_005c96a0 (the per-unit callback) runs under unicorn through
upstream's tools/emu/emu.py on a parameter block laid out as FUN_005c9800 builds it
(+0 calc2, +4 calc1, +8 StaticFieldMin or 0, +0xc length, +0x10 EType 2 = ltng); the damage
record it hands to FUN_0057c6c0 is captured (hook), then game.exe's own FUN_0057c1e0 (the
resistance step: FUN_0057c060, FUN_0057c140, FUN_0057bf80 per element, FUN_0057be00 with the
pierce, FUN_0057bf10 absorb, FUN_0063ee90 hireling test) runs on it natively. The life step
of FUN_0057c6c0 (0x57c852..0x57c874: life - record+0x4c, under 0x100 is 0) is applied here.
Hooked (not the claim): FUN_00625480 / FUN_00625500 stat getters (life 6, ltng res 0x29,
pierce 0x14e on the attacker; every other stat 0), FUN_00625d10 max life, FUN_00611d30 the
DifficultyLevels row (+0 ResistPenalty, +0x40 StaticFieldMin), FUN_00451f30 / FUN_0063edc0 /
FUN_0063e990 / FUN_0063e940 / FUN_00639df0 (flag and state tests: 0), FUN_005c0c30 events,
FUN_0057cee0 (the get-hit event). Units: attacker type 0 (a player), defender type 1 class 19.

d2d side: the REAL Fight::spot of the build under test (`test_game --verify-static`, the
mode in verify_static_field_mode.inc pasted into a throwaway worktree's test_game.cpp).

    python verify_static_field.py gen N SEED > cases.tsv
    python verify_static_field.py game cases.tsv > game.tsv          # D2_MPQ_DIR with bin/game.exe
    build/tests/test_game --verify-static cases.tsv | grep -E '^[0-9]+ ' > d2d.tsv
    python verify_static_field.py compare cases.tsv game.tsv d2d.tsv

cases.tsv: id phase difficulty expansion maxhp life res pierce slvl pct low floor
(phase 0: real Skills.txt / DifficultyLevels values, pct/low/floor are what game.exe reads;
phase 1: synthetic calc1 / calc2 / StaticFieldMin, set on both sides). Output: id life_after.
"""
import random
import struct
import sys

STAT_LIFE, STAT_LTNG_RES, STAT_LTNG_PIERCE = 6, 0x29, 0x14E
REAL_PCT, REAL_LOW = 25, 0          # Skills.txt Static Field calc1 "par4" (Param4 25), calc2 "par3" (Param3 0)
FLOORS = (0, 33, 50)                # DifficultyLevels StaticFieldMin
PENALTY = (0, -40, -100)            # DifficultyLevels ResistPenalty (+0; a monster target never reads it)


def gen(n, seed):
    rng = random.Random(seed)
    rows = []

    def add(phase, diff, exp, maxhp, life, res, pierce, slvl=1, pct=REAL_PCT, low=REAL_LOW, floor=None):
        if floor is None:
            floor = FLOORS[diff] if exp else 0
        rows.append((phase, diff, exp, maxhp, max(1, min(life, maxhp << 8)), res, pierce, slvl, pct, low, floor))

    # boundaries, exhaustive: every life in 256ths for small max life, the floor edges for larger,
    # resistances around 0 / -100 / 100 with and without pierce, both game types, all difficulties
    edge_res = (-250, -101, -100, -99, -60, -1, 0, 1, 40, 99, 100, 101, 250)
    edge_pierce = (0, 1, 40, 99, 100, 150)
    for diff in range(3):
        for exp in (0, 1):
            for res in (-100, -1, 0, 1, 100):
                for maxhp in (1, 2, 3, 4):
                    for life in range(1, (maxhp << 8) + 1):
                        add(0, diff, exp, maxhp, life, res, 0)
            for res in edge_res:
                for pierce in edge_pierce:
                    for maxhp in (100, 101, 299, 300, 3000, 65535):
                        for pct_of in (33, 50):
                            edge = maxhp * pct_of // 100
                            for whole in (edge - 1, edge, edge + 1):
                                for frac in (0, 1, 255):
                                    add(0, diff, exp, maxhp, (whole << 8) + frac, res, pierce)
                        add(0, diff, exp, maxhp, maxhp << 8, res, pierce)
    # synthetic calc1 / calc2 / floor: the life - 1 cap (pct >= 75), the least, MulDiv's big branches
    for pct in (0, 1, 74, 75, 76, 99, 100, 101, 200, 1000):
        for low in (0, 1, 255, 256, 0x10000):
            for life in (1, 2, 3, 4, 5, 256, 257, 511, 512, 5000):
                for res in (-100, -1, 0, 50):
                    add(1, rng.randrange(3), 1, 20000, life, res, rng.choice((0, 0, 30)), 1, pct, low, rng.choice((0, 33, 50, 99, 100)))
    # random: N on top of the boundary cases
    n += len(rows)
    while len(rows) < n:
        diff, exp = rng.randrange(3), rng.choice((0, 1, 1))
        maxhp = rng.choice((rng.randint(1, 50), rng.randint(1, 2000), rng.randint(1, 60000), rng.randint(60000, 2000000)))
        life = rng.choice((rng.randint(1, maxhp << 8), (rng.randint(0, maxhp) << 8) + rng.choice((0, 1, 128, 255)),
                           (maxhp * rng.choice((33, 50)) // 100 + rng.randint(-2, 2)) << 8))
        res = rng.choice((rng.randint(-100, 99), rng.randint(-300, 300), rng.randint(95, 200), 0))
        pierce = rng.choice((0, 0, rng.randint(0, 100), rng.randint(0, 300)))
        if rng.random() < 0.1:
            add(1, diff, exp, maxhp, life, res, pierce, rng.randint(1, 40), rng.choice((rng.randint(0, 100), rng.randint(0, 400))),
                rng.choice((0, rng.randint(0, 0x4000))), rng.choice((0, 33, 50, rng.randint(0, 100))))
        else:
            add(0, diff, exp, maxhp, life, res, pierce, rng.randint(1, 60))
    for i, r in enumerate(rows):
        print(i, *r, sep="\t")


def game(path):
    sys.path.insert(0, ".")          # run from upstream's tools/emu
    from emu import Emu
    emu = Emu()
    game_ = emu.alloc(0x200)
    attacker = emu.alloc(0x200)
    defender = emu.alloc(0x200)
    diffrec = emu.alloc(0x100)
    params = emu.alloc(0x40)
    ctx = emu.alloc(0x40)
    emu.w32(attacker, 0)                  # player
    emu.w32(defender, 1)                  # monster
    emu.w32(defender + 4, 19)             # fallen1's class: not a hireling (FUN_0063ee90)
    st = {}
    cap = {}

    def stat(e):
        unit, sid = e.arg(0), e.arg(1)
        return st.get((unit, sid), 0) & 0xFFFFFFFF

    emu.hook(0x625480, stat, 3)
    emu.hook(0x625500, stat, 3)
    emu.hook(0x625d10, lambda e: st[("max", e.arg(0))] & 0xFFFFFFFF, 1)
    emu.hook(0x611d30, lambda e: diffrec, 1)
    for a, n in ((0x63edc0, 1), (0x63e990, 1), (0x63e940, 1), (0x639df0, 2), (0x63e9f0, 2)):
        emu.hook(a, lambda e: 0, n)
    emu.hook(0x451f30, lambda e: 0, 0)
    emu.hook(0x5c0c30, lambda e: 0, 3)
    emu.hook(0x57cee0, lambda e: 0, 2)

    def apply(e):                          # FUN_0057c6c0(ecx game, edx attacker; unit, 1, record)
        cap["unit"], cap["rec"] = e.arg(0), e.arg(2)
        cap["copy"] = e.read(e.arg(2), 0x70)
        return 0

    emu.hook(0x57c6c0, apply, 3)
    record = emu.alloc(0x70)
    for line in open(path):
        f = line.split()
        if not f:
            continue
        cid, phase, diff, exp, maxhp, life, res, pierce, slvl, pct, low, floor = map(int, f)
        emu.mu.mem_write(game_ + 0x6d, bytes([diff]))
        emu.w32(game_ + 0x70, exp)
        emu.w32(diffrec, PENALTY[diff])
        emu.w32(diffrec + 0x40, FLOORS[diff])
        st.clear()
        st[(defender, STAT_LIFE)] = life
        st[("max", defender)] = maxhp << 8
        st[(defender, STAT_LTNG_RES)] = res
        st[(attacker, STAT_LTNG_PIERCE)] = pierce
        # FUN_005c9800's block: +0 calc2, +4 calc1, +8 the floor (expansion only), +0xc length, +0x10 EType
        for off, v in ((0, low), (4, pct), (8, floor if exp else 0), (0xC, 0), (0x10, 2)):
            emu.w32(params + off, v)
        emu.w32(ctx, game_)
        emu.w32(ctx + 4, attacker)
        emu.w32(ctx + 0xC, params)
        cap.clear()
        emu.call(0x5c96a0, ecx=ctx, edx=defender)
        if "copy" not in cap:
            print(cid, life)               # no record: nothing happens to its life
            continue
        emu.mu.mem_write(record, cap["copy"])
        emu.call(0x57c1e0, defender, record, ecx=game_, edx=attacker)
        total = struct.unpack("<i", emu.read(record + 0x4C, 4))[0]
        after = life
        if total > 0:                      # 0x57c852: total > 0 -> life - total, under 0x100 is 0
            after = life - total
            if after < 0x100:
                after = 0
        print(cid, after)


def compare(cases, a, b):
    rows = {l.split()[0]: l.split() for l in open(cases) if l.strip()}
    ga = dict(l.split()[:2] for l in open(a) if l.strip())
    gb = dict(l.split()[:2] for l in open(b) if l.strip())
    flags = [l for l in open(b) if "RESFAIL" in l]
    bad = [k for k in rows if ga.get(k) != gb.get(k)]
    per = {}
    for k in bad:
        per[rows[k][1]] = per.get(rows[k][1], 0) + 1
    print(f"cases {len(rows)} game {len(ga)} d2d {len(gb)} mismatches {len(bad)} per-phase {per} resfail {len(flags)}")
    for k in bad[:15]:
        print("  ", " ".join(rows[k]), "game", ga.get(k), "d2d", gb.get(k))


if __name__ == "__main__":
    cmd = sys.argv[1] if len(sys.argv) > 1 else ""
    if cmd == "gen":
        gen(int(sys.argv[2]), int(sys.argv[3]))
    elif cmd == "game":
        game(sys.argv[2])
    elif cmd == "compare":
        compare(*sys.argv[2:5])
    else:
        sys.exit(__doc__)
