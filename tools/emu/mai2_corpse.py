"""monsterai2 (D9): the Fallen Shaman's corpse test, game.exe's own callback FUN_005f1380 run on random candidates, against the rule
d2d's ai.cpp:880-884 applies for a unique shaman (a Fallen / Fallen Shaman class corpse, not unique, ai_distance < 80).

    PYTHONPATH=<upstream tools/emu>:tools/emu D2_MPQ_DIR=<dir> python mai2_corpse.py [cases] [seed]

The callback (called by the room search FUN_005dd0b0 for every unit near, type 1, only when the shaman's AI flags are aware and not
bit 4, FUN_005a0180 8 / 4) takes (candidate, ctx), ctx +0xc = the limit the think wrote (aip4 squared, 0x5f14c8..0x5f14d2). Hooked around
it (answers outside the claim): the class (FUN_00463860), the corpse bookkeeping (FUN_006259b0, FUN_0063a770, FUN_0063a320: none),
the squared distance (FUN_005b0bd0, as upstream's mon_ais.py hooks it: dx^2 + dy^2), the flag test (FUN_005a0180 on the candidate).
The callback's own code checks: type 1, class in 0x13 / 0x3a, corpse bit (+0xc4 bit 1), mode DD (+0x10 == 0xc), squared distance <= limit,
not flagged 0xc. Prints `corpse cases game_accepts d2d_accepts disagree` and the disagreements by cause. Never runs a game."""
import random
import struct
import sys

import emu
from unicorn.x86_const import UC_X86_REG_ECX, UC_X86_REG_EDX

CB = 0x5f1380


def ai_distance(dx, dy):                                    # monsters.hpp:398
    dx, dy = abs(dx), abs(dy)
    return (dy + dx * 2) // 2 if dy < dx else (dx + dy * 2) // 2


def main():
    cases = int(sys.argv[1]) if len(sys.argv) > 1 else 6000
    rng = random.Random(int(sys.argv[2]) if len(sys.argv) > 2 else 1)
    e = emu.Emu()
    cur = {}
    data = e.alloc(0x1000)
    e.w32(0x744304, data); e.w32(data + 0xa80, 0x200)
    me, cand, ctx = e.alloc(0x200), e.alloc(0x200), e.alloc(0x40)
    e.w32(me, 1); e.w32(cand, 1)
    e.hook(0x463860, lambda e: cur["class"], 0)
    e.hook(0x6259b0, lambda e: 0, 1)
    e.hook(0x63a770, lambda e: 0, 1)
    e.hook(0x63a320, lambda e: 0, 1)
    def squared(e):                                         # FUN_005b0bd0 (ECX, EDX: the two units)
        return (cur["dx"]) ** 2 + (cur["dy"]) ** 2
    e.hook(0x5b0bd0, squared, 0)
    e.hook(0x5a0180, lambda e: int(bool(cur["flags"] & e.mu.reg_read(UC_X86_REG_EDX))), 0)
    seen, agree, game_ok, d2d_ok = {}, 0, 0, 0
    bad = {}
    for _ in range(cases):
        aip4 = rng.choice((24, 24, 24, 26, 10, 40))
        cur["class"] = rng.choice((0x13, 0x13, 0x3a, 0x3a, 0x10, 0x12, 0x21))
        cur["flags"] = rng.choice((0, 0, 0, 4, 8, 0xc))
        corpse_bit = rng.random() < 0.9
        mode = rng.choice((0xc, 0xc, 0xc, 1, 0))
        reach = rng.choice((10, 30, 60, 100))
        cur["dx"], cur["dy"] = rng.randint(-reach, reach), rng.randint(-reach, reach)
        e.w32(cand + 0x10, mode); e.w32(cand + 0xc4, 2 if corpse_bit else 0)
        for k in range(8): e.w32(ctx + 4 * k, 0)
        e.w32(ctx + 0xc, aip4 * aip4)
        e.call(CB, cand, ctx, edx=me)
        game = e.r32(ctx + 4) == 1
        # d2d's unique-shaman rule (ai.cpp:880-884) on the same candidate: a dead (DD), unraised one of base 19 / 58, not unique;
        # `unique` there is Boss::unique / superunique, which flag bit 8 stands for (bit 4: a champion, left to the game's mask 0xc).
        d2d = mode == 0xc and corpse_bit and cur["class"] in (0x13, 0x3a) and not (cur["flags"] & 8) and ai_distance(cur["dx"], cur["dy"]) < 80
        read = mode == 0xc and corpse_bit and cur["class"] in (0x13, 0x3a) and not (cur["flags"] & 0xc) and cur["dx"] ** 2 + cur["dy"] ** 2 <= aip4 * aip4
        misread = bad.get("my reading of the callback differs from the run", 0) + (read != game)
        if read != game: bad["my reading of the callback differs from the run"] = misread
        game_ok += game; d2d_ok += d2d
        if game == d2d: agree += 1
        else:
            near = cur["dx"] ** 2 + cur["dy"] ** 2 <= aip4 * aip4
            cause = ("flag 4 (champion): game rejects" if cur["flags"] & 4 and not game else
                     "within 80 by ai_distance, beyond aip4 squared: d2d takes it" if d2d and not near else
                     "other")
            bad[cause] = bad.get(cause, 0) + 1
    print(f"corpse {cases} {game_ok} {d2d_ok} {cases - agree}")
    for cause, n in sorted(bad.items()): print("  disagree", n, cause)
    return 0


if __name__ == "__main__":
    sys.exit(main())
