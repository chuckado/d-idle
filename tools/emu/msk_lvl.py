#!/usr/bin/env python3
"""monsterskill oracle R1: the skill levels game.exe gives a spawned monster (FUN_00573cb0, the loop at 0x5740e0..0x574135).
Spawns every Act 1 / Act 2 spawn-list row (levels.txt Act 0 / 1: mon, nmon, umon) with game.exe's own FUN_00555230 in a booted Blood Moor game at
difficulty 0 / 1 / 2 (a fresh game each), then for each skill slot of the row finds the unit's skill record with FUN_006439b0 and reads its level with
FUN_006442a0 (flag 0 and 1). The reading to compare: Sk*lvl + DifficultyLevels MonsterSkillBonus (0 / 3 / 7), present only when Sk*lvl > 0.
Also prints the unit's mode word for the slots, the ShamanFire / UnHolyBolt row that FUN_005cb0c0 would pick (stem + first + TransLvl).

  tools/emu/run-on-worker.sh --ttl 2400 msk_lvl.py > /tmp/d2agent-monsterskill/lvl.txt
"""
import os
import subprocess
import sys

import mm_launch

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
        return 0


def main():
    mons = {m['Id']: m for m in table('monstats')}
    sk = {s['skill']: num(s['Id']) for s in table('skills')}
    ids = set()
    for l in table('levels'):
        if l.get('Act', '') in ('0', '1') and l.get('Name'):
            for k in range(1, 11):
                for p in ('mon', 'nmon', 'umon'):
                    if l.get('%s%d' % (p, k), '') in mons:
                        ids.add(l['%s%d' % (p, k)])
    rows = sorted(ids, key=lambda i: int(mons[i]['hcIdx']))
    cases = bad = absent_bad = 0
    out = []
    for diff in (0, 1, 2):
        g = mm_launch.G(diff)
        e = g.e
        for n, mid in enumerate(rows):
            m = mons[mid]
            slots = [(k, m['Skill%d' % k], num(m['Sk%dlvl' % k])) for k in range(1, 9) if m.get('Skill%d' % k)]
            if not slots:
                continue
            u, x, y = g.unit(int(m['hcIdx']), n)
            if not u:
                out.append('spawn failed %s' % mid)
                continue
            for k, name, lv in slots:
                want = lv + (0, 3, 7)[diff] if lv > 0 else None
                ent = e.call(0x6439b0, -1 & 0xFFFFFFFF, ecx=u, edx=sk[name])
                got = e.call(0x6442a0, u, ent, 1) if ent else None
                raw = e.r32(ent + 0x28) if ent else None
                cases += 1
                ok = got == want
                if not ok:
                    bad += 1
                out.append('%s diff %d %s slot %d %s Sk*lvl %d want %s got %s raw+0x28 %s %s' % ('ok ' if ok else 'BAD', diff, mid, k, name, lv, want, got, raw, ''))
    print('\n'.join(out))
    print('RESULT cases %d (rows with skill slots x 3 difficulties), mismatches %d' % (cases, bad))


main()
