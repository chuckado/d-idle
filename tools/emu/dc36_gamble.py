# SPDX-License-Identifier: GPL-3.0-or-later
"""game.exe's gamble price (FUN_0062efb0 mode 2 -> FUN_00629370) as the oracle.

(Scratch oracle of card decomp36. It imports skill_tips / price_oracle of upstream's tools/emu: run it on the builder with
that directory on PYTHONPATH; the d2d side is the fork branch named in docs/research/dc36-lane-b.md.)

    python3 dc36_gamble.py probe                      # does the price depend on anything but the code and clvl?
    python3 dc36_gamble.py gen > cases.tsv            # every code a gamble tab can hold x clvl 1..99 x reduced 0 / 10
    python3 dc36_gamble.py game cases.tsv > game.tsv  # `id price`
A case is `id code clvl reduced`.
"""
import struct
import sys

import price_oracle as po


class G(po.Oracle):
    def gprice(self, code, clvl, reduced, quality=4, ident=0, eth=0, npc="gheed", diff=0, mode=2, w30=1):
        e = self.e
        mem = bytearray(0x100)
        struct.pack_into("<I", mem, 0, quality)
        struct.pack_into("<I", mem, 0x18, (0x10 if ident else 0) | (0x400000 if eth else 0))
        struct.pack_into("<i", mem, 0x28, -1)
        struct.pack_into("<H", mem, 0x30, w30)
        e.mu.mem_write(self.data, bytes(mem))
        e.w32(self.item + 4, self.cls[code])
        self.istats = {}
        self.pstats = {po.STAT_CLVL: clvl, po.STAT_REDUCED: reduced}
        r = e.call(0x62efb0, diff, self.quests, self.hc[npc], mode, ecx=self.player, edx=self.item)
        return r - (1 << 32) if r & 0x80000000 else r


def codes(g):
    """The bases a gamble tab can hold: gamble.txt's codes and the exceptional / elite versions they upgrade to (FUN_005786a0)."""
    import struct
    rows = dict(g.items)
    out = set(r.get("code", "").strip() for r in po.txt(g.e, "gamble")) - {""}
    for code in list(out):
        for off in (0x88, 0x8c):
            up = struct.pack("<I", g.e.r32(rows[code] + off)).decode("latin1").strip("\0 ")
            if up and up != "0": out.add(up)
    return [c for c, _ in g.items if c in out and c in g.cls]


def main():
    a = sys.argv[1:]
    if not a: sys.exit(__doc__)
    g = G()
    if a[0] == "probe":
        for code in ("cap", "rin", "amu", "jav", "bkf", "lsd", "gsd", "7wb"):
            if code not in g.cls: continue
            base = g.gprice(code, 30, 0)
            var = {(q, i, et, n, d): g.gprice(code, 30, 0, q, i, et, n, d) for q in (1, 2, 4, 6, 7) for i in (0, 1) for et in (0, 1)
                   for n in ("gheed", "akara", "charsi") for d in (0, 2)}
            print(code, base, sorted(set(var.values())), [g.gprice(code, 30, 0, w30=w) for w in (0, 1, 2, 50, 99)],
                  [g.gprice(code, c, 0) for c in (1, 5, 6, 30, 99)])
    elif a[0] == "gen":
        for code in codes(g):
            for clvl in range(1, 100):
                for red in (0, 10):
                    print(f"{code}.{clvl}.{red}", code, clvl, red)
    elif a[0] == "game":
        for line in open(a[1]):
            f = line.split()
            if f: print(f[0], g.gprice(f[1], int(f[2]), int(f[3])))


if __name__ == "__main__":
    main()
