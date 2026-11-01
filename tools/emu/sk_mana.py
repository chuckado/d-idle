# SPDX-License-Identifier: GPL-3.0-or-later
"""game.exe's player mana check and payment as the oracle: FUN_0056c160 (the
check the skill start runs, flag 0 of FUN_0056f640) and FUN_0056bfe0 (the
payment, in the start after a srvstfunc or in the do FUN_0056f7f0 at the
action frame).

    python sk_mana.py rows                    # the mana columns of every skill row, as read
    python sk_mana.py gen 40000 1 > cases.tsv
    python sk_mana.py game cases.tsv > game.tsv
    test_sk_mana --sk-mana cases.tsv > d2d.tsv  # tools/emu/sk_mana_test.cpp
    diff game.tsv d2d.tsv

A case is `id skill lvl mana` (mana in 256ths). `game` prints `id check paid`:
check = FUN_0056c160's answer for a player whose current skill is the case's
skill at the case's level with the case's mana, paid = what FUN_0056bfe0 then
takes off stat 8 (0 when the check failed: the start refuses the skill and
nothing is cast). `game --raw` adds FUN_0056bfe0's own return value.

Hooked (not the claim): FUN_00620250 (the unit's current skill: a fake
record for the check, none for the payment so the item-charge branch is
skipped), FUN_00643ad0 (the record's item owner: -1, not a charge),
FUN_00644140 (the record's Skills.txt row), FUN_006442a0 (its level),
FUN_00625480 (stat 8: the case's mana), FUN_0063a400 (shapeshifted: no),
FUN_00639df0 (state 0x72 blood_mana: no) and FUN_006272b0 (the stat add:
recorded). The arithmetic (Skills +0x186 minmana, +0x188 manashift, +0x18a
mana, +0x18c lvlmana) runs natively.

`gen N SEED`: every skill row x levels 1..60 x the mana values around the
thresholds (0, formula - 1, formula, formula + 1, minmana x 256 - 1, x 256,
x 256 + 1), then N random cases (levels 1..99, mana 0..0x20000).
"""
import random
import sys

import skill_tips

RECORD = 0x23c
FAKE = 0x30000000


class Oracle:
    def __init__(self):
        self.e = e = skill_tips.boot()
        tables = e.r32(0x744304)
        self.skills, self.count = e.r32(tables + 0xb98), e.r32(tables + 0xba0)
        self.unit = e.alloc(0x200)                       # dword 0 = 0: a player
        self.skill = self.lvl = self.mana = 0
        self.current = FAKE
        self.added = None
        e.hook(0x620250, lambda e: self.current, 1)
        e.hook(0x643ad0, lambda e: 0xFFFFFFFF, 1)
        e.hook(0x644140, lambda e: self.skills + self.skill * RECORD, 0)
        e.hook(0x6442a0, lambda e: self.lvl, 3)
        e.hook(0x625480, lambda e: self.mana if e.arg(1) == 8 else 0, 3)
        e.hook(0x63a400, lambda e: 0, 1)
        e.hook(0x639df0, lambda e: 0, 2)
        e.hook(0x6272b0, self.add, 4)

    def add(self, e):
        if e.arg(1) == 8:
            v = e.arg(2)
            self.added = v - (1 << 32) if v & 0x80000000 else v
        return 0

    def row(self, skill):
        e, r = self.e, self.skills + skill * RECORD
        s16 = lambda a: int.from_bytes(e.read(a, 2), "little", signed=True)
        return {"startmana": s16(r + 0x184), "minmana": s16(r + 0x186), "manashift": e.read(r + 0x188, 1)[0],
                "mana": s16(r + 0x18a), "lvlmana": s16(r + 0x18c), "srvst": s16(r + 0x2c), "srvdo": s16(r + 0x2e),
                "usemanaondo": (e.read(r + 8, 1)[0] & 0x20) != 0}

    def run(self, skill, lvl, mana):
        self.skill, self.lvl, self.mana = skill, lvl, mana
        self.current = FAKE
        check = self.e.call(0x56c160, ecx=self.unit) & 0xFFFFFFFF
        self.current, self.added = 0, None
        ret = self.e.call(0x56bfe0, skill, lvl, ecx=0, edx=self.unit) & 0xFFFFFFFF
        paid = -self.added if self.added is not None else 0
        return check, paid, ret


def formula(r, lvl):
    return (r["mana"] + r["lvlmana"] * max(lvl - 1, 0)) << (r["manashift"] & 31)


def main():
    cmd = sys.argv[1] if len(sys.argv) > 1 else "rows"
    if cmd == "gen":
        n, seed = int(sys.argv[2]), int(sys.argv[3])
        o, rng, cid = Oracle(), random.Random(seed), 0
        for skill in range(o.count):
            r = o.row(skill)
            for lvl in range(1, 61):
                f, m = formula(r, lvl), r["minmana"] << 8
                for mana in sorted({0, f - 1, f, f + 1, m - 1, m, m + 1}):
                    if mana < 0: continue
                    print(cid, skill, lvl, mana); cid += 1
        for _ in range(n):
            print(cid, rng.randrange(o.count), rng.randint(1, 99), rng.randrange(0x20000)); cid += 1
        return
    o = Oracle()
    if cmd == "rows":
        for skill in range(o.count):
            r = o.row(skill)
            if r["mana"] or r["lvlmana"] or r["minmana"] or r["startmana"]:
                print(skill, " ".join(f"{k}={v}" for k, v in r.items()))
        return
    if cmd == "game":
        raw = "--raw" in sys.argv
        for line in open(sys.argv[2]):
            cid, skill, lvl, mana = map(int, line.split())
            check, paid, ret = o.run(skill, lvl, mana)
            if not check: paid = 0
            print(cid, int(check != 0), paid, *( [ret] if raw else []))
        return
    raise SystemExit(__doc__)


if __name__ == "__main__":
    main()
