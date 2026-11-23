# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp23: where does d2d's player path leave game.exe's on a real map? Per start of DUMP whose first path
differs (ww_path.py compare), run on the SAME blocked bits, with no walk-back (free tile unit only):
  game  FUN_00679ed0 (toward, then the search) and FUN_0067b850 alone, natively;
  port  the fork's tools/emu/moves.py player_path / search_path (line for line copies of rules::player_path),
        imported from $WW_PYPATH (the fork checkout's tools/emu).
    ww_run.sh ww_search.py DUMP KEYFILE     (KEYFILE: lines `seed level warp dir`)
"""
import struct
import sys

import emu
import moves
import ww_path


def main():
    dump, keys = sys.argv[1], [tuple(l.split()) for l in open(sys.argv[2]) if l.strip()]
    grids, starts = ww_path.parse(dump)
    D = {tuple(f[1:5]): f for f in starts}
    rig = ww_path.Rig()
    e = rig.e
    for key in keys:
        f = D[key]
        ux, uy, vel, bits = grids[key]
        rig.setup(bits)
        sx, sy = int(f[5]) >> 16, int(f[6]) >> 16
        x, y, tx, ty = sx - ux + ww_path.BASE, sy - uy + ww_path.BASE, ww_path.BASE, ww_path.BASE
        blocked = lambda a, b: rig.blocked(a, b)
        # the same ctx moves.py builds
        p, unit, ctx = rig.path, rig.unit, e.alloc(0x40)
        e.w32(p, x << 16 | 0x8000); e.w32(p + 4, y << 16 | 0x8000); e.w32(p + 0x3c, 7)
        e.w32(ctx, y << 16 | x); e.w32(ctx + 4, ty << 16 | tx); e.w32(ctx + 8, 1)
        e.mu.mem_write(ctx + 0x14, bytes([1])); e.w32(ctx + 0x18, 0x49); e.w32(ctx + 0x30, p)
        out = {}
        for name, addr in (("game-player", 0x679ed0), ("game-search", 0x67b850), ("game-toward", 0x679c80)):
            e.w32(p + 0x24, 0); e.w32(p + 0x28, 0)
            n = e.call(addr, ecx=ctx, limit=5000000)
            out[name] = [struct.unpack("<HH", e.read(p + 0x9c + 4 * i, 4)) for i in range(n)]
        out["port-player"] = moves.player_path(x, y, tx, ty, 1, True, blocked)
        out["port-search"] = moves.search_path(x, y, tx, ty, True, blocked)
        out["port-toward"] = moves.toward_path(x, y, tx, ty, 0x49, 1, blocked)
        rel = lambda pts: [(a - ww_path.BASE + ux, b - ww_path.BASE + uy) for a, b in pts]
        print("key", key, "start", (sx, sy), "unit", (ux, uy))
        print("  d2d     ", f[8:])
        for name in ("game-player", "port-player", "game-search", "port-search", "game-toward", "port-toward"):
            print("  %-12s" % name, rel(out[name]))


if __name__ == "__main__":
    main()
