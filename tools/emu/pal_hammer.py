# SPDX-License-Identifier: GPL-3.0-or-later
"""palskill: Blessed Hammer's flight by game.exe's own code.

    python3 pal_hammer.py [level] [dx dy]     # trace: frame, x, y (16.16 and subtiles), distance

Builds the missile as the do FUN_005d0040 does (FUN_0059fa30 with flags 0x20,
the `blessedhammer` row (92), the skill (112) and its level; then
FUN_00648cf0(path, 14) and FUN_00649970(path, missile, 0), at 0x5d0107 and
0x5d0113) from a Quill Rat shooter in missiles.py's Blood Moor game, then steps
the missile with its pSrvDoFunc (table 0x73c768, row +0xc = 1, FUN_005ae1f0)
for 130 frames and prints the position after each. The hit function is
hooked to record and pass (no foe). Runs next to upstream's tools/emu."""
import sys

import missiles


def main():
    level = int(sys.argv[1]) if len(sys.argv) > 1 else 1
    dx, dy = (int(sys.argv[2]), int(sys.argv[3])) if len(sys.argv) > 3 else (20, 0)
    game = missiles.Game()
    e = game.e
    import random
    rng = random.Random(5)
    for _ in range(1000):
        room = rng.choice(game.rooms)
        x, y = room[1] + rng.randrange(room[3]), room[2] + rng.randrange(room[4])
        if all(game.collision(x + a, y + b) & 0x1f0f == 0 for a in range(-12, 13) for b in range(-12, 13)): break
    shooter = game.unit(63, x, y)
    src, sx, sy, _ = shooter
    mid = 92
    p = e.alloc(0x60)
    fields = {0: 0x20, 1: src, 2: src, 4: mid, 5: sx, 6: sy, 7: sx + dx, 8: sy + dy, 0xb: 112, 0xc: level}
    for i, v in fields.items(): e.w32(p + 4 * i, v)
    m = e.call(0x59fa30, ecx=game.g, edx=p)
    if not m: raise SystemExit("no missile")
    path = e.r32(m + 0x2c)
    e.call(0x648cf0, path, 14)
    e.call(0x649970, path, m, 0)
    rng_frames = e.s32(e.r32(m + 0x14) + 0x10) << 16 >> 16
    fn = e.r32(0x73c768 + 4 * game.row(mid, 0xc, 2))
    print("#start", sx, sy, "range", rng_frames, "fn", hex(fn), "vel", game.row(mid, 0x9a, 1), "maxvel", game.row(mid, 0x9c, 1), "accel", game.row(mid, 0x9e, 2))
    game.foe = 0
    game.hit.clear()
    for frame in range(1, 131):
        e.call(fn, ecx=game.g, edx=m)
        x16, y16 = e.s32(path), e.s32(path + 4)
        left = e.s32(e.r32(m + 0x14) + 0x10) << 16 >> 16
        print(frame, x16, y16, f"{(x16 - sx * 65536) / 65536:.3f} {(y16 - sy * 65536) / 65536:.3f}", left)
        if left < 1: break


if __name__ == "__main__":
    main()
