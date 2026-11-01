#!/usr/bin/env python3
"""MONSTERMISSILE probe: does the missile element level bracket (FUN_0064b100 / 0064b1d0 / 0064b2a0) bite at levels 1 / 4 / 8?

Copy next to missiles.py (tools/emu/mm_run.sh style) and run: it spawns a Sand Maggot (goospit1: EMin 4 / Emax 4 / ELen 200,
MinELev 3 / MaxELev 3 / ELevLen 0, HitShift 1) and a Fallen Shaman (shafire1, MinELev 10), calls the element functions directly with
level 1, 4, 8 and prints min, max, length (256ths, as FUN_0064b860 stores them).
"""
import sys

import mm_launch

G = mm_launch.G


def sg(v): return v - (1 << 32) if v & 0x80000000 else v


def main():
    g = G(0)
    e = g.e
    for who, cls, mid in (("sandmaggot1", 70, 70), ("fallenshaman1", 58, 22)):
        u, x, y = g.unit(cls, 0)
        mu = None
        for lvl in (1, 4, 8):
            p = e.alloc(0x60)
            for i, v in {0: 0x21, 1: u, 4: mid, 5: x, 6: y, 7: x + 8, 8: y + 3, 0xc: lvl}.items(): e.w32(p + 4 * i, v)
            mu = e.call(0x59fa30, ecx=g.g, edx=p)
            out = [sg(e.call(f, mu, u, mid, lvl)) for f in (0x64b100, 0x64b1d0, 0x64b2a0)]
            phys = [sg(e.call(f, mu, u, mid, lvl)) for f in (0x64af20, 0x64aff0)]
            print(who, "row", mid, "level", lvl, "element min/max/len", out, "phys min/max (no owner part)", phys)


if __name__ == "__main__": main()
