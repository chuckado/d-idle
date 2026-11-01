# SPDX-License-Identifier: GPL-3.0-or-later
"""Verifier (card verifyfreeze): a cold missile's freeze on a monster by game.exe's own code,
against d2d's REAL Fight::strike (test mode verify_freeze_mode.inc). Written without the author's
oracle (tools/emu/freeze_rules.py of fix/freeze-rules).

    PYTHONPATH=<d2decomped>/tools/emu python verify_freeze.py N SEED > game.txt
    build/tests/test_verify_freeze --verify-freeze SEED N | diff game.txt -
    PYTHONPATH=... python verify_freeze.py --game-only     # the game.exe-side checks d2d can't run

A case line is `id diff coldeffect bosscol kind half res pierce roll freeze prior_chill prior_freeze
fdiv cdiv` and its outcome `chill freeze`: the ticks from now each state ends, or `-`.
kind: 0 plain (monster data +0x16 = 0x00), 1 champion (0x0d), 2 unique (0x09), 3 superunique (0x0b),
4 minion (0x10), 5 ghostly champion (0x4d), 6 act boss as Blood Raven (0x08): the flag words
game.exe's own spawn path gives them (verify_freeze_spawn.py). `half` is d2d's Monster::half_freeze
(the Countess / Blood Raven call), which game.exe's record never sees: state 118 is not read on
this path. fdiv / cdiv: DifficultyLevels MonsterFreezeDivisor (+0x14) / MonsterColdDivisor (+0x18).

Run for real: FUN_0057c6c0 (the damage record on the monster, flags 1, resist on) with
FUN_0057c1e0, FUN_0057c060, FUN_0057c140, the element table at 0x732980 through FUN_0057bf80 /
FUN_0057be00 (row 5 +0x30, row 6 +0x34: cold resist 0x2b less pierce 0x14f), then
FUN_0057af80 (chill) and FUN_0057b230 (freeze) with FUN_00639df0 (state 0x36), FUN_0063e9f0 +
FUN_00451f80 (MonStats +0xc & [0x6ce280]), FUN_005a0180 (+0x16 & 8), FUN_0063ee90 (hireling
classes), FUN_0057af30 (coldeffect, MonStats +0x168 + difficulty) and the divisor idiv.
Hooked (not the claim): stat getters / setter (FUN_00625480 / FUN_00625500 / FUN_00627260),
max life FUN_00625d10, the difficulty row FUN_00611d30 (a 0x40-byte row holding the case's
divisors), FUN_00554010 (the game), FUN_00625b70 (state bits: the case's state 0x36 only),
FUN_00639db0 (state set: recorded), the state-list calls FUN_006256b0 / FUN_00625330 /
FUN_00625310 / FUN_006251f0 (they record each state's end), events FUN_005417d0 /
FUN_00540e60, speed stats, the shatter roll, unit-kind predicates other than the hireling one
(FUN_0063e990 / FUN_0063e940 / FUN_0063edc0), FUN_0057c6c0's life / stun / knockback side calls.
MonStats is a fake table: class 19 (fallen1) and the hireling classes, row byte +0xc and +0x168..
written per case.
"""
import sys

import emu

FRAME = 1000
PLAIN = 19
HIRELINGS = (271, 338, 359, 560, 561)
FLAGS = (0x00, 0x0d, 0x09, 0x0b, 0x10, 0x4d, 0x08)
FDIV = (1, 2, 4)                       # 1.14d DifficultyLevels MonsterFreezeDivisor
CDIV = (1, 2, 4)                       # MonsterColdDivisor


def xorshift(state):
    x = state[0]
    x ^= (x << 13) & 0xFFFFFFFF
    x ^= x >> 17
    x ^= (x << 5) & 0xFFFFFFFF
    state[0] = x
    return x


class Game:
    def __init__(self):
        e = self.e = emu.Emu()
        self.stats, self.ends, self.prior, self.states, self.last_set = {}, {}, {}, set(), None
        self.game = e.alloc(0x100); e.w32(self.game + 0x70, 1)
        e.w32(self.game + 0xa8, FRAME)
        self.attacker = e.alloc(0x100)                                   # type 0: a player
        self.target = e.alloc(0x100)
        self.mdata = e.alloc(0x100)
        self.rec = e.alloc(0x80)
        self.diffrow = e.alloc(0x40)
        self.tables = e.alloc(0x1000)
        self.rows = e.alloc(600 * 0x1a8)
        e.w32(self.tables + 0xa78, self.rows); e.w32(self.tables + 0xa80, 600)
        e.w32(self.tables + 0xc4, 0x200)                                 # states count (FUN_00639db0 / FUN_00639df0 bounds)
        e.w32(0x744304, self.tables)
        stat = lambda e: self.stats.get((e.arg(0), e.arg(1)), 0)
        e.hook(0x625480, stat, 3)
        e.hook(0x625500, stat, 3)
        e.hook(0x627260, lambda e: self.stats.__setitem__((e.arg(0), e.arg(1)), e.arg(2)) or 0, 4)
        e.hook(0x625d10, lambda e: 100000 << 8, 1)
        e.hook(0x611d30, lambda e: self.diffrow, 1)
        e.hook(0x554010, lambda e: self.game, 0)
        e.hook(0x625b70, lambda e: 1 if (e.arg(0), e.arg(1)) in self.states else 0, 2)
        e.hook(0x639db0, lambda e: self.set_state(e.arg(0), e.arg(1), e.arg(2)), 3)
        e.hook(0x6256b0, lambda e: 0x1000 + e.arg(1) if e.arg(1) in self.prior or e.arg(1) in self.ends else 0, 2)
        e.hook(0x625330, lambda e: self.ends.get(e.arg(0) - 0x1000, self.prior.get(e.arg(0) - 0x1000)), 1)
        e.hook(0x625310, lambda e: self.ends.__setitem__(e.arg(0) - 0x1000, e.arg(1)) or 0, 2)
        e.hook(0x6251f0, lambda e: self.new_list(e.arg(2)), 5)
        for addr, n in ((0x63e990, 1), (0x63e940, 1), (0x63edc0, 1), (0x5c0c30, 3), (0x5347d0, 7), (0x5347e0, 3), (0x61ab00, 1),
                        (0x5541b0, 0), (0x621d50, 2), (0x57aa60, 0), (0x57aaa0, 0), (0x57aae0, 2), (0x57ac50, 3), (0x57add0, 3),
                        (0x5d6410, 0), (0x5417d0, 4), (0x540e60, 2), (0x6252d0, 2), (0x625ce0, 2), (0x626e10, 3), (0x6270b0, 4),
                        (0x623f50, 3), (0x4638a0, 0)):
            e.hook(addr, lambda e: 0, n)
        for addr, n in ((0x554200, 1), (0x620bb0, 1), (0x457490, 0)):
            e.hook(addr, lambda e: 1, n)

    def set_state(self, unit, state, on):
        if unit == self.target and on: self.last_set = state
        return 0

    def new_list(self, end):
        self.ends[self.last_set] = end
        return 0x1000 + self.last_set

    def run(self, diff, cold_effect, bosscol, flags, res, pierce, roll, freeze, prior_chill, prior_freeze, fdiv, cdiv,
            cls=PLAIN, length=0, extra_states=(), extra_stats=()):
        e = self.e
        self.stats = {(self.target, 6): 100000 << 8, (self.target, 7): 100000 << 8,
                      (self.target, 0x2b): res & 0xFFFFFFFF, (self.attacker, 0x14f): pierce & 0xFFFFFFFF}
        for k, v in extra_stats: self.stats[(self.target, k)] = v
        self.states = {(self.target, s) for s in extra_states}
        self.prior = {}
        if prior_chill: self.prior[0xb] = FRAME + prior_chill
        if prior_freeze: self.prior[1] = FRAME + prior_freeze
        self.ends, self.last_set = {}, None
        e.mu.mem_write(self.game + 0x6d, bytes([diff]))
        e.w32(self.diffrow + 0x14, fdiv); e.w32(self.diffrow + 0x18, cdiv)
        row = self.rows + cls * 0x1a8
        e.mu.mem_write(row, b"\0" * 0x1a8)
        e.mu.mem_write(row + 0xc, bytes([0x40 if bosscol else 0]))
        e.mu.mem_write(row + 0x168, bytes((cold_effect & 0xFF if d == diff else 0) for d in range(3)))
        e.mu.mem_write(self.target, b"\0" * 0x100)
        e.w32(self.target, 1); e.w32(self.target + 4, cls); e.w32(self.target + 0x14, self.mdata)
        e.mu.mem_write(self.mdata, b"\0" * 0x100)
        e.mu.mem_write(self.mdata + 0x16, bytes([flags & 0xFF, flags >> 8]))
        e.mu.mem_write(self.rec, b"\0" * 0x80)
        e.w32(self.rec, 1); e.w32(self.rec + 4, 1)
        e.w32(self.rec + 0x24, roll); e.w32(self.rec + 0x30, length); e.w32(self.rec + 0x34, freeze); e.w32(self.rec + 0x60, 0x30)
        e.call(0x57c6c0, self.target, 1, self.rec, ecx=self.game, edx=self.attacker)

        def ticks(state):
            end = self.ends.get(state, self.prior.get(state))
            return str(end - FRAME) if end is not None and end - FRAME > 0 else "-"
        return f"{ticks(0xb)} {ticks(1)}"


def cases(n, seed):
    """Same cases, same order, as test_verify_freeze --verify-freeze SEED N."""
    i = 0
    kinds_half = [(k, 0) for k in (0, 1, 2, 4, 5)] + [(k, h) for k in (3, 6) for h in (0, 1)]
    for diff in range(3):
        for ce in (-50, -1, 0, 1, 25):
            for kind, half in kinds_half:
                for bcol in (0, 1):
                    for res in (-150, -100, -1, 0, 1, 50, 99, 100, 101):
                        for fr in (0, 1, 2, 3, 4, 5, 7, 50, 100, 375):
                            yield (f"E{i}", diff, ce, bcol, kind, half, res, 0, 256, fr, 0, 0, FDIV[diff], CDIV[diff]); i += 1
        for pierce in (1, 40, 100):
            for res in (50, 99, 100, 150):
                for fr in (50, 150):
                    for kind in (0, 1):
                        yield (f"E{i}", diff, -50, 0, kind, 0, res, pierce, 256, fr, 0, 0, FDIV[diff], CDIV[diff]); i += 1
        for pc in (0, 1, 37, 1000):
            for pf in (0, 1, 37, 1000):
                for fr in (0, 2, 50, 375):
                    for kind in (0, 1, 4):
                        for ce in (-50, 25):
                            yield (f"E{i}", diff, ce, 0, kind, 0, 0, 0, 256, fr, pc, pf, FDIV[diff], CDIV[diff]); i += 1
    for fdiv in (1, 2, 3, 5, 7):
        for cdiv in (1, 2, 3, 4):
            for fr in (1, 2, 3, 4, 5, 9, 50, 101):
                for ce in (-50, -1, 25):
                    for kind in (0, 1):
                        yield (f"E{i}", 0, ce, 0, kind, 0, 0, 0, 256, fr, 0, 0, fdiv, cdiv); i += 1
    for fr in (0xFFFFF, 0x100000, 0x100001, 10000000):
        for res in (-100, -50, 0, 50):
            for kind in (0, 1):
                yield (f"E{i}", 0, -50, 0, kind, 0, res, 0, 256, fr, 0, 0, 1, 1); i += 1
    for kind in (0, 1, 4):
        yield (f"E{i}", 1, -50, 0, kind, 0, 0, 0, 0, 50, 0, 0, FDIV[1], CDIV[1]); i += 1
    state = [(seed & 0xFFFFFFFF) or 1]
    effects = (-50, -50, -50, -33, -25, 0, 10, -100)
    for k in range(n):
        diff = xorshift(state) % 3
        ce = effects[xorshift(state) % 8] if xorshift(state) % 4 else xorshift(state) % 201 - 100
        bcol = 1 if xorshift(state) % 8 == 0 else 0
        kind = xorshift(state) % 7
        half = xorshift(state) % 2 if kind in (3, 6) else (1 if xorshift(state) % 16 == 0 else 0)
        r = xorshift(state) % 100
        if r < 10: res = 100 + xorshift(state) % 51
        elif r < 25: res = -(xorshift(state) % 151)
        else: res = xorshift(state) % 100
        pierce = 0 if res >= 100 or xorshift(state) % 4 else xorshift(state) % 61
        roll = 1 + xorshift(state) % 0x4000
        fr = 0 if xorshift(state) % 10 == 0 else 1 + xorshift(state) % 500
        pc = 0 if xorshift(state) % 10 < 7 else 1 + xorshift(state) % 600
        pf = 0 if xorshift(state) % 10 < 7 else 1 + xorshift(state) % 600
        fdiv = FDIV[diff] if xorshift(state) % 5 else 1 + xorshift(state) % 8
        cdiv = CDIV[diff] if xorshift(state) % 5 else 1 + xorshift(state) % 8
        yield (f"R{k}", diff, ce, bcol, kind, half, res, pierce, roll, fr, pc, pf, fdiv, cdiv)


def game_only(g):
    """Checks on game.exe's side alone (d2d's skill missiles don't strike mercs; no 1.14d monster has
    state 0x36 at a freeze or stats 0x99 / 0x76): every flag byte, hirelings, state 0x36, stats."""
    out, bad = [], 0
    for diff in range(3):
        for ce in (-50, 0, 25):
            for bcol in (0, 1):
                for flags in range(256):
                    got = g.run(diff, ce, bcol, flags, 0, 0, 256, 100, 0, 0, FDIV[diff], CDIV[diff])
                    want = g.run(diff, ce, bcol, flags & 8, 0, 0, 256, 100, 0, 0, FDIV[diff], CDIV[diff])
                    if got != want: bad += 1; out.append(f"flagbyte d{diff} ce{ce} b{bcol} {flags:#04x}: {got} vs {want}")
    out.append(f"flag bytes 0..255: only bit 8 matters: {3 * 3 * 2 * 256 - bad} same, {bad} differ")
    for cls in HIRELINGS:
        for flags in (0, 8, 0x10):
            for ce in (-50, 25, 0):
                out.append(f"hireling class {cls} flags {flags:#04x} ce {ce}: {g.run(1, ce, 0, flags, 0, 0, 256, 100, 0, 0, 2, 2, cls=cls)}")
    for ce in (-50, 25):
        out.append(f"state 0x36 plain ce {ce}: {g.run(1, ce, 0, 0, 0, 0, 256, 100, 0, 0, 2, 2, extra_states=(0x36,))}")
        out.append(f"state 0x36 champion ce {ce}: {g.run(1, ce, 0, 0x0d, 0, 0, 256, 100, 0, 0, 2, 2, extra_states=(0x36,))}")
        out.append(f"state 118 plain ce {ce}: {g.run(1, ce, 0, 0, 0, 0, 256, 100, 0, 0, 2, 2, extra_states=(0x76,))}")
        out.append(f"stat 0x76 plain ce {ce}: {g.run(1, ce, 0, 0, 0, 0, 256, 100, 0, 0, 2, 2, extra_stats=((0x76, 1),))}")
        out.append(f"stat 0x99 plain ce {ce}: {g.run(1, ce, 0, 0, 0, 0, 256, 100, 0, 0, 2, 2, extra_stats=((0x99, 1),))}")
        out.append(f"cold length 40 + freeze 100, champion ce {ce}: {g.run(1, ce, 0, 0x0d, 0, 0, 256, 100, 0, 0, 2, 2, length=40)}")
        out.append(f"cold length 400 + freeze 100, champion ce {ce}: {g.run(1, ce, 0, 0x0d, 0, 0, 256, 100, 0, 0, 2, 2, length=400)}")
        out.append(f"cold length 40 + freeze 100, plain ce {ce}: {g.run(1, ce, 0, 0, 0, 0, 256, 100, 0, 0, 2, 2, length=40)}")
    # divisor 0 isn't run: FUN_0057b230 asserts (0x0057b2fd, line 0x254) and game.exe exits; d2d freezes for nothing.
    return out


def compare(game_file, d2d_file):
    """Mismatches by the claim that owns the case: C4 d2d's half_freeze set, C3 a special monster
    (flag 8 kinds or MonStats boss), C5 the rest; `big` a length over 0x100000 (MulDiv's
    divide-first, not a claim)."""
    def rows(f):
        with open(f) as lines: return [l.split() for l in lines if l[:1] in "ER" and l[1:2].isdigit()]
    game, d2d = rows(game_file), rows(d2d_file)
    if len(game) != len(d2d): raise SystemExit(f"{len(game)} game lines vs {len(d2d)} d2d lines")
    total, bad = {}, {}
    for g, d in zip(game, d2d):
        if g[:14] != d[:14]: raise SystemExit(f"generators differ at {g[0]} / {d[0]}")
        kind, bcol, half, fr = int(g[4]), int(g[3]), int(g[5]), int(g[9])
        cls = "big" if fr > 0x100000 else "C4" if half else "C3" if kind in (1, 2, 3, 5, 6) or bcol else "C5"
        total[cls] = total.get(cls, 0) + 1
        if g[14:16] != d[14:16] or len(d) > 16: bad[cls] = bad.get(cls, 0) + 1
    print(" ".join(f"{c} {bad.get(c, 0)}/{total[c]}" for c in sorted(total)), f"all {sum(bad.values())}/{len(game)}")


def main():
    if sys.argv[1:2] == ["--compare"]:
        compare(sys.argv[2], sys.argv[3])
        return
    g = Game()
    if sys.argv[1:2] == ["--game-only"]:
        print("\n".join(game_only(g)))
        return
    n = int(sys.argv[1]) if len(sys.argv) > 1 else 100000
    seed = int(sys.argv[2], 0) if len(sys.argv) > 2 else 1
    total = 0
    for case in cases(n, seed):
        _, diff, ce, bcol, kind, half, res, pierce, roll, fr, pc, pf, fdiv, cdiv = case
        print(" ".join(map(str, case)), g.run(diff, ce, bcol, FLAGS[kind], res, pierce, roll, fr, pc, pf, fdiv, cdiv))
        total += 1
    print(f"{total} cases", file=sys.stderr)


if __name__ == "__main__": main()
