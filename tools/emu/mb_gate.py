#!/usr/bin/env python3
"""decomp43 (P07): which MonStats.txt column is FUN_00622720's gate A (monster record byte +0xf & mask 0x6ce270)?

Reads the compiled MonStats table (game.exe memory, 0x744304 -> +0xa78, stride 0x1a8) in a booted game and, for every 0/1 column of
monstats.txt, counts the rows where the gate bit of the dword at +0xc equals the column. Prints the columns that agree on (nearly) all
rows, plus the MonStats2 layer 7 (SH) data of every row: the flag column, the SHv list, and game.exe's own count byte (+0x15+7) and
variant bytes (+0x26 + 7*12 + i) of the compiled MonStats2 row (table 0x744304 -> +0xa90, stride 0x134).
"""
import csv
import io
import sys

import emu
import mm_launch


def table(name):
    raw = emu.mpq_cache().read('data\\global\\excel\\%s.txt' % name).decode('latin-1')
    r = list(csv.reader(io.StringIO(raw), delimiter='\t'))
    return [dict(zip(r[0], x)) for x in r[1:] if x and x[0] and x[0] != 'Expansion']   # the separator row is not a record


def main():
    g = mm_launch.G(0)
    e = g.e
    items(e)
    mask = e.r32(0x6ce270) & 0xff
    t = e.r32(0x744304)
    base, n = e.r32(t + 0xa78), e.r32(t + 0xa80)
    rows = table('monstats')
    print('mask', hex(mask), 'rows game', n, 'rows txt', len(rows))
    gate = [bool((e.r32(base + i * 0x1a8 + 0xc) >> 24) & mask) for i in range(min(n, len(rows)))]
    print('gate A rows:', [rows[i]['Id'] for i, v in enumerate(gate) if v])
    for col in rows[0]:
        vals = [r[col] for r in rows[:len(gate)]]
        dis = [rows[i]['Id'] for i, v in enumerate(vals) if (v not in ('', '0')) != gate[i]]
        if len(dis) <= 6:
            print('column', col, 'disagrees on', len(dis), dis)
    for bit in range(32):
        if [bool(e.r32(base + i * 0x1a8 + 0xc) >> bit & 1) for i in range(len(gate))] == gate:
            print('dword +0xc bit', bit, 'is the gate')
    ms2 = {r['Id']: r for r in table('monstats2')}
    b2, n2 = e.r32(t + 0xa90), e.r32(t + 0xa98)
    ms2rows = list(ms2.values())
    print('monstats2 game rows', n2, 'txt rows', len(ms2rows))
    for i in range(min(n, len(rows))):
        w = e.r16(base + i * 0x1a8 + 0x18)
        if w >= n2:
            continue
        r2 = b2 + w * 0x134
        cnt = e.read(r2 + 0x15 + 7, 1)[0]
        var = list(e.read(r2 + 0x26 + 7 * 12, 12))[:cnt]
        row = rows[i]
        txt = ms2.get(row['MonStatsEx'], {})
        print('S', i, row['Id'], 'gate', int(gate[i]), 'sh_flag', txt.get('SH'), 'SHv', txt.get('SHv', ''), 'game_count', cnt, 'game_var_bytes', var, 'mBL', txt.get('mBL'))



def items(e):
    """every component code's item record (FUN_00633640) and its word +0x11e, beside the txt row of that code."""
    rows = {}
    for name in ('armor', 'weapons', 'misc'):
        for i, r in enumerate(table(name)):
            rows.setdefault(r.get('alternategfx') or r['code'], []).append((name, i, r))
            rows.setdefault(r['code'], []).append((name, i, r))
    types = [r.get('Code') for r in table('ItemTypes')]
    cnt = e.s32(0x96bcb4)
    tmp = e.alloc(16)
    for i in range(cnt):
        code = e.r32(e.r32(0x96bcac) + 4 * i)
        name = code.to_bytes(4, 'little').decode('latin-1').rstrip('\0 ')
        rec = e.call(0x633640, code, tmp)
        if not rec:
            continue
        w = e.r16(rec + 0x11e)
        txt = rows.get(name, [])
        print('I', name, 'w11e', w, [(n, r['code'], r.get('type'), r.get('component'), r.get('wclass')) for n, _, r in txt[:2]])
    print('ItemTypes row 2 =', types[2], 'row 51 =', types[51] if len(types) > 51 else None)


main()
