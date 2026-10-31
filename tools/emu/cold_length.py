# SPDX-License-Identifier: GPL-3.0-or-later
"""The chill length a cold hit hands to the chill (FUN_0057af80), by game.exe's own code.

    uv run python cold_length.py [N] [SEED] > game.txt     (default 100000 random cases, seed 1)

Runs the real FUN_0057c6c0 (apply a damage record to a monster) on a record
with cold damage (+0x24, 256ths) and a cold length (+0x30, ticks), as
FUN_0056c8e0 fills it for a cold skill, and prints the length FUN_0057c6c0
passes to FUN_0057af80 (its second argument; hooked here and recorded).
Inside it run for real: FUN_0057c140 (stat 0x99 'cannot be frozen' zeroes
the cold and freeze lengths, stat 0x76 'half freeze duration' halves them),
FUN_0057c1e0's loop over the element table at 0x732980 (entry 5: field 0x30
through cold resistance 0x2b less the attacker's cold pierce 0x14f, as
FUN_0057bf80 / FUN_0057be00 do the damage) and the call itself, made
whatever the damage came to (0057c8a0..0057c8a8: no test on +0x24).

Hooked (not the claim), as tools/emu/life_apply.py: the stat getters
(FUN_00625480 and FUN_00625500, a dict) and setter, max life, the unit-kind
predicates, difficulty row, events, logging, the other side effects.

FUN_0057af80 itself (read, V): nothing at length <= 0; a monster's
MonStats coldeffect (+0x168 + difficulty, signed byte) 0 = never chilled,
negative: length / DifficultyLevels MonsterColdDivisor (+0x18); a player
-50, no divisor; at least 1 tick; a running chill (state 0xb) is only
lengthened. d2d's land() does the same (components/game/fight/strike.cpp,
the 'Chilled' block), so the oracle compares what reaches it.

A case: `id roll len res pierce cbf hfd` (roll 256ths, len ticks, cbf / hfd
the two stats). To stderr: how many cases a line-for-line copy of d2d's
missile_blow (components/rules/combat.hpp, etype 2: chill_ticks =
resisted(elen, res) only when the resisted damage is a whole point or
more; neither stat 0x99 nor 0x76 read) gets wrong, all and with the two
stats at 0.
"""
import sys

import life_apply
from static_field import xorshift, s32


class Game(life_apply.Game):
    def __init__(self):
        super().__init__()
        e = self.e
        self.chill = None
        e.hook(0x625500, lambda e: self.stats.get((e.arg(0), e.arg(1)), 0), 3)
        e.hook(0x57af80, lambda e: setattr(self, "chill", s32(e.arg(1))) or 0, 2)

    def cold(self, roll, length, res, pierce, cbf, hfd):
        e = self.e
        self.reset(1000, 0, 0)
        self.stats[(self.target, 0x2b)] = res & 0xFFFFFFFF
        self.stats[(self.attacker, 0x14f)] = pierce
        self.stats[(self.target, 0x99)] = cbf
        self.stats[(self.target, 0x76)] = hfd
        self.chill = None
        e.mu.mem_write(self.rec, b"\0" * 0x80)
        e.w32(self.rec, 1); e.w32(self.rec + 4, 1)
        e.w32(self.rec + 0x24, roll); e.w32(self.rec + 0x30, length); e.w32(self.rec + 0x60, 0x30)
        e.call(0x57c6c0, self.target, 1, self.rec, ecx=self.game, edx=self.attacker)
        return "none" if self.chill is None else str(max(self.chill, 0))


def resisted(dmg, res): return 0 if res >= 100 else (dmg * (100 - res)) // 100


def d2d(roll, length, res, pierce, cbf, hfd):
    from static_field import resisted_fine
    eff = max(res - pierce, -100) if res < 100 else res
    fine = resisted_fine(roll, eff)
    return str(resisted(length, eff) if fine >> 8 > 0 else 0)


def cases(n, seed):
    i = 0
    for res in (-150, -100, -50, -1, 0, 1, 25, 50, 75, 90, 99, 100, 150):
        for roll in (1, 128, 255, 256, 257, 511, 512, 1024, 2560):
            for length in (1, 2, 3, 50, 99, 100, 101, 150, 200, 375):
                yield (f"B{i}", roll, length, res, 0, 0, 0); i += 1
    state = [seed or 1]
    for i in range(n):
        roll = 1 + xorshift(state) % (0x800 if xorshift(state) % 2 else 0x4000)
        length = 1 + xorshift(state) % 500
        r = xorshift(state) % 100
        if r < 10: res = 100 + xorshift(state) % 51
        elif r < 25: res = -(xorshift(state) % 151)
        else: res = xorshift(state) % 100
        pierce = 0 if res >= 100 or xorshift(state) % 4 else xorshift(state) % 61
        k = xorshift(state) % 10
        cbf, hfd = (1, 0) if k == 0 else (0, 1) if k == 1 else (0, 0)
        yield (f"R{i}", roll, length, res, pierce, cbf, hfd)


def main():
    n = int(sys.argv[1]) if len(sys.argv) > 1 else 100000
    seed = int(sys.argv[2]) if len(sys.argv) > 2 else 1
    g = Game()
    bad = plain = plain_bad = total = 0
    for case in cases(n, seed):
        got = g.cold(*case[1:])
        total += 1
        miss = d2d(*case[1:]) != got
        bad += miss
        if case[5] == 0 and case[6] == 0: plain += 1; plain_bad += miss
        print(" ".join(map(str, case)), got)
    print(f"{total} cases: d2d wrong on {bad}; with stats 0x99 / 0x76 at 0: {plain_bad} of {plain}", file=sys.stderr)


if __name__ == "__main__": main()
