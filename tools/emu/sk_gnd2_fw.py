# SPDX-License-Identifier: GPL-3.0-or-later
"""SKGROUND2: Fire Wall's two makers and what they drop, by game.exe's own code.

    python sk_gnd2_fw.py [CASES] [SEED]     # prints `G <case> <lvl> <frame> <dx> <dy>` per sub-missile

Runs next to upstream's tools/emu (missiles.py's Game: a Blood Moor game, every
room up). The cast (do 24, FUN_005c9ea0, read) puts two `firewallmaker` rows
(68) at the target point T aimed at T + (-dy, +dx) and T + (+dy, -dx), where
(dx, dy) = caster - T (params flags 0x21, owner, skill 51, the level), and one
`firewall` row at T. The makers are made here by the REAL FUN_0059fa30 with
exactly those params and stepped by their own missile do (row +0xc = 6,
FUN_005ae680: it drops SubMissile1 `firewall` at its own subtile each frame
through FUN_0059fa30, which is hooked after the maker exists to record the
params: row, start subtile, level) for the maker's Range 7 + 2 lvl frames.
`dx dy` are the sub-missile's subtile minus T; `frame` counts from 1.
The d2d side: tools/emu/sk_gnd2_test.cpp --fw (the same cases).
"""
import random
import sys

import missiles

MAKER, WALL = 68, 69


def main():
    n = int(sys.argv[1]) if len(sys.argv) > 1 else 200
    seed = int(sys.argv[2]) if len(sys.argv) > 2 else 1
    rng = random.Random(seed)
    game = missiles.Game()
    e = game.e
    spawns = []
    orig = bytes(e.read(0x59fa30, 5))

    def on_make(emu):
        p = emu.mu.reg_read(__import__("unicorn").x86_const.UC_X86_REG_EDX)
        spawns.append((emu.r32(p), emu.r32(p + 16), emu.s32(p + 20), emu.s32(p + 24), emu.r32(p + 0x30)))
        return 0

    def hooked(on):
        if on:
            e.hook(0x59fa30, on_make, 0)
        else:
            e.mu.mem_write(0x59fa30, orig)
            e.mu.ctl_remove_cache(0x59fa30, 0x59fa35)

    sites = []
    for room in game.rooms:
        rx, ry, rw, rh = room[1:5]
        for _ in range(120):
            x, y = rx + rng.randrange(rw), ry + rng.randrange(rh)
            if all(game.collision(x + dx, y + dy) & 0x1f0f == 0 for dx in range(-14, 15, 2) for dy in range(-14, 15, 2)):
                sites.append((x, y))
                break
    rng.shuffle(sites)
    print("# sites", len(sites), file=sys.stderr)
    done = 0
    for case in range(n):
        tx, ty = sites[case % len(sites)]
        lvl = rng.randint(1, 20)
        dist = rng.randint(3, 25)
        ang = rng.random() * 6.2831853
        import math
        cx, cy = tx + round(dist * math.cos(ang)), ty + round(dist * math.sin(ang))
        shooter = game.unit(63, cx, cy)
        if not shooter:
            continue
        src, cx, cy, _ = shooter
        dxc, dyc = cx - tx, cy - ty
        aims = [(tx - dyc, ty + dxc), (tx + dyc, ty - dxc)]
        print(f"C {case} {lvl} {tx} {ty} {cx} {cy}")
        for k, (ax, ay) in enumerate(aims):
            hooked(False)
            p = e.alloc(0x60)
            for i, v in {0: 0x21, 1: src, 4: MAKER, 5: tx, 6: ty, 7: ax, 8: ay, 0xb: 51, 0xc: lvl}.items():
                e.w32(p + 4 * i, v)
            m = e.call(0x59fa30, ecx=game.g, edx=p)
            if not m:
                print(f"# no maker {case} {k}", file=sys.stderr)
                continue
            fn = e.r32(0x73c768 + 4 * game.row(MAKER, 0xc, 2))
            frames = e.s32(e.r32(m + 0x14) + 0x10) << 16 >> 16
            hooked(True)
            for frame in range(1, 80):
                spawns.clear()
                e.call(fn, ecx=game.g, edx=m)
                for flags, row, sx, sy, lv in spawns:
                    print(f"G {case} {lvl} {k} {frame} {sx - tx} {sy - ty} row {row} flags {flags:#x} lv {lv}")
                if e.s32(e.r32(m + 0x14) + 0x10) << 16 >> 16 < 1:
                    break
            hooked(False)
            print(f"E {case} {k} frames {frames}")
        done += 1
    print("done", done, file=sys.stderr)


if __name__ == "__main__":
    main()
