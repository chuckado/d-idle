#!/usr/bin/env python3
"""decomp43 (P07 monster block): what FUN_00622720's two gates read, from game.exe's own memory (emulator).

    PYTHONPATH=<upstream tools/emu> python3 mb_explore.py
prints: the global mask 0x6ce270, the list at 0x744580 (count 0x744584), and for every component code in the table
0x96bcac (count 0x96bcb4) the item record FUN_00633640 finds (stride 0x1a8, table 0x96ca5c) and its word +0x11e.
"""
import mm_launch


def main():
    g = mm_launch.G(0)
    e = g.e
    print('mask', hex(e.r32(0x6ce270)))
    n = e.r32(0x744584)
    print('list count', n, [hex(e.r32(0x744580 + 4 * i)) for i in range(n)])
    cnt = e.s32(0x96bcb4)
    print('component codes', cnt)
    tmp = e.alloc(16)
    for i in range(cnt):
        code = e.r32(e.r32(0x96bcac) + 4 * i)
        name = code.to_bytes(4, 'little').decode('latin-1').rstrip('\0 ')
        rec = e.call(0x633640, code, tmp)
        w = e.r16(rec + 0x11e) if rec else None
        print('C', i, repr(name), hex(code), 'rec', hex(rec), 'w11e', w)


main()
