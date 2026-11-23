# SPDX-License-Identifier: GPL-3.0-or-later
"""game.exe's shrine row roll (FUN_0054f9d0 -> FUN_0054f770) as the oracle.

(Scratch oracle of card decomp38. It imports skill_tips of upstream's tools/emu: run it on the builder with that
directory on PYTHONPATH; the d2d side is the fork branch named in docs/research/dc38-lane-d.md.)

    python3 dc38_shrine.py gen N SEED > cases.tsv
    python3 dc38_shrine.py game cases.tsv > game.tsv
    test_shrine_roll --shrine-roll cases.tsv > d2d.tsv ; diff game.tsv d2d.tsv

A case is `parm0 level_id seed_lo seed_hi` (objects.txt Parm0: 0 any row, 1 health, 2 mana, 3 the rest; the
Levels.txt id the shrine is made on; the object seed). `game` prints the case plus `row seed_lo seed_hi`: the
Shrines.txt row stored in the object's data +4 and the seed after the roll.
FUN_0054f9d0 runs for real, with the class lists FUN_00546c60 builds (real, its allocator and seed init hooked);
hooked: FUN_0040b430 (pool alloc), FUN_00650e30 / FUN_00650e40 (their seed init), FUN_00681ef0 (memset),
FUN_0061a1b0 (the room's level id: the case's), FUN_00621bb0 (the call that applies the row).
"""
import random
import sys



def gen(n, seed):
    r = random.Random(seed)
    for _ in range(n):
        parm0 = r.choice((0, 1, 2, 3, 3, 3))
        print(parm0, r.randrange(0, 137), r.getrandbits(32), r.getrandbits(32))


def game(path):
    import skill_tips
    from unicorn.x86_const import UC_X86_REG_EDX
    e = skill_tips.boot()
    game_rec = e.alloc(0x2000)
    e.w32(game_rec + 0x1c, 0x1000)                    # the pool handle: unused by the allocator hook
    e.w32(game_rec + 0xd0, 12345); e.w32(game_rec + 0xd4, 6)
    level = [0]
    e.hook(0x40b430, lambda e: e.alloc(max(e.mu.reg_read(UC_X86_REG_EDX), 4)), 3)
    e.hook(0x650e30, lambda e: 0, 0)
    e.hook(0x650e40, lambda e: 0, 0)
    e.hook(0x681ef0, lambda e: 0, 0)               # cdecl: the caller pops
    e.call(0x546c60, ecx=game_rec)
    e.hook(0x61a1b0, lambda e: level[0], 1)
    e.hook(0x621bb0, lambda e: 0, 2)
    unit = e.alloc(0x100); data = e.alloc(0x40)
    e.w32(unit + 0x14, data)
    row = e.alloc(0x400); room = e.alloc(0x40)
    seedp = e.r32(game_rec + 0x10f0)               # the game's object seed is the first 8 bytes of the class block (FUN_00546fa0)
    arg = e.alloc(0x40)
    e.w32(arg, game_rec); e.w32(arg + 4, unit); e.w32(arg + 8, room); e.w32(arg + 0xc, seedp); e.w32(arg + 0x10, row)
    for line in open(path):
        p0, lv, lo, hi = (int(x) for x in line.split())
        level[0] = lv
        e.w32(row + 0x178, p0)
        e.w32(seedp, lo); e.w32(seedp + 4, hi)
        e.mu.mem_write(data, bytes(0x40))
        e.call(0x54f9d0, ecx=arg)
        print(p0, lv, lo, hi, e.read(data + 4, 1)[0], e.r32(seedp), e.r32(seedp + 4))


if __name__ == "__main__":
    a = sys.argv[1:]
    if a[0] == "gen":
        gen(int(a[1]), int(a[2]))
    elif a[0] == "game":
        game(a[1])
