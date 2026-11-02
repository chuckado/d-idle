"""monsterskill: the bot's shooter model (internal/combat/threat.go ShooterPeak, mirrored line for line from monstats columns) against the
tables, for every Act 1 / Act 2 spawn-list row that fires a missile (MissA1 / MissA2 / MissS1 / MissC set, or a skill whose srvdofunc is 85
or that has srvmissilea / srvmissile). D data only (patched stack). Columns: rangedtype (the bot's m.Ranged), the bot's peak and unblockable
per difficulty (raw column numbers: percent of the monster-level damage for a ratio row, not points), and the same peak in points
(monlvl L-DM x percent / 100 at the row's level, the rule of docs/research/monster-attack-trace.md).
Run: tools/emu/run-on-worker.sh msk_bot.py > /tmp/d2agent-monsterskill/bot.txt"""
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
    return [dict(zip(h, l.split('\t'))) for l in L[i + 1:]]


def num(s):
    try:
        return int(s)
    except (ValueError, TypeError):
        return 0


def main():
    mons = {m['Id']: m for m in table('monstats')}
    skl = {s['skill']: s for s in table('skills')}
    monlvl = {num(r['Level']): r for r in table('monlvl')}
    ids = set()
    for l in table('levels'):
        if l.get('Act', '') in ('0', '1') and l.get('Name'):
            for k in range(1, 11):
                for p in ('mon', 'nmon', 'umon'):
                    if l.get('%s%d' % (p, k), '') in mons:
                        ids.add(l['%s%d' % (p, k)])
    ids |= {b for b in ('andariel', 'bloodraven', 'radament', 'summoner') if b in mons}
    sfx = ('', '(N)', '(H)')
    print('id ai ranged | source | bot peak raw N/NM/H | unblockable N/NM/H | peak points N/NM/H (L-DM x pct) | NoRatio')
    nfire = nrng_blank = nrng_set_nofire = 0
    for mid in sorted(ids, key=lambda i: int(mons[i]['hcIdx'])):
        m = mons[mid]
        src = [c for c in ('MissA1', 'MissA2', 'MissS1', 'MissC') if m.get(c)]
        for k in range(1, 9):
            s = skl.get(m.get('Skill%d' % k, ''))
            if s and (num(s.get('srvdofunc')) == 85 or s.get('srvmissilea') or s.get('srvmissile')):
                src.append('Skill%d:%s' % (k, s['skill']))
        ranged = m.get('rangedtype', '') not in ('', '0')
        if not src:
            if ranged:
                nrng_set_nofire += 1
                print(mid, m['AI'], 'ranged=1', '| no missile source (rangedtype set)')
            continue
        nfire += 1
        if not ranged:
            nrng_blank += 1
        peaks, unb, pts = [], [], []
        for d in range(3):
            phys = max(num(m['A1MaxD' + sfx[d]]), num(m['A2MaxD' + sfx[d]]))
            elem = max(num(m['El%dMaxD' % k]) for k in (1, 2, 3))      # the Go loader reads the base-difficulty El column only
            peak = max(phys, num(m['S1MaxD' + sfx[d]]), elem)
            peaks.append(peak)
            unb.append(phys == 0 and elem > 0)
            lv = num(m['Level' + sfx[d]])
            ldm = num(monlvl.get(min(lv, 110), {}).get(('L-DM', 'L-DM(N)', 'L-DM(H)')[d]))
            pts.append(0 if num(m.get('NoRatio')) else peak * ldm // 100)
            if num(m.get('NoRatio')):
                pts[-1] = peak
        print(mid, m['AI'], 'ranged=%d' % ranged, '|', ' '.join(src), '|', '/'.join(map(str, peaks)), '|', '/'.join('u' if u else '-' for u in unb),
              '|', '/'.join(map(str, pts)), '|', m.get('NoRatio') or '0')
    print('missile-firing rows %d, of them rangedtype blank (bot: melee) %d; rangedtype set without a missile source %d' % (nfire, nrng_blank, nrng_set_nofire))


main()
