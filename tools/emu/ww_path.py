# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp23 oracle: a player's walk to a clicked warp's tile unit, in game.exe.

    ww_path.py game DUMP > game.tsv     # DUMP = test_warp_walk --dump N  (the fork's fid/warp-walk-r2)
    ww_path.py compare DUMP game.tsv    # d2d's first path against game.exe's, per start; the outcome table

What runs for real (game.exe 1.14d under unicorn): FUN_00649970 (the path request) with
FUN_006498a0 (the end = the tile unit's own subtile: a static path's +0xc / +0x10, type 5),
FUN_00648cf0 (path type 7), FUN_00648120 and FUN_00648050 (a blocked end walked back, a
player's push-on), FUN_00679ed0 (the player's pather: toward, then the search), FUN_0064fe40,
FUN_00650840 (a walker's frame) and FUN_00641530 (the unit distance).
Hooked (the oracle's inputs, not claims): FUN_0064d910 and FUN_0064ff90 answer from the
dump's blocked bits (the same bits d2d's player walk asks), FUN_00463740 (the room lookup:
always a room), FUN_00649560 / FUN_00649400 (stamps: none), FUN_00620510 (unit size: the
player 2, a tile unit 0, as docs/research/re/drlg.md reads it), FUN_006503f0 (the chase
check: go on; a tile unit never moves).

The walk model (read in FUN_00548b00 case 5, FUN_00548a50, FUN_0057f190, FUN_00580c20, the
fork's collision.md "The pending interact's fields"): a walk is a path request plus frames
until FUN_00650840 returns 0; when it ends, FUN_00548b00 asks the unit distance (taken at 4
or less); out of reach it makes ONE more walk (pending +0x150), then goes idle.
"""
import struct
import sys

import emu

W = 64                                   # the dump's window: 129 x 129 subtiles round the tile unit
SIDE = 2 * W + 1
BASE = 1 << 14                           # the walk's subtile frame: the unit at BASE, BASE


class Rig:
    def __init__(self):
        self.e = e = emu.Emu()
        self.bits = None
        self.ux = self.uy = 0
        self.outside = 0
        e.hook(0x64d910, self._wall5, 5)
        e.hook(0x64ff90, self._wall_frame, 2)
        e.hook(0x463740, lambda cur: 1, 1)
        e.hook(0x649560, lambda cur: 0, 2)
        e.hook(0x649400, lambda cur: 0, 1)
        e.hook(0x620510, lambda cur: 2 if cur.r32(cur.arg(0)) == 0 else 0, 1)
        e.hook(0x6503f0, lambda cur: 1, 0)
        self.mark = e.brk

    def blocked(self, x, y):
        gx, gy = x - BASE + W, y - BASE + W
        if not (0 <= gx < SIDE and 0 <= gy < SIDE):
            self.outside += 1
            return True                    # off the window: counted, answered blocked
        return self.bits[gy * SIDE + gx] == "1"

    def _wall5(self, cur):
        return int(self.blocked(cur.arg(1) & 0xffff, cur.arg(2) & 0xffff))

    def _wall_frame(self, cur):
        v = cur.arg(0)
        return 0 if self.blocked(v & 0xffff, v >> 16) else 1

    # -- one case ---------------------------------------------------------------
    def setup(self, bits):
        e = self.e
        e.brk = self.mark
        self.bits = bits
        self.unit, self.path = e.alloc(0x100), e.alloc(0x200)
        self.tile, self.tile_path = e.alloc(0x100), e.alloc(0x40)
        e.w32(self.tile, 5); e.w32(self.tile + 0x2c, self.tile_path)
        e.w32(self.tile_path + 0xc, BASE); e.w32(self.tile_path + 0x10, BASE)
        e.w32(self.unit, 0); e.w32(self.unit + 0x2c, self.path)
        p = self.path
        room = e.alloc(0x100)               # FUN_00647fb0 tests every point against the room's subtile rectangle: all of the window
        e.w32(room + 0x4c, 0); e.w32(room + 0x50, 0); e.w32(room + 0x54, 0x10000); e.w32(room + 0x58, 0x10000)
        e.w32(p + 0x30, self.unit); e.w32(p + 0x1c, room); e.w32(p + 0x58, self.tile)
        e.w32(p + 0x50, 0x1c09); e.w32(p + 0x48, 0)
        e.mu.mem_write(p + 0x91, bytes((0x49, 0x46)))
        e.w32(p + 0x5c, 5)

    def place(self, fx, fy, velocity):
        e, p = self.e, self.path
        e.w32(p + 0, fx & 0xffffffff); e.w32(p + 4, fy & 0xffffffff)
        e.w32(p + 0x7c, velocity)

    def pos(self):
        return self.e.r32(self.path), self.e.r32(self.path + 4)

    def request(self):
        e, p = self.e, self.path
        e.call(0x648cf0, p, 7)
        e.w32(p + 0x10, 0)               # FUN_006498a0 sets it from the tile unit
        n = e.call(0x649970, p, self.unit, 0, limit=2000000) & 0xffffffff
        pts = [(e.r32(p + 0x9c + 4 * i) & 0xffff, e.r32(p + 0x9c + 4 * i) >> 16) for i in range(e.r32(p + 0x28))]
        return n, pts

    def frames(self):
        e, p = self.e, self.path
        for k in range(600):
            if not e.call(0x650840, self.unit, 0x400, limit=2000000) & 1: return k
        return -1

    def distance(self):
        return self.e.call(0x641530, self.unit, self.tile) & 0xffffffff


def parse(path):
    grids, starts = {}, []
    for line in open(path):
        f = line.split()
        if not f: continue
        if f[0] == "G": grids[tuple(f[1:4])] = (int(f[4]), int(f[5]), int(f[6]), f[7])
        elif f[0] == "D": starts.append(f)
        elif f[0] == "W": pass
    return grids, starts


def game(dump):
    grids, starts = parse(dump)
    rig = Rig()
    for f in starts:
        seed, level, warp, d = f[1:5]
        ux, uy, vel, bits = grids[(seed, level, warp)]
        sx, sy = int(f[5]), int(f[6])        # 16.16 subtiles, absolute
        rig.setup(bits)
        rig.outside = 0
        fx, fy = sx - (ux << 16) + (BASE << 16), sy - (uy << 16) + (BASE << 16)
        v = (vel << 8) * 100 // 100
        row = [seed, level, warp, d]
        walks = []
        taken = 0
        for walk in (1, 2):
            rig.place(fx, fy, v) if walk == 1 else None
            n, pts = rig.request()
            first = pts if walk == 1 else first
            steps = rig.frames() if n else 0
            fx, fy = rig.pos()
            dist = rig.distance()
            walks.append((n, steps, fx, fy, dist))
            if dist <= 4: taken = walk; break
        pts1 = " ".join(f"{a - BASE + ux},{b - BASE + uy}" for a, b in first)
        w = ";".join(f"{n}/{s}/{(x >> 16) - BASE + ux},{(y >> 16) - BASE + uy}/{dd}" for n, s, x, y, dd in walks)
        print("\t".join(map(str, row + [taken, len(first), pts1, w, rig.outside])))


def compare(dump, gamefile):
    grids, starts = parse(dump)
    g = {}
    for line in open(gamefile):
        c = line.rstrip("\n").split("\t")
        g[tuple(c[:4])] = c
    same = diff = 0
    for f in starts:
        key = tuple(f[1:5])
        d2d_pts = f[8:]
        c = g.get(key)
        if not c: continue
        gp = c[6].split() if c[6] else []
        if gp == d2d_pts: same += 1
        else:
            diff += 1
            if diff <= 10: print("DIFF", key, "d2d", d2d_pts[:8], "game", gp[:8])
    print(f"first path: {same} equal, {diff} differ of {same + diff}")


if __name__ == "__main__":
    if sys.argv[1] == "game": game(sys.argv[2])
    elif sys.argv[1] == "compare": compare(sys.argv[2], sys.argv[3])
    else: sys.exit(__doc__)
