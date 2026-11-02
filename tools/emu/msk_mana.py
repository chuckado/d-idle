#!/usr/bin/env python3
"""monsterskill oracle R2: does a MONSTER pay mana? game.exe's own FUN_0056c160 (the check) and FUN_0056bfe0 (the payment), run for a
unit whose type dword is 1 (a monster) and, as the control, 0 (a player), over every Skills.txt row x levels 1..60 x mana 0 and mana 0x20000.
Reuses tools/emu/sk_mana.py's Oracle (same hooks: the skill record, its level and stat 8 are fed; the check and payment logic is game.exe's).

  tools/emu/run-on-worker.sh --ttl 1500 msk_mana.py > /tmp/d2agent-monsterskill/mana.txt

Prints: cases, and for the monster: check != 0 count, payments > 0 count; same for the player control (where the formula bites).
"""
import sk_mana


def main():
    o = sk_mana.Oracle()
    rows = [o.row(s) for s in range(o.count)]
    res = {}
    for utype in (1, 0):
        o.e.w32(o.unit, utype)
        n = checks = paid_n = 0
        for skill in range(o.count):
            for lvl in range(1, 61):
                for mana in (0, 0x20000):
                    check, paid, ret = o.run(skill, lvl, mana)
                    n += 1
                    checks += check != 0
                    if check and paid:
                        paid_n += 1
        res[utype] = (n, checks, paid_n)
        print('unit type', utype, 'cases', n, 'check passed', checks, 'payments > 0', paid_n)
    costly = sum(1 for r in rows if r['mana'] or r['lvlmana'])
    print('skill rows', o.count, 'rows with a mana cost', costly)
    m = res[1]
    print('RESULT monster: cases %d, check refused %d, paid something %d' % (m[0], m[0] - m[1], m[2]))


main()
