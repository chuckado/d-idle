"""monsterskill: the monster skill slots of every Act 1 / Act 2 spawn-list row (levels.txt Act 0 and 1: mon1..10, nmon1..10, umon1..10)
plus their skills.txt rows (srvdofunc, srvmissilea, Delay, mana, range, state flags). D data only (patched stack, mpq-table.sh).
Run: tools/emu/run-on-worker.sh msk_skills.py > /tmp/d2agent-monsterskill/skills.txt"""
import os
import subprocess

here = os.path.dirname(os.path.abspath(__file__))


def table(name):
    out = subprocess.run(['bash', os.path.join(here, 'mpq-table.sh'), name], capture_output=True, text=True).stdout
    L = [x for x in out.split('\n') if x.strip()]
    i = 0
    while L[i].split('\t')[0] not in ('Id', 'skill', 'Name', 'Level', 'Act'):
        i += 1
    h = L[i].split('\t')
    return h, [dict(zip(h, l.split('\t'))) for l in L[i + 1:]]


_, mons = table('monstats')
_, skl = table('skills')
lh, lv = table('levels')
mons = {m['Id']: m for m in mons}
sk = {s['skill']: s for s in skl}
ids = {}
for l in lv:
    a = l.get('Act', '')
    if a in ('0', '1') and l.get('Name'):
        for k in range(1, 11):
            for p in ('mon', 'nmon', 'umon'):
                v = l.get('%s%d' % (p, k), '')
                if v in mons:
                    ids.setdefault(v, set()).add(l['Name'])
print('rows in Act 1/2 spawn lists:', len(ids))
cols = ['srvdofunc', 'srvmissilea', 'srvmissile', 'cltmissile', 'range', 'Delay', 'mana', 'lvlmana', 'prgstack', 'passive', 'attackrank', 'aurastate', 'interrupt', 'durability']
print('SKILL COLUMNS', [c for c in cols if c in skl[0]])
seen = {}
for mid in sorted(ids, key=lambda i: int(mons[i]['hcIdx'])):
    m = mons[mid]
    sl = []
    for n in range(1, 9):
        s = m.get('Skill%d' % n, '')
        if s:
            sl.append('%d:%s mode=%s lvl=%s' % (n, s, m.get('Sk%dmode' % n), m.get('Sk%dlvl' % n)))
            seen.setdefault(s, []).append(mid)
    print('ROW', mid, m['AI'], 'MissA1=%s MissA2=%s MissS1=%s MissC=%s' % (m['MissA1'], m['MissA2'], m['MissS1'], m['MissC']), '|', '; '.join(sl))
print()
for s, who in sorted(seen.items()):
    r = sk.get(s)
    print('SKILL', s, 'rows=%d' % len(who), who[:4], {c: r.get(c) for c in cols if c in r} if r else 'NOT IN skills.txt')
