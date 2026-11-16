# SPDX-License-Identifier: GPL-3.0-or-later
"""game.exe's death losses (FUN_00535ab0 -> FUN_005357d0 gold, FUN_005359f0 experience) as the oracle.

(Scratch oracle of card decomp36. It imports skill_tips / price_oracle of upstream's tools/emu: run it on the builder with
that directory on PYTHONPATH; the d2d side is the fork branch named in docs/research/dc36-lane-b.md.)

    python3 dc36_death.py gen N SEED > cases.tsv
    python3 dc36_death.py game cases.tsv [killer] > game.tsv
    test_death_penalty --death-penalty cases.tsv > d2d.tsv ; diff game.tsv d2d.tsv

A case is `id class diff level exp gold bank`; `game` prints `id gold bank exp xploss goldlost pile`
(stats 0xe, 0xf, 0xd after; the experience the corpse is told it lost (FUN_005391e0's argument);
stat 0xaf goldlost; the pile FUN_00535510 is asked to drop). The game record is a single-player one
(+0x6a = 3, difficulty byte +0x6d). killer: 0 none (default), 1 a monster without owner.
FUN_005357d0 / FUN_005359f0 / FUN_00535ab0 and the stat setter and getter run for real; hooked are
FUN_00535510 (the drop, records its amount), FUN_0058f0d0 (a monster's owner: none), FUN_005531c0 and
FUN_005391e0 (the corpse record that takes the lost experience).
"""
import random
import sys

import skill_tips
from unicorn.x86_const import UC_X86_REG_EDX


class Death:
    def __init__(self):
        self.e = e = skill_tips.boot()
        self.game = e.alloc(0x200)
        self.unit = e.alloc(0x200)
        self.killer = e.alloc(0x200)
        self.stat_list = e.alloc(0x100)
        self.rec = e.alloc(0x40)
        e.w32(self.unit + 0x5C, self.stat_list)
        e.w32(self.unit + 0xc, 0x1234)
        e.w32(self.killer, 1)                    # a monster
        self.pile = 0
        self.xploss = 0
        e.hook(0x535510, self.on_pile, 2)
        e.hook(0x58f0d0, lambda e: 0, 0)
        e.hook(0x5531c0, lambda e: self.rec, 1)
        e.hook(0x5391e0, self.on_xp, 0)

    def on_pile(self, e):
        self.pile = e.arg(1)
        return 1

    def on_xp(self, e):
        self.xploss = e.mu.reg_read(UC_X86_REG_EDX)
        return 0

    def thresholds(self, cls, level):
        return self.e.call(0x611800, cls, level)

    def get(self, stat): return self.e.call(0x625480, self.unit, stat, 0) & 0xFFFFFFFF
    def put(self, stat, v): self.e.call(0x627260, self.unit, stat, v, 0)

    def die(self, cls, diff, level, exp, gold, bank, killer=0):
        e = self.e
        e.mu.mem_write(self.stat_list, bytes(0x100))
        e.w32(self.stat_list + 0x10, 0x80000000)
        e.w32(self.unit, 0); e.w32(self.unit + 4, cls)
        e.mu.mem_write(self.game, bytes(0x200))
        e.mu.mem_write(self.game + 0x6a, bytes([3]))
        e.mu.mem_write(self.game + 0x6d, bytes([diff]))
        for stat, v in ((12, level), (13, exp), (14, gold), (15, bank)): self.put(stat, v)
        self.pile = 0; self.xploss = 0
        e.call(0x535ab0, self.killer if killer else 0, ecx=self.game, edx=self.unit)
        return (self.get(14), self.get(15), self.get(13), self.xploss, self.get(0xaf), self.pile)


def gen(n, seed):
    d = Death()
    r = random.Random(seed)
    pct = {}
    for diff in range(3):
        row = d.e.call(0x611d30, diff)
        pct[diff] = d.e.r32(row + 4)
    sys.stderr.write(f"DeathExpPenalty per difficulty (row +4): {pct}\n")
    out = []
    for i in range(n):
        cls, diff = r.randrange(7), r.randrange(3)
        level = r.choice([r.randint(2, 99)] * 3 + [1, 2, 3, 98, 99])
        low, high = d.thresholds(cls, level - 1), d.thresholds(cls, level)
        span = (high - low) & 0xffffffff
        loss = pct[diff] * span // 100
        exp = r.choice([low, low + 1, low + 2, low + max(loss - 1, 0), low + loss, low + loss + 1, low + loss + 2,
                        r.randint(low, max(low, high - 1)), high - 1])     # a player's experience is at least its level's start
        exp = max(exp, low) & 0xffffffff
        cap, floor = level * 10000, level * 500
        gold = r.choice([0, 1, 99, 100, 101, floor - 1, floor, floor + 1, cap // 2, cap, r.randint(0, cap), r.randint(0, 2 * floor + 2000)])
        bank = r.choice([0, 1, 500, floor, floor * 3, 100000, 2500000, r.randint(0, 2500000)])
        out.append(f"c{i} {cls} {diff} {level} {exp} {max(gold, 0)} {max(bank, 0)}")
    return out


def main():
    a = sys.argv[1:]
    if not a: sys.exit(__doc__)
    if a[0] == "gen":
        print("\n".join(gen(int(a[1]), int(a[2]))))
    elif a[0] == "game":
        d = Death()
        killer = int(a[2]) if len(a) > 2 else 0
        for line in open(a[1]):
            f = line.split()
            if f: print(f[0], *d.die(*map(int, f[1:7]), killer=killer))


if __name__ == "__main__":
    main()
