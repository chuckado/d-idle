# SPDX-License-Identifier: GPL-3.0-or-later
"""palskill: how often one Blessed Hammer hits a monster standing in its orbit,
by game.exe's own hit function (FUN_005adf10, LastCollide / unit filters
native). A hostile Quill Rat with a huge life stands DIST subtiles east of the
shooter; one hammer is sent as pal_hammer.py does; the frames on which the
hit event FUN_005adb00 ran (the unit was hit) are printed.

    python3 pal_hammer_hits.py [dist dx dy]  -> `hits N frames f1 f2 ...`

Runs next to upstream's tools/emu (imports missiles.py)."""
import random
import sys

import drlg
import missiles
import monsters


class Game(missiles.Game):
    """missiles.Game without its hook on FUN_005adf10: the hit function runs natively."""

    def __init__(self, seed=3, lid=2):
        e = self.e = drlg.boot()
        self.g = monsters.new_game(e, seed)
        act = monsters.act_of(e, self.g, lid)
        lvl = drlg.find_level(e, e.r32(act + 0x48), lid)
        drlg.bring_up(e, lvl, seed, "list")
        self.rooms = []
        room1 = e.call(0x61a180, act)
        while room1:
            c = e.r32(room1 + 0x20)
            self.rooms.append((room1, *(e.s32(c + 4 * i) for i in range(4)), c + 0x24))
            room1 = e.r32(room1 + 0x7c)
        self.rows = e.r32(e.r32(0x744304) + 0xb64)
        self.hit, self.foe = [], 0
        e.hook(0x554200, lambda e: int(e.arg(0) == self.foe and self.foe != 0), 1)


def main():
    dist = int(sys.argv[1]) if len(sys.argv) > 1 else 4
    dx, dy = (int(sys.argv[2]), int(sys.argv[3])) if len(sys.argv) > 3 else (20, 0)
    game = Game()
    e = game.e
    rng = random.Random(5)
    for _ in range(1000):
        room = rng.choice(game.rooms)
        x, y = room[1] + rng.randrange(room[3]), room[2] + rng.randrange(room[4])
        if all(game.collision(x + a, y + b) & 0x1f0f == 0 for a in range(-14, 15) for b in range(-14, 15)): break
    src, sx, sy, _ = game.unit(63, x, y)
    foe = game.unit(63, x + dist, y)
    if not foe: raise SystemExit("no foe")
    fu = foe[0]
    for stat in (6, 7):
        e.call(0x627260, fu, stat, 0x7f000000, 0)
    game.foe = fu
    frames, current = [], [0]
    e.hook(0x5adb00, lambda e: (frames.append(current[0]), 0)[1], 0)
    p = e.alloc(0x60)
    fields = {0: 0x20, 1: src, 2: src, 4: 92, 5: sx, 6: sy, 7: sx + dx, 8: sy + dy, 0xb: 112, 0xc: 1}
    for i, v in fields.items(): e.w32(p + 4 * i, v)
    m = e.call(0x59fa30, ecx=game.g, edx=p)
    path = e.r32(m + 0x2c)
    e.call(0x648cf0, path, 14)
    e.call(0x649970, path, m, 0)
    fn = e.r32(0x73c768 + 4 * game.row(92, 0xc, 2))
    game.hit.clear()
    for frame in range(1, 131):
        current[0] = frame
        e.call(fn, ecx=game.g, edx=m)
        if e.s32(e.r32(m + 0x14) + 0x10) << 16 >> 16 < 1: break
    print("dist", dist, "hits", len(frames), "frames", *frames)


if __name__ == "__main__":
    main()
