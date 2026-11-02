"""monsterskill: scan monstats for skill slots whose Sk*lvl is <= 0 or blank (game.exe 0x5740eb: `test al, al; jle skip` = the skill is never added),
and count the slots. Run: tools/emu/run-on-worker.sh msk_scan.py > /tmp/d2agent-monsterskill/scan.txt"""
import os
import subprocess

here = os.path.dirname(os.path.abspath(__file__))
out = subprocess.run(['bash', os.path.join(here, 'mpq-table.sh'), 'monstats'], capture_output=True, text=True).stdout
L = [x for x in out.split('\n') if x.strip()]
i = 0
while not L[i].startswith('Id\t'):
    i += 1
h = L[i].split('\t')
rows = [dict(zip(h, l.split('\t'))) for l in L[i + 1:]]
tot = bad = 0
for r in rows:
    for n in range(1, 9):
        s = r.get('Skill%d' % n, '')
        if not s:
            continue
        tot += 1
        lv = r.get('Sk%dlvl' % n, '')
        if not lv.strip() or int(lv) <= 0:
            bad += 1
            print('NOLVL', r['Id'], r['AI'], n, s, r.get('Sk%dmode' % n), repr(lv), 'enabled=%s' % r.get('enabled'))
print('slots', tot, 'with lvl<=0 or blank', bad, 'rows', len(rows))
