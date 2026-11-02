"""monsterskill: print the monster skill slots of monstats rows (D, 1.14d patched stack).
Run: tools/emu/run-on-worker.sh msk_rows.py ID_PREFIX,ID_PREFIX,...  > /tmp/d2agent-monsterskill/out.txt
Also: msk_rows.py table:skills KEY_PREFIX,... prints rows of any excel table whose first column starts with a key."""
import os
import subprocess
import sys

scratch = os.environ.get('D2_SCRATCH', '/tmp/d2agent-monsterskill')
here = os.path.dirname(os.path.abspath(__file__))


def table(name):
    out = subprocess.run(['bash', os.path.join(here, 'mpq-table.sh'), name], capture_output=True, text=True).stdout
    L = [x for x in out.split('\n') if x.strip()]
    i = 0
    while not (L[i].startswith('Id\t') or L[i].startswith('skill\t') or L[i].startswith('Skill\t') or L[i].split('\t')[0] in ('Id', 'skill', 'Skill', 'Name')):
        i += 1
    h = L[i].split('\t')
    return h, [dict(zip(h, l.split('\t'))) for l in L[i + 1:]]


def main():
    a = sys.argv[1]
    if a.startswith('table:'):
        h, rows = table(a[6:])
        want = sys.argv[2].split(',')
        cols = sys.argv[3].split(',') if len(sys.argv) > 3 else h
        print('\t'.join(cols))
        for r in rows:
            if any(r[h[0]].startswith(w) for w in want):
                print('\t'.join(r.get(c, '?') for c in cols))
        return
    h, rows = table('monstats')
    want = a.split(',')
    for r in rows:
        if any(r['Id'].startswith(w) for w in want):
            out = []
            for i in range(1, 9):
                if r['Skill%d' % i]:
                    out.append('%s/%s/%s' % (r['Skill%d' % i], r['Sk%dmode' % i], r['Sk%dlvl' % i]))
            print(r['Id'], r['AI'], 'A1=%s A2=%s S1=%s SC=%s' % (r['MissA1'], r['MissA2'], r['MissS1'], r['MissC']),
                  'SkDmg=%s' % r['SkillDamage'], 'ranged=%s melee=%s' % (r['rangedtype'], r['isMelee']),
                  'aidel=%s/%s/%s' % (r['aidel'], r['aidel(N)'], r['aidel(H)']), 'skills:', ' '.join(out),
                  'aip:', ','.join(r['aip%d' % i] for i in range(1, 9)),
                  'N:', ','.join(r['aip%d(N)' % i] for i in range(1, 9)),
                  'H:', ','.join(r['aip%d(H)' % i] for i in range(1, 9)), 'En=%s' % r['enabled'])


main()
