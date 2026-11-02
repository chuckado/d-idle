#!/usr/bin/env python3
"""monsterskill: which branch of FUN_005fd470 (the monster's "may I use this skill now" test, V read 0x5fd470..0x5fd77d) each Act 1 / Act 2
monster skill takes. The branch is chosen by fields of the Skills record game.exe loaded (read from game memory after the loaders ran):
  id 0xa7 -> never; record +8 & mask[0] (byte at 0x6ce268) -> the target-corpse branch; id 0xa4 -> needs a target; srvdofunc (+0x2e) 0x4d / 0x4e
  -> the Leap spot test; id 0xc7, 0xb8; srvdofunc 0x43 -> the path test (mask 0xc01); id 0xcb -> the turret test (0x805); everything else -> 1.
Run: tools/emu/run-on-worker.sh msk_usable.py > /tmp/d2agent-monsterskill/usable.txt"""
import os
import subprocess

import skill_tips

here = os.path.dirname(os.path.abspath(__file__))


def table(name):
    out = subprocess.run(['bash', os.path.join(here, 'mpq-table.sh'), name], capture_output=True, text=True).stdout
    L = [x for x in out.split('\n') if x.strip()]
    i = 0
    while L[i].split('\t')[0] not in ('Id', 'skill', 'Name', 'Level', 'Act'):
        i += 1
    h = L[i].split('\t')
    return [dict(zip(h, l.split('\t'))) for l in L[i + 1:]]


def num(s):
    try:
        return int(s)
    except ValueError:
        return -1


def main():
    e = skill_tips.boot()
    tables = e.r32(0x744304)
    base, count = e.r32(tables + 0xb98), e.r32(tables + 0xba0)
    mask0 = e.read(0x6ce268, 1)[0]
    print('skills', count, 'mask[0] =', hex(mask0))
    mons = {m['Id']: m for m in table('monstats')}
    skl = {s['skill']: num(s['Id']) for s in table('skills')}
    used = {}
    for l in table('levels'):
        if l.get('Act', '') in ('0', '1') and l.get('Name'):
            for k in range(1, 11):
                for p in ('mon', 'nmon', 'umon'):
                    mid = l.get('%s%d' % (p, k), '')
                    if mid in mons:
                        m = mons[mid]
                        for n in range(1, 9):
                            if m.get('Skill%d' % n):
                                used.setdefault(m['Skill%d' % n], set()).add(mid)
    for name in sorted(used):
        sid = skl[name]
        r = base + sid * 0x23c
        flag8 = e.read(r + 8, 1)[0]
        do = int.from_bytes(e.read(r + 0x2e, 2), 'little')
        if sid == 0xa7: br = 'never'
        elif flag8 & mask0: br = 'corpse+line (flag +8 bit 0)'
        elif sid == 0xa4: br = 'needs a target'
        elif do in (0x4d, 0x4e): br = 'leap spot'
        elif sid == 0xc7: br = 'id 0xc7'
        elif sid == 0xb8: br = 'id 0xb8'
        elif do == 0x43: br = 'path 0xc01'
        elif sid == 0xcb: br = 'turret 0x805'
        else: br = 'always usable (returns 1)'
        print('%-18s id %3d (0x%x) +8=%#04x srvdofunc %3d  -> %s  rows %d: %s' % (name, sid, sid, flag8, do, br, len(used[name]), ' '.join(sorted(used[name])[:5])))


main()
