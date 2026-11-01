#!/usr/bin/env python3
"""Independent verifier for upstream candidate fix/static-field-chill (d2lod-bot card verifychill).

Written before reading the author's oracle (tools/emu/chill_length.py): the game.exe rule was
re-derived from disasm.py reads of FUN_0057c6c0, FUN_0057c1e0, FUN_0057c140, FUN_0057be00,
FUN_0057bf80, FUN_0057bf10, FUN_00483360, FUN_0057af80, FUN_0057af30, FUN_0057b230,
FUN_0063ee90, FUN_0063e9f0, FUN_005a0180, FUN_005a89a0 and the table at 0x732980.

game side: game.exe's own FUN_0057c6c0 runs under unicorn (upstream tools/emu/emu.py) on a
damage record (+0x24 cold damage, +0x30 cold length, +0x34 freeze length, the rest 0), with a
player attacker (type 0) and a monster defender (type 1). Native, unhooked: FUN_0057c6c0,
FUN_0057c1e0 with FUN_0057c060 / FUN_0057c140 / FUN_0057be00 / FUN_0057bf80 / FUN_0057bf10 /
FUN_00483360, FUN_0057af80, FUN_0057b230, FUN_0057af30, FUN_0063ee90, FUN_0063e9f0,
FUN_005a0180. Hooked (not the claim; every one listed): FUN_00625480 / FUN_00625500 stat
getters (defender: life 6 = 10^6 points, cold res 0x2b; attacker: cold pierce 0x14f; every
other stat 0), FUN_00627260 stat setter (no-op), FUN_00611d30 the DifficultyLevels row
(+0 ResistPenalty, +0x14 MonsterFreezeDivisor, +0x18 MonsterColdDivisor, from the MPQ's
DifficultyLevels.txt), FUN_00451f80 the MonStats row of a class (a fake row: +0xc boss bit =
[0x6ce280], +0x168 + difficulty the case's coldeffect; the same rows hang off the fake
[0x744304] table that FUN_0057af80 / FUN_0057af30 index directly), FUN_00457490 (1),
FUN_00554200 (1), FUN_00620bb0 (1), FUN_0061ab00 (0), FUN_005541b0 (0), FUN_00621d50,
FUN_0057c420, FUN_0057a980, FUN_0057aa60, FUN_0057aaa0, FUN_0057aae0 (stun, record +0x44 = 0),
FUN_0057ac50, FUN_0057add0, FUN_005d6410, FUN_00540e60, FUN_005c0c30 events, FUN_005347d0 /
FUN_005347e0 logging, FUN_0063e990 / FUN_0063e940 / FUN_00639df0 (0), FUN_004638a0 (1),
FUN_00554010 (the game), and the state-list calls that only carry the end frame:
FUN_006256b0 (the unit's list for a state, or 0), FUN_006251f0 (a new list: its end),
FUN_006252d0 (its state id), FUN_00625330 / FUN_00625310 (get / set the end), FUN_00639db0,
FUN_005417d0, FUN_00625ce0, FUN_00626e10, FUN_006270b0, FUN_00623f50 (no-ops).
Output: the ticks to the end of state 0xb (chill) and of state 1 (freeze), or `-`.

d2d side: test_game --verify-chill (verify_chill_mode.inc pasted into a throwaway worktree's
test_game.cpp): the REAL missile_blow + Fight::land (phase 0) or Fight::strike with a freeze
(phase 1).

    python verify_chill.py gen N SEED > cases.tsv
    python verify_chill.py game cases.tsv > game.tsv      # D2_MPQ_DIR with bin/game.exe
    python verify_chill.py patch TEST_GAME_CPP            # paste the mode
    build/tests/test_game --verify-chill cases.tsv | grep -E '^[0-9]+ ' > d2d.tsv
    python verify_chill.py compare cases.tsv game.tsv d2d.tsv

cases.tsv: id phase difficulty coldeffect res pierce roll length prior uflag bcol hire freeze boss
(roll in 256ths; uflag: monster data +0x16; bcol: the MonStats boss bit; hire: a hireling
class on the game side only; boss: d2d's Boss kind, matched to uflag).
"""
import os
import random
import struct
import sys

STAT_LIFE, STAT_COLD_RES, STAT_COLD_PIERCE = 6, 0x2B, 0x14F
FALLEN, HIRELINGS = 19, (0x10F, 0x152, 0x167, 0x230, 0x231)
FRAME = 2000
# d2d Boss kinds (rules/uniques.hpp) and the monster-data flag bits the common D2 layout gives
# them; only the game side reads uflag, only the d2d side reads boss.
BOSS_FLAGS = {0: 0, 1: 0x4, 2: 0x8, 3: 0xA, 4: 0x10}


def gen(n, seed):
    rng = random.Random(seed)
    rows = []

    def add(phase, diff, ce, res, pierce, roll, length, prior=0, boss=0, bcol=0, hire=0, freeze=0):
        rows.append((phase, diff, ce, res, pierce, roll, length, prior, BOSS_FLAGS[boss], bcol, hire, freeze, boss))

    # phase 0, exhaustive boundaries: whole-point edges of the roll x lengths x resistances
    rolls = (-256, -1, 0, 1, 127, 255, 256, 257, 511, 512, 25600)
    lengths = (-50, -1, 0, 1, 2, 3, 4, 5, 8, 99, 100, 101, 199, 200, 201, 1000)
    reses = (-300, -150, -101, -100, -99, -50, -1, 0, 1, 2, 33, 50, 67, 75, 95, 98, 99, 100, 101, 150, 250)
    for diff in range(3):
        for ce in (-128, -100, -50, -1, 0, 1, 50, 127):
            for res in reses:
                for roll in rolls:
                    for length in lengths:
                        add(0, diff, ce, res, 0, roll, length)
        for res in reses:
            for pierce in (1, 2, 50, 99, 100, 101, 250):
                for roll in (1, 255, 256, 2560):
                    for length in (1, 2, 50, 101, 375):
                        add(0, diff, -50, res, pierce, roll, length)
        for prior in (1, 2, 24, 25, 26, 49, 50, 51, 400):
            for res in (0, 50, 99):
                for roll in (1, 255, 512):
                    for length in (1, 25, 50, 51, 100):
                        add(0, diff, rng.choice((-50, -50, -25, 10)), res, 0, roll, length, prior)
        for boss in BOSS_FLAGS:
            for bcol in (0, 1):
                for roll in (1, 255, 256):
                    for length in (0, 1, 50):
                        add(0, diff, -50, rng.choice((0, 50, 99)), 0, roll, length, rng.choice((0, 30)), boss, bcol)
    # phase 0, huge lengths (resisted() vs FUN_00483360 over 0x100000, int overflow)
    for length in (0xFFFFF, 0x100000, 0x100001, 0x123457, 0x7FFFFF, 0xA3D70A, 0xA3D70B, 0xFFFFFF):
        for res in (-100, -50, -1, 1, 50, 99):
            add(0, rng.randrange(3), -50, res, 0, 128, length)
    # phase 1 (freeze via FUN_0057b230 / Fight::strike): every kind x boss bit x hireling
    for diff in range(3):
        for boss in BOSS_FLAGS:
            for bcol in (0, 1):
                for hire in (0, 1):
                    for res in (0, 50, 100):
                        for freeze in (0, 1, 2, 7, 50):
                            for ce in (-50, 0, 25):
                                add(1, diff, ce, res, 0, 128, 0, 0, boss, bcol, hire, freeze)
    n += len(rows)
    while len(rows) < n:
        diff = rng.randrange(3)
        ce = rng.choice((-50, -50, -50, -33, -25, -75, 0, rng.randint(-128, 127)))
        res = rng.choice((rng.randint(-100, 99), rng.randint(-200, 300), rng.randint(90, 110), 0, rng.randint(0, 100)))
        pierce = rng.choice((0, 0, 0, rng.randint(0, 100), rng.randint(0, 300)))
        roll = rng.choice((rng.randint(0, 255), rng.randint(1, 1024), rng.randint(1, 0x10000), rng.randint(-512, 512)))
        length = rng.choice((rng.randint(0, 25), rng.randint(1, 500), rng.randint(1, 5000), rng.randint(-20, 20)))
        prior = rng.choice((0, 0, 0, rng.randint(1, 600)))
        boss = rng.choice((0, 0, 0, 1, 2, 3, 4))
        bcol = int(rng.random() < 0.1)
        if rng.random() < 0.05:
            add(1, diff, ce, res, pierce, roll, 0, prior, boss, bcol, int(rng.random() < 0.2), rng.randint(0, 200))
        else:
            add(0, diff, ce, res, pierce, roll, length, prior, boss, bcol)
    for i, r in enumerate(rows):
        print(i, *r, sep="\t")


def diff_levels(emu):
    text = emu.mpq_file("data\\global\\excel\\DifficultyLevels.txt").decode("latin-1").splitlines()
    head = text[0].split("\t")
    rows = [dict(zip(head, line.split("\t"))) for line in text[1:4]]
    return [(int(r["ResistPenalty"]), int(r["MonsterFreezeDivisor"]), int(r["MonsterColdDivisor"])) for r in rows]


def game(path):
    sys.path.insert(0, ".")          # run from upstream's tools/emu
    from emu import Emu
    emu = Emu()
    levels = diff_levels(emu)
    boss_mask = emu.r32(0x6CE280) & 0xFF
    assert boss_mask, "[0x6ce280] is 0 in the image"
    game_ = emu.alloc(0x200)
    attacker = emu.alloc(0x200)
    defender = emu.alloc(0x200)
    mdata = emu.alloc(0x100)
    diffrec = emu.alloc(0x100)
    record = emu.alloc(0x80)
    nrows = 0x240
    monstats = emu.alloc(0x1A8 * nrows)
    tables = emu.alloc(0x1000)
    emu.w32(tables + 0xA78, monstats)
    emu.w32(tables + 0xA80, nrows)
    emu.w32(0x744304, tables)
    emu.w32(attacker, 0)
    emu.w32(defender, 1)
    emu.w32(defender + 0x14, mdata)
    emu.w32(game_ + 0xA8, FRAME)
    emu.w32(game_ + 0x70, 1)
    st = {}
    lists = {}                         # list address -> [state, end]
    unit_list = {}                     # (unit, state) -> list address

    def stat(e):
        return st.get((e.arg(0), e.arg(1)), 0) & 0xFFFFFFFF

    def reg(e, name):
        import unicorn.x86_const as x
        return e.mu.reg_read(getattr(x, "UC_X86_REG_" + name))

    def new_list(e):                   # FUN_006251f0(pool, 2, end, owner type, owner id)
        a = e.alloc(0x40)
        lists[a] = [None, e.arg(2)]
        return a

    def set_state(e):                  # FUN_006252d0(list, state)
        lists[e.arg(0)][0] = e.arg(1)
        return 0

    def attach(e):                     # FUN_00626e10(unit, list, 1)
        unit_list[(e.arg(0), lists[e.arg(1)][0])] = e.arg(1)
        return 0

    def set_end(e):
        lists[e.arg(0)][1] = e.arg(1)
        return 0

    for a, fn, n in (
        (0x625480, stat, 3), (0x625500, stat, 3), (0x627260, lambda e: 0, 4),
        (0x611d30, lambda e: diffrec, 1),
        (0x451F80, lambda e: monstats + 0x1A8 * reg(e, "ECX") if reg(e, "ECX") < nrows else 0, 0),
        (0x457490, lambda e: 1, 0), (0x554200, lambda e: 1, 1), (0x620BB0, lambda e: 1, 1),
        (0x61AB00, lambda e: 0, 1), (0x5541B0, lambda e: 0, 0), (0x621D50, lambda e: 0, 2),
        (0x57C420, lambda e: 0, 1), (0x57A980, lambda e: 0, 1), (0x57AA60, lambda e: 0, 0),
        (0x57AAA0, lambda e: 0, 0), (0x57AAE0, lambda e: 0, 2), (0x57AC50, lambda e: 0, 3),
        (0x57ADD0, lambda e: 0, 3), (0x5D6410, lambda e: 0, 0), (0x540E60, lambda e: 0, 2),
        (0x5C0C30, lambda e: 0, 3), (0x5347D0, lambda e: 0, 7), (0x5347E0, lambda e: 0, 3),
        (0x63E990, lambda e: 0, 1), (0x63E940, lambda e: 0, 1), (0x639DF0, lambda e: 0, 2),
        (0x4638A0, lambda e: 1, 0), (0x554010, lambda e: game_, 0),
        (0x6256B0, lambda e: unit_list.get((e.arg(0), e.arg(1)), 0), 2),
        (0x6251F0, new_list, 5), (0x6252D0, set_state, 2), (0x626E10, attach, 3),
        (0x625330, lambda e: lists[e.arg(0)][1], 1), (0x625310, set_end, 2),
        (0x639DB0, lambda e: 0, 3), (0x5417D0, lambda e: 0, 4), (0x625CE0, lambda e: 0, 2),
        (0x6270B0, lambda e: 0, 4), (0x623F50, lambda e: 0, 3),
    ):
        emu.hook(a, fn, n)

    for line in open(path):
        f = line.split()
        if not f:
            continue
        cid, phase, diff, ce, res, pierce, roll, length, prior, uflag, bcol, hire, freeze, _boss = map(int, f)
        cls = HIRELINGS[cid % len(HIRELINGS)] if hire else FALLEN
        emu.mu.mem_write(game_ + 0x6D, bytes([diff]))
        penalty, fdiv, cdiv = levels[diff]
        emu.w32(diffrec, penalty)
        emu.w32(diffrec + 0x14, fdiv)
        emu.w32(diffrec + 0x18, cdiv)
        row = monstats + 0x1A8 * cls
        emu.mu.mem_write(row, b"\0" * 0x1A8)
        emu.mu.mem_write(row + 0xC, bytes([boss_mask if bcol else 0]))
        emu.mu.mem_write(row + 0x168 + diff, struct.pack("<b", ce))
        emu.w32(defender + 4, cls)
        emu.w32(defender + 0x20, 0x1234567 + cid)
        emu.w32(defender + 0x24, 0)
        emu.mu.mem_write(mdata + 0x16, struct.pack("<H", uflag))
        st.clear()
        st[(defender, STAT_LIFE)] = 1000000 << 8
        st[(defender, STAT_COLD_RES)] = res
        st[(attacker, STAT_COLD_PIERCE)] = pierce
        lists.clear()
        unit_list.clear()
        if prior > 0:
            a = emu.alloc(0x40)
            lists[a] = [0xB, FRAME + prior]
            unit_list[(defender, 0xB)] = a
        emu.mu.mem_write(record, b"\0" * 0x80)
        emu.w32(record + 0x24, roll)
        emu.w32(record + 0x30, length)
        emu.w32(record + 0x34, freeze)
        emu.call(0x57C6C0, defender, 1, record, ecx=game_, edx=attacker)

        def end(state):
            a = unit_list.get((defender, state))
            return str(lists[a][1] - FRAME) if a and lists[a][1] > FRAME else "-"
        print(cid, end(0xB), end(1))


def classify(r, g, d):
    """Why a phase's line differs, by the rules read from game.exe (verifier section)."""
    phase, ce, res, roll, length = int(r[1]), int(r[3]), int(r[4]), int(r[6]), int(r[7])
    uflag, bcol, hire, freeze = int(r[9]), int(r[10]), int(r[11]), int(r[12])
    if phase == 1:
        special = bool(uflag & 8) or bcol or hire
        if special:
            return "C3 freeze->chill (unique flag 8 / MonStats boss / hireling)"
        return "C5 freeze itself (d2d: unresisted, stun + chill, no freeze divisor; game: freeze only, ce < 0 only)"
    if abs(length) >= 0x100000:
        return "R-big length >= 0x100000 (resisted() vs MulDiv / int overflow)"
    if roll <= 0:
        return "R-ehi missile_blow returns before the cold when the max roll <= 0"
    return "OTHER"


def compare(cases, a, b):
    rows = {l.split()[0]: l.split() for l in open(cases) if l.strip()}
    ga = {l.split()[0]: l.split()[1:3] for l in open(a) if l.strip()}
    gb = {l.split()[0]: l.split()[1:3] for l in open(b) if l.strip()}
    flags = [l for l in open(b) if "RESFAIL" in l]
    bad = [k for k in rows if ga.get(k) != gb.get(k)]
    per = {}
    for k in bad:
        c = classify(rows[k], ga.get(k), gb.get(k))
        per.setdefault(c, []).append(k)
    p0 = sum(1 for k in rows if rows[k][1] == "0")
    print(f"cases {len(rows)} (phase 0 {p0}, phase 1 {len(rows) - p0}) game {len(ga)} d2d {len(gb)} mismatches {len(bad)} resfail {len(flags)}")
    for c, ks in sorted(per.items()):
        print(f"  {len(ks):7d}  {c}")
        for k in ks[:4]:
            print("           ", " ".join(rows[k]), "game", ga.get(k), "d2d", gb.get(k))


def patch(path):
    mode = open(os.path.join(os.path.dirname(os.path.abspath(__file__)), "verify_chill_mode.inc")).read()
    src = open(path).read()
    anchor = "character.panel = panel_stats(*data, character.header, character.items, character.stats);\n"
    assert anchor in src and "--verify-chill" not in src
    open(path, "w").write(src.replace(anchor, anchor + mode, 1))


if __name__ == "__main__":
    cmd = sys.argv[1] if len(sys.argv) > 1 else ""
    if cmd == "gen":
        gen(int(sys.argv[2]), int(sys.argv[3]))
    elif cmd == "game":
        game(sys.argv[2])
    elif cmd == "compare":
        compare(*sys.argv[2:5])
    elif cmd == "patch":
        patch(sys.argv[2])
    else:
        sys.exit(__doc__)
