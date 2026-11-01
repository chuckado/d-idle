# SPDX-License-Identifier: GPL-3.0-or-later
"""Verifier (card verifymods): a champion's, unique's, superunique's or minion's level, life,
defence, experience and speed by game.exe's own spawn code, against d2d's REAL make_boss
(test mode verify_mods_mode.inc). Written without the author's oracle (boss_level.py of
fix/boss-stats).

    PYTHONPATH=<d2decomped>/tools/emu python verify_mods.py N SEED > game.txt
    build/tests/test_verify_mods --verify-mods SEED N | diff game.txt -

Both sides generate the same cases from SEED & 0xffffffff (xorshift32, `cases` below and the
C++ twin). A line: `i row diff kind mods seed | level life def exp vel`, or `i row ... | -` when the
row has no Level for that difficulty (the spawn would not place it).
kind: u unique (fixed mods 1..4 then its own), s superunique (the same path), c champion
(fixed 1..4 then 16 / 36 ghostly / 37 fanatic / 38 possessed / 39 berserker), m minion (its
leader's fixed mods 1..4 with flag 0, as FUN_0058f380 -> FUN_005a2100 hand them), p plain.

Run for real: FUN_00573cb0 (the spawn's stat init: MonStats level, FUN_006538a0 mask 7, the
life roll FUN_0045c3e0 on the unit seed), FUN_005a2120 (fixed mods 0x6e2168, then the unit's
own list at monster data +0x1c, through the handler table 0x73c008: FUN_005a0dc0 /
FUN_005a0d20, FUN_005a0e40, FUN_005a0e80, FUN_005a1080, FUN_005a11f0, FUN_005a1230,
FUN_005a1280, FUN_005a17e0, FUN_005a1910, FUN_005a1370, FUN_005a1990 & co), DifficultyLevels
(FUN_00611d30) and MonUMod constants (FUN_005a00f0) from the loaded tables, FUN_00463860 /
FUN_00463900. A minion: FUN_00573cb0, then the handler of each of the leader's fixed mods with
flag 0 (FUN_005a2100's call).
Hooked (not the claim): the stat list (FUN_00627260 set, FUN_006272b0 add, FUN_00625480 /
FUN_006253b0 get, FUN_00625d10 max life: one dict per unit, no layers), FUN_00554010 (the game),
FUN_00573930 (player-count bonuses: zero, one player), FUN_005739d0, FUN_0063ee90 (hireling: no),
FUN_00626d40, FUN_0063eef0, FUN_006251f0, FUN_00626e10, FUN_0063abd0, FUN_00536b20, FUN_006438b0,
FUN_0056deb0, FUN_006439f0, FUN_00644340, FUN_00573470 (no MonStats2 chance list),
FUN_005d6b60, FUN_0058f380 (group walk: the minion case runs the handlers itself).
The game: classic level rule (+0x70 = 0: no area level in nightmare / hell, as d2d has none),
the L- MonLvl columns (+0x74 = 1, as d2d reads).
Fields: level stat 12; life stat 7 >> 8; defence stat 31 with item_armor_percent (stat 16)
applied as d2d keeps it (ac + ac * pct / 100, truncated); experience stat 13; velocity % stat 67 less the 75 FUN_00573cb0 sets (d2d's boss_speed is the bonus).
"""
import sys

import drlg
import emu

KINDS = "uscmp"
CHAMP = (16, 36, 37, 38, 39)
UMODS = (5, 6, 8, 9, 17, 18, 25, 27, 28)      # unique mods with a stat init
ROWS = 734


def xorshift(state):
    x = state[0]
    x ^= (x << 13) & 0xFFFFFFFF
    x ^= x >> 17
    x ^= (x << 5) & 0xFFFFFFFF
    state[0] = x
    return x


def cases(n, seed):
    """The C++ twin (verify_mods_mode.inc) draws the same numbers in the same order."""
    st = [(seed & 0xFFFFFFFF) or 1]
    edge_rows = (0, 1, 2, 3, 5, 19, 156, 229, 250, 300, 400, 500, 571, 650, 700, 733)
    for i in range(n):
        row = edge_rows[i % len(edge_rows)] if i < 3 * 16 * len(edge_rows) else xorshift(st) % ROWS
        diff = (i // len(edge_rows)) % 3 if i < 3 * 16 * len(edge_rows) else xorshift(st) % 3
        kind = KINDS[xorshift(st) % len(KINDS)]
        mods = []
        if kind == "c":
            mods = [CHAMP[xorshift(st) % len(CHAMP)]]
        elif kind in "us":
            for _ in range(xorshift(st) % 4):
                m = UMODS[xorshift(st) % len(UMODS)]
                if m not in mods: mods.append(m)
        unit_seed = xorshift(st)
        yield i, row, diff, kind, mods, unit_seed


class Game:
    def __init__(self):
        e = self.e = drlg.boot()
        self.stats = {}
        self.game = e.alloc(0x200)
        e.w32(self.game + 0x70, 0)
        e.w32(self.game + 0x74, 1)
        self.tables = e.r32(0x744304)
        self.nrows = e.r32(self.tables + 0xa80)
        st = lambda e: self.stats.setdefault(e.arg(0), {}).get(e.arg(1), 0)
        def put(e):
            self.stats.setdefault(e.arg(0), {})[e.arg(1)] = e.arg(2) - (1 << 32) if e.arg(2) & 0x80000000 else e.arg(2)
            return 0
        def add(e):
            d = self.stats.setdefault(e.arg(0), {})
            v = e.arg(2) - (1 << 32) if e.arg(2) & 0x80000000 else e.arg(2)
            d[e.arg(1)] = d.get(e.arg(1), 0) + v
            return 0
        e.hook(0x627260, put, 4)
        e.hook(0x6272b0, add, 4)
        e.hook(0x625480, lambda e: st(e) & 0xFFFFFFFF, 3)
        e.hook(0x6253b0, lambda e: st(e) & 0xFFFFFFFF, 3)
        e.hook(0x625d10, lambda e: self.stats.setdefault(e.arg(0), {}).get(7, 0) & 0xFFFFFFFF, 1)
        e.hook(0x554010, lambda e: self.game, 0)
        def bonus(e):
            out = e.mu.reg_read(emu.UC_X86_REG_EDX)
            e.mu.mem_write(out, bytes(0x14))       # +0 life %, +4 exp %, +8 skill levels, +0x10 stat 0x64
            return 0
        e.hook(0x573930, bonus, 2)
        for addr, n in ((0x5739d0, 0), (0x63ee90, 1), (0x626d40, 4), (0x63eef0, 3), (0x6251f0, 5), (0x626e10, 3),
                        (0x63abd0, 2), (0x536b20, 1), (0x6438b0, 1), (0x56deb0, 2), (0x6439f0, 0), (0x644340, 2),
                        (0x573470, 0), (0x5d6b60, 5), (0x58f380, 2)):
            e.hook(addr, lambda e: 0, n)

    def unit(self, row, seed, flags, mods):
        e = self.e
        u, md = e.alloc(0x100), e.alloc(0x80)
        e.w32(u + 0, 1); e.w32(u + 4, row); e.w32(u + 0xc, 0x100 + row); e.w32(u + 0x14, md)
        e.w32(u + 0x20, seed); e.w32(u + 0x24, 666)
        e.w32(md + 0, e.r32(self.tables + 0xa78) + row * 0x1a8)
        e.mu.mem_write(md + 0x16, bytes([flags & 0xff, flags >> 8]))
        e.mu.mem_write(md + 0x1c, bytes(mods[:9]) + bytes(9 - len(mods[:9])))
        self.stats[u] = {}
        return u

    def level(self, row, diff):
        return self.e.r16(self.e.r32(self.tables + 0xa78) + row * 0x1a8 + 0xaa + 2 * diff)

    def run(self, row, diff, kind, mods, seed):
        e = self.e
        if row >= self.nrows or self.level(row, diff) == 0: return None
        e.mu.mem_write(self.game + 0x6d, bytes([diff]))
        flags = {"u": 0x09, "s": 0x0b, "c": 0x0d, "m": 0x10, "p": 0}[kind]   # FUN_005a0320 / FUN_005a48c0 / FUN_005a0c00
        u = self.unit(row, seed, flags, mods if kind in "usc" else [])
        e.call(0x573cb0, self.game, 0, 0, regs={"esi": u})
        if kind in "usc":
            e.call(0x5a2120, self.game, u, 0)
        elif kind == "m":
            for m in (1, 2, 3, 4):
                h = e.r32(0x73c008 + 4 * m)
                if h: e.call(h, 0, ecx=u, edx=m)
        s = self.stats[u]
        ac, pct = s.get(31, 0), s.get(16, 0)
        ac = ac + int(ac * pct / 100)
        return s.get(12, 0), s.get(7, 0) >> 8, ac, s.get(13, 0), s.get(67, 0) - 75   # FUN_00573cb0 sets stat 67 to 75


def main():
    args = sys.argv[1:]
    if len(args) != 2: sys.exit(__doc__)
    n, seed = int(args[0]), int(args[1], 0)
    g = Game()
    for i, row, diff, kind, mods, useed in cases(n, seed):
        got = g.run(row, diff, kind, mods, useed)
        head = f"{i} {row} {diff} {kind} {'.'.join(map(str, mods)) or '-'} {useed}"
        print(f"{head} | {' '.join(map(str, got))}" if got else f"{head} | -")


if __name__ == "__main__":
    main()
