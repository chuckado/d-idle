#!/usr/bin/env python3
"""decomp43 (P07): FUN_00664860's Act 2 branch. In a level of act index 1 the variant code of a monster whose MonStats word +2
(its BaseId) is 0 (skeleton1's row) or 0xaa comes from the tables at 0x7489a8 / 0x748a18 (a pointer per layer, a code per
variant index) instead of the MonStats2 list. Prints those tables for layer 7 and the block chance of skeleton rows in an Act 1
level (2) and an Act 2 level (argument) for every variant index.
    mb_act2.py [LEVEL]   (default 41)
"""
import sys

import mm_launch


def code(v):
    return v.to_bytes(4, 'little').decode('latin-1').rstrip('\0 ')


def main():
    lid = 0
    g = mm_launch.G(0, lid=2)
    e = g.e
    t = e.r32(0x744304)
    base, n = e.r32(t + 0xa78), e.r32(t + 0xa80)
    for name, tab in (('0x7489a8', 0x7489a8), ('0x748a18', 0x748a18)):
        for layer in range(16):
            p = e.r32(tab + layer * 4)
            if p:
                print('table', name, 'layer', layer, [code(e.r32(p + 4 * i)) for i in range(8)])
    for level in (2,) + tuple(int(a) for a in sys.argv[1:]):
        try:
            g = mm_launch.G(0, lid=level) if level != 2 else g
        except Exception as ex:
            print('level', level, 'could not be brought up:', type(ex).__name__, str(ex)[:80])
            continue
        e = g.e
        for row in (0, 4, 19, 83, 170, 648, 649, 725):
            u, x, y = g.unit(row, 0)
            data = e.r32(u + 0x14)
            res = []
            for idx in range(6):
                e.mu.mem_write(data + 0xb, bytes([idx]))
                res.append(e.call(0x622720, u, 0) & 0xffffffff)
            print('level', level, 'row', row, 'chance by idx 0..5', res)


main()
