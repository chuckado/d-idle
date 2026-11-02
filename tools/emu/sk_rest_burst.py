# SPDX-License-Identifier: GPL-3.0-or-later
"""SORCREST: which subtile offsets a Fire Ball (hit function 1, FUN_005a9a70) or
Glacial Spike (hit function 13, FUN_005aa8b0) burst takes in, by game.exe's own code.

    python sk_rest_burst.py [fireball|glacialspike [SPAN [LEVEL]]]   # prints `B size class a b hit` per case

Runs next to upstream's tools/emu (missiles.py's Game: a Blood Moor game). Per
case a monster unit of a class (size 1..3) stands at a free subtile F; the
missile row is made by the REAL FUN_0059fa30 (flags 0x21, owner a shooter unit,
start F - (a, b), skill id and LEVEL set: params +0x2c / +0x30) and its SrvHitFunc (row +0x?, table 0x73c840) is
called natively with the missile. The unit search (FUN_0056b7e0), the rooms
walk and the radius test run natively; the hostile test is hooked as in
missiles.py (the case's foe only) and the record application FUN_0056b9c0 is
hooked to note the unit and the record's address. `hit` = 1 when the foe got
the record. Also prints `R` lines: the record's fire / cold word for the case
(one record per burst is checked by the read; here only the foe is in the
room, so a shared record cannot be seen).
"""
import random
import sys

import missiles

ROWS = {"fireball": 62, "glacialspike": 96}
HITIDX = {"fireball": 1, "glacialspike": 13}   # Missiles.txt pSrvHitFunc (D)
SKILL = {"fireball": 47, "glacialspike": 55}
HITFUNC = 0x73c840   # SrvHitFunc table (FUN_005a9a70 is entry 1)


def main():
    name = sys.argv[1] if len(sys.argv) > 1 else "fireball"
    span = int(sys.argv[2]) if len(sys.argv) > 2 else 7
    level = int(sys.argv[3]) if len(sys.argv) > 3 else 1
    row = ROWS[name]
    game = missiles.Game()
    e = game.e
    rng = random.Random(1)
    applied = []

    def on_apply(en):
        applied.append((en.arg(0), en.arg(1)))
        return 1
    e.hook(0x56b9c0, on_apply, 3)
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
    for cls in (190, 19, 10, 63, 58, 2, 5, 38, 156, *range(0, 200)):
        if all(len(v) >= 1 for v in found.values()): break
        u = game.unit(cls, *probe)
        if u and u[3] in found and len(found[u[3]]) < 1:
            found[u[3]].append(cls)
    print("# classes by size:", found, file=sys.stderr)
    hitfn = e.r32(HITFUNC + 4 * HITIDX[name])
    print("# SrvHitFunc index", HITIDX[name], "-> %#x" % hitfn, "sHitPar1", game.row(row, 0x4c, 4), file=sys.stderr)
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
                    if game.room_at(sx, sy) is None: continue
                    p = e.alloc(0x60)
                    for i, v in {0: 0x21, 1: src, 4: row, 5: sx, 6: sy, 7: sx + 10, 8: sy, 0xb: SKILL[name], 0xc: level}.items(): e.w32(p + 4 * i, v)
                    m = e.call(0x59fa30, ecx=game.g, edx=p)
                    if not m: print(f"B {size} {cls} {a} {b} nomissile"); continue
                    del applied[:]
                    e.call(hitfn, 0, ecx=game.g, edx=m)
                    print(f"B {size} {cls} {a} {b} {1 if any(u == fu for u, _ in applied) else 0}")
    print("done", file=sys.stderr)


if __name__ == "__main__":
    main()
