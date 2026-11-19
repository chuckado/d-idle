# SPDX-License-Identifier: GPL-3.0-or-later
"""game.exe's per-frame life rule for poison and bleeding (FUN_005a6920), the DoT sub-point death rule (card decomp39).

(Scratch oracle; it imports upstream's tools/emu `skill_tips` and `emu`: run it on the builder with that directory
as the working directory.)

    python3 dc39_dot.py run            # prints one line per (life, net): `L net life_after died`, then the summary

A monster unit with life stat 6 = L (256ths), max life stat 7, and stat 0x4a (the regeneration a frame, 256ths) =
net is run through FUN_005a6920 for real (the stat getter and setter run natively); hooked: the state test
FUN_00639df0 (no state), FUN_0063a750 (no stat 0x4a share), the unit's room FUN_00620bb0 (a fake room; FUN_0061ab00
runs for real on it), the event posts FUN_005417d0 / FUN_00540e60, the life percent FUN_005a5650 / FUN_00571a10 and
the death handlers FUN_0057ccb0 / FUN_005c0c30 (a call of FUN_0057ccb0 is `died`).
"""
import sys

import skill_tips

TICK = 0x5A6920


def main():
    e = skill_tips.boot()
    game, unit = e.alloc(0x200), e.alloc(0x200)
    stat_list = e.alloc(0x100)
    room, lvl, info = e.alloc(0x80), e.alloc(0x100), e.alloc(0x400)
    e.w32(room + 0x10, lvl); e.w32(lvl + 0x58, info)                 # FUN_0061ab00: room -> +0x10 -> +0x58 -> +0x1d0 = the level id
    e.w32(unit, 1); e.w32(unit + 4, 0x10); e.w32(unit + 0x10, 1)    # a monster in mode 1 (NU)
    e.w32(unit + 0x5C, stat_list)
    e.w32(unit + 0xC, 0x1234)
    st = {"died": 0}

    def got(stat): return e.call(0x625480, unit, stat, 0) & 0xFFFFFFFF

    def put(stat, v): e.call(0x627260, unit, stat, v & 0xFFFFFFFF, 0)

    e.hook(0x639DF0, lambda _e: 0, 2)
    e.hook(0x63A750, lambda _e: 0, 1)
    e.hook(0x620BB0, lambda _e: room, 1)
    e.hook(0x5417D0, lambda _e: 0, 4)
    e.hook(0x540E60, lambda _e: 0, 2)
    e.hook(0x5A5650, lambda _e: 0, 0)
    e.hook(0x571A10, lambda _e: 0, 0)
    e.hook(0x5C0C30, lambda _e: 0, 3)
    e.hook(0x57CCB0, lambda _e: st.__setitem__("died", 1) or 0, 1)
    total = 0
    lines = []
    skipped_levels = []
    # the town guard: a damaging tick on a life under one point does nothing when FUN_0061ab00 (FUN_006426a0 on the level id) says so
    for level_id in range(1, 0x6E):
        e.w32(info + 0x1D0, level_id)
        e.mu.mem_write(stat_list, bytes(0x100)); e.w32(stat_list + 0x10, 0x80000000)
        put(7, 100 * 256); put(6, 128); put(0x4A, -1)
        e.call(TICK, unit, ecx=game, edx=unit)
        if got(6) == 128:
            skipped_levels.append(level_id)
    print("levels whose tick on a life under 256 is skipped:", skipped_levels)
    e.w32(info + 0x1D0, 2)                                           # the Blood Moor
    for max_life in (100 * 256,):
        for life in (1, 2, 100, 255, 256, 257, 300, 383, 384, 511, 512, 513, 1000, 5000):
            for net in (-1, -2, -50, -100, -128, -255, -256, -257, -300, -1000):
                e.mu.mem_write(stat_list, bytes(0x100))
                e.w32(stat_list + 0x10, 0x80000000)
                put(7, max_life); put(6, life); put(0x4A, net)
                st["died"] = 0
                e.call(TICK, unit, ecx=game, edx=unit)
                after = got(6)
                lines.append((life, net, after, st["died"]))
    for life, net, after, died in lines:
        total += 1
        # the rule read at 0x5a69d4..0x5a69f3: life += net; under 1 (256ths) is 0 and the monster dies
        want = life + net if life + net >= 1 else 0
        assert after == want and died == (1 if want == 0 else 0), (life, net, after, died, want)
        print(life, net, after, died)
    survivors = [x for x in lines if x[2] > 0 and x[2] < 256]
    print(f"{total} cases, life under one point (256ths) kept alive by the tick: {len(survivors)}, all equal life + net")


if __name__ == "__main__":
    main()
