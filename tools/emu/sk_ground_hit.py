# SPDX-License-Identifier: GPL-3.0-or-later
"""SKGROUND2 D4: which subtile offsets a standing Blizzard shard hits, by game.exe's own code.

    python sk_ground_hit.py [SPAN [ROW [FRAMES]]]    # ROW blizzard1 (default) or firewall; prints `H size class a b hit first count` per case

Runs next to upstream's tools/emu (missiles.py's Game: a Blood Moor game,
every room up). Per case: a monster unit of a class (size 1..3) stands at a
free subtile F; a `blizzard1` missile (row 159, the shard) is made by the REAL
FUN_0059fa30 (params flags 3, owner a shooter unit, row, start = F - (a, b),
level 1) and stepped by its own missile do (row +0xc, do 3 = FUN_005ae480 ->
FUN_005ae1f0: its unit search FUN_00641cb0 in the Size-2 footprint) for the
row's Range (9) frames. FUN_005adf10 (the hit) is hooked to record the unit and
end nothing else (the case's monster is the only hostile one, FUN_00554200
hooked, as in missiles.py). The 0x40 the standing shard marks in the room's
collision is cleared after each case.

`hit` is 1 when FUN_005adf10 was called with the monster at any of the frames,
`frame` the first one. The d2d side: tools/emu/sk_ground_test.cpp --hit.
"""
import random
import sys

import missiles

ROWS = {"blizzard1": 159, "firewall": 69}
SHARD = 159


def main():
    global SHARD
    span = int(sys.argv[1]) if len(sys.argv) > 1 else 7
    SHARD = ROWS[sys.argv[2]] if len(sys.argv) > 2 else 159
    nframes = int(sys.argv[3]) if len(sys.argv) > 3 else 9
    game = missiles.Game()
    e = game.e
    rng = random.Random(1)
    # monster classes per size: try a spread, keep the first three per size
    found = {1: [], 2: [], 3: []}
    free = []
    for room in game.rooms:
        rx, ry, rw, rh = room[1:5]
        for _ in range(60):
            x, y = rx + rng.randrange(rw), ry + rng.randrange(rh)
            if all(game.collision(x + dx, y + dy) & 0x1f0f == 0 for dx in range(-12, 13, 3) for dy in range(-12, 13, 3)):
                free.append((x, y))
                break
    rng.shuffle(free)
    probe = free.pop()
    sites = iter(free)
    for cls in (19, 63, 58, 2, 5, 38, 156, *range(0, 200)):
        if all(len(v) >= 2 for v in found.values()): break
        u = game.unit(cls, *probe)
        if u and u[3] in found and len(found[u[3]]) < 2:
            found[u[3]].append(cls)
    print("# classes by size:", found, file=sys.stderr)
    saved_row = game.row  # noqa
    rows = game.rows
    rng_range = game.row(SHARD, 0x96, 2)
    print("# shard Range", rng_range, file=sys.stderr)
    for size, classes in found.items():
        for cls in classes:
            x, y = next(sites)
            foe = game.unit(cls, x, y)
            if not foe: continue
            fu, fx, fy, fsize = foe
            game.foe = fu
            shooter = game.unit(63, fx + 14, fy + 14) or game.unit(63, fx - 14, fy - 14)
            if not shooter: print("# no shooter", cls, file=sys.stderr); continue
            src = shooter[0]
            for a in range(-span, span + 1):
                for b in range(-span, span + 1):
                    sx, sy = fx - a, fy - b
                    r = game.room_at(sx, sy)
                    if r is None: continue
                    idx = r[5] + 2 * ((sy - r[2]) * r[3] + (sx - r[1]))
                    word = e.r16(idx)
                    p = e.alloc(0x60)
                    for i, v in {0: 3, 1: src, 4: SHARD, 5: sx, 6: sy, 7: sx, 8: sy, 0xc: 1}.items(): e.w32(p + 4 * i, v)
                    m = e.call(0x59fa30, ecx=game.g, edx=p)
                    if not m: print(f"H {size} {cls} {a} {b} nomissile -"); continue
                    fn = e.r32(0x73c768 + 4 * game.row(SHARD, 0xc, 2))
                    game.hit.clear()
                    first = -1
                    total = 0
                    for frame in range(1, nframes + 1):
                        n = len(game.hit)
                        e.call(fn, ecx=game.g, edx=m)
                        got = sum(1 for h in game.hit[n:] if h[0] == fu or h[1] == fu)
                        if got and first < 0: first = frame
                        total += got
                    print(f"H {size} {cls} {a} {b} {1 if first >= 0 else 0} {first} {total}")
                    # undo the shard's own mark in the room's collision
                    for rr in game.rooms:
                        pass
                    e.mu.mem_write(idx, word.to_bytes(2, "little"))
    print("done", file=sys.stderr)


if __name__ == "__main__":
    main()
