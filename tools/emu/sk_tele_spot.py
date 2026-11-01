# SPDX-License-Identifier: GPL-3.0-or-later
"""game.exe's player Teleport destination rule as the oracle: srvdofunc 27
(FUN_005ca360) and the warp it calls (FUN_00554ea0: the target room, the
nearest free spot FUN_0064e7b0 -> FUN_0064dea0, the collision test
FUN_0064d9b0), run natively on synthetic rooms; compared with a line copy of
d2d's `Fight::spot` Teleport branch (components/game/fight/strike.cpp,
upstream main 72132b99).

    python sk_tele_spot.py gen 20000 1 > cases.tsv
    python sk_tele_spot.py game cases.tsv > game.tsv
    python sk_tele_spot.py d2d cases.tsv > d2d.tsv     # the line copy ("before")
    python sk_tele_spot.py diff game.tsv d2d.tsv
    python sk_tele_spot.py probe                       # where the snap lands round a blocked target

The world: 5 x 5 rooms of 40 x 40 subtiles (outdoor rooms are 8 x 8 tiles);
the caster stands in the centre room (2, 2). A room's near list is the room
and the up to 8 rooms touching it. A case is
`id tele cx cy tx ty n {x y bits}...`: the level's Teleport column, the
caster's and the target's subtile, then the non-zero collision words.
`game` prints `id do landed x y`: FUN_005ca360's answer and where
FUN_00650be0 put the unit (`- -` when nothing moved).

Hooked (not the claim): FUN_0056d2c0 (the target point: the case's), the
caster's room FUN_00620bb0, FUN_0061a1b0 / FUN_0061db70 (a fake Levels row:
byte +4 = Teleport), FUN_006229f0 (the clear-line test for Teleport 2:
recorded, answers "clear"), FUN_00619790 (a room's near list),
FUN_0061a010 (a room's collision grid: {x0, y0, w, h, ..., +0x20 words}),
FUN_00620510 (the unit's size: 2, game.exe's answer for a player, read in
its jump table at 0x62057c), FUN_00650be0 (the position set: recorded),
FUN_0064c040 and FUN_00554670 (the room move bookkeeping). The unit is
typed a monster (dword 0 = 1) so FUN_00554ea0 takes its short tail; the
player's tail (client packets, the 20-step ring, the 50-frame warp spot,
the pets' warp FUN_005754b0) sets no position. Runs natively: FUN_005ca360,
FUN_00554ea0, FUN_00463740, FUN_0064e7b0, FUN_0064dea0, FUN_0064d9b0 and
the footprint test under it.
"""
import random
import sys

from unicorn.x86_const import UC_X86_REG_ECX

import emu

ROOM, GRID = 40, 5
SIDE = ROOM * GRID
MASK = 0x1c09


class Oracle:
    def __init__(self):
        self.e = e = emu.Emu()
        self.rooms, self.coll = {}, {}
        for i in range(GRID):
            for j in range(GRID):
                r, c = e.alloc(0x100), e.alloc(0x40)
                for off, v in ((0x4c, i * ROOM), (0x50, j * ROOM), (0x54, ROOM), (0x58, ROOM)):
                    e.w32(r + off, v)
                for base in (0, 0x10):
                    for k, v in enumerate((i * ROOM, j * ROOM, ROOM, ROOM)):
                        e.w32(c + base + 4 * k, v)
                e.w32(c + 0x20, e.alloc(ROOM * ROOM * 2))
                self.rooms[(i, j)], self.coll[r] = r, (i, j, c)
        self.near = {}
        for (i, j), r in self.rooms.items():
            near = [self.rooms[(a, b)] for a in range(i - 1, i + 2) for b in range(j - 1, j + 2) if (a, b) in self.rooms]
            arr = e.alloc(4 * len(near))
            for k, n in enumerate(near): e.w32(arr + 4 * k, n)
            self.near[r] = (arr, len(near))
        self.unit, self.path, self.levelrow = e.alloc(0x200), e.alloc(0x100), e.alloc(0x100)
        e.w32(self.unit, 1)
        e.w32(self.unit + 0x2c, self.path)
        self.xy = (0, 0)
        self.here = self.rooms[(2, 2)]
        self.moved = None
        self.line = None
        e.hook(0x619790, self.near_list, 3)
        e.hook(0x61a010, lambda e: self.coll[e.mu.reg_read(UC_X86_REG_ECX)][2], 0)
        e.hook(0x620bb0, lambda e: self.here, 1)
        e.hook(0x620510, lambda e: 2, 1)
        e.hook(0x650be0, self.set_pos, 5)
        e.hook(0x64c040, lambda e: 0, 1)
        e.hook(0x554670, lambda e: 0, 1)
        e.hook(0x56d2c0, self.target, 2)
        e.hook(0x61a1b0, lambda e: 0x1000, 1)
        e.hook(0x61db70, lambda e: self.levelrow, 1)
        e.hook(0x6229f0, self.line_test, 4)

    def near_list(self, e):
        arr, n = self.near[e.arg(0)]
        e.w32(e.arg(1), arr)
        e.w32(e.arg(2), n)
        return 0

    def set_pos(self, e):
        self.moved = (e.arg(3), e.arg(4))
        return 1

    def target(self, e):
        e.w32(e.arg(0), self.xy[0])
        e.w32(e.arg(1), self.xy[1])
        return 1

    def line_test(self, e):
        self.line = (e.arg(1), e.arg(2), e.arg(3))
        return 0

    def load(self, words):
        for (i, j), r in self.rooms.items():
            buf = bytearray(ROOM * ROOM * 2)
            for (x, y), v in words.items():
                if i * ROOM <= x < (i + 1) * ROOM and j * ROOM <= y < (j + 1) * ROOM:
                    k = ((y - j * ROOM) * ROOM + (x - i * ROOM)) * 2
                    buf[k:k + 2] = v.to_bytes(2, "little")
            self.e.mu.mem_write(self.e.r32(self.coll[r][2] + 0x20), bytes(buf))

    def run(self, tele, cx, cy, tx, ty, words):
        self.load(words)
        self.e.mu.mem_write(self.levelrow + 4, bytes([tele]))
        self.here = self.rooms[(cx // ROOM, cy // ROOM)]
        self.xy, self.moved, self.line = (tx & 0xFFFFFFFF, ty & 0xFFFFFFFF), None, None
        ret = self.e.call(0x5ca360, 0, 0, ecx=0x2000, edx=self.unit) & 0xFFFFFFFF
        return ret, self.moved


def d2d(tele, cx, cy, tx, ty, words):
    """Line copy of strike.cpp's Teleport branch on the same grid: refused at
    Teleport 0, in town (not modelled: the case is a field level) or when
    `unit_blocked` (the plus of 5 subtiles, mask 0x0c09; off the level
    blocked) hits the target; else the player is put AT the target. The
    clear-line sampling for Teleport 2 is left out (cases use 0 and 1)."""
    if tele == 0: return 0, None
    def blocked(x, y):
        return not (0 <= x < SIDE and 0 <= y < SIDE) or words.get((x, y), 0) & 0x0c09
    if any(blocked(tx + dx, ty + dy) for dx, dy in ((0, 0), (-1, 0), (1, 0), (0, -1), (0, 1))): return 0, None
    return 1, (tx, ty)


def gen(n, seed):
    rng = random.Random(seed)
    for cid in range(n):
        words = {}
        for _ in range(rng.randint(0, 40)):                      # walls, doors, objects
            x, y = rng.randrange(SIDE), rng.randrange(SIDE)
            w, h = rng.randint(1, 12), rng.randint(1, 12)
            bits = rng.choice((0x1, 0x9, 0x5, 0x400, 0x800, 0x4, 0x8))
            for a in range(x, min(x + w, SIDE)):
                for b in range(y, min(y + h, SIDE)):
                    words[(a, b)] = words.get((a, b), 0) | bits
        for _ in range(rng.randint(0, 12)):                      # monsters' unit bit
            x, y = rng.randrange(SIDE), rng.randrange(SIDE)
            words[(x, y)] = words.get((x, y), 0) | 0x1000
        cx, cy = rng.randrange(2 * ROOM, 3 * ROOM), rng.randrange(2 * ROOM, 3 * ROOM)
        words.pop((cx, cy), None)
        mode = rng.random()
        if mode < 0.4:                                           # an ordinary hop: up to 40 away
            tx, ty = cx + rng.randint(-40, 40), cy + rng.randint(-40, 40)
        elif mode < 0.7:                                         # anywhere on the map
            tx, ty = rng.randrange(SIDE), rng.randrange(SIDE)
        elif mode < 0.85:                                        # at a monster
            mons = [p for p, v in words.items() if v & 0x1000] or [(cx + 5, cy)]
            tx, ty = rng.choice(mons)
        else:                                                    # past the map
            tx, ty = rng.randint(-30, SIDE + 30), rng.choice((-5, SIDE + 5))
        tele = 0 if rng.random() < 0.05 else 1
        flat = " ".join(f"{x} {y} {v}" for (x, y), v in sorted(words.items()))
        print(cid, tele, cx, cy, tx, ty, len(words), flat)


def parse(line):
    f = list(map(int, line.split()))
    cid, tele, cx, cy, tx, ty, n = f[:7]
    words = {(f[7 + 3 * k], f[8 + 3 * k]): f[9 + 3 * k] for k in range(n)}
    return cid, tele, cx, cy, tx, ty, words


def show(cid, ret, pos):
    print(cid, int(ret != 0), *(pos if pos else ("-", "-")))


def main():
    cmd = sys.argv[1] if len(sys.argv) > 1 else ""
    if cmd == "gen":
        return gen(int(sys.argv[2]), int(sys.argv[3]))
    if cmd == "d2d":
        for line in open(sys.argv[2]):
            cid, *rest = parse(line)
            show(cid, *d2d(*rest))
        return
    if cmd == "game":
        o = Oracle()
        for line in open(sys.argv[2]):
            cid, *rest = parse(line)
            ret, pos = o.run(*rest)
            show(cid, ret, pos)
        return
    if cmd == "diff":
        cases = {}
        if len(sys.argv) > 4:
            for line in open(sys.argv[4]):
                p = parse(line)
                cases[p[0]] = p
        g = [line.split() for line in open(sys.argv[2])]
        d = [line.split() for line in open(sys.argv[3])]
        kinds = {}
        for a, b in zip(g, d):
            if a == b: k = "same"
            elif a[1] == "1" and b[1] == "0": k = "game lands, d2d refuses"
            elif a[1] == "0" and b[1] == "1": k = "game refuses, d2d lands"
            else: k = "both land, other spot"
            if k != "same" and cases:
                _, tele, cx, cy, tx, ty, words = cases[int(a[0])]
                far = max(abs(tx // ROOM - cx // ROOM), abs(ty // ROOM - cy // ROOM)) > 1
                on_unit = bool(words.get((tx, ty), 0) & 0x1000)
                k += " (target room not near)" if far else " (target on a monster)" if on_unit else " (target blocked)" if a[1] == "1" and b[1] == "0" else ""
            kinds[k] = kinds.get(k, 0) + 1
        print(f"{len(g)} cases")
        for k, v in sorted(kinds.items(), key=lambda kv: -kv[1]): print(f"  {v:7d}  {k}")
        return
    if cmd == "probe":
        o = Oracle()
        cx, cy = 100, 100
        for r in (0, 1, 2, 3, 5):
            words = {(110 + dx, 100 + dy): 1 for dx in range(-r, r + 1) for dy in range(-r, r + 1)}
            ret, pos = o.run(1, cx, cy, 110, 100, words)
            print(f"target (110,100), blocked square radius {r}: do {ret} landed {pos}")
        ret, pos = o.run(1, cx, cy, 110, 100, {(110, 100): 0x1000})
        print(f"target on a unit bit 0x1000: do {ret} landed {pos}")
        for bits in (0x4, 0x8, 0x400, 0x800, 0x2, 0x10, 0x100, 0x2000):
            ret, pos = o.run(1, cx, cy, 110, 100, {(110, 100): bits})
            print(f"target word {bits:#x}: do {ret} landed {pos}")
        for dx in (1, 2):
            ret, pos = o.run(1, cx, cy, 110, 100, {(110 + dx, 100): 1})
            print(f"wall {dx} east of the target: landed {pos}")
        for t in ((199, 100), (120, 100), (79, 100), (40, 40), (39, 39), (159, 159), (160, 160), (-1, 100)):
            ret, pos = o.run(1, cx, cy, *t, {})
            print(f"open map, target {t}: do {ret} landed {pos}")
        ret, pos = o.run(0, cx, cy, 110, 100, {})
        print(f"Teleport 0: do {ret} landed {pos}")
        ret, pos = o.run(2, cx, cy, 110, 100, {})
        print(f"Teleport 2: do {ret} landed {pos}, line test args {o.line and tuple(hex(v) for v in o.line)}")
        return
    raise SystemExit(__doc__)


if __name__ == "__main__":
    main()
