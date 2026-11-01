# SPDX-License-Identifier: GPL-3.0-or-later
"""SKGROUND2: do two Fire Wall missiles on one monster both hit it every frame (game.exe)?
    python sk_ground_fw2.py     # prints `K <n missiles> <class> <frame> <hits that frame>`
A size-2 monster (class 19) with 1, 2 or 3 `firewall` missiles (row 69) in the subtiles whose
Size-1 footprints cover it (the monster's subtile and its left / right neighbours); every
missile is stepped each frame by its own do (FUN_005ae520); FUN_005adf10 hooked to count
the calls on the monster. d2d's `burned` list lets a monster be burned once a frame per skill."""
import sys

import missiles

ROW = 69


def main():
    game = missiles.Game()
    e = game.e
    site = next((r for r in game.rooms if r[3] > 60 and r[4] > 60), game.rooms[0])
    x, y = site[1] + 30, site[2] + 30
    for cls in (190, 19, 10):
        foe = game.unit(cls, x, y)
        fu, fx, fy, size = foe
        game.foe = fu
        shooter = game.unit(63, fx + 14, fy + 14)
        src = shooter[0]
        for n in (1, 2, 3):
            offs = [(0, 0), (-1, 0), (1, 0)][:n]
            ms = []
            for dx, dy in offs:
                p = e.alloc(0x60)
                for i, v in {0: 3, 1: src, 4: ROW, 5: fx + dx, 6: fy + dy, 7: fx + dx, 8: fy + dy, 0xc: 1}.items():
                    e.w32(p + 4 * i, v)
                ms.append(e.call(0x59fa30, ecx=game.g, edx=p))
            fn = e.r32(0x73c768 + 4 * game.row(ROW, 0xc, 2))
            for frame in range(1, 7):
                game.hit.clear()
                for m in ms:
                    e.call(fn, ecx=game.g, edx=m)
                hits = sum(1 for h in game.hit if h[0] == fu or h[1] == fu)
                print(f"K {n} class {cls} (size {size}) frame {frame} hits {hits}")


if __name__ == "__main__":
    main()
